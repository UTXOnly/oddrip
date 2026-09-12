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
	"strings"
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

// wsUpdateReplier serves one update_subscription command: it decodes the
// command and writes whatever frames reply(id, params) returns, in order.
func wsUpdateReplier(reply func(id int, params types.UpdateSubscriptionParams) []map[string]interface{}) func(*websocket.Conn) {
	return func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd types.UpdateSubscriptionCommand
		if json.Unmarshal(data, &cmd) != nil || cmd.Cmd != "update_subscription" {
			return
		}
		for _, frame := range reply(cmd.ID, cmd.Params) {
			body, _ := json.Marshal(frame)
			if conn.WriteMessage(websocket.TextMessage, body) != nil {
				return
			}
		}
		wsDrain(conn)
	}
}

func wsTestCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// The spec answers get_snapshot with orderbook_snapshot frames that carry no
// command id; the call must complete on the first one for the subscription.
func TestWS_UpdateSubscription_GetSnapshot_SnapshotOnly(t *testing.T) {
	snapshot := map[string]interface{}{
		"market_ticker":  "FED-23DEC-T3.00",
		"market_id":      "9b0f6b43-5b68-4f9f-9f02-9a2d1b8ac1a1",
		"yes_dollars_fp": [][]string{{"0.0800", "300.00"}},
	}
	ws := wsTestConnect(t, wsTestServer(t, wsUpdateReplier(func(id int, p types.UpdateSubscriptionParams) []map[string]interface{} {
		if p.Action != types.WSUpdateSubscriptionGetSnapshot || len(p.Sids) != 1 || p.Sids[0] != 7 {
			return []map[string]interface{}{{"id": id, "type": "error", "msg": map[string]interface{}{"code": 1, "msg": "bad command"}}}
		}
		return []map[string]interface{}{
			// A snapshot for a different sid must not satisfy the waiter.
			{"type": "orderbook_snapshot", "sid": 8, "seq": 1, "msg": snapshot},
			{"type": "orderbook_snapshot", "sid": 7, "seq": 12, "msg": snapshot},
		}
	})))
	resp, err := ws.UpdateSubscription(wsTestCtx(t), types.UpdateSubscriptionParams{
		Sids:          []int{7},
		MarketTickers: []string{"FED-23DEC-T3.00"},
		Action:        types.WSUpdateSubscriptionGetSnapshot,
	})
	if err != nil {
		t.Fatalf("UpdateSubscription(get_snapshot): %v", err)
	}
	if resp.Type != types.WSTypeOrderbookSnapshot || resp.SID != 7 || resp.Seq != 12 || resp.Msg != nil {
		t.Fatalf("resp = %+v", resp)
	}
	// Both frames still reach the consumer.
	var got []int
	for len(got) < 2 {
		select {
		case m := <-ws.Messages():
			if m.Type == types.WSTypeOrderbookSnapshot {
				got = append(got, m.SID)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("snapshots on Messages(): got %v", got)
		}
	}
	if got[0] != 8 || got[1] != 7 {
		t.Fatalf("snapshot sids on Messages() = %v", got)
	}
}

func TestWS_UpdateSubscription_GetSnapshot_OKReply(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsUpdateReplier(func(id int, p types.UpdateSubscriptionParams) []map[string]interface{} {
		return []map[string]interface{}{{"id": id, "sid": 7, "seq": 3, "type": "ok", "msg": map[string]interface{}{"market_tickers": []string{"A", "B"}}}}
	})))
	sid := 7
	resp, err := ws.UpdateSubscription(wsTestCtx(t), types.UpdateSubscriptionParams{
		SID: &sid, MarketTickers: []string{"A"}, Action: types.WSUpdateSubscriptionGetSnapshot,
	})
	if err != nil {
		t.Fatalf("UpdateSubscription: %v", err)
	}
	if resp.Type != types.WSTypeOK || resp.SID != 7 || resp.Msg == nil || len(resp.Msg.MarketTickers) != 2 {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestWS_UpdateSubscription_GetSnapshot_ServerError(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsUpdateReplier(func(id int, p types.UpdateSubscriptionParams) []map[string]interface{} {
		return []map[string]interface{}{{"id": id, "type": "error", "msg": map[string]interface{}{"code": 6, "msg": "Invalid subscription id"}}}
	})))
	sid := 99
	_, err := ws.UpdateSubscription(wsTestCtx(t), types.UpdateSubscriptionParams{
		SID: &sid, MarketTickers: []string{"A"}, Action: types.WSUpdateSubscriptionGetSnapshot,
	})
	var wsErr *WSError
	if !errors.As(err, &wsErr) || wsErr.Code != 6 {
		t.Fatalf("err = %v, want *WSError code 6", err)
	}
}

