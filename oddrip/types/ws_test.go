package types

import (
	"encoding/json"
	"testing"
)

func TestSubscribeCommand_Marshal(t *testing.T) {
	cmd := SubscribeCommand{
		ID:  1,
		Cmd: "subscribe",
		Params: SubscribeParams{
			Channels:     []string{WSChannelTicker, WSChannelOrderbookDelta},
			MarketTicker: "FED-23DEC-T3.00",
		},
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if json.Unmarshal(data, &m) != nil {
		t.Fatal("unmarshal")
	}
	if m["cmd"] != "subscribe" || m["id"].(float64) != 1 {
		t.Errorf("unexpected cmd or id: %v", m)
	}
	params := m["params"].(map[string]interface{})
	channels := params["channels"].([]interface{})
	if len(channels) != 2 || params["market_ticker"] != "FED-23DEC-T3.00" {
		t.Errorf("params: %v", params)
	}
}

func TestSubscribeCommand_PythValue_UnderlyingTickers(t *testing.T) {
	cmd := SubscribeCommand{
		ID:  4,
		Cmd: "subscribe",
		Params: SubscribeParams{
			Channels:          []string{WSChannelPythValue},
			UnderlyingTickers: []string{"Metal.XAU/USD"},
		},
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SubscribeCommand
	if json.Unmarshal(data, &decoded) != nil {
		t.Fatal("unmarshal")
	}
	if decoded.Params.Channels[0] != WSChannelPythValue || len(decoded.Params.UnderlyingTickers) != 1 {
		t.Fatalf("decoded: %+v", decoded.Params)
	}
}

func TestUpdateSubscriptionParams_GetSnapshot(t *testing.T) {
	cmd := UpdateSubscriptionCommand{
		ID:  3,
		Cmd: "update_subscription",
		Params: UpdateSubscriptionParams{
			Sids:          []int{456},
			Action:        WSUpdateSubscriptionGetSnapshot,
			MarketTickers: []string{"MKT-1"},
		},
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var decoded UpdateSubscriptionCommand
	if json.Unmarshal(data, &decoded) != nil {
		t.Fatal("unmarshal")
	}
	if decoded.Params.Action != "get_snapshot" || len(decoded.Params.MarketTickers) != 1 {
		t.Fatalf("decoded: %+v", decoded.Params)
	}
}

func TestUpdateSubscriptionParams_Action(t *testing.T) {
	cmd := UpdateSubscriptionCommand{
		ID:  2,
		Cmd: "update_subscription",
		Params: UpdateSubscriptionParams{
			SID:           intPtr(10),
			Action:        WSUpdateSubscriptionAddMarkets,
			MarketTickers: []string{"TICK-A", "TICK-B"},
		},
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var decoded UpdateSubscriptionCommand
	if json.Unmarshal(data, &decoded) != nil {
		t.Fatal("unmarshal")
	}
	if decoded.Params.Action != "add_markets" || *decoded.Params.SID != 10 || len(decoded.Params.MarketTickers) != 2 {
		t.Errorf("decoded: %+v", decoded)
	}
}

func TestUpdateSubscriptionParams_SubscribeUnderlyings(t *testing.T) {
	cmd := UpdateSubscriptionCommand{
		ID:  5,
		Cmd: "update_subscription",
		Params: UpdateSubscriptionParams{
			SID:               intPtr(1),
			Action:            WSUpdateSubscriptionSubscribeUnderlyings,
			UnderlyingTickers: []string{"Metal.XAG/USD"},
		},
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var decoded UpdateSubscriptionCommand
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Params.Action != "subscribe_underlyings" || decoded.Params.UnderlyingTickers[0] != "Metal.XAG/USD" {
		t.Fatalf("decoded: %+v", decoded.Params)
	}
}

func TestPythValueMsg_Unmarshal(t *testing.T) {
	const payload = `{
		"underlying_ticker": "Metal.XAU/USD",
		"value_usd": "2365.12345000",
		"source_ts_ms": 1710000000100,
		"received_at": 1710000000123
	}`
	var msg PythValueMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.UnderlyingTicker != "Metal.XAU/USD" || msg.ValueUSD != "2365.12345000" {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestMarketLifecycleV2Msg_Unmarshal_PriceRanges(t *testing.T) {
	const payload = `{
		"market_ticker": "INXD-23SEP14-B4487",
		"event_type": "price_level_structure_updated",
		"price_level_structure": "deci_cent",
		"price_ranges": [{"start": "0.0000", "end": "1.0000", "step": "0.0010"}]
	}`
	var msg MarketLifecycleV2Msg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.EventType != "price_level_structure_updated" || len(msg.PriceRanges) != 1 || msg.PriceRanges[0].Step != "0.0010" {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func TestMarketLifecycleV2Msg_Unmarshal_MetadataUpdated(t *testing.T) {
	const payload = `{
		"market_ticker": "KXBTC-25APR30-T0915-B95000",
		"event_type": "metadata_updated",
		"strike_type": "between",
		"floor_strike": 95000,
		"cap_strike": 95250,
		"yes_sub_title": "above $95,000"
	}`
	var msg MarketLifecycleV2Msg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.StrikeType != "between" || msg.FloorStrike == nil || *msg.FloorStrike != 95000 || msg.YesSubTitle == "" {
		t.Fatalf("unexpected: %+v", msg)
	}
}

func intPtr(i int) *int { return &i }

func TestCFBenchmarksValue5HzMsg_Unmarshal(t *testing.T) {
	const payload = `{"index_id":"BRTI","value_usd":"65000.12345678","source_ts_ms":1716300000200,"received_at":1716300000250,"data":"{\"raw\":true}"}`
	var out CFBenchmarksValue5HzMsg
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.IndexID != "BRTI" || out.ValueUSD != "65000.12345678" || out.SourceTsMs != 1716300000200 {
		t.Fatalf("unexpected: %+v", out)
	}
	if out.Data != `{"raw":true}` {
		t.Fatalf("data: %q", out.Data)
	}
}

func TestCFBenchmarksValueMsg_Unmarshal(t *testing.T) {
	const payload = `{"index_id":"BRTI","received_at":1716300000250,"data":"{}","avg_60s_data":{"index_id":"BRTI","value_usd":"64999.00000000","source_ts_ms":1716300000000}}`
	var out CFBenchmarksValueMsg
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.Avg60sData == nil || out.Avg60sData.ValueUSD != "64999.00000000" {
		t.Fatalf("avg: %+v", out.Avg60sData)
	}
	if out.Last60sWindowedAverage15Min != nil {
		t.Fatalf("unexpected 15min average: %+v", out.Last60sWindowedAverage15Min)
	}
}

func TestSubscribeParams_IndexIDsMarshal(t *testing.T) {
	data, err := json.Marshal(SubscribeParams{
		Channels: []string{WSChannelCFBenchmarksValue5Hz},
		IndexIDs: []string{"BRTI", "ETHUSD_RTI"},
	})
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"channels":["cfbenchmarks_value_5hz"],"index_ids":["BRTI","ETHUSD_RTI"]}`
	if string(data) != want {
		t.Fatalf("marshal: %s", data)
	}
}
