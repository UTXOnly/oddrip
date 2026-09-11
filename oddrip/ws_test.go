package oddrip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

func TestConnectWS_NoAuth(t *testing.T) {
	client := New()
	ctx := context.Background()

	_, err := client.ConnectWS(ctx)
	if err != ErrWSAuthRequired {
		t.Fatalf("ConnectWS without auth: got %v", err)
	}
}

func TestConnectWS_Subscribe_Integration(t *testing.T) {
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd struct {
			ID     int `json:"id"`
			Params struct {
				Channels []string `json:"channels"`
			} `json:"params"`
		}
		if json.Unmarshal(data, &cmd) != nil {
			return
		}
		for i, ch := range cmd.Params.Channels {
			body, _ := json.Marshal(map[string]interface{}{
				"id":   cmd.ID,
				"type": "subscribed",
				"msg":  map[string]interface{}{"channel": ch, "sid": i + 1},
			})
			conn.WriteMessage(websocket.TextMessage, body)
		}
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	client := New(Auth(&mockWSAuth{}))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := client.ConnectWS(ctx,
		WSScheme("ws"),
		WSHost(u.Host),
		WSPath("/"),
	)
	if err != nil {
		t.Fatalf("ConnectWS: %v", err)
	}
	defer ws.Close()

	subs, err := ws.Subscribe(ctx, types.SubscribeParams{
		Channels: []string{types.WSChannelTicker},
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if len(subs) != 1 {
		t.Errorf("expected 1 subscribed, got %d", len(subs))
	}
	if len(subs) > 0 && (subs[0].Msg.Channel != types.WSChannelTicker || subs[0].Msg.SID != 1) {
		t.Errorf("subscribed: channel=%s sid=%d", subs[0].Msg.Channel, subs[0].Msg.SID)
	}
}

func TestConnectWS_DialFails(t *testing.T) {
	client := New(Auth(&mockWSAuth{}))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.ConnectWS(ctx, WSHost("invalid.invalid"))
	if err == nil {
		t.Fatal("expected dial error")
	}
}

func TestWSError_Error(t *testing.T) {
	e := &WSError{Code: 8, Message: "Unknown channel name"}
	if e.Error() != "ws error 8: Unknown channel name" {
		t.Errorf("WSError.Error(): %s", e.Error())
	}
}

func TestWSChannelConstants(t *testing.T) {
	if types.WSChannelTicker != "ticker" || types.WSChannelOrderbookDelta != "orderbook_delta" {
		t.Errorf("channel constants wrong")
	}
	if types.WSChannelMultivariateLifecycle != "multivariate_market_lifecycle" {
		t.Errorf("multivariate lifecycle channel: %q", types.WSChannelMultivariateLifecycle)
	}
	if types.WSChannelPythValue != "pyth_value" {
		t.Errorf("pyth_value channel: %q", types.WSChannelPythValue)
	}
	if types.WSUpdateSubscriptionGetSnapshot != "get_snapshot" {
		t.Errorf("get_snapshot action: %q", types.WSUpdateSubscriptionGetSnapshot)
	}
	if types.WSUpdateSubscriptionSubscribeUnderlyings != "subscribe_underlyings" {
		t.Errorf("subscribe_underlyings action: %q", types.WSUpdateSubscriptionSubscribeUnderlyings)
	}
	if types.WSChannelCFBenchmarksValue != "cfbenchmarks_value" || types.WSChannelCFBenchmarksValue5Hz != "cfbenchmarks_value_5hz" {
		t.Errorf("cfbenchmarks channels: %q %q", types.WSChannelCFBenchmarksValue, types.WSChannelCFBenchmarksValue5Hz)
	}
	if types.WSUpdateSubscriptionIndexList != "indexlist" || types.WSUpdateSubscriptionSubscribeIndices != "subscribe_indices" {
		t.Errorf("index actions: %q %q", types.WSUpdateSubscriptionIndexList, types.WSUpdateSubscriptionSubscribeIndices)
	}
}

func TestUpdateSubscription_RejectsUnknownAction(t *testing.T) {
	ws := &WSConn{}
	if _, err := ws.UpdateSubscription(context.Background(), types.UpdateSubscriptionParams{Action: "not_an_action"}); err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestWS_Subscribe_MultiChannel(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsSubscribeEcho))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channels := []string{types.WSChannelTicker, types.WSChannelOrderbookDelta}
	start := time.Now()
	subs, err := ws.Subscribe(ctx, types.SubscribeParams{Channels: channels})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Subscribe took %v", d)
	}
	if len(subs) != len(channels) {
		t.Fatalf("expected %d subscribed, got %d", len(channels), len(subs))
	}
	for i, s := range subs {
		if s.Msg.Channel != channels[i] || s.Msg.SID != i+1 {
			t.Errorf("subs[%d]: channel=%s sid=%d", i, s.Msg.Channel, s.Msg.SID)
		}
	}
}