func TestWS_UpdateSubscription_GetSnapshot_RequiresSID(t *testing.T) {
	ws := &WSConn{}
	_, err := ws.UpdateSubscription(context.Background(), types.UpdateSubscriptionParams{
		MarketTickers: []string{"A"}, Action: types.WSUpdateSubscriptionGetSnapshot,
	})
	if err == nil {
		t.Fatal("expected error when get_snapshot has no sid")
	}
}

func TestWS_UpdateSubscription_AddMarkets_OK(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsUpdateReplier(func(id int, p types.UpdateSubscriptionParams) []map[string]interface{} {
		if p.Action != types.WSUpdateSubscriptionAddMarkets {
			return []map[string]interface{}{{"id": id, "type": "error", "msg": map[string]interface{}{"code": 1, "msg": "bad command"}}}
		}
		return []map[string]interface{}{
			{"id": id, "sid": 456, "seq": 222, "type": "ok", "msg": map[string]interface{}{"market_tickers": []string{"MARKET-1", "MARKET-2", "MARKET-3"}}},
			{"type": "orderbook_snapshot", "sid": 456, "seq": 223, "msg": map[string]interface{}{"market_ticker": "MARKET-3", "market_id": "x"}},
		}
	})))
	resp, err := ws.UpdateSubscription(wsTestCtx(t), types.UpdateSubscriptionParams{
		Sids: []int{456}, MarketTickers: []string{"MARKET-3"}, Action: types.WSUpdateSubscriptionAddMarkets,
	})
	if err != nil {
		t.Fatalf("UpdateSubscription: %v", err)
	}
	if resp.Type != types.WSTypeOK || resp.ID == 0 || resp.SID != 456 || resp.Seq != 222 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Msg == nil || len(resp.Msg.MarketTickers) != 3 {
		t.Fatalf("resp.Msg = %+v", resp.Msg)
	}
}

// indexlist / underlying_list replies carry the command id but a list-specific
// type; the list itself decodes into OKMsg.
func TestWS_UpdateSubscription_IndexList(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsUpdateReplier(func(id int, p types.UpdateSubscriptionParams) []map[string]interface{} {
		return []map[string]interface{}{{"id": id, "sid": 3, "seq": 9, "type": "cfbenchmarks_value_indexlist", "msg": map[string]interface{}{"index_ids": []string{"BRTI", "ETHUSD_RTI"}}}}
	})))
	sid := 3
	resp, err := ws.UpdateSubscription(wsTestCtx(t), types.UpdateSubscriptionParams{SID: &sid, Action: types.WSUpdateSubscriptionIndexList})
	if err != nil {
		t.Fatalf("UpdateSubscription: %v", err)
	}
	if resp.Type != types.WSTypeCFBenchmarksValueIndexList || resp.Msg == nil || len(resp.Msg.IndexIDs) != 2 || resp.Msg.IndexIDs[0] != "BRTI" {
		t.Fatalf("resp = %+v msg=%+v", resp, resp.Msg)
	}
}

// A get_snapshot waiter must not survive the call: a later snapshot on the
// same sid goes only to Messages().
func TestWS_UpdateSubscription_GetSnapshot_WaiterRemoved(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsUpdateReplier(func(id int, p types.UpdateSubscriptionParams) []map[string]interface{} {
		return []map[string]interface{}{{"type": "orderbook_snapshot", "sid": 7, "seq": 1, "msg": map[string]interface{}{"market_ticker": "A", "market_id": "x"}}}
	})))
	sid := 7
	if _, err := ws.UpdateSubscription(wsTestCtx(t), types.UpdateSubscriptionParams{SID: &sid, MarketTickers: []string{"A"}, Action: types.WSUpdateSubscriptionGetSnapshot}); err != nil {
		t.Fatalf("UpdateSubscription: %v", err)
	}
	ws.pendMu.Lock()
	n := len(ws.snapshots) + len(ws.pending)
	ws.pendMu.Unlock()
	if n != 0 {
		t.Fatalf("waiters left registered: %d", n)
	}
}

