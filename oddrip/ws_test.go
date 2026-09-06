package oddrip

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

type mockWSAuth struct{}

func (m *mockWSAuth) Apply(req *http.Request) error {
	req.Header.Set("KALSHI-ACCESS-KEY", "test")
	req.Header.Set("KALSHI-ACCESS-SIGNATURE", "test")
	req.Header.Set("KALSHI-ACCESS-TIMESTAMP", "0")
	return nil
}
