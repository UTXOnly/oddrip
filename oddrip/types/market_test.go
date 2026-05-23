package types

import (
	"encoding/json"
	"testing"
)

func TestTrade_Unmarshal_OpenAPIRequiredFields(t *testing.T) {
	const payload = `{
		"trade_id": "t1",
		"ticker": "MKT-A",
		"count_fp": "1.00",
		"yes_price_dollars": "0.5500",
		"no_price_dollars": "0.4500",
		"taker_side": "yes",
		"created_time": "2026-03-20T12:00:00Z"
	}`
	var tr Trade
	if err := json.Unmarshal([]byte(payload), &tr); err != nil {
		t.Fatal(err)
	}
	if tr.TradeID != "t1" || tr.Ticker != "MKT-A" || tr.CountFp != "1.00" {
		t.Fatalf("unexpected: %+v", tr)
	}
	if tr.YesPriceDollars != "0.5500" || tr.NoPriceDollars != "0.4500" || tr.CreatedTime == "" {
		t.Fatalf("dollars/time: %+v", tr)
	}
}

func TestTrade_Unmarshal_TakerOutcomeSide(t *testing.T) {
	const payload = `{
		"trade_id": "t1",
		"ticker": "MKT-A",
		"count_fp": "1.00",
		"yes_price_dollars": "0.5500",
		"no_price_dollars": "0.4500",
		"taker_side": "yes",
		"taker_outcome_side": "yes",
		"taker_book_side": "bid",
		"created_time": "2026-03-20T12:00:00Z"
	}`
	var tr Trade
	if err := json.Unmarshal([]byte(payload), &tr); err != nil {
		t.Fatal(err)
	}
	if tr.TakerOutcomeSide != "yes" || tr.TakerBookSide != "bid" {
		t.Fatalf("direction: %+v", tr)
	}
}

func TestMarket_Unmarshal_OccurrenceDatetime(t *testing.T) {
	const payload = `{
		"ticker": "MKT",
		"event_ticker": "EVT",
		"market_type": "binary",
		"title": "t",
		"subtitle": "s",
		"yes_sub_title": "y",
		"no_sub_title": "n",
		"created_time": "2026-01-01T00:00:00Z",
		"updated_time": "2026-01-01T00:00:00Z",
		"open_time": "2026-01-01T00:00:00Z",
		"close_time": "2026-01-02T00:00:00Z",
		"expiration_time": "2026-01-02T00:00:00Z",
		"latest_expiration_time": "2026-01-02T00:00:00Z",
		"settlement_timer_seconds": 0,
		"status": "active",
		"response_price_units": "usd_cent",
		"notional_value_dollars": "1.0000",
		"yes_bid_dollars": "0.5000",
		"yes_ask_dollars": "0.5100",
		"no_bid_dollars": "0.4900",
		"no_ask_dollars": "0.5000",
		"last_price_dollars": "0.5000",
		"volume_fp": "0.00",
		"volume_24h_fp": "0.00",
		"open_interest_fp": "0.00",
		"result": "",
		"can_close_early": false,
		"fractional_trading_enabled": true,
		"expiration_value": "",
		"rules_primary": "",
		"rules_secondary": "",
		"tick_size": 1,
		"price_level_structure": "linear_cent",
		"price_ranges": [],
		"occurrence_datetime": "2026-01-01T12:00:00Z"
	}`
	var m Market
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatal(err)
	}
	if m.OccurrenceDatetime == nil || *m.OccurrenceDatetime != "2026-01-01T12:00:00Z" {
		t.Fatalf("occurrence_datetime: %+v", m.OccurrenceDatetime)
	}
}
