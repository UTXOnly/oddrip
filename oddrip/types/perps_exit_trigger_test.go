package types

import (
	"encoding/json"
	"testing"
)

func TestPerpsExitTriggerTypes_Constants(t *testing.T) {
	cases := []struct{ got, want string }{
		{ExitTriggerKindBracket, "bracket"},
		{ExitTriggerKindTrailing, "trailing"},
		{ExitTriggerStatusPendingOnEntry, "pending_on_entry"},
		{ExitTriggerStatusActive, "active"},
		{ExitTriggerStatusFilled, "filled"},
		{ExitTriggerStatusFailed, "failed"},
		{ExitTriggerStatusCanceled, "canceled"},
		{ExitTriggerStatusUnknown, "unknown"},
		{ExitTriggerLegStopLoss, "stop_loss"},
		{ExitTriggerLegTakeProfit, "take_profit"},
		{ExitTriggerReasonUserCanceled, "user_canceled"},
		{ExitTriggerReasonPositionClosed, "position_closed"},
		{ExitTriggerReasonEntryNotFilled, "entry_not_filled"},
		{ExitTriggerReasonOrderRejected, "order_rejected"},
		{ExitTriggerReasonPositionFlipped, "position_flipped"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("constant %q, want %q", c.got, c.want)
		}
	}
}

// Every ExitTrigger property from the spec. The spec has no example payload,
// so this is built from the schema's property list.
func TestPerpsExitTriggerTypes_ExitTriggerUnmarshal(t *testing.T) {
	const bracket = `{
		"id":"trg-1","ticker":"KXBTCPERP","kind":"bracket","status":"filled",
		"status_reason":"position_closed","triggered_leg":"stop_loss","triggered_order_id":"ord-9",
		"anchor_order_id":"ord-1","client_trigger_id":"ctid-1",
		"stop_loss_price":"60000.0000","take_profit_price":"70000.0000",
		"count":"2.00","filled_count":"2.00",
		"created_time":"2026-01-01T00:00:00Z","updated_time":"2026-01-01T00:05:00Z"
	}`
	var b ExitTrigger
	if err := json.Unmarshal([]byte(bracket), &b); err != nil {
		t.Fatal(err)
	}
	if b.ID != "trg-1" || b.Ticker != "KXBTCPERP" || b.Kind != ExitTriggerKindBracket || b.Status != ExitTriggerStatusFilled ||
		b.StatusReason != ExitTriggerReasonPositionClosed || b.TriggeredLeg != ExitTriggerLegStopLoss ||
		b.TriggeredOrderID != "ord-9" || b.AnchorOrderID != "ord-1" || b.ClientTriggerID != "ctid-1" ||
		b.StopLossPrice != "60000.0000" || b.TakeProfitPrice != "70000.0000" ||
		b.Count != "2.00" || b.FilledCount != "2.00" ||
		b.CreatedTime != "2026-01-01T00:00:00Z" || b.UpdatedTime != "2026-01-01T00:05:00Z" {
		t.Fatalf("bracket: %+v", b)
	}
	if b.TrailBps != 0 || b.TrailAmount != "" || b.WatermarkPrice != "" || b.EffectiveStopPrice != "" {
		t.Fatalf("bracket has trailing fields: %+v", b)
	}
	if _, err := ParseTime(b.UpdatedTime); err != nil {
		t.Fatalf("updated_time: %v", err)
	}

	const trailing = `{"exit_triggers":[{
		"id":"trg-2","ticker":"KXBTCPERP","kind":"trailing","status":"active",
		"trail_amount":"500.0000","trail_bps":150,"watermark_price":"65000.0000","effective_stop_price":"64500.0000",
		"count":"0.00","filled_count":"0.00",
		"created_time":"2026-01-01T00:00:00Z","updated_time":"2026-01-01T00:00:00Z"
	}]}`
	var resp GetExitTriggersResponse
	if err := json.Unmarshal([]byte(trailing), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.ExitTriggers) != 1 {
		t.Fatalf("resp: %+v", resp)
	}
	tr := resp.ExitTriggers[0]
	if tr.Kind != ExitTriggerKindTrailing || tr.Status != ExitTriggerStatusActive || tr.TrailAmount != "500.0000" ||
		tr.TrailBps != 150 || tr.WatermarkPrice != "65000.0000" || tr.EffectiveStopPrice != "64500.0000" ||
		tr.Count != "0.00" || tr.StopLossPrice != "" || tr.StatusReason != "" || tr.TriggeredLeg != "" {
		t.Fatalf("trailing: %+v", tr)
	}

	var empty GetExitTriggersResponse
	if err := json.Unmarshal([]byte(`{"exit_triggers":[]}`), &empty); err != nil {
		t.Fatal(err)
	}
	if empty.ExitTriggers == nil || len(empty.ExitTriggers) != 0 {
		t.Fatalf("empty: %+v", empty)
	}
}

