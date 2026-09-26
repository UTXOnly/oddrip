package types

import (
	"encoding/json"
	"testing"
	"time"
)

// The perps spec has no examples for these schemas; payloads follow the
// property lists, with null for the nullable fields.

func TestPerpsOrderTypes_MarginOrder(t *testing.T) {
	const payload = `{
		"order_id": "o1",
		"user_id": "u1",
		"client_order_id": "c1",
		"ticker": "KXBTCPERP",
		"side": "ask",
		"last_update_reason": "Amend",
		"price": "105.2500",
		"fill_count": "1.00",
		"remaining_count": "2.00",
		"expiration_time": null,
		"created_time": "2026-01-02T03:04:05Z",
		"last_update_time": "2026-01-02T03:04:06.5Z",
		"self_trade_prevention_type": null,
		"cancel_order_on_pause": false,
		"order_group_id": "og1",
		"order_source": "system",
		"order_reason": "take_profit_stop_loss"
	}`
	var o MarginOrder
	if err := json.Unmarshal([]byte(payload), &o); err != nil {
		t.Fatal(err)
	}
	if o.OrderID != "o1" || o.UserID != "u1" || o.ClientOrderID != "c1" || o.Ticker != "KXBTCPERP" || o.Side != BookSideAsk {
		t.Fatalf("ids: %+v", o)
	}
	if o.LastUpdateReason != MarginLastUpdateReasonAmend || o.Price != "105.2500" || o.FillCount != "1.00" || o.RemainingCount != "2.00" {
		t.Fatalf("state: %+v", o)
	}
	if o.ExpirationTime != nil || o.SelfTradePreventionType != nil {
		t.Fatalf("nullable fields not nil: %+v", o)
	}
	if o.CreatedTime == nil || o.LastUpdateTime == nil {
		t.Fatalf("timestamps: %+v", o)
	}
	ts, err := ParseTime(*o.LastUpdateTime)
	if err != nil || !ts.Equal(time.Date(2026, 1, 2, 3, 4, 6, 500_000_000, time.UTC)) {
		t.Fatalf("last_update_time: %v, %v", ts, err)
	}
	if o.CancelOrderOnPause == nil || *o.CancelOrderOnPause {
		t.Fatalf("cancel_order_on_pause: %v", o.CancelOrderOnPause)
	}
	if o.OrderGroupID != "og1" || o.OrderSource != OrderSourceSystem || o.OrderReason != OrderReasonTakeProfitStopLoss {
		t.Fatalf("source: %+v", o)
	}

	// Required fields only; the empty last_update_reason is a spec enum value.
	const minimal = `{"order_id":"o2","user_id":"u1","client_order_id":"","ticker":"KXBTCPERP","side":"bid",
		"last_update_reason":"","price":"1.00","fill_count":"0.00","remaining_count":"1.00",
		"self_trade_prevention_type":"maker"}`
	var m MarginOrder
	if err := json.Unmarshal([]byte(minimal), &m); err != nil {
		t.Fatal(err)
	}
	if m.LastUpdateReason != "" || m.CreatedTime != nil || m.CancelOrderOnPause != nil || m.OrderSource != "" {
		t.Fatalf("minimal: %+v", m)
	}
	if m.SelfTradePreventionType == nil || *m.SelfTradePreventionType != SelfTradeMaker {
		t.Fatalf("self_trade_prevention_type: %v", m.SelfTradePreventionType)
	}
}

func TestPerpsOrderTypes_GetMarginOrdersResponse(t *testing.T) {
	const payload = `{"orders":[{"order_id":"o1","user_id":"u1","client_order_id":"c1","ticker":"KXBTCPERP","side":"bid",
		"last_update_reason":"Trade","price":"1.00","fill_count":"1.00","remaining_count":"0.00"}],"cursor":"abc"}`
	var out GetMarginOrdersResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.Cursor != "abc" || len(out.Orders) != 1 || out.Orders[0].LastUpdateReason != MarginLastUpdateReasonTrade {
		t.Fatalf("unexpected: %+v", out)
	}

	var one GetMarginOrderResponse
	if err := json.Unmarshal([]byte(`{"order":{"order_id":"o1","ticker":"KXBTCPERP"}}`), &one); err != nil {
		t.Fatal(err)
	}
	if one.Order.OrderID != "o1" || one.Order.Ticker != "KXBTCPERP" {
		t.Fatalf("unexpected: %+v", one)
	}
}

