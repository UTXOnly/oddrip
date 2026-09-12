package oddrip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

const defaultWSHost = "external-api-ws.kalshi.com"
const defaultWSPath = "/trade-api/ws/v2"

const (
	defaultWSBufferSize   = 4096
	defaultWSPingInterval = 30 * time.Second
	defaultWSReadTimeout  = 90 * time.Second
	defaultWSWriteTimeout = 10 * time.Second
	// wsCloseWait bounds how long Close waits for the read loop after the
	// socket is closed. The read loop exits as soon as the socket does, so
	// this is a backstop, not an expected wait.
	wsCloseWait = 5 * time.Second
)

var (
	ErrWSClosed       = errors.New("websocket closed")
	ErrWSAuthRequired = errors.New("websocket requires auth")
	ErrWSSlowConsumer = errors.New("websocket consumer too slow")
	// ErrWSMalformedFrame is the terminal error when a text frame is not valid
	// JSON; Err() wraps it with the decode error.
	ErrWSMalformedFrame = errors.New("websocket malformed frame")
	// ErrWSWriteTimeout is the terminal error when a command frame could not
	// be written within WSWriteTimeout or the caller's context deadline,
	// whichever came first; Err() wraps it with the socket error. A timed-out
	// write leaves the socket unusable, so the connection is failed and
	// Messages() closes, the same as ErrWSSlowConsumer.
	ErrWSWriteTimeout = errors.New("websocket write timeout")
)

type WSConn struct {
	conn         *websocket.Conn
	auth         AuthProvider
	host         string
	path         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	nextID       atomic.Int64
	mu           sync.Mutex
	closed       bool
	readErr      error
	// writeSem serializes data-frame writes (gorilla allows one WriteMessage
	// at a time). It is a 1-slot channel rather than a mutex so a waiter can
	// give up when its context ends or the connection dies. Control frames
	// (ping, pong, close) go through WriteControl, which gorilla serializes
	// internally and which is safe alongside a data write in progress.
	writeSem  chan struct{}
	pendMu    sync.Mutex
	pending   map[int]chan *wsEnvelope   // command replies, by command id
	snapshots map[int][]chan *wsEnvelope // get_snapshot waiters, by sid
	msgChan   chan *types.WSMessage
	readDone  chan struct{}
}

type wsEnvelope struct {
	ID   int             `json:"id,omitempty"`
	Type string          `json:"type"`
	SID  int             `json:"sid,omitempty"`
	Seq  int             `json:"seq,omitempty"`
	Msg  json.RawMessage `json:"msg,omitempty"`
}

type WSOption func(*wsOpts)

