package oddrip

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

func TestPerpsWS_DefaultURL(t *testing.T) {
	if got, want := wsConfig(perpsWSOptions(nil)).url(), "wss://external-api-margin-ws.kalshi.com/trade-api/ws/v2/margin"; got != want {
		t.Fatalf("default perps dial URL = %q, want %q", got, want)
	}
	// Demo is a host override; the perps path stays.
	if got, want := wsConfig(perpsWSOptions([]WSOption{WSHost("external-api-margin-ws.demo.kalshi.co")})).url(), "wss://external-api-margin-ws.demo.kalshi.co/trade-api/ws/v2/margin"; got != want {
		t.Fatalf("demo perps dial URL = %q, want %q", got, want)
	}
	// ConnectWS keeps the event-contract default.
	if got, want := wsConfig(nil).url(), "wss://external-api-ws.kalshi.com/trade-api/ws/v2"; got != want {
		t.Fatalf("event-contract dial URL = %q, want %q", got, want)
	}
}

func TestPerpsWS_OptionsOverrideDefaults(t *testing.T) {
	cfg := wsConfig(perpsWSOptions([]WSOption{WSScheme("ws"), WSHost("127.0.0.1:9"), WSPath("/custom"), WSBufferSize(8)}))
	if got, want := cfg.url(), "ws://127.0.0.1:9/custom"; got != want {
		t.Fatalf("overridden dial URL = %q, want %q", got, want)
	}
	if cfg.bufferSize != 8 {
		t.Fatalf("bufferSize = %d, want 8", cfg.bufferSize)
	}

	// Against a live server: overriding only the host still dials the perps
	// path, and a WSPath override replaces it.
	srv, paths := perpsWSPathServer(t)
	perpsWSTestConnect(t, srv)
	if got := perpsWSNextPath(t, paths); got != defaultPerpsWSPath {
		t.Fatalf("dialed path = %q, want %q", got, defaultPerpsWSPath)
	}
	perpsWSTestConnect(t, srv, WSPath("/other"))
	if got := perpsWSNextPath(t, paths); got != "/other" {
		t.Fatalf("dialed path with WSPath override = %q, want /other", got)
	}
}

func TestPerpsWS_NoAuth(t *testing.T) {
	if _, err := New().ConnectPerpsWS(context.Background()); err != ErrWSAuthRequired {
		t.Fatalf("ConnectPerpsWS without auth: got %v, want ErrWSAuthRequired", err)
	}
}