// Two get_snapshot calls in flight on the same sid both complete on the next
// snapshot; neither starves the other.
func TestWS_UpdateSubscription_GetSnapshot_ConcurrentSameSID(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		for i := 0; i < 2; i++ {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
		body, _ := json.Marshal(map[string]interface{}{"type": "orderbook_snapshot", "sid": 7, "seq": 1, "msg": map[string]interface{}{"market_ticker": "A", "market_id": "x"}})
		conn.WriteMessage(websocket.TextMessage, body)
		wsDrain(conn)
	}))
	sid := 7
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			_, err := ws.UpdateSubscription(wsTestCtx(t), types.UpdateSubscriptionParams{SID: &sid, MarketTickers: []string{"A"}, Action: types.WSUpdateSubscriptionGetSnapshot})
			errs <- err
		}()
	}
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("UpdateSubscription: %v", err)
		}
	}
}

// A channel rejected after others were accepted: the accepted sids are live
// on the server and must be returned alongside the error.
func TestWS_Subscribe_PartialFailure(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd struct {
			ID int `json:"id"`
		}
		json.Unmarshal(data, &cmd)
		for _, frame := range []map[string]interface{}{
			{"id": cmd.ID, "type": "subscribed", "msg": map[string]interface{}{"channel": "ticker", "sid": 1}},
			{"id": cmd.ID, "type": "error", "msg": map[string]interface{}{"code": 8, "msg": "Unknown channel name"}},
		} {
			body, _ := json.Marshal(frame)
			conn.WriteMessage(websocket.TextMessage, body)
		}
		wsDrain(conn)
	}))
	subs, err := ws.Subscribe(wsTestCtx(t), types.SubscribeParams{Channels: []string{"ticker", "bogus"}})
	var wsErr *WSError
	if !errors.As(err, &wsErr) || wsErr.Code != 8 {
		t.Fatalf("err = %v, want *WSError code 8", err)
	}
	if len(subs) != 1 || subs[0].Msg.SID != 1 || subs[0].Msg.Channel != "ticker" {
		t.Fatalf("partial subs = %+v, want the accepted ticker sid", subs)
	}
}

func TestWS_Subscribe_MalformedReply(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd struct {
			ID int `json:"id"`
		}
		json.Unmarshal(data, &cmd)
		body, _ := json.Marshal(map[string]interface{}{"id": cmd.ID, "type": "subscribed", "msg": "not-an-object"})
		conn.WriteMessage(websocket.TextMessage, body)
		wsDrain(conn)
	}))
	subs, err := ws.Subscribe(wsTestCtx(t), types.SubscribeParams{Channels: []string{"ticker"}})
	if err == nil {
		t.Fatalf("expected decode error, got subs=%+v", subs)
	}
	var wsErr *WSError
	if errors.As(err, &wsErr) {
		t.Fatalf("decode failure must not be reported as a server *WSError: %v", err)
	}
}

func TestWS_Unsubscribe_OnePerSID(t *testing.T) {
	// The server confirms each sid separately (AsyncAPI unsubscribedResponse:
	// {"id":102,"sid":2,"seq":7,"type":"unsubscribed"}), so Unsubscribe waits for
	// len(sids) id-matched replies.
	type unsubCmd struct {
		ID     int    `json:"id"`
		Cmd    string `json:"cmd"`
		Params struct {
			Sids []int `json:"sids"`
		} `json:"params"`
	}
	sent := make(chan unsubCmd, 1)
	ws := wsTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd unsubCmd
		json.Unmarshal(data, &cmd)
		sent <- cmd
		for i, sid := range cmd.Params.Sids {
			body, _ := json.Marshal(map[string]interface{}{"id": cmd.ID, "sid": sid, "seq": 7 + i, "type": "unsubscribed"})
			if conn.WriteMessage(websocket.TextMessage, body) != nil {
				return
			}
		}
		wsDrain(conn)
	}))

	start := time.Now()
	if err := ws.Unsubscribe(wsTestCtx(t), []int{1, 2}); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Unsubscribe took %v", d)
	}
	cmd := <-sent
	if cmd.Cmd != "unsubscribe" || len(cmd.Params.Sids) != 2 || cmd.Params.Sids[0] != 1 || cmd.Params.Sids[1] != 2 {
		t.Fatalf("command sent: %+v", cmd)
	}
	if err := ws.Unsubscribe(wsTestCtx(t), nil); err == nil {
		t.Fatal("empty sids should be rejected before sending")
	}
}

