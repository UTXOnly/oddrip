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

const defaultWSHost = "api.elections.kalshi.com"
const defaultWSPath = "/trade-api/ws/v2"

const (
	defaultWSBufferSize   = 4096
	defaultWSPingInterval = 30 * time.Second
	defaultWSReadTimeout  = 90 * time.Second
)

var (
	ErrWSClosed       = errors.New("websocket closed")
	ErrWSAuthRequired = errors.New("websocket requires auth")
	ErrWSSlowConsumer = errors.New("websocket consumer too slow")
)

type WSConn struct {
	conn        *websocket.Conn
	auth        AuthProvider
	host        string
	path        string
	readTimeout time.Duration
	nextID      atomic.Int64
	mu          sync.Mutex
	closed      bool
	readErr     error
	writeMu     sync.Mutex
	pendMu      sync.Mutex
	pending     map[int]chan *wsEnvelope
	msgChan     chan *types.WSMessage
	readDone    chan struct{}
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

func (c *Client) ConnectWS(ctx context.Context, opts ...WSOption) (*WSConn, error) {
	if c.auth == nil {
		return nil, ErrWSAuthRequired
	}
	cfg := wsOpts{
		scheme:       "wss",
		host:         defaultWSHost,
		path:         defaultWSPath,
		bufferSize:   defaultWSBufferSize,
		pingInterval: defaultWSPingInterval,
		readTimeout:  defaultWSReadTimeout,
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
	u := url.URL{Scheme: cfg.scheme, Host: cfg.host, Path: cfg.path}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if err := c.auth.Apply(req); err != nil {
		return nil, err
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.DialContext(ctx, u.String(), req.Header)
	if err != nil {
		return nil, fmt.Errorf("ws dial: %w", err)
	}
	ws := &WSConn{
		conn:        conn,
		auth:        c.auth,
		host:        cfg.host,
		path:        cfg.path,
		readTimeout: cfg.readTimeout,
		pending:     make(map[int]chan *wsEnvelope),
		msgChan:     make(chan *types.WSMessage, cfg.bufferSize),
		readDone:    make(chan struct{}),
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
			continue
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
	ws.pending = make(map[int]chan *wsEnvelope)
	ws.pendMu.Unlock()
}

func (ws *WSConn) nextIDVal() int {
	return int(ws.nextID.Add(1))
}

func (ws *WSConn) sendAndWait(ctx context.Context, id int, payload interface{}, expectCount int) ([]*wsEnvelope, error) {
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
	ws.pendMu.Lock()
	ws.pending[id] = ch
	ws.pendMu.Unlock()
	ws.mu.Unlock()
	defer func() {
		ws.pendMu.Lock()
		delete(ws.pending, id)
		ws.pendMu.Unlock()
	}()

	ws.writeMu.Lock()
	err = ws.conn.WriteMessage(websocket.TextMessage, data)
	ws.writeMu.Unlock()
	if err != nil {
		if e := ws.Err(); e != nil {
			return nil, e
		}
		return nil, err
	}
	var out []*wsEnvelope
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ws.readDone:
			return nil, ws.closedErr()
		case env, ok := <-ch:
			if !ok {
				return nil, ws.closedErr()
			}
			if env.Type == "error" {
				var errMsg types.ErrorMsg
				if len(env.Msg) > 0 {
					json.Unmarshal(env.Msg, &errMsg)
				}
				return nil, &WSError{Code: errMsg.Code, Message: errMsg.Msg}
			}
			out = append(out, env)
			if expectCount <= 0 || len(out) >= expectCount {
				return out, nil
			}
		}
	}
}

func (ws *WSConn) closedErr() error {
	if err := ws.Err(); err != nil {
		return err
	}
	return ErrWSClosed
}

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
	envs, err := ws.sendAndWait(ctx, id, cmd, len(params.Channels))
	if err != nil {
		return nil, err
	}
	result := make([]types.SubscribedResponse, 0, len(envs))
	for _, env := range envs {
		if env.Type != "subscribed" {
			continue
		}
		var m types.SubscribedMsg
		if len(env.Msg) > 0 {
			json.Unmarshal(env.Msg, &m)
		}
		result = append(result, types.SubscribedResponse{
			ID:   env.ID,
			Type: "subscribed",
			Msg:  m,
		})
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
	_, err := ws.sendAndWait(ctx, id, cmd, len(sids))
	return err
}

func (ws *WSConn) ListSubscriptions(ctx context.Context) (*types.ListSubscriptionsResponse, error) {
	id := ws.nextIDVal()
	cmd := types.ListSubscriptionsCommand{ID: id, Cmd: "list_subscriptions"}
	envs, err := ws.sendAndWait(ctx, id, cmd, 1)
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
		json.Unmarshal(env.Msg, &list.Msg)
	}
	return &list, nil
}

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
	id := ws.nextIDVal()
	cmd := types.UpdateSubscriptionCommand{ID: id, Cmd: "update_subscription", Params: params}
	envs, err := ws.sendAndWait(ctx, id, cmd, 1)
	if err != nil {
		return nil, err
	}
	if len(envs) == 0 {
		return nil, errors.New("no response")
	}
	env := envs[0]
	var ok types.OKResponse
	ok.ID = env.ID
	ok.SID = env.SID
	ok.Seq = env.Seq
	ok.Type = env.Type
	if len(env.Msg) > 0 {
		ok.Msg = &types.OKMsg{}
		json.Unmarshal(env.Msg, ok.Msg)
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
// the terminal read error, ErrWSSlowConsumer, or ErrWSClosed after Close().
func (ws *WSConn) Err() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.readErr
}

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
	if !healthy {
		ws.conn.Close()
		<-ws.readDone
		return nil
	}
	ws.writeMu.Lock()
	err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	ws.writeMu.Unlock()
	if e := ws.conn.Close(); e != nil && err == nil {
		err = e
	}
	select {
	case <-ws.readDone:
		return err
	case <-time.After(5 * time.Second):
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
