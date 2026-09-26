package oddrip

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

const perpsBracketTriggerJSON = `{
	"id":"trg-1","ticker":"KXBTCPERP","kind":"bracket","status":"active",
	"stop_loss_price":"60000.0000","take_profit_price":"70000.0000",
	"count":"0.00","filled_count":"0.00",
	"created_time":"2026-01-01T00:00:00Z","updated_time":"2026-01-01T00:00:01Z"
}`

const perpsAnchoredTriggerJSON = `{
	"id":"trg-2","ticker":"KXBTCPERP","kind":"bracket","status":"pending_on_entry",
	"anchor_order_id":"ord-1","client_trigger_id":"ctid-1",
	"stop_loss_price":"60000.0000",
	"count":"0.00","filled_count":"0.00",
	"created_time":"2026-01-01T00:00:00Z","updated_time":"2026-01-01T00:00:00Z"
}`

func TestPerpsExitTriggers_SetIsolated(t *testing.T) {
	client, ct := newCaptureClient(200, perpsBracketTriggerJSON)
	got, err := client.Perps.ExitTriggers.SetIsolated(context.Background(), "KXBTCPERP", &types.SetIsolatedExitTriggerRequest{
		Kind: types.ExitTriggerKindBracket, StopLossPrice: "60000.0000", TakeProfitPrice: "70000.0000",
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/margin/isolated/positions/KXBTCPERP/exit_trigger")
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"kind":"bracket","stop_loss_price":"60000.0000","take_profit_price":"70000.0000"}`)
	if got.ID != "trg-1" || got.Status != types.ExitTriggerStatusActive || got.StopLossPrice != "60000.0000" ||
		got.TakeProfitPrice != "70000.0000" || got.Count != "0.00" {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, `{
		"id":"trg-3","ticker":"KXBTCPERP","kind":"trailing","status":"active","trail_bps":250,
		"count":"0.00","filled_count":"0.00","created_time":"2026-01-01T00:00:00Z","updated_time":"2026-01-01T00:00:00Z"
	}`)
	got, err = client.Perps.ExitTriggers.SetIsolated(context.Background(), "KXBTCPERP", &types.SetIsolatedExitTriggerRequest{
		Kind: types.ExitTriggerKindTrailing, TrailBps: 250,
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertBody(t, `{"kind":"trailing","trail_bps":250}`)
	if got.Kind != types.ExitTriggerKindTrailing || got.TrailBps != 250 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsExitTriggers_GetIsolated(t *testing.T) {
	client, ct := newCaptureClient(200, `{"exit_triggers":[`+perpsBracketTriggerJSON+`]}`)
	got, err := client.Perps.ExitTriggers.GetIsolated(context.Background(), "KXBTCPERP", &types.IsolatedMarginExitTriggerOpts{Kind: types.ExitTriggerKindBracket})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/isolated/positions/KXBTCPERP/exit_trigger")
	ct.assertQuery(t, map[string]string{"kind": "bracket"})
	if len(got.ExitTriggers) != 1 || got.ExitTriggers[0].ID != "trg-1" {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, `{"exit_triggers":[]}`)
	got, err = client.Perps.ExitTriggers.GetIsolated(context.Background(), "KXBTCPERP", nil)
	if err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
	if got.ExitTriggers == nil || len(got.ExitTriggers) != 0 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsExitTriggers_DeleteIsolated(t *testing.T) {
	client, ct := newCaptureClient(204, ``)
	if err := client.Perps.ExitTriggers.DeleteIsolated(context.Background(), "KXBTCPERP", &types.IsolatedMarginExitTriggerOpts{Kind: types.ExitTriggerKindTrailing}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/isolated/positions/KXBTCPERP/exit_trigger")
	ct.assertQuery(t, map[string]string{"kind": "trailing"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}

	client, ct = newCaptureClient(204, ``)
	if err := client.Perps.ExitTriggers.DeleteIsolated(context.Background(), "KXBTCPERP", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsExitTriggers_SetCross(t *testing.T) {
	client, ct := newCaptureClient(200, perpsAnchoredTriggerJSON)
	got, err := client.Perps.ExitTriggers.SetCross(context.Background(), "KXBTCPERP", &types.SetCrossExitTriggerRequest{
		AnchorOrderID: "ord-1", ClientTriggerID: "ctid-1", StopLossPrice: "60000.0000",
	}, &types.CrossMarginExitTriggerSubaccountOpts{Subaccount: ptrOf(2)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/margin/cross/positions/KXBTCPERP/exit_trigger")
	ct.assertQuery(t, map[string]string{"subaccount": "2"})
	ct.assertBody(t, `{"anchor_order_id":"ord-1","client_trigger_id":"ctid-1","stop_loss_price":"60000.0000"}`)
	if got.ID != "trg-2" || got.Status != types.ExitTriggerStatusPendingOnEntry || got.AnchorOrderID != "ord-1" || got.ClientTriggerID != "ctid-1" {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, perpsBracketTriggerJSON)
	if _, err := client.Perps.ExitTriggers.SetCross(context.Background(), "KXBTCPERP", &types.SetCrossExitTriggerRequest{
		Count: "5", ClientTriggerID: "ctid-2", TakeProfitPrice: "70000.0000",
	}, nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"count":"5","client_trigger_id":"ctid-2","take_profit_price":"70000.0000"}`)
}

