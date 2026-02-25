package oddrip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/oddrip/client/oddrip/types"
)

const defaultWSHost = "api.elections.kalshi.com"
const defaultWSPath = "/trade-api/ws/v2"

var (
	ErrWSClosed     = errors.New("websocket closed")
	ErrWSAuthRequired = errors.New("websocket requires auth")
)

type WSConn struct {
	conn     *websocket.Conn
	auth     AuthProvider
	host     string
	path     string
	nextID   atomic.Int64
	mu       sync.Mutex
	closed   bool
	readErr  error
	pendMu   sync.Mutex
	pending  map[int]chan *wsEnvelope
	msgChan  chan *types.WSMessage
	readDone chan struct{}
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
	scheme string
	host  string
	path  string
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

func (c *Client) ConnectWS(ctx context.Context, opts ...WSOption) (*WSConn, error) {
	if c.auth == nil {
		return nil, ErrWSAuthRequired
	}
	cfg := wsOpts{scheme: "wss", host: defaultWSHost, path: defaultWSPath}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.scheme == "" {
		cfg.scheme = "wss"
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
		conn:     conn,
		auth:     c.auth,
		host:     cfg.host,
		path:     cfg.path,
		pending:  make(map[int]chan *wsEnvelope),
		msgChan:  make(chan *types.WSMessage, 256),
		readDone: make(chan struct{}),
	}
	ws.nextID.Store(1)
	go ws.readLoop()
	return ws, nil
}

func (ws *WSConn) readLoop() {
	defer close(ws.readDone)
	defer close(ws.msgChan)
	for {
		_, data, err := ws.conn.ReadMessage()
		if err != nil {
			ws.mu.Lock()
			ws.readErr = err
			ws.mu.Unlock()
			ws.drainPending(err)
			return
		}
		var env wsEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		ws.pendMu.Lock()
		ch, ok := ws.pending[env.ID]
		delete(ws.pending, env.ID)
		ws.pendMu.Unlock()
		if ok && ch != nil {
			select {
			case ch <- &env:
			default:
			}
		}
		msg := &types.WSMessage{Type: env.Type, SID: env.SID, Seq: env.Seq, Msg: env.Msg}
		select {
		case ws.msgChan <- msg:
		default:
		}
	}
}

func (ws *WSConn) drainPending(err error) {
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
	if ws.closed {
		ws.mu.Unlock()
		return nil, ErrWSClosed
	}
	ch := make(chan *wsEnvelope, 8)
	ws.pendMu.Lock()
	ws.pending[id] = ch
	ws.pendMu.Unlock()
	ws.mu.Unlock()
	defer func() {
		ws.pendMu.Lock()
		delete(ws.pending, id)
		ws.pendMu.Unlock()
	}()

	if err := ws.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, err
	}
	var out []*wsEnvelope
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case env, ok := <-ch:
			if !ok {
				ws.mu.Lock()
				e := ws.readErr
				ws.mu.Unlock()
				if e != nil {
					return nil, e
				}
				return nil, ErrWSClosed
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
	if params.Action != "add_markets" && params.Action != "delete_markets" {
		return nil, errors.New("action must be add_markets or delete_markets")
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

func (ws *WSConn) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.closed {
		return nil
	}
	ws.closed = true
	err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if e := ws.conn.Close(); e != nil && err == nil {
		err = e
	}
	<-ws.readDone
	return err
}

type WSError struct {
	Code    int
	Message string
}

func (e *WSError) Error() string {
	return fmt.Sprintf("ws error %d: %s", e.Code, e.Message)
}