type wsOpts struct {
	scheme       string
	host         string
	path         string
	bufferSize   int
	pingInterval time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func WSScheme(scheme string) WSOption {
	return func(o *wsOpts) {
		o.scheme = scheme
	}
}

func WSHost(host string) WSOption {
	return func(o *wsOpts) {
		o.host = host
	}
}

func WSPath(path string) WSOption {
	return func(o *wsOpts) {
		o.path = path
	}
}

// WSBufferSize sets the Messages() buffer. If the consumer lets it fill, the
// connection fails with ErrWSSlowConsumer rather than dropping messages.
func WSBufferSize(n int) WSOption {
	return func(o *wsOpts) {
		o.bufferSize = n
	}
}

// WSPingInterval sets how often a keepalive ping is sent. <= 0 disables pings.
func WSPingInterval(d time.Duration) WSOption {
	return func(o *wsOpts) {
		o.pingInterval = d
	}
}

// WSReadTimeout fails the connection if nothing (data, ping, or pong) is read
// for this long. <= 0 disables the read deadline.
func WSReadTimeout(d time.Duration) WSOption {
	return func(o *wsOpts) {
		o.readTimeout = d
	}
}

// WSWriteTimeout bounds every socket write. A command frame that cannot be
// written within this long (or the caller's context deadline, if sooner)
// fails the connection with an error wrapping ErrWSWriteTimeout; the close
// frame sent by Close is bounded by it too. Writes are never unbounded:
// <= 0 uses the default of 10s.
func WSWriteTimeout(d time.Duration) WSOption {
	return func(o *wsOpts) {
		o.writeTimeout = d
	}
}

// wsConfig applies opts over the defaults and normalizes them.
func wsConfig(opts []WSOption) wsOpts {
	cfg := wsOpts{
		scheme:       "wss",
		host:         defaultWSHost,
		path:         defaultWSPath,
		bufferSize:   defaultWSBufferSize,
		pingInterval: defaultWSPingInterval,
		readTimeout:  defaultWSReadTimeout,
		writeTimeout: defaultWSWriteTimeout,
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.scheme == "" {
		cfg.scheme = "wss"
	}
	if cfg.bufferSize <= 0 {
		cfg.bufferSize = defaultWSBufferSize
	}
	if cfg.writeTimeout <= 0 {
		cfg.writeTimeout = defaultWSWriteTimeout
	}
	return cfg
}

// url is the dial destination.
func (o wsOpts) url() string {
	return (&url.URL{Scheme: o.scheme, Host: o.host, Path: o.path}).String()
}

func (c *Client) ConnectWS(ctx context.Context, opts ...WSOption) (*WSConn, error) {
	if c.auth == nil {
		return nil, ErrWSAuthRequired
	}
	cfg := wsConfig(opts)
	u := cfg.url()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	if err := c.auth.Apply(req); err != nil {
		return nil, err
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.DialContext(ctx, u, req.Header)
	if err != nil {
		return nil, fmt.Errorf("ws dial: %w", err)
	}
	ws := &WSConn{
		conn:         conn,
		auth:         c.auth,
		host:         cfg.host,
		path:         cfg.path,
		readTimeout:  cfg.readTimeout,
		writeTimeout: cfg.writeTimeout,
		writeSem:     make(chan struct{}, 1),
		pending:      make(map[int]chan *wsEnvelope),
		snapshots:    make(map[int][]chan *wsEnvelope),
		msgChan:      make(chan *types.WSMessage, cfg.bufferSize),
		readDone:     make(chan struct{}),
	}
	ws.nextID.Store(1)
	ws.resetDeadline()
	conn.SetPongHandler(func(string) error {
		ws.resetDeadline()
		return nil
	})
	conn.SetPingHandler(func(data string) error {
		ws.resetDeadline()
		err := conn.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(time.Second))
		var ne net.Error
		if errors.Is(err, websocket.ErrCloseSent) || errors.As(err, &ne) {
			return nil
		}
		return err
	})
	go ws.readLoop()
	if cfg.pingInterval > 0 {
		go ws.keepalive(cfg.pingInterval)
	}
	return ws, nil
}

func (ws *WSConn) resetDeadline() {
	if ws.readTimeout > 0 {
		ws.conn.SetReadDeadline(time.Now().Add(ws.readTimeout))
	}
}

func (ws *WSConn) setErr(err error) {
	ws.mu.Lock()
	if ws.readErr == nil {
		ws.readErr = err
	}
	ws.mu.Unlock()
}

func (ws *WSConn) keepalive(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ws.readDone:
			return
		case <-t.C:
			ws.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
		}
	}
}

func (ws *WSConn) readLoop() {
	defer close(ws.readDone)
	defer close(ws.msgChan)
	defer ws.drainPending()
	for {
		_, data, err := ws.conn.ReadMessage()
		if err != nil {
			ws.setErr(err)
			return
		}
		ws.resetDeadline()
		var env wsEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			// The server only sends JSON text frames. Anything else means the
			// stream is corrupt; fail loudly like the slow-consumer path rather
			// than skip it and leave the connection looking healthy.
			ws.setErr(fmt.Errorf("%w: %v", ErrWSMalformedFrame, err))
			ws.conn.Close()
			return
		}
		if env.ID != 0 {
			ws.pendMu.Lock()
			ch := ws.pending[env.ID]
			ws.pendMu.Unlock()
			if ch != nil {
				select {
				case ch <- &env:
				default:
				}
			}
		}
		// get_snapshot is answered with orderbook_snapshot frames that carry no
		// command id, so its waiter is keyed by sid instead.
		if env.Type == types.WSTypeOrderbookSnapshot && env.SID != 0 {
			ws.pendMu.Lock()
			waiters := ws.snapshots[env.SID]
			ws.pendMu.Unlock()
			for _, ch := range waiters {
				select {
				case ch <- &env:
				default:
				}
			}
		}
		msg := &types.WSMessage{Type: env.Type, SID: env.SID, Seq: env.Seq, Msg: env.Msg}
		select {
		case ws.msgChan <- msg:
		default:
			ws.setErr(ErrWSSlowConsumer)
			ws.conn.Close()
			return
		}
	}
}