func TestPerpsOrderTypes_CreateMarginOrderRequest(t *testing.T) {
	req := CreateMarginOrderRequest{
		Ticker: "KXBTCPERP", ClientOrderID: "cid", Side: BookSideBid, Count: "10", Price: "0.5600",
		TimeInForce: TimeInForceIOC, SelfTradePreventionType: SelfTradeTakerAtCross,
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"ticker":"KXBTCPERP","client_order_id":"cid","side":"bid","count":"10","price":"0.5600","time_in_force":"immediate_or_cancel","self_trade_prevention_type":"taker_at_cross"}`
	if string(b) != want {
		t.Fatalf("required only:\n got %s\nwant %s", b, want)
	}

	f := false
	sub := 0
	req.ExpirationTime = 1700000000
	req.PostOnly = &f
	req.CancelOrderOnPause = &f
	req.ReduceOnly = &f
	req.Subaccount = &sub
	req.OrderGroupID = "og1"
	b, err = json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	const wantAll = `{"ticker":"KXBTCPERP","client_order_id":"cid","side":"bid","count":"10","price":"0.5600","time_in_force":"immediate_or_cancel","self_trade_prevention_type":"taker_at_cross","expiration_time":1700000000,"post_only":false,"cancel_order_on_pause":false,"reduce_only":false,"subaccount":0,"order_group_id":"og1"}`
	if string(b) != wantAll {
		t.Fatalf("all fields:\n got %s\nwant %s", b, wantAll)
	}
}

func TestPerpsOrderTypes_CreateMarginOrderResponse(t *testing.T) {
	const payload = `{"order_id":"o1","client_order_id":"cid","fill_count":"4.00","remaining_count":"6.00","average_fill_price":"0.5600","average_fee_paid":"0.0070"}`
	var out CreateMarginOrderResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.OrderID != "o1" || out.ClientOrderID != "cid" || out.FillCount != "4.00" || out.RemainingCount != "6.00" ||
		out.AverageFillPrice != "0.5600" || out.AverageFeePaid != "0.0070" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestPerpsOrderTypes_CancelMarginOrderResponse(t *testing.T) {
	var out CancelMarginOrderResponse
	if err := json.Unmarshal([]byte(`{"order_id":"o1","client_order_id":"c1","reduced_by":"3.00"}`), &out); err != nil {
		t.Fatal(err)
	}
	if out.OrderID != "o1" || out.ClientOrderID != "c1" || out.ReducedBy != "3.00" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestPerpsOrderTypes_DecreaseMarginOrder(t *testing.T) {
	by := "2.50"
	b, err := json.Marshal(DecreaseMarginOrderRequest{ReduceBy: &by})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"reduce_by":"2.50"}` {
		t.Fatalf("marshal: %s", b)
	}

	var out DecreaseMarginOrderResponse
	if err := json.Unmarshal([]byte(`{"order_id":"o1","client_order_id":"c1","remaining_count":"7.50"}`), &out); err != nil {
		t.Fatal(err)
	}
	if out.OrderID != "o1" || out.ClientOrderID != "c1" || out.RemainingCount != "7.50" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestPerpsOrderTypes_AmendMarginOrder(t *testing.T) {
	b, err := json.Marshal(AmendMarginOrderRequest{Ticker: "KXBTCPERP", Side: BookSideAsk, Price: "0.5700", Count: "5"})
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"ticker":"KXBTCPERP","side":"ask","price":"0.5700","count":"5"}`
	if string(b) != want {
		t.Fatalf("marshal:\n got %s\nwant %s", b, want)
	}

	var out AmendMarginOrderResponse
	const payload = `{"order_id":"o1","client_order_id":"c2","remaining_count":"3.00","fill_count":"2.00","average_fill_price":"0.5700","average_fee_paid":null}`
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.OrderID != "o1" || out.ClientOrderID != "c2" || out.RemainingCount == nil || *out.RemainingCount != "3.00" ||
		out.FillCount == nil || *out.FillCount != "2.00" || out.AverageFillPrice == nil || *out.AverageFillPrice != "0.5700" ||
		out.AverageFeePaid != nil {
		t.Fatalf("unexpected: %+v", out)
	}
}

// Perps order groups reuse the event-contract order-group types; these
// payloads follow the perps schemas.
func TestPerpsOrderTypes_OrderGroupReuse(t *testing.T) {
	var list GetOrderGroupsResponse
	if err := json.Unmarshal([]byte(`{"order_groups":[{"id":"og1","contracts_limit_fp":"10.00","is_auto_cancel_enabled":true,"exchange_index":0}]}`), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.OrderGroups) != 1 || list.OrderGroups[0].ID != "og1" || list.OrderGroups[0].ContractsLimitFp != "10.00" || !list.OrderGroups[0].IsAutoCancelEnabled {
		t.Fatalf("list: %+v", list)
	}

	var one GetOrderGroupResponse
	if err := json.Unmarshal([]byte(`{"is_auto_cancel_enabled":false,"contracts_limit_fp":"5.00","orders":["o1"],"exchange_index":0}`), &one); err != nil {
		t.Fatal(err)
	}
	if one.IsAutoCancelEnabled || one.ContractsLimitFp != "5.00" || len(one.Orders) != 1 {
		t.Fatalf("get: %+v", one)
	}

	var created CreateOrderGroupResponse
	if err := json.Unmarshal([]byte(`{"order_group_id":"og1","subaccount":0}`), &created); err != nil {
		t.Fatal(err)
	}
	if created.OrderGroupID != "og1" || created.Subaccount != 0 {
		t.Fatalf("create: %+v", created)
	}

	fp := "10.00"
	b, err := json.Marshal(CreateOrderGroupRequest{ContractsLimitFp: &fp})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"contracts_limit_fp":"10.00"}` {
		t.Fatalf("create request: %s", b)
	}
	lim := int64(3)
	b, err = json.Marshal(UpdateOrderGroupLimitRequest{ContractsLimit: &lim})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"contracts_limit":3}` {
		t.Fatalf("limit request: %s", b)
	}
}
