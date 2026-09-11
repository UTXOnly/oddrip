package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWSMessage_Decode(t *testing.T) {
	const envelope = `{"type":"ticker","sid":11,"msg":{"market_ticker":"FED-23DEC-T3.00","market_id":"9b0f6b43-5b68-4f9f-9f02-9a2d1b8ac1a1","price_dollars":"0.480","yes_bid_dollars":"0.450","yes_ask_dollars":"0.530","volume_fp":"33896.00","open_interest_fp":"20422.00","dollar_volume":16948,"dollar_open_interest":10211,"yes_bid_size_fp":"300.00","yes_ask_size_fp":"150.00","last_trade_size_fp":"25.00","ts":1669149841,"ts_ms":1669149841000,"time":"2022-11-22T20:44:01Z"}}`
	var m WSMessage
	if err := json.Unmarshal([]byte(envelope), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != WSTypeTicker || m.SID != 11 {
		t.Fatalf("envelope: %+v", m)
	}
	var tick TickerMsg
	if err := m.Decode(&tick); err != nil {
		t.Fatal(err)
	}
	if tick.MarketTicker != "FED-23DEC-T3.00" || tick.YesBidDollars != "0.450" || tick.DollarVolume != 16948 || tick.TsMs != 1669149841000 || tick.Time != "2022-11-22T20:44:01Z" {
		t.Fatalf("ticker: %+v", tick)
	}
}

func TestWSMessage_Decode_EmptyMsg(t *testing.T) {
	m := WSMessage{Type: WSTypeUnsubscribed, SID: 2}
	var v struct{}
	err := m.Decode(&v)
	if err == nil || !strings.Contains(err.Error(), "unsubscribed") {
		t.Fatalf("want error naming the type, got %v", err)
	}
}

func TestWSMessage_Decode_ErrorMsg(t *testing.T) {
	const envelope = `{"id":123,"type":"error","msg":{"code":7,"msg":"Unknown subscription ID"}}`
	var m WSMessage
	if err := json.Unmarshal([]byte(envelope), &m); err != nil {
		t.Fatal(err)
	}
	var e ErrorMsg
	if err := m.Decode(&e); err != nil {
		t.Fatal(err)
	}
	if m.Type != WSTypeError || e.Code != 7 || e.Msg != "Unknown subscription ID" {
		t.Fatalf("error: %+v", e)
	}
}

func TestOrderbookSnapshotMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"market_ticker": "FED-23DEC-T3.00",
		"market_id": "9b0f6b43-5b68-4f9f-9f02-9a2d1b8ac1a1",
		"yes_dollars_fp": [["0.0800", "300.00"], ["0.2200", "333.00"]],
		"no_dollars_fp": [["0.5400", "20.00"], ["0.5600", "146.00"]]
	}`
	var msg OrderbookSnapshotMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if len(msg.YesDollarsFp) != 2 || len(msg.NoDollarsFp) != 2 {
		t.Fatalf("levels: %+v", msg)
	}
	if msg.YesDollarsFp[1] != (OrderbookLevel{PriceDollars: "0.2200", CountFp: "333.00"}) {
		t.Fatalf("yes[1]: %+v", msg.YesDollarsFp[1])
	}
	if msg.NoDollarsFp[0] != (OrderbookLevel{PriceDollars: "0.5400", CountFp: "20.00"}) {
		t.Fatalf("no[0]: %+v", msg.NoDollarsFp[0])
	}
}

func TestOrderbookSnapshotMsg_Unmarshal_OneSided(t *testing.T) {
	const payload = `{"market_ticker":"MKT","market_id":"id","no_dollars_fp":[["0.9900","1.00"]]}`
	var msg OrderbookSnapshotMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.YesDollarsFp != nil || len(msg.NoDollarsFp) != 1 {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestOrderbookLevel_RoundTrip(t *testing.T) {
	in := OrderbookLevel{PriceDollars: "0.0800", CountFp: "300.00"}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `["0.0800","300.00"]` {
		t.Fatalf("marshal: %s", data)
	}
	var out OrderbookLevel
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("round trip: %+v", out)
	}
}

func TestOrderbookLevel_Unmarshal_Malformed(t *testing.T) {
	for _, in := range []string{
		`["0.0800"]`,
		`["0.0800","300.00","extra"]`,
		`[]`,
		`[8, 300]`,
		`{"price":"0.0800","count":"300.00"}`,
		`"0.0800"`,
		`["", "300.00"]`,
		`["0.0800", ""]`,
		`null`,
	} {
		var l OrderbookLevel
		if err := json.Unmarshal([]byte(in), &l); err == nil {
			t.Errorf("%s: want error, got %+v", in, l)
		}
	}
}

func TestOrderbookDeltaMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"market_ticker": "FED-23DEC-T3.00",
		"market_id": "9b0f6b43-5b68-4f9f-9f02-9a2d1b8ac1a1",
		"price_dollars": "0.960",
		"delta_fp": "-54.00",
		"side": "yes",
		"ts": "2022-11-22T20:44:01Z",
		"ts_ms": 1669149841000
	}`
	var msg OrderbookDeltaMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.PriceDollars != "0.960" || msg.DeltaFp != "-54.00" || msg.Side != OrderSideYes || msg.TsMs == nil || *msg.TsMs != 1669149841000 {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.ClientOrderID != "" || msg.Subaccount != nil {
		t.Fatalf("optional fields should be unset: %+v", msg)
	}
}