func (ws *WSConn) drainPending() {
	ws.pendMu.Lock()
	for _, ch := range ws.pending {
		close(ch)
	}
	for _, waiters := range ws.snapshots {
		for _, ch := range waiters {
			close(ch)
		}
	}
	ws.pending = make(map[int]chan *wsEnvelope)
	ws.snapshots = make(map[int][]chan *wsEnvelope)
	ws.pendMu.Unlock()
}

func (ws *WSConn) nextIDVal() int {
	return int(ws.nextID.Add(1))
}

// sendAndWait writes payload and collects expectCount id-matched replies. If
// snapshotSID is non-zero the call also completes on the first
// orderbook_snapshot frame for that sid, which is how get_snapshot is answered.
// An error reply ends the wait; the replies collected before it are returned
// alongside the *WSError so callers can report partial success.
//
// The write is bounded by ctx and writeTimeout (see write); nothing here waits
// past the caller's deadline.
func (ws *WSConn) sendAndWait(ctx context.Context, id int, payload interface{}, expectCount int, snapshotSID int) ([]*wsEnvelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	ws.mu.Lock()
	if ws.readErr != nil {
		err := ws.readErr
		ws.mu.Unlock()
		return nil, err
	}
	ch := make(chan *wsEnvelope, max(expectCount, 1))
	var snap chan *wsEnvelope // nil blocks forever in the select below
	ws.pendMu.Lock()
	ws.pending[id] = ch
	if snapshotSID != 0 {
		snap = make(chan *wsEnvelope, 1)
		ws.snapshots[snapshotSID] = append(ws.snapshots[snapshotSID], snap)
	}
	ws.pendMu.Unlock()
	ws.mu.Unlock()
	defer func() {
		ws.pendMu.Lock()
		delete(ws.pending, id)
		if snap != nil {
			ws.removeSnapshotWaiter(snapshotSID, snap)
		}
		ws.pendMu.Unlock()
	}()

	if err := ws.write(ctx, data); err != nil {
		return nil, err
	}
	var out []*wsEnvelope
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ws.readDone:
			return nil, ws.closedErr()
		case env, ok := <-snap:
			if !ok {
				return nil, ws.closedErr()
			}
			return []*wsEnvelope{env}, nil
		case env, ok := <-ch:
			if !ok {
				return nil, ws.closedErr()
			}
			if env.Type == types.WSTypeError {
				var errMsg types.ErrorMsg
				if len(env.Msg) > 0 {
					json.Unmarshal(env.Msg, &errMsg)
				}
				return out, &WSError{Code: errMsg.Code, Message: errMsg.Msg}
			}
			out = append(out, env)
			if expectCount <= 0 || len(out) >= expectCount {
				return out, nil
			}
		}
	}
}

// write sends one text frame. Waiting for the write slot is bounded by ctx
// (and by the connection dying); the socket write itself by the sooner of the
// ctx deadline and writeTimeout.
//
// A caller whose ctx ends while another command holds the slot gets ctx.Err()
// and the connection is untouched: its frame never started. Once a frame is
// being written, a timeout from either bound leaves gorilla's writer state
// corrupt (every later write would fail), so the connection is failed with an
// error wrapping ErrWSWriteTimeout; Messages() closes and Err() reports it.
// The caller gets ctx.Err() if its own deadline was the cause, else that
// error. Other write errors are returned as-is: the read loop surfaces the
// socket failure behind them on its own. A timeout is the case it cannot see,
// since a peer that has stopped reading may keep sending.
func (ws *WSConn) write(ctx context.Context, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case ws.writeSem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	case <-ws.readDone:
		return ws.closedErr()
	}
	defer func() { <-ws.writeSem }()

	deadline := time.Now().Add(ws.writeTimeout)
	ctxBound := false
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline, ctxBound = d, true
	}
	ws.conn.SetWriteDeadline(deadline)
	err := ws.conn.WriteMessage(websocket.TextMessage, data)
	if err == nil {
		return nil
	}
	if e := ws.Err(); e != nil {
		return e
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		ws.setErr(fmt.Errorf("%w: %v", ErrWSWriteTimeout, err))
		ws.conn.Close()
		if cerr := ctx.Err(); cerr != nil {
			return cerr
		}
		if ctxBound {
			// The socket deadline was the caller's. Its context timer can
			// fire a moment after the socket's, so ctx.Err() may still be
			// nil here; the write was cut at that deadline regardless.
			return context.DeadlineExceeded
		}
		return ws.Err()
	}
	return err
}