// Subscribe on a perps connection returns the subscribed sid, and the book
// frames that follow reach Messages() and decode into the Margin* types.
func TestPerpsWS_Subscribe_OrderbookSnapshot(t *testing.T) {
	sent := make(chan types.SubscribeCommand, 1)
	ws := perpsWSTestConnect(t, wsTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var cmd types.SubscribeCommand
		json.Unmarshal(data, &cmd)
		sent <- cmd
		for _, frame := range []map[string]interface{}{
			{"id": cmd.ID, "type": "subscribed", "msg": map[string]interface{}{"channel": "orderbook_delta", "sid": 1}},
			{"type": "orderbook_snapshot", "sid": 1, "seq": 1, "msg": map[string]interface{}{
				"market_ticker": "KXBTCPERP",
				"bid":           [][]string{{"64990.0000", "1.50"}},
				"ask":           [][]string{{"65010.0000", "2.00"}, {"65020.0000", "4.00"}},
			}},
			{"type": "orderbook_delta", "sid": 1, "seq": 2, "msg": map[string]interface{}{
				"market_ticker": "KXBTCPERP", "price": "64990.0000", "delta": "-1.50", "side": "bid", "ts_ms": 1700000000123,
			}},
		} {
			body, _ := json.Marshal(frame)
			if conn.WriteMessage(websocket.TextMessage, body) != nil {
				return
			}
		}
		wsDrain(conn)
	}))

	subs, err := ws.Subscribe(wsTestCtx(t), types.SubscribeParams{
		Channels:      []string{types.WSChannelOrderbookDelta},
		MarketTickers: []string{"KXBTCPERP"},
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if len(subs) != 1 || subs[0].Msg.Channel != types.WSChannelOrderbookDelta || subs[0].Msg.SID != 1 {
		t.Fatalf("subscribed = %+v", subs)
	}
	cmd := <-sent
	if cmd.Cmd != "subscribe" || !reflect.DeepEqual(cmd.Params.Channels, []string{"orderbook_delta"}) || !reflect.DeepEqual(cmd.Params.MarketTickers, []string{"KXBTCPERP"}) {
		t.Fatalf("command sent: %+v", cmd)
	}

	// The subscribed reply reaches Messages() too; skip to the book frames.
	snap := perpsWSNextOfType(t, ws, types.WSTypeOrderbookSnapshot)
	if snap.SID != 1 || snap.Seq != 1 {
		t.Fatalf("snapshot frame = %+v", snap)
	}
	var book types.MarginOrderbookSnapshotMsg
	if err := snap.Decode(&book); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if book.MarketTicker != "KXBTCPERP" || len(book.Bid) != 1 || len(book.Ask) != 2 || book.Ask[1] != (types.OrderbookLevel{PriceDollars: "65020.0000", CountFp: "4.00"}) {
		t.Fatalf("snapshot = %+v", book)
	}

	delta := perpsWSNextOfType(t, ws, types.WSTypeOrderbookDelta)
	if delta.SID != 1 || delta.Seq != 2 {
		t.Fatalf("delta frame = %+v", delta)
	}
	var d types.MarginOrderbookDeltaMsg
	if err := delta.Decode(&d); err != nil {
		t.Fatalf("decode delta: %v", err)
	}
	if d.Price != "64990.0000" || d.Delta != "-1.50" || d.Side != types.BookSideBid || d.TsMs == nil || *d.TsMs != 1700000000123 {
		t.Fatalf("delta = %+v", d)
	}
}

// get_snapshot on a perps connection returns on the first orderbook_snapshot
// for the sid. The perps spec says the snapshot carries the command id when
// the command had one; the call completes either way.
func TestPerpsWS_UpdateSubscription_GetSnapshot(t *testing.T) {
	for _, withID := range []bool{false, true} {
		name := "snapshot without id"
		if withID {
			name = "snapshot with command id"
		}
		t.Run(name, func(t *testing.T) {
			ws := perpsWSTestConnect(t, wsTestServer(t, wsUpdateReplier(func(id int, p types.UpdateSubscriptionParams) []map[string]interface{} {
				if p.Action != types.WSUpdateSubscriptionGetSnapshot || p.SID == nil || *p.SID != 5 {
					return []map[string]interface{}{{"id": id, "type": "error", "msg": map[string]interface{}{"code": 13, "msg": "Unsupported action"}}}
				}
				frame := map[string]interface{}{"type": "orderbook_snapshot", "sid": 5, "seq": 9, "msg": map[string]interface{}{
					"market_ticker": "KXBTCPERP",
					"bid":           [][]string{{"64990.0000", "1.50"}},
				}}
				if withID {
					frame["id"] = id
				}
				return []map[string]interface{}{frame}
			})))
			sid := 5
			resp, err := ws.UpdateSubscription(wsTestCtx(t), types.UpdateSubscriptionParams{
				SID:           &sid,
				MarketTickers: []string{"KXBTCPERP"},
				Action:        types.WSUpdateSubscriptionGetSnapshot,
			})
			if err != nil {
				t.Fatalf("UpdateSubscription(get_snapshot): %v", err)
			}
			if resp.Type != types.WSTypeOrderbookSnapshot || resp.SID != 5 || resp.Seq != 9 || resp.Msg != nil {
				t.Fatalf("resp = %+v", resp)
			}
			m := perpsWSNextOfType(t, ws, types.WSTypeOrderbookSnapshot)
			var book types.MarginOrderbookSnapshotMsg
			if err := m.Decode(&book); err != nil {
				t.Fatalf("decode snapshot: %v", err)
			}
			if m.SID != 5 || book.MarketTicker != "KXBTCPERP" || len(book.Bid) != 1 || book.Ask != nil {
				t.Fatalf("snapshot on Messages() = %+v %+v", m, book)
			}
		})
	}
}

// perpsWSTestConnect dials srv through ConnectPerpsWS, overriding only the
// scheme and host so the perps default path is used.
func perpsWSTestConnect(t *testing.T, srv *httptest.Server, opts ...WSOption) *WSConn {
	t.Helper()
	u, _ := url.Parse(srv.URL)
	client := New(Auth(&mockWSAuth{}))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ws, err := client.ConnectPerpsWS(ctx, append([]WSOption{WSScheme("ws"), WSHost(u.Host)}, opts...)...)
	if err != nil {
		t.Fatalf("ConnectPerpsWS: %v", err)
	}
	t.Cleanup(func() { ws.Close() })
	return ws
}

// perpsWSPathServer is a WebSocket server that reports the path of each
// handshake it accepts.
func perpsWSPathServer(t *testing.T) (*httptest.Server, <-chan string) {
	t.Helper()
	paths := make(chan string, 4)
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths <- r.URL.Path
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		wsDrain(conn)
	}))
	t.Cleanup(srv.Close)
	return srv, paths
}

func perpsWSNextPath(t *testing.T, paths <-chan string) string {
	t.Helper()
	select {
	case p := <-paths:
		return p
	case <-time.After(2 * time.Second):
		t.Fatal("no handshake reached the server")
		return ""
	}
}

// perpsWSNextOfType returns the next message of type typ on Messages(),
// skipping others.
func perpsWSNextOfType(t *testing.T, ws *WSConn, typ string) *types.WSMessage {
	t.Helper()
	timeout := time.After(2 * time.Second)
	for {
		select {
		case m, ok := <-ws.Messages():
			if !ok {
				t.Fatalf("Messages() closed before a %s frame: %v", typ, ws.Err())
			}
			if m.Type == typ {
				return m
			}
		case <-timeout:
			t.Fatalf("no %s frame on Messages()", typ)
			return nil
		}
	}
}
