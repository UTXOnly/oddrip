package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

// The perps AsyncAPI has an example only for order_group_updates; the other
// payloads here are built from each schema's property list.

// perpsWSRoundTrip decodes payload into v, re-encodes it, and checks the result
// carries exactly the payload's keys and values, so a field the struct drops
// or renames fails the test.
func perpsWSRoundTrip(t *testing.T, payload string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(payload), v); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var want, got map[string]any
	if err := json.Unmarshal([]byte(payload), &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip:\n got %s\nwant %s", out, payload)
	}
}

func TestPerpsWS_OrderbookSnapshotMsg(t *testing.T) {
	const frame = `{"type":"orderbook_snapshot","id":4,"sid":2,"seq":1,"msg":{"market_ticker":"KXBTCPERP","bid":[["64990.0000","1.50"],["64980.0000","3.00"]],"ask":[["65010.0000","2.00"]]}}`
	var m WSMessage
	if err := json.Unmarshal([]byte(frame), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != WSTypeOrderbookSnapshot || m.SID != 2 || m.Seq != 1 {
		t.Fatalf("envelope: %+v", m)
	}
	var msg MarginOrderbookSnapshotMsg
	if err := m.Decode(&msg); err != nil {
		t.Fatal(err)
	}
	if msg.MarketTicker != "KXBTCPERP" || len(msg.Bid) != 2 || len(msg.Ask) != 1 {
		t.Fatalf("snapshot: %+v", msg)
	}
	if msg.Bid[1] != (OrderbookLevel{PriceDollars: "64980.0000", CountFp: "3.00"}) || msg.Ask[0] != (OrderbookLevel{PriceDollars: "65010.0000", CountFp: "2.00"}) {
		t.Fatalf("levels: bid=%+v ask=%+v", msg.Bid, msg.Ask)
	}
	perpsWSRoundTrip(t, string(m.Msg), &MarginOrderbookSnapshotMsg{})

	var oneSided MarginOrderbookSnapshotMsg
	perpsWSRoundTrip(t, `{"market_ticker":"KXBTCPERP","ask":[["65010.0000","2.00"]]}`, &oneSided)
	if oneSided.Bid != nil || len(oneSided.Ask) != 1 {
		t.Fatalf("one-sided: %+v", oneSided)
	}
}

func TestPerpsWS_OrderbookDeltaMsg(t *testing.T) {
	const payload = `{
		"market_ticker": "KXBTCPERP",
		"price": "64990.0000",
		"delta": "-1.50",
		"side": "bid",
		"last_update_reason": "Trade",
		"client_order_id": "c-1",
		"ts_ms": 1700000000123,
		"subaccount": 0
	}`
	var msg MarginOrderbookDeltaMsg
	perpsWSRoundTrip(t, payload, &msg)
	if msg.Price != "64990.0000" || msg.Delta != "-1.50" || msg.Side != BookSideBid || msg.LastUpdateReason != MarginLastUpdateReasonTrade || msg.ClientOrderID != "c-1" {
		t.Fatalf("delta: %+v", msg)
	}
	if msg.TsMs == nil || *msg.TsMs != 1700000000123 || msg.Subaccount == nil || *msg.Subaccount != 0 {
		t.Fatalf("delta: %+v", msg)
	}

	var bare MarginOrderbookDeltaMsg
	perpsWSRoundTrip(t, `{"market_ticker":"KXBTCPERP","price":"65010.0000","delta":"2.00","side":"ask"}`, &bare)
	if bare.Side != BookSideAsk || bare.LastUpdateReason != "" || bare.ClientOrderID != "" || bare.TsMs != nil || bare.Subaccount != nil {
		t.Fatalf("optional fields should be unset: %+v", bare)
	}
}

func TestPerpsWS_TickerMsg(t *testing.T) {
	const payload = `{
		"market_ticker": "KXBTCPERP",
		"price": "65000.0000",
		"bid": "64990.0000",
		"ask": "65010.0000",
		"bid_size_fp": "1.50",
		"ask_size_fp": "2.00",
		"last_trade_size_fp": "0.25",
		"volume": "1200.00",
		"volume_notional_value_dollars": "78000000.0000",
		"volume_24h": "300.00",
		"volume_24h_notional_value_dollars": "19500000.0000",
		"open_interest": "450.00",
		"open_interest_notional_value_dollars": "29250000.0000",
		"reference_price": {"price": "65002.1234", "ts_ms": 1700000000100},
		"settlement_mark_price": {"price": "65001.0000", "ts_ms": 1700000000110},
		"liquidation_mark_price": {"price": "65000.5000", "ts_ms": 1700000000120},
		"funding_rate": {"rate": -0.000125, "next_funding_time_ms": 1700003600000, "ts_ms": 1700000000000},
		"ts_ms": 1700000000123
	}`
	var msg MarginTickerMsg
	perpsWSRoundTrip(t, payload, &msg)
	if msg.Price != "65000.0000" || msg.Bid != "64990.0000" || msg.Ask != "65010.0000" || msg.BidSizeFp != "1.50" || msg.AskSizeFp != "2.00" || msg.LastTradeSizeFp != "0.25" {
		t.Fatalf("ticker: %+v", msg)
	}
	if msg.Volume != "1200.00" || msg.Volume24h != "300.00" || msg.Volume24hNotionalValueDollars != "19500000.0000" || msg.OpenInterest != "450.00" || msg.OpenInterestNotionalValueDollars != "29250000.0000" || msg.TsMs != 1700000000123 {
		t.Fatalf("ticker: %+v", msg)
	}
	if msg.ReferencePrice == nil || *msg.ReferencePrice != (TickerPrice{Price: "65002.1234", TsMs: 1700000000100}) {
		t.Fatalf("reference_price: %+v", msg.ReferencePrice)
	}
	if msg.SettlementMarkPrice == nil || msg.SettlementMarkPrice.Price != "65001.0000" || msg.LiquidationMarkPrice == nil || msg.LiquidationMarkPrice.TsMs != 1700000000120 {
		t.Fatalf("mark prices: %+v %+v", msg.SettlementMarkPrice, msg.LiquidationMarkPrice)
	}
	if msg.FundingRate == nil || *msg.FundingRate != (MarginTickerFundingRate{Rate: -0.000125, NextFundingTimeMs: 1700003600000, TsMs: 1700000000000}) {
		t.Fatalf("funding_rate: %+v", msg.FundingRate)
	}

	const required = `{"market_ticker":"KXBTCPERP","price":"65000.0000","bid":"64990.0000","ask":"65010.0000","bid_size_fp":"1.50","ask_size_fp":"2.00","last_trade_size_fp":"0.25","volume":"1200.00","volume_notional_value_dollars":"78000000.0000","volume_24h":"300.00","volume_24h_notional_value_dollars":"19500000.0000","open_interest":"450.00","open_interest_notional_value_dollars":"29250000.0000","ts_ms":1700000000123}`
	var bare MarginTickerMsg
	perpsWSRoundTrip(t, required, &bare)
	if bare.ReferencePrice != nil || bare.SettlementMarkPrice != nil || bare.LiquidationMarkPrice != nil || bare.FundingRate != nil {
		t.Fatalf("optional prices should be nil: %+v", bare)
	}
}

func TestPerpsWS_TradeMsg(t *testing.T) {
	const payload = `{"trade_id":"3f1c2a9e-8d4b-4c1a-9e7f-2b6d5a4c3e21","market_ticker":"KXBTCPERP","price":"65000.0000","count":"0.25","taker_side":"ask","ts_ms":1700000000123}`
	var msg MarginTradeMsg
	perpsWSRoundTrip(t, payload, &msg)
	if msg.TradeID != "3f1c2a9e-8d4b-4c1a-9e7f-2b6d5a4c3e21" || msg.Price != "65000.0000" || msg.Count != "0.25" || msg.TakerSide != BookSideAsk || msg.TsMs != 1700000000123 {
		t.Fatalf("trade: %+v", msg)
	}
}

func TestPerpsWS_FillMsg(t *testing.T) {
	const payload = `{
		"trade_id": "3f1c2a9e-8d4b-4c1a-9e7f-2b6d5a4c3e21",
		"order_id": "7a6b5c4d-3e2f-4a1b-8c9d-0e1f2a3b4c5d",
		"client_order_id": "c-1",
		"market_ticker": "KXBTCPERP",
		"is_taker": true,
		"side": "bid",
		"ts_ms": 1700000000123,
		"price": "65000.0000",
		"count": "0.25",
		"fee_cost": "0.1625",
		"post_position": "1.25",
		"subaccount": 2,
		"order_source": "system"
	}`
	var msg MarginFillMsg
	perpsWSRoundTrip(t, payload, &msg)
	if msg.OrderID != "7a6b5c4d-3e2f-4a1b-8c9d-0e1f2a3b4c5d" || msg.ClientOrderID != "c-1" || !msg.IsTaker || msg.Side != BookSideBid || msg.TsMs != 1700000000123 {
		t.Fatalf("fill: %+v", msg)
	}
	if msg.Price != "65000.0000" || msg.Count != "0.25" || msg.FeeCost != "0.1625" || msg.PostPosition != "1.25" || msg.Subaccount == nil || *msg.Subaccount != 2 || msg.OrderSource != OrderSourceSystem {
		t.Fatalf("fill: %+v", msg)
	}

	const required = `{"trade_id":"3f1c2a9e-8d4b-4c1a-9e7f-2b6d5a4c3e21","order_id":"7a6b5c4d-3e2f-4a1b-8c9d-0e1f2a3b4c5d","market_ticker":"KXBTCPERP","is_taker":false,"side":"ask","ts_ms":1700000000123,"price":"65000.0000","count":"0.25","fee_cost":"0.0000","post_position":"-0.25","order_source":"user"}`
	var bare MarginFillMsg
	perpsWSRoundTrip(t, required, &bare)
	if bare.IsTaker || bare.ClientOrderID != "" || bare.Subaccount != nil || bare.OrderSource != OrderSourceUser {
		t.Fatalf("fill without optional fields: %+v", bare)
	}
}

func TestPerpsWS_UserOrderMsg(t *testing.T) {
	const payload = `{
		"order_id": "7a6b5c4d-3e2f-4a1b-8c9d-0e1f2a3b4c5d",
		"user_id": "1b2c3d4e-5f6a-4b7c-8d9e-0f1a2b3c4d5e",
		"client_order_id": "c-1",
		"ticker": "KXBTCPERP",
		"side": "bid",
		"price": "64990.0000",
		"fill_count": "0.50",
		"remaining_count": "1.00",
		"self_trade_prevention_type": "taker_at_cross",
		"order_group_id": "og_123",
		"expiration_ts_ms": 1700003600000,
		"created_ts_ms": 1700000000000,
		"last_updated_ts_ms": 1700000000123,
		"last_update_reason": "ReduceOnlyCancel",
		"subaccount_number": 0,
		"order_source": "user"
	}`
	var msg MarginUserOrderMsg
	perpsWSRoundTrip(t, payload, &msg)
	if msg.OrderID != "7a6b5c4d-3e2f-4a1b-8c9d-0e1f2a3b4c5d" || msg.UserID != "1b2c3d4e-5f6a-4b7c-8d9e-0f1a2b3c4d5e" || msg.ClientOrderID != "c-1" || msg.Ticker != "KXBTCPERP" || msg.Side != BookSideBid {
		t.Fatalf("user order: %+v", msg)
	}
	if msg.Price != "64990.0000" || msg.FillCount != "0.50" || msg.RemainingCount != "1.00" || msg.SelfTradePreventionType != SelfTradeTakerAtCross || msg.OrderGroupID != "og_123" {
		t.Fatalf("user order: %+v", msg)
	}
	if msg.ExpirationTsMs == nil || *msg.ExpirationTsMs != 1700003600000 || msg.CreatedTsMs == nil || *msg.CreatedTsMs != 1700000000000 || msg.LastUpdatedTsMs == nil || *msg.LastUpdatedTsMs != 1700000000123 {
		t.Fatalf("user order timestamps: %+v", msg)
	}
	if msg.LastUpdateReason != MarginLastUpdateReasonReduceOnlyCancel || msg.SubaccountNumber == nil || *msg.SubaccountNumber != 0 || msg.OrderSource != OrderSourceUser {
		t.Fatalf("user order: %+v", msg)
	}

	// created_ts_ms is required but nullable; optional fields are absent.
	const nullCreated = `{"order_id":"7a6b5c4d-3e2f-4a1b-8c9d-0e1f2a3b4c5d","user_id":"1b2c3d4e-5f6a-4b7c-8d9e-0f1a2b3c4d5e","client_order_id":"","ticker":"KXBTCPERP","side":"ask","price":"65010.0000","fill_count":"0.00","remaining_count":"2.00","created_ts_ms":null,"order_source":"system"}`
	var bare MarginUserOrderMsg
	if err := json.Unmarshal([]byte(nullCreated), &bare); err != nil {
		t.Fatal(err)
	}
	if bare.CreatedTsMs != nil || bare.ExpirationTsMs != nil || bare.LastUpdatedTsMs != nil || bare.SubaccountNumber != nil || bare.LastUpdateReason != "" || bare.OrderSource != OrderSourceSystem {
		t.Fatalf("user order without optional fields: %+v", bare)
	}
}

// The perps order_group_updates payload matches the event-contract one field
// for field, so it decodes into OrderGroupUpdatesMsg. Frame is the
// orderGroupLimitUpdated example from perps_asyncapi.yaml.
func TestPerpsWS_OrderGroupUpdatesMsg_SpecExample(t *testing.T) {
	const frame = `{"type":"order_group_updates","sid":21,"seq":7,"msg":{"event_type":"limit_updated","order_group_id":"og_123","contracts_limit_fp":"150.00","ts_ms":1700000000123}}`
	var m WSMessage
	if err := json.Unmarshal([]byte(frame), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != WSTypeOrderGroupUpdates || m.SID != 21 || m.Seq != 7 {
		t.Fatalf("envelope: %+v", m)
	}
	var msg OrderGroupUpdatesMsg
	if err := m.Decode(&msg); err != nil {
		t.Fatal(err)
	}
	if msg != (OrderGroupUpdatesMsg{EventType: "limit_updated", OrderGroupID: "og_123", ContractsLimitFp: "150.00", TsMs: 1700000000123}) {
		t.Fatalf("msg: %+v", msg)
	}
	perpsWSRoundTrip(t, string(m.Msg), &OrderGroupUpdatesMsg{})
}

func TestPerpsWS_MarginErrorMsg(t *testing.T) {
	const frame = `{"type":"error","id":9,"sid":3,"seq":4,"msg":{"code":26,"msg":"Subscription market limit exceeded","market_ticker":"KXBTCPERP","market_tickers":["KXBTCPERP","KXETHPERP"]}}`
	var m WSMessage
	if err := json.Unmarshal([]byte(frame), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != WSTypeError || m.SID != 3 || m.Seq != 4 {
		t.Fatalf("envelope: %+v", m)
	}
	var e MarginErrorMsg
	if err := m.Decode(&e); err != nil {
		t.Fatal(err)
	}
	if e.Code != 26 || e.Msg != "Subscription market limit exceeded" || e.MarketTicker != "KXBTCPERP" || !reflect.DeepEqual(e.MarketTickers, []string{"KXBTCPERP", "KXETHPERP"}) {
		t.Fatalf("error msg: %+v", e)
	}
	perpsWSRoundTrip(t, string(m.Msg), &MarginErrorMsg{})

	b, err := json.Marshal(MarginErrorMsg{Code: 7, Msg: "Unknown subscription ID"})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"code":7,"msg":"Unknown subscription ID"}` {
		t.Fatalf("marshal without tickers: %s", b)
	}
}

// Perps commands are built from the shared SubscribeParams and
// UpdateSubscriptionParams; unset event-contract-only fields stay off the wire.
func TestPerpsWS_CommandMarshal(t *testing.T) {
	yes := true
	b, err := json.Marshal(SubscribeCommand{ID: 1, Cmd: "subscribe", Params: SubscribeParams{
		Channels:            []string{WSChannelTicker, WSChannelOrderbookDelta},
		MarketTickers:       []string{"KXBTCPERP"},
		SendInitialSnapshot: &yes,
		SkipTickerAck:       &yes,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"id":1,"cmd":"subscribe","params":{"channels":["ticker","orderbook_delta"],"market_tickers":["KXBTCPERP"],"send_initial_snapshot":true,"skip_ticker_ack":true}}`; string(b) != want {
		t.Fatalf("subscribe:\n got %s\nwant %s", b, want)
	}

	sid := 3
	b, err = json.Marshal(UpdateSubscriptionCommand{ID: 2, Cmd: "update_subscription", Params: UpdateSubscriptionParams{
		SID:           &sid,
		MarketTickers: []string{"KXBTCPERP"},
		Action:        WSUpdateSubscriptionGetSnapshot,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"id":2,"cmd":"update_subscription","params":{"sid":3,"market_tickers":["KXBTCPERP"],"action":"get_snapshot"}}`; string(b) != want {
		t.Fatalf("update_subscription:\n got %s\nwant %s", b, want)
	}
}

// Every channel and message type in perps_asyncapi.yaml already has a
// constant; perps reuses them.
func TestPerpsWS_ChannelAndTypeConstants(t *testing.T) {
	for _, tc := range []struct{ got, want string }{
		{WSChannelOrderbookDelta, "orderbook_delta"},
		{WSChannelTicker, "ticker"},
		{WSChannelTrade, "trade"},
		{WSChannelFill, "fill"},
		{WSChannelUserOrders, "user_orders"},
		{WSChannelOrderGroup, "order_group_updates"},
		{WSTypeSubscribed, "subscribed"},
		{WSTypeUnsubscribed, "unsubscribed"},
		{WSTypeOK, "ok"},
		{WSTypeError, "error"},
		{WSTypeOrderbookSnapshot, "orderbook_snapshot"},
		{WSTypeOrderbookDelta, "orderbook_delta"},
		{WSTypeTicker, "ticker"},
		{WSTypeTrade, "trade"},
		{WSTypeFill, "fill"},
		{WSTypeUserOrder, "user_order"},
		{WSTypeOrderGroupUpdates, "order_group_updates"},
		{WSUpdateSubscriptionAddMarkets, "add_markets"},
		{WSUpdateSubscriptionDeleteMarkets, "delete_markets"},
		{WSUpdateSubscriptionGetSnapshot, "get_snapshot"},
	} {
		if tc.got != tc.want {
			t.Errorf("got %q want %q", tc.got, tc.want)
		}
	}
}