func TestOrderbookDeltaMsg_Unmarshal_OwnOrder(t *testing.T) {
	const payload = `{"market_ticker":"MKT","market_id":"id","price_dollars":"0.5000","delta_fp":"10.00","side":"no","client_order_id":"my-1","subaccount":3,"ts_ms":1}`
	var msg OrderbookDeltaMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.ClientOrderID != "my-1" || msg.Subaccount == nil || *msg.Subaccount != 3 {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestTradeMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"trade_id": "d91bc706-ee49-470d-82d8-11418bda6fed",
		"market_ticker": "HIGHNY-22DEC23-B53.5",
		"yes_price_dollars": "0.3600",
		"no_price_dollars": "0.6400",
		"count_fp": "136.00",
		"taker_side": "no",
		"taker_outcome_side": "no",
		"taker_book_side": "ask",
		"is_block_trade": false,
		"ts": 1669149841,
		"ts_ms": 1669149841000
	}`
	var msg TradeMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.TradeID != "d91bc706-ee49-470d-82d8-11418bda6fed" || msg.YesPriceDollars != "0.3600" || msg.NoPriceDollars != "0.6400" || msg.CountFp != "136.00" {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.TakerOutcomeSide != OutcomeSideNo || msg.TakerBookSide != BookSideAsk || msg.IsBlockTrade || msg.TsMs != 1669149841000 {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestFillMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"trade_id": "d91bc706-ee49-470d-82d8-11418bda6fed",
		"order_id": "ee587a1c-8b87-4dcf-b721-9f6f790619fa",
		"market_ticker": "HIGHNY-22DEC23-B53.5",
		"exchange_index": 2,
		"is_taker": true,
		"side": "yes",
		"yes_price_dollars": "0.7500",
		"count_fp": "278.00",
		"fee_cost": "0.010000",
		"action": "buy",
		"ts": 1671899397,
		"ts_ms": 1671899397000,
		"post_position_fp": "500.00",
		"purchased_side": "yes",
		"outcome_side": "yes",
		"book_side": "bid",
		"subaccount": 3
	}`
	var msg FillMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.OrderID != "ee587a1c-8b87-4dcf-b721-9f6f790619fa" || msg.ExchangeIndex != 2 || !msg.IsTaker || msg.YesPriceDollars != "0.7500" || msg.CountFp != "278.00" {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.FeeCost != "0.010000" || msg.PostPositionFp != "500.00" || msg.OutcomeSide != OutcomeSideYes || msg.BookSide != BookSideBid {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.Subaccount == nil || *msg.Subaccount != 3 || msg.ClientOrderID != "" || msg.TsMs != 1671899397000 {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestMarketPositionMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"user_id": "user123",
		"market_ticker": "FED-23DEC-T3.00",
		"position_fp": "100.00",
		"position_cost_dollars": "50.0000",
		"realized_pnl_dollars": "10.0000",
		"fees_paid_dollars": "1.0000",
		"position_fee_cost_dollars": "0.5000",
		"volume_fp": "15.00"
	}`
	var msg MarketPositionMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.UserID != "user123" || msg.PositionFp != "100.00" || msg.PositionCostDollars != "50.0000" || msg.RealizedPnlDollars != "10.0000" {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.FeesPaidDollars != "1.0000" || msg.PositionFeeCostDollars != "0.5000" || msg.VolumeFp != "15.00" || msg.Subaccount != nil {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestUserOrderMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"order_id": "ee587a1c-8b87-4dcf-b721-9f6f790619fa",
		"user_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"ticker": "FED-23DEC-T3.00",
		"exchange_index": 2,
		"status": "resting",
		"side": "yes",
		"is_yes": true,
		"outcome_side": "yes",
		"book_side": "bid",
		"yes_price_dollars": "0.3500",
		"fill_count_fp": "0.00",
		"remaining_count_fp": "10.00",
		"initial_count_fp": "10.00",
		"taker_fill_cost_dollars": "0.000000",
		"maker_fill_cost_dollars": "0.000000",
		"taker_fees_dollars": "0.000000",
		"maker_fees_dollars": "0.000000",
		"client_order_id": "my-order-1",
		"order_group_id": "og_123",
		"self_trade_prevention_type": "taker_at_cross",
		"created_time": "2024-12-01T10:00:00Z",
		"created_ts_ms": 1733047200000,
		"expiration_time": "2024-12-01T11:00:00Z",
		"expiration_ts_ms": 1733050800000,
		"subaccount_number": 0
	}`
	var msg UserOrderMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.OrderID != "ee587a1c-8b87-4dcf-b721-9f6f790619fa" || msg.Ticker != "FED-23DEC-T3.00" || msg.Status != OrderStatusResting || !msg.IsYes {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.YesPriceDollars != "0.3500" || msg.RemainingCountFp != "10.00" || msg.InitialCountFp != "10.00" || msg.MakerFeesDollars != "0.000000" {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.ClientOrderID != "my-order-1" || msg.OrderGroupID != "og_123" || msg.SelfTradePreventionType != SelfTradeTakerAtCross || msg.CreatedTsMs != 1733047200000 {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.ExpirationTsMs == nil || *msg.ExpirationTsMs != 1733050800000 || msg.LastUpdatedTsMs != nil || msg.SubaccountNumber == nil || *msg.SubaccountNumber != 0 {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestOrderGroupUpdatesMsg_Unmarshal(t *testing.T) {
	const payload = `{"event_type":"limit_updated","order_group_id":"og_123","contracts_limit_fp":"150.00","ts_ms":1733047200000}`
	var msg OrderGroupUpdatesMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.EventType != "limit_updated" || msg.OrderGroupID != "og_123" || msg.ContractsLimitFp != "150.00" || msg.TsMs != 1733047200000 {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestMultivariateMarketLifecycleMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"market_ticker": "KXMVE-TEST-EVENT-M1",
		"event_type": "created",
		"exchange_index": 0,
		"open_ts": 1773936000,
		"close_ts": 1774022400,
		"additional_metadata": {
			"name": "MVE One",
			"title": "Market 1",
			"yes_sub_title": "YES 1",
			"no_sub_title": "NO 1",
			"rules_primary": "Rule 1",
			"rules_secondary": "Rule 2",
			"can_close_early": true,
			"event_ticker": "KXMVE-TEST-EVENT",
			"expected_expiration_ts": 1774029600
		}
	}`
	var msg MultivariateMarketLifecycleMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.EventType != "created" || msg.ExchangeIndex == nil || *msg.ExchangeIndex != 0 || msg.OpenTs == nil || *msg.OpenTs != 1773936000 || msg.CloseTs == nil {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.AdditionalMetadata == nil || msg.AdditionalMetadata.EventTicker != "KXMVE-TEST-EVENT" || msg.AdditionalMetadata.CanCloseEarly == nil || !*msg.AdditionalMetadata.CanCloseEarly {
		t.Fatalf("metadata: %+v", msg.AdditionalMetadata)
	}
	if msg.Result != "" || msg.SettledTs != nil || msg.IsDeactivated != nil {
		t.Fatalf("optional fields should be unset: %+v", msg)
	}
}

func TestMultivariateMarketLifecycleMsg_Unmarshal_Determined(t *testing.T) {
	const payload = `{"market_ticker":"KXMVE-TEST-EVENT-M1","event_type":"determined","result":"yes","determination_ts":1774022400,"settlement_value":"1.0000"}`
	var msg MultivariateMarketLifecycleMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.EventType != "determined" || msg.Result != "yes" || msg.DeterminationTs == nil || msg.SettlementValue != "1.0000" || msg.AdditionalMetadata != nil {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestEventLifecycleMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"event_ticker": "KXQUICKSETTLE-26JAN25H2150",
		"exchange_index": 0,
		"title": "What will 1+1 equal on Jan 25 at 21:50?",
		"subtitle": "Jan 25 at 21:50",
		"collateral_return_type": "MECNET",
		"series_ticker": "KXQUICKSETTLE"
	}`
	var msg EventLifecycleMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.EventTicker != "KXQUICKSETTLE-26JAN25H2150" || msg.CollateralReturnType != "MECNET" || msg.SeriesTicker != "KXQUICKSETTLE" || msg.StrikeDate != nil {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestEventFeeUpdateMsg_Unmarshal(t *testing.T) {
	var set EventFeeUpdateMsg
	if err := json.Unmarshal([]byte(`{"event_ticker":"KXBTCD-26MAY2018","fee_type_override":"quadratic","fee_multiplier_override":1}`), &set); err != nil {
		t.Fatal(err)
	}
	if set.FeeTypeOverride == nil || *set.FeeTypeOverride != "quadratic" || set.FeeMultiplierOverride == nil || *set.FeeMultiplierOverride != 1 {
		t.Fatalf("set: %+v", set)
	}
	var cleared EventFeeUpdateMsg
	if err := json.Unmarshal([]byte(`{"event_ticker":"KXBTCD-26MAY2018","fee_type_override":null,"fee_multiplier_override":null}`), &cleared); err != nil {
		t.Fatal(err)
	}
	if cleared.EventTicker != "KXBTCD-26MAY2018" || cleared.FeeTypeOverride != nil || cleared.FeeMultiplierOverride != nil {
		t.Fatalf("cleared: %+v", cleared)
	}
}

func TestRFQCreatedMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"id": "rfq_456",
		"creator_id": "",
		"market_ticker": "KXMVE-24DEC-COMBO",
		"event_ticker": "KXMVE-24DEC-EVENT",
		"target_cost_dollars": "100.0000",
		"created_ts": "2024-12-01T10:00:00Z",
		"mve_collection_ticker": "KXMVE-24DEC",
		"mve_selected_legs": [
			{"event_ticker": "KXEVENTA-24DEC", "market_ticker": "KXEVENTA-24DEC-YES", "side": "yes", "yes_settlement_value_dollars": "1.0000"},
			{"event_ticker": "KXEVENTB-24DEC", "market_ticker": "KXEVENTB-24DEC-YES", "side": "no"}
		]
	}`
	var msg RFQCreatedMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.ID != "rfq_456" || msg.TargetCostDollars != "100.0000" || msg.ContractsFp != "" || msg.CreatedTs != "2024-12-01T10:00:00Z" || msg.MveCollectionTicker != "KXMVE-24DEC" {
		t.Fatalf("unexpected: %+v", msg)
	}
	if len(msg.MveSelectedLegs) != 2 || msg.MveSelectedLegs[0].YesSettlementValueDollars == nil || *msg.MveSelectedLegs[0].YesSettlementValueDollars != "1.0000" || msg.MveSelectedLegs[1].YesSettlementValueDollars != nil {
		t.Fatalf("legs: %+v", msg.MveSelectedLegs)
	}
}

func TestRFQDeletedMsg_Unmarshal(t *testing.T) {
	const payload = `{"id":"rfq_123","creator_id":"comm_abc123","market_ticker":"FED-23DEC-T3.00","event_ticker":"FED-23DEC","contracts_fp":"100.00","target_cost_dollars":"0.35","deleted_ts":"2024-12-01T10:05:00Z"}`
	var msg RFQDeletedMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.ID != "rfq_123" || msg.CreatorID != "comm_abc123" || msg.ContractsFp != "100.00" || msg.DeletedTs != "2024-12-01T10:05:00Z" {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestQuoteCreatedMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"quote_id": "quote_456",
		"rfq_id": "rfq_123",
		"quote_creator_id": "comm_def456",
		"rfq_creator_id": "comm_abc123",
		"market_ticker": "FED-23DEC-T3.00",
		"event_ticker": "FED-23DEC",
		"yes_bid_dollars": "0.35",
		"no_bid_dollars": "0.65",
		"yes_contracts_offered_fp": "100.00",
		"no_contracts_offered_fp": "200.00",
		"rfq_target_cost_dollars": "0.35",
		"created_ts": "2024-12-01T10:02:00Z",
		"subaccount": 3
	}`
	var msg QuoteCreatedMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.QuoteID != "quote_456" || msg.RFQID != "rfq_123" || msg.YesBidDollars != "0.35" || msg.NoBidDollars != "0.65" || msg.NoContractsOfferedFp != "200.00" {
		t.Fatalf("unexpected: %+v", msg)
	}
	if msg.Subaccount == nil || *msg.Subaccount != 3 || msg.CreatedTs != "2024-12-01T10:02:00Z" {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestQuoteAcceptedMsg_Unmarshal(t *testing.T) {
	const payload = `{"quote_id":"quote_456","rfq_id":"rfq_123","quote_creator_id":"comm_def456","rfq_creator_id":"comm_abc123","market_ticker":"FED-23DEC-T3.00","event_ticker":"FED-23DEC","yes_bid_dollars":"0.35","no_bid_dollars":"0.65","accepted_side":"yes","contracts_accepted_fp":"50.00","yes_contracts_offered_fp":"100.00","no_contracts_offered_fp":"200.00","rfq_target_cost_dollars":"0.35","subaccount":3}`
	var msg QuoteAcceptedMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.AcceptedSide != OrderSideYes || msg.ContractsAcceptedFp != "50.00" || msg.RFQTargetCostDollars != "0.35" || msg.Subaccount == nil {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestQuoteExecutedMsg_Unmarshal(t *testing.T) {
	const payload = `{"quote_id":"quote_456","rfq_id":"rfq_123","quote_creator_id":"a1b2c3d4e5f6...","rfq_creator_id":"f6e5d4c3b2a1...","order_id":"order_789","client_order_id":"my_client_order_123","market_ticker":"FED-23DEC-T3.00","executed_ts":"2024-12-01T10:05:00Z","subaccount":3}`
	var msg QuoteExecutedMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.OrderID != "order_789" || msg.ClientOrderID != "my_client_order_123" || msg.ExecutedTs != "2024-12-01T10:05:00Z" || msg.Subaccount == nil || *msg.Subaccount != 3 {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestWSTypeConstants(t *testing.T) {
	for _, tc := range []struct{ got, want string }{
		{WSTypeSubscribed, "subscribed"},
		{WSTypeError, "error"},
		{WSTypeOrderbookSnapshot, "orderbook_snapshot"},
		{WSTypeOrderbookDelta, WSChannelOrderbookDelta},
		{WSTypeTicker, WSChannelTicker},
		{WSTypeTrade, WSChannelTrade},
		{WSTypeFill, WSChannelFill},
		{WSTypeMarketPosition, "market_position"},
		{WSTypeMarketLifecycleV2, WSChannelMarketLifecycle},
		{WSTypeMultivariateMarketLifecycle, WSChannelMultivariateLifecycle},
		{WSTypeOrderGroupUpdates, WSChannelOrderGroup},
		{WSTypeUserOrder, "user_order"},
		{WSTypePythValue, WSChannelPythValue},
		{WSTypeCFBenchmarksValue5HzIndexList, "cfbenchmarks_value_5hz_indexlist"},
	} {
		if tc.got != tc.want {
			t.Errorf("got %q want %q", tc.got, tc.want)
		}
	}
}