func TestWS_Subscribe_ManyChannels(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsSubscribeEcho))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channels := make([]string, 12)
	for i := range channels {
		channels[i] = fmt.Sprintf("ch%d", i)
	}
	subs, err := ws.Subscribe(ctx, types.SubscribeParams{Channels: channels})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if len(subs) != 12 {
		t.Fatalf("expected 12 subscribed, got %d", len(subs))
	}
	for i, s := range subs {
		if s.Msg.Channel != channels[i] || s.Msg.SID != i+1 {
			t.Errorf("subs[%d]: channel=%s sid=%d", i, s.Msg.Channel, s.Msg.SID)
		}
	}
}

func TestWS_Subscribe_Concurrent(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsSubscribeEcho))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const n = 20
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			subs, err := ws.Subscribe(ctx, types.SubscribeParams{Channels: []string{fmt.Sprintf("ch%d", i)}})
			if err == nil && len(subs) != 1 {
				err = fmt.Errorf("got %d subscribed", len(subs))
			}
			errs[i] = err
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("Subscribe %d: %v", i, err)
		}
	}
}

func TestWS_SlowConsumer(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		for i := 0; i < 100; i++ {
			body, _ := json.Marshal(map[string]interface{}{
				"type": "ticker", "sid": 1, "seq": i, "msg": map[string]interface{}{},
			})
			if conn.WriteMessage(websocket.TextMessage, body) != nil {
				return
			}
		}
		wsDrain(conn)
	}), WSBufferSize(4))

	select {
	case <-ws.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done() did not close")
	}
	if err := ws.Err(); err != ErrWSSlowConsumer {
		t.Fatalf("Err() = %v, want ErrWSSlowConsumer", err)
	}
	n := 0
	for range ws.Messages() {
		n++
	}
	if n > 4 {
		t.Errorf("buffered %d messages, want <= 4", n)
	}
	if err := ws.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	if err := ws.Err(); err != ErrWSSlowConsumer {
		t.Errorf("Err() after Close = %v, want ErrWSSlowConsumer", err)
	}
}

func TestWS_ReadTimeout(t *testing.T) {
	block := make(chan struct{})
	t.Cleanup(func() { close(block) })
	ws := wsTestConnect(t, wsTestServer(t, func(*websocket.Conn) { <-block }),
		WSReadTimeout(200*time.Millisecond), WSPingInterval(50*time.Millisecond))

	start := time.Now()
	select {
	case <-ws.Done():
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("Done() did not close")
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("dead connection detected after %v", d)
	}
	err := ws.Err()
	var ne net.Error
	if !errors.As(err, &ne) || !ne.Timeout() {
		t.Fatalf("Err() = %v, want net timeout", err)
	}
	if _, ok := <-ws.Messages(); ok {
		t.Error("Messages() not closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, serr := ws.Subscribe(ctx, types.SubscribeParams{Channels: []string{types.WSChannelTicker}}); serr != err {
		t.Errorf("Subscribe after timeout = %v, want %v", serr, err)
	}
}

func TestWS_Keepalive_Healthy(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsDrain),
		WSReadTimeout(200*time.Millisecond), WSPingInterval(50*time.Millisecond))

	select {
	case <-ws.Done():
		t.Fatalf("connection dropped: %v", ws.Err())
	case <-time.After(time.Second):
	}
	if err := ws.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil", err)
	}
	if err := ws.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	if err := ws.Err(); err != ErrWSClosed {
		t.Errorf("Err() after Close = %v, want ErrWSClosed", err)
	}
}