// removeSnapshotWaiter must be called with pendMu held.
func (ws *WSConn) removeSnapshotWaiter(sid int, ch chan *wsEnvelope) {
	waiters := ws.snapshots[sid]
	for i, w := range waiters {
		if w == ch {
			waiters = append(waiters[:i], waiters[i+1:]...)
			break
		}
	}
	if len(waiters) == 0 {
		delete(ws.snapshots, sid)
	} else {
		ws.snapshots[sid] = waiters
	}
}

func (ws *WSConn) closedErr() error {
	if err := ws.Err(); err != nil {
		return err
	}
	return ErrWSClosed
}

// Subscribe sends a subscribe command and returns one SubscribedResponse per
// channel. The server answers each channel separately, so if it rejects one
// channel after accepting others the accepted subscriptions are returned
// together with the *WSError; they are live and must be unsubscribed or used.
func (ws *WSConn) Subscribe(ctx context.Context, params types.SubscribeParams) ([]types.SubscribedResponse, error) {
	if len(params.Channels) == 0 {
		return nil, errors.New("channels required")
	}
	id := ws.nextIDVal()
	cmd := types.SubscribeCommand{
		ID:     id,
		Cmd:    "subscribe",
		Params: params,
	}
	envs, err := ws.sendAndWait(ctx, id, cmd, len(params.Channels), 0)
	var result []types.SubscribedResponse
	for _, env := range envs {
		if env.Type != types.WSTypeSubscribed {
			continue
		}
		var m types.SubscribedMsg
		if len(env.Msg) > 0 {
			if derr := json.Unmarshal(env.Msg, &m); derr != nil {
				return result, fmt.Errorf("ws: decode subscribed reply: %w", derr)
			}
		}
		result = append(result, types.SubscribedResponse{
			ID:   env.ID,
			Type: types.WSTypeSubscribed,
			Msg:  m,
		})
	}
	if err != nil {
		return result, err
	}
	if result == nil {
		result = []types.SubscribedResponse{}
	}
	return result, nil
}

func (ws *WSConn) Unsubscribe(ctx context.Context, sids []int) error {
	if len(sids) == 0 {
		return errors.New("sids required")
	}
	id := ws.nextIDVal()
	cmd := types.UnsubscribeCommand{ID: id, Cmd: "unsubscribe"}
	cmd.Params.Sids = sids
	_, err := ws.sendAndWait(ctx, id, cmd, len(sids), 0)
	return err
}

func (ws *WSConn) ListSubscriptions(ctx context.Context) (*types.ListSubscriptionsResponse, error) {
	id := ws.nextIDVal()
	cmd := types.ListSubscriptionsCommand{ID: id, Cmd: "list_subscriptions"}
	envs, err := ws.sendAndWait(ctx, id, cmd, 1, 0)
	if err != nil {
		return nil, err
	}
	if len(envs) == 0 {
		return nil, errors.New("no response")
	}
	env := envs[0]
	var list types.ListSubscriptionsResponse
	list.ID = env.ID
	list.Type = env.Type
	if len(env.Msg) > 0 {
		if err := json.Unmarshal(env.Msg, &list.Msg); err != nil {
			return nil, fmt.Errorf("ws: decode list_subscriptions reply: %w", err)
		}
	}
	return &list, nil
}