func TestPerpsExitTriggers_GetCross(t *testing.T) {
	client, ct := newCaptureClient(200, `{"exit_triggers":[`+perpsBracketTriggerJSON+`,`+perpsAnchoredTriggerJSON+`]}`)
	got, err := client.Perps.ExitTriggers.GetCross(context.Background(), "KXBTCPERP", &types.CrossMarginExitTriggerOpts{Subaccount: ptrOf(0), Kind: types.ExitTriggerKindBracket})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/cross/positions/KXBTCPERP/exit_trigger")
	ct.assertQuery(t, map[string]string{"subaccount": "0", "kind": "bracket"})
	if len(got.ExitTriggers) != 2 || got.ExitTriggers[1].AnchorOrderID != "ord-1" {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, `{"exit_triggers":[]}`)
	if _, err := client.Perps.ExitTriggers.GetCross(context.Background(), "KXBTCPERP", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsExitTriggers_DeleteCross(t *testing.T) {
	client, ct := newCaptureClient(204, ``)
	if err := client.Perps.ExitTriggers.DeleteCross(context.Background(), "KXBTCPERP", &types.CrossMarginExitTriggerOpts{Subaccount: ptrOf(3), Kind: types.ExitTriggerKindTrailing}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/cross/positions/KXBTCPERP/exit_trigger")
	ct.assertQuery(t, map[string]string{"subaccount": "3", "kind": "trailing"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}

	client, ct = newCaptureClient(204, ``)
	if err := client.Perps.ExitTriggers.DeleteCross(context.Background(), "KXBTCPERP", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsExitTriggers_UpdateCross(t *testing.T) {
	client, ct := newCaptureClient(200, perpsBracketTriggerJSON)
	got, err := client.Perps.ExitTriggers.UpdateCross(context.Background(), "KXBTCPERP", "trg-1", &types.UpdateExitTriggerRequest{
		StopLossPrice: "61000.0000",
	}, &types.CrossMarginExitTriggerSubaccountOpts{Subaccount: ptrOf(1)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/margin/cross/positions/KXBTCPERP/exit_trigger/trg-1")
	ct.assertQuery(t, map[string]string{"subaccount": "1"})
	ct.assertBody(t, `{"stop_loss_price":"61000.0000"}`)
	if got.ID != "trg-1" {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, perpsBracketTriggerJSON)
	if _, err := client.Perps.ExitTriggers.UpdateCross(context.Background(), "KXBTCPERP", "trg-1", &types.UpdateExitTriggerRequest{
		StopLossPrice: "61000.0000", TakeProfitPrice: "72000.0000",
	}, nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"stop_loss_price":"61000.0000","take_profit_price":"72000.0000"}`)
}

func TestPerpsExitTriggers_DeleteCrossByID(t *testing.T) {
	client, ct := newCaptureClient(204, ``)
	if err := client.Perps.ExitTriggers.DeleteCrossByID(context.Background(), "KXBTCPERP", "trg-1", &types.CrossMarginExitTriggerSubaccountOpts{Subaccount: ptrOf(4)}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/cross/positions/KXBTCPERP/exit_trigger/trg-1")
	ct.assertQuery(t, map[string]string{"subaccount": "4"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}

	client, ct = newCaptureClient(204, ``)
	if err := client.Perps.ExitTriggers.DeleteCrossByID(context.Background(), "KXBTCPERP", "trg-1", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsExitTriggers_NilRequest(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	if _, err := client.Perps.ExitTriggers.SetIsolated(ctx, "KXBTCPERP", nil); err == nil {
		t.Fatal("SetIsolated: expected error for nil request")
	}
	if _, err := client.Perps.ExitTriggers.SetCross(ctx, "KXBTCPERP", nil, nil); err == nil {
		t.Fatal("SetCross: expected error for nil request")
	}
	if _, err := client.Perps.ExitTriggers.UpdateCross(ctx, "KXBTCPERP", "trg-1", nil, nil); err == nil {
		t.Fatal("UpdateCross: expected error for nil request")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestPerpsExitTriggers_NoLeg(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	if _, err := client.Perps.ExitTriggers.SetIsolated(ctx, "KXBTCPERP", &types.SetIsolatedExitTriggerRequest{Kind: types.ExitTriggerKindBracket}); err == nil {
		t.Fatal("SetIsolated: expected error for a request with no leg")
	}
	if _, err := client.Perps.ExitTriggers.SetCross(ctx, "KXBTCPERP", &types.SetCrossExitTriggerRequest{Count: "1", ClientTriggerID: "ctid-1"}, nil); err == nil {
		t.Fatal("SetCross: expected error for a request with no leg")
	}
	if _, err := client.Perps.ExitTriggers.UpdateCross(ctx, "KXBTCPERP", "trg-1", &types.UpdateExitTriggerRequest{}, nil); err == nil {
		t.Fatal("UpdateCross: expected error for a request with no leg")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
	// A trailing stop needs only trail_bps.
	if _, err := client.Perps.ExitTriggers.SetIsolated(ctx, "KXBTCPERP", &types.SetIsolatedExitTriggerRequest{Kind: types.ExitTriggerKindTrailing, TrailBps: 50}); err != nil {
		t.Fatal(err)
	}
}

func TestPerpsExitTriggers_EmptyPathParam(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	s := client.Perps.ExitTriggers
	calls := map[string]func() error{
		"SetIsolated": func() error {
			_, err := s.SetIsolated(ctx, "", &types.SetIsolatedExitTriggerRequest{StopLossPrice: "1"})
			return err
		},
		"GetIsolated":    func() error { _, err := s.GetIsolated(ctx, "", nil); return err },
		"DeleteIsolated": func() error { return s.DeleteIsolated(ctx, "", nil) },
		"SetCross": func() error {
			_, err := s.SetCross(ctx, "", &types.SetCrossExitTriggerRequest{StopLossPrice: "1"}, nil)
			return err
		},
		"GetCross":    func() error { _, err := s.GetCross(ctx, "", nil); return err },
		"DeleteCross": func() error { return s.DeleteCross(ctx, "", nil) },
		"UpdateCross empty ticker": func() error {
			_, err := s.UpdateCross(ctx, "", "trg-1", &types.UpdateExitTriggerRequest{StopLossPrice: "1"}, nil)
			return err
		},
		"UpdateCross empty trigger": func() error {
			_, err := s.UpdateCross(ctx, "KXBTCPERP", "", &types.UpdateExitTriggerRequest{StopLossPrice: "1"}, nil)
			return err
		},
		"DeleteCrossByID empty ticker":  func() error { return s.DeleteCrossByID(ctx, "", "trg-1", nil) },
		"DeleteCrossByID empty trigger": func() error { return s.DeleteCrossByID(ctx, "KXBTCPERP", "", nil) },
	}
	for name, call := range calls {
		if err := call(); !errors.Is(err, ErrEmptyPathParam) {
			t.Errorf("%s: err = %v, want ErrEmptyPathParam", name, err)
		}
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

// Exit-trigger writes are PUT/DELETE and keep the idempotent retry policy: a
// replace is idempotent, and an appending SetCross carries ClientTriggerID,
// which the server replays instead of creating a second trigger.
func TestPerpsExitTriggers_RetryPolicy(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		call func(*Client) error
	}{
		{"SetCross append", func(c *Client) error {
			_, err := c.Perps.ExitTriggers.SetCross(ctx, "KXBTCPERP", &types.SetCrossExitTriggerRequest{
				Count: "1", ClientTriggerID: "ctid-1", StopLossPrice: "60000.0000",
			}, nil)
			return err
		}},
		{"SetIsolated", func(c *Client) error {
			_, err := c.Perps.ExitTriggers.SetIsolated(ctx, "KXBTCPERP", &types.SetIsolatedExitTriggerRequest{StopLossPrice: "60000.0000"})
			return err
		}},
		{"DeleteCrossByID", func(c *Client) error { return c.Perps.ExitTriggers.DeleteCrossByID(ctx, "KXBTCPERP", "trg-1", nil) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, calls := countingServer(t, 503)
			err := tc.call(client)
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != 503 {
				t.Fatalf("err = %v, want *APIError 503", err)
			}
			if got := atomic.LoadInt32(calls); got != 3 {
				t.Fatalf("attempts = %d, want 3", got)
			}
		})
	}
}