func TestWS_ListSubscriptions_OK(t *testing.T) {
	// AsyncAPI listSubscriptionsResponse: type is "ok" and msg is an array.
	ws := wsTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd struct {
			ID  int    `json:"id"`
			Cmd string `json:"cmd"`
		}
		json.Unmarshal(data, &cmd)
		if cmd.Cmd != "list_subscriptions" {
			return
		}
		body, _ := json.Marshal(map[string]interface{}{
			"id":   cmd.ID,
			"type": "ok",
			"msg": []map[string]interface{}{
				{"channel": "orderbook_delta", "sid": 1},
				{"channel": "ticker", "sid": 2},
				{"channel": "fill", "sid": 3},
			},
		})
		conn.WriteMessage(websocket.TextMessage, body)
		wsDrain(conn)
	}))

	list, err := ws.ListSubscriptions(wsTestCtx(t))
	if err != nil {
		t.Fatalf("ListSubscriptions: %v", err)
	}
	if list.Type != types.WSTypeOK || list.ID == 0 {
		t.Fatalf("reply envelope: %+v", list)
	}
	want := []types.ListSubscriptionsItem{{Channel: "orderbook_delta", SID: 1}, {Channel: "ticker", SID: 2}, {Channel: "fill", SID: 3}}
	if len(list.Msg) != len(want) {
		t.Fatalf("msg: %+v", list.Msg)
	}
	for i := range want {
		if list.Msg[i] != want[i] {
			t.Errorf("msg[%d] = %+v, want %+v", i, list.Msg[i], want[i])
		}
	}
}

func TestWS_MalformedFrame_FailsConnection(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		good, _ := json.Marshal(map[string]interface{}{"type": "ticker", "sid": 1, "seq": 1, "msg": map[string]interface{}{"market_ticker": "A"}})
		if conn.WriteMessage(websocket.TextMessage, good) != nil {
			return
		}
		if conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ticker","sid":1,`)) != nil {
			return
		}
		wsDrain(conn)
	}))

	var got []*types.WSMessage
	for msg := range ws.Messages() {
		got = append(got, msg)
	}
	if len(got) != 1 || got[0].Type != types.WSTypeTicker {
		t.Fatalf("messages before the bad frame: %+v", got)
	}
	select {
	case <-ws.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done() did not close")
	}
	err := ws.Err()
	if !errors.Is(err, ErrWSMalformedFrame) {
		t.Fatalf("Err() = %v, want ErrWSMalformedFrame", err)
	}
	if err == ErrWSMalformedFrame {
		t.Fatalf("Err() should wrap the decode error, got bare sentinel")
	}
	if _, err := ws.Subscribe(wsTestCtx(t), types.SubscribeParams{Channels: []string{"ticker"}}); !errors.Is(err, ErrWSMalformedFrame) {
		t.Fatalf("Subscribe after failure = %v, want ErrWSMalformedFrame", err)
	}
	if err := ws.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestWSWriteTimeout_Option(t *testing.T) {
	for _, tc := range []struct {
		opt  time.Duration
		want time.Duration
	}{
		{0, defaultWSWriteTimeout},
		{-1, defaultWSWriteTimeout},
		{3 * time.Second, 3 * time.Second},
	} {
		ws := wsTestConnect(t, wsTestServer(t, wsDrain), WSWriteTimeout(tc.opt))
		if ws.writeTimeout != tc.want {
			t.Errorf("WSWriteTimeout(%v): writeTimeout = %v, want %v", tc.opt, ws.writeTimeout, tc.want)
		}
	}
	ws := wsTestConnect(t, wsTestServer(t, wsDrain))
	if ws.writeTimeout != defaultWSWriteTimeout {
		t.Errorf("default writeTimeout = %v, want %v", ws.writeTimeout, defaultWSWriteTimeout)
	}
}