// UpdateSubscription sends an update_subscription command and returns the
// server's reply. The reply Type depends on the action:
//
//   - add_markets, delete_markets, subscribe_underlyings, unsubscribe_underlyings,
//     subscribe_indices, unsubscribe_indices: "ok"; Msg holds the full list
//     after the update.
//   - underlying_list / indexlist: "pyth_value_underlying_list" or
//     "cfbenchmarks_value_indexlist" / "cfbenchmarks_value_5hz_indexlist";
//     Msg.UnderlyingTickers or Msg.IndexIDs holds the list.
//   - get_snapshot: the server answers with orderbook_snapshot frames on
//     Messages(), which carry no command id. The call returns once the first
//     snapshot for the subscription (or an id-matched ok/error) arrives, with
//     Type "orderbook_snapshot" and the frame's SID/Seq; the snapshots
//     themselves are read from Messages().
//
// Server-side rejections are returned as *WSError.
func (ws *WSConn) UpdateSubscription(ctx context.Context, params types.UpdateSubscriptionParams) (*types.OKResponse, error) {
	switch params.Action {
	case types.WSUpdateSubscriptionAddMarkets,
		types.WSUpdateSubscriptionDeleteMarkets,
		types.WSUpdateSubscriptionGetSnapshot,
		types.WSUpdateSubscriptionSubscribeUnderlyings,
		types.WSUpdateSubscriptionUnsubscribeUnderlyings,
		types.WSUpdateSubscriptionUnderlyingList,
		types.WSUpdateSubscriptionSubscribeIndices,
		types.WSUpdateSubscriptionUnsubscribeIndices,
		types.WSUpdateSubscriptionIndexList:
	default:
		return nil, errors.New("action must be add_markets, delete_markets, get_snapshot, subscribe_underlyings, unsubscribe_underlyings, underlying_list, subscribe_indices, unsubscribe_indices, or indexlist")
	}
	snapshotSID := 0
	if params.Action == types.WSUpdateSubscriptionGetSnapshot {
		switch {
		case params.SID != nil:
			snapshotSID = *params.SID
		case len(params.Sids) == 1:
			snapshotSID = params.Sids[0]
		default:
			return nil, errors.New("get_snapshot requires sid or a single-element sids")
		}
	}
	id := ws.nextIDVal()
	cmd := types.UpdateSubscriptionCommand{ID: id, Cmd: "update_subscription", Params: params}
	envs, err := ws.sendAndWait(ctx, id, cmd, 1, snapshotSID)
	if err != nil {
		return nil, err
	}
	if len(envs) == 0 {
		return nil, errors.New("no response")
	}
	env := envs[0]
	ok := types.OKResponse{ID: id, SID: env.SID, Seq: env.Seq, Type: env.Type}
	if env.Type != types.WSTypeOrderbookSnapshot && len(env.Msg) > 0 {
		ok.Msg = &types.OKMsg{}
		if err := json.Unmarshal(env.Msg, ok.Msg); err != nil {
			return nil, fmt.Errorf("ws: decode %s reply: %w", env.Type, err)
		}
	}
	return &ok, nil
}

func (ws *WSConn) Messages() <-chan *types.WSMessage {
	return ws.msgChan
}

// Done is closed once the read loop has exited; Messages() is closed by then.
func (ws *WSConn) Done() <-chan struct{} {
	return ws.readDone
}

// Err is nil while the connection is healthy. After the read loop exits it is
// the terminal read error, ErrWSSlowConsumer, an error wrapping
// ErrWSMalformedFrame or ErrWSWriteTimeout, or ErrWSClosed after Close().
func (ws *WSConn) Err() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.readErr
}

// Close sends a close frame, closes the socket, and waits for the read loop
// to exit. It is idempotent, and pending commands return ErrWSClosed.
//
// Close is bounded even when the peer has stopped reading: it returns within
// about WSWriteTimeout plus five seconds, and in practice as soon as the
// close frame is written or skipped. The close frame is a control write with
// its own deadline, so Close never queues behind a command write; if one is
// in progress the frame is skipped (a jammed socket would not deliver it
// anyway) and the socket is closed directly, which also unblocks that write.
func (ws *WSConn) Close() error {
	ws.mu.Lock()
	if ws.closed {
		ws.mu.Unlock()
		return nil
	}
	ws.closed = true
	healthy := ws.readErr == nil
	if healthy {
		ws.readErr = ErrWSClosed
	}
	ws.mu.Unlock()
	var err error
	if healthy {
		select {
		case ws.writeSem <- struct{}{}:
			err = ws.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(ws.writeTimeout))
			<-ws.writeSem
		default:
			// A command write is in flight and bounded by its own deadline;
			// do not wait for it.
		}
	}
	if e := ws.conn.Close(); e != nil && err == nil && healthy {
		err = e
	}
	select {
	case <-ws.readDone:
		return err
	case <-time.After(wsCloseWait):
		return fmt.Errorf("close: read loop did not exit: %w", err)
	}
}

type WSError struct {
	Code    int
	Message string
}

func (e *WSError) Error() string {
	return fmt.Sprintf("ws error %d: %s", e.Code, e.Message)
}