func TestPerpsExitTriggerTypes_RequestMarshal(t *testing.T) {
	bps := 250
	cases := []struct {
		name string
		v    any
		want string
	}{
		{"isolated empty", SetIsolatedExitTriggerRequest{}, `{}`},
		{"isolated stop-loss only", SetIsolatedExitTriggerRequest{StopLossPrice: "60000.0000"}, `{"stop_loss_price":"60000.0000"}`},
		{"isolated bracket", SetIsolatedExitTriggerRequest{Kind: ExitTriggerKindBracket, StopLossPrice: "60000.0000", TakeProfitPrice: "70000.0000"},
			`{"kind":"bracket","stop_loss_price":"60000.0000","take_profit_price":"70000.0000"}`},
		{"isolated trailing bps", SetIsolatedExitTriggerRequest{Kind: ExitTriggerKindTrailing, TrailBps: bps},
			`{"kind":"trailing","trail_bps":250}`},
		{"isolated trailing amount", SetIsolatedExitTriggerRequest{Kind: ExitTriggerKindTrailing, TrailAmount: "500.00"},
			`{"kind":"trailing","trail_amount":"500.00"}`},
		{"cross empty", SetCrossExitTriggerRequest{}, `{}`},
		{"cross replace", SetCrossExitTriggerRequest{TakeProfitPrice: "70000.0000"}, `{"take_profit_price":"70000.0000"}`},
		{"cross append count", SetCrossExitTriggerRequest{Count: "3", ClientTriggerID: "ctid-1", StopLossPrice: "60000.0000"},
			`{"count":"3","client_trigger_id":"ctid-1","stop_loss_price":"60000.0000"}`},
		{"cross append anchored", SetCrossExitTriggerRequest{AnchorOrderID: "ord-1", ClientTriggerID: "ctid-2", Kind: ExitTriggerKindBracket, StopLossPrice: "60000.0000", TakeProfitPrice: "70000.0000"},
			`{"anchor_order_id":"ord-1","client_trigger_id":"ctid-2","kind":"bracket","stop_loss_price":"60000.0000","take_profit_price":"70000.0000"}`},
		{"cross trailing", SetCrossExitTriggerRequest{Kind: ExitTriggerKindTrailing, TrailBps: bps}, `{"kind":"trailing","trail_bps":250}`},
		{"update both legs", UpdateExitTriggerRequest{StopLossPrice: "61000.0000", TakeProfitPrice: "72000.0000"},
			`{"stop_loss_price":"61000.0000","take_profit_price":"72000.0000"}`},
		{"update one leg", UpdateExitTriggerRequest{TakeProfitPrice: "72000.0000"}, `{"take_profit_price":"72000.0000"}`},
	}
	for _, tc := range cases {
		b, err := json.Marshal(tc.v)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if string(b) != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, b, tc.want)
		}
	}
}