// A caller whose context ends while another command holds the write slot gets
// ctx.Err() without its frame ever starting, and the connection stays healthy
// for the next command.
func TestWS_Write_ContextEndsWaitingForSlot(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsSubscribeEcho))
	params := types.SubscribeParams{Channels: []string{types.WSChannelTicker}}
	ws.writeSem <- struct{}{} // another command is mid-write

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ws.Subscribe(cancelled, params); err != context.Canceled {
		t.Fatalf("Subscribe with cancelled ctx = %v, want context.Canceled", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := ws.Subscribe(ctx, params)
	if err != context.DeadlineExceeded {
		t.Fatalf("Subscribe while slot held = %v, want context.DeadlineExceeded", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Subscribe returned after %v, want promptly after its 50ms deadline", d)
	}
	if err := ws.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil: giving up on the slot must not fail the connection", err)
	}
	ws.pendMu.Lock()
	n := len(ws.pending)
	ws.pendMu.Unlock()
	if n != 0 {
		t.Errorf("waiters left registered: %d", n)
	}

	<-ws.writeSem
	subs, err := ws.Subscribe(wsTestCtx(t), params)
	if err != nil || len(subs) != 1 {
		t.Fatalf("Subscribe after slot released: subs=%+v err=%v", subs, err)
	}
}

// Cancelled and live callers contending for the write slot: every live call
// completes and every cancelled one returns ctx.Err() without being sent.
func TestWS_Subscribe_ConcurrentWithCancelled(t *testing.T) {
	ws := wsTestConnect(t, wsTestServer(t, wsSubscribeEcho))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cancelled, cancelNow := context.WithCancel(ctx)
	cancelNow()

	const n = 20
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := ctx
			if i%2 == 1 {
				c = cancelled
			}
			subs, err := ws.Subscribe(c, types.SubscribeParams{Channels: []string{fmt.Sprintf("ch%d", i)}})
			if err == nil && len(subs) != 1 {
				err = fmt.Errorf("got %d subscribed", len(subs))
			}
			errs[i] = err
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if i%2 == 1 {
			if err != context.Canceled {
				t.Errorf("cancelled Subscribe %d = %v, want context.Canceled", i, err)
			}
		} else if err != nil {
			t.Errorf("Subscribe %d: %v", i, err)
		}
	}
	if err := ws.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

// A write that cannot complete within WSWriteTimeout fails the connection:
// the command returns an error wrapping ErrWSWriteTimeout, Err() reports it,
// and Messages() closes, the same path as a slow consumer.
func TestWS_WriteTimeout_FailsConnection(t *testing.T) {
	ws := wsTestConnect(t, wsNoReadServer(t), WSWriteTimeout(200*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	_, err := ws.Subscribe(ctx, wsJamParams())
	if !errors.Is(err, ErrWSWriteTimeout) {
		t.Fatalf("Subscribe = %v, want ErrWSWriteTimeout", err)
	}
	if err == ErrWSWriteTimeout {
		t.Fatalf("Subscribe should wrap the socket error, got bare sentinel")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("write failed after %v with a 200ms write timeout", d)
	}
	select {
	case <-ws.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done() did not close")
	}
	if e := ws.Err(); !errors.Is(e, ErrWSWriteTimeout) || e == ErrWSWriteTimeout {
		t.Fatalf("Err() = %v, want wrapped ErrWSWriteTimeout", e)
	}
	if _, ok := <-ws.Messages(); ok {
		t.Error("Messages() not closed")
	}
	start = time.Now()
	if _, serr := ws.Subscribe(wsTestCtx(t), types.SubscribeParams{Channels: []string{types.WSChannelTicker}}); !errors.Is(serr, ErrWSWriteTimeout) {
		t.Errorf("Subscribe after failure = %v, want ErrWSWriteTimeout", serr)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Subscribe after failure took %v, want immediate", d)
	}
	if err := ws.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	if e := ws.Err(); !errors.Is(e, ErrWSWriteTimeout) {
		t.Errorf("Err() after Close = %v, want ErrWSWriteTimeout", e)
	}
}

// The caller's context deadline bounds the write too: with the default write
// timeout, a command with a short deadline against a peer that is not reading
// returns ctx.Err() at that deadline. The frame was cut off mid-write, which
// leaves the socket unusable, so the connection is failed as well.
func TestWS_WriteTimeout_CallerDeadline(t *testing.T) {
	ws := wsTestConnect(t, wsNoReadServer(t))
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := ws.Subscribe(ctx, wsJamParams())
	if err != context.DeadlineExceeded {
		t.Fatalf("Subscribe = %v, want context.DeadlineExceeded", err)
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("Subscribe returned after %v with a 200ms deadline", d)
	}
	select {
	case <-ws.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done() did not close")
	}
	if e := ws.Err(); !errors.Is(e, ErrWSWriteTimeout) {
		t.Fatalf("Err() = %v, want ErrWSWriteTimeout", e)
	}
}

// While one command's write is blocked in the socket, another command with a
// shorter deadline gives up on the write slot at that deadline instead of
// queueing behind the blocked write until it times out.
func TestWS_Write_BlockedWriterDoesNotHoldOthers(t *testing.T) {
	ws := wsTestConnect(t, wsNoReadServer(t), WSWriteTimeout(time.Second))

	first := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := ws.Subscribe(ctx, wsJamParams())
		first <- err
	}()
	wsWaitSlotHeld(t, ws)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := ws.Subscribe(ctx, types.SubscribeParams{Channels: []string{types.WSChannelTicker}})
	if err != context.DeadlineExceeded {
		t.Fatalf("second Subscribe = %v, want context.DeadlineExceeded", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("second Subscribe returned after %v, want promptly after its 100ms deadline", d)
	}
	select {
	case err := <-first:
		if !errors.Is(err, ErrWSWriteTimeout) {
			t.Fatalf("blocked Subscribe = %v, want ErrWSWriteTimeout", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("blocked Subscribe did not return")
	}
}

// Close does not wait for a command write that is stuck in the socket: it
// skips the close frame, closes the socket, and that write returns.
func TestWS_Close_BoundedWithStuckWriter(t *testing.T) {
	ws := wsTestConnect(t, wsNoReadServer(t), WSWriteTimeout(2*time.Second))

	stuck := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := ws.Subscribe(ctx, wsJamParams())
		stuck <- err
	}()
	wsWaitSlotHeld(t, ws)

	start := time.Now()
	if err := ws.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Close took %v with a command write stuck", d)
	}
	select {
	case err := <-stuck:
		if err != ErrWSClosed {
			t.Errorf("stuck Subscribe after Close = %v, want ErrWSClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stuck Subscribe did not return after Close")
	}
	select {
	case <-ws.Done():
	default:
		t.Error("Done() not closed")
	}
	if err := ws.Err(); err != ErrWSClosed {
		t.Errorf("Err() = %v, want ErrWSClosed", err)
	}
}

// With the socket already full and no command in flight, the close frame
// itself cannot be delivered; Close still returns within the write timeout.
func TestWS_Close_BoundedWhenSocketFull(t *testing.T) {
	ws := wsTestConnect(t, wsNoReadServer(t), WSWriteTimeout(200*time.Millisecond))
	// Fill the socket underneath gorilla. The peer never reads, so it does
	// not matter that these bytes are not WebSocket frames.
	nc := ws.conn.NetConn()
	nc.SetWriteDeadline(time.Now().Add(200 * time.Millisecond))
	if _, err := nc.Write(make([]byte, wsJamSize)); err == nil {
		t.Fatal("expected the raw write to block and time out")
	}
	nc.SetWriteDeadline(time.Time{})

	start := time.Now()
	err := ws.Close()
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("Close took %v with the socket full and a 200ms write timeout", d)
	}
	// The close frame is a few bytes; whether it squeezes into the jammed
	// send buffer depends on kernel drain timing. Either it fits (nil) or
	// its control write hits the 200ms deadline; anything else is a bug.
	var ne net.Error
	if err != nil && (!errors.As(err, &ne) || !ne.Timeout()) {
		t.Errorf("Close = %v, want nil or the close frame's write timeout", err)
	}
	select {
	case <-ws.Done():
	default:
		t.Error("Done() not closed")
	}
	if err := ws.Err(); err != ErrWSClosed {
		t.Errorf("Err() = %v, want ErrWSClosed", err)
	}
	if err := ws.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

// wsJamSize is larger than any default loopback send+receive buffer pair
// (macOS autotunes each to at most 4MB), so writing it to a peer that never
// reads blocks until the write deadline.
const wsJamSize = 16 << 20

func wsJamParams() types.SubscribeParams {
	return types.SubscribeParams{
		Channels:      []string{types.WSChannelTicker},
		MarketTickers: []string{strings.Repeat("x", wsJamSize)},
	}
}

// wsNoReadServer accepts the connection and never reads from it.
func wsNoReadServer(t *testing.T) *httptest.Server {
	t.Helper()
	block := make(chan struct{})
	t.Cleanup(func() { close(block) })
	return wsTestServer(t, func(*websocket.Conn) { <-block })
}

// wsWaitSlotHeld returns once a command holds the write slot.
func wsWaitSlotHeld(t *testing.T, ws *WSConn) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for len(ws.writeSem) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("no command took the write slot")
		}
		time.Sleep(time.Millisecond)
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