func TestWS_Close_Idempotent(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsSubscribeEcho))

	if err := ws.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := ws.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if err := ws.Err(); err != ErrWSClosed {
		t.Errorf("Err() = %v, want ErrWSClosed", err)
	}
	select {
	case <-ws.Done():
	default:
		t.Error("Done() not closed")
	}
	if _, ok := <-ws.Messages(); ok {
		t.Error("Messages() not closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	_, err := ws.Subscribe(ctx, types.SubscribeParams{Channels: []string{types.WSChannelTicker}})
	if err != ErrWSClosed {
		t.Errorf("Subscribe after Close = %v, want ErrWSClosed", err)
	}
	if d := time.Since(start); d > 100*time.Millisecond {
		t.Errorf("Subscribe after Close took %v", d)
	}
}

func TestWS_Subscribe_ServerError(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd struct {
			ID int `json:"id"`
		}
		json.Unmarshal(data, &cmd)
		body, _ := json.Marshal(map[string]interface{}{
			"id":   cmd.ID,
			"type": "error",
			"msg":  map[string]interface{}{"code": 8, "msg": "Unknown channel name"},
		})
		conn.WriteMessage(websocket.TextMessage, body)
		wsDrain(conn)
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := ws.Subscribe(ctx, types.SubscribeParams{Channels: []string{"bogus"}})
	var wsErr *WSError
	if !errors.As(err, &wsErr) {
		t.Fatalf("Subscribe: %v, want *WSError", err)
	}
	if wsErr.Code != 8 || wsErr.Message != "Unknown channel name" {
		t.Errorf("WSError = %+v", wsErr)
	}
	if err := ws.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

func wsTestServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		handler(conn)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func wsTestConnect(t *testing.T, srv *httptest.Server, opts ...WSOption) *WSConn {
	t.Helper()
	u, _ := url.Parse(srv.URL)
	client := New(Auth(&mockWSAuth{}))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ws, err := client.ConnectWS(ctx, append([]WSOption{WSScheme("ws"), WSHost(u.Host), WSPath("/")}, opts...)...)
	if err != nil {
		t.Fatalf("ConnectWS: %v", err)
	}
	t.Cleanup(func() { ws.Close() })
	return ws
}

func wsSubscribeEcho(conn *websocket.Conn) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd struct {
			ID     int `json:"id"`
			Params struct {
				Channels []string `json:"channels"`
			} `json:"params"`
		}
		if json.Unmarshal(data, &cmd) != nil {
			return
		}
		for i, ch := range cmd.Params.Channels {
			body, _ := json.Marshal(map[string]interface{}{
				"id":   cmd.ID,
				"type": "subscribed",
				"msg":  map[string]interface{}{"channel": ch, "sid": i + 1},
			})
			if conn.WriteMessage(websocket.TextMessage, body) != nil {
				return
			}
		}
	}
}

func wsDrain(conn *websocket.Conn) {
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

type mockWSAuth struct{}

func (m *mockWSAuth) Apply(req *http.Request) error {
	req.Header.Set("KALSHI-ACCESS-KEY", "test")
	req.Header.Set("KALSHI-ACCESS-SIGNATURE", "test")
	req.Header.Set("KALSHI-ACCESS-TIMESTAMP", "0")
	return nil
}
