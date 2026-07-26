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
		"taker_outcome_side": "yes",
		"taker_book_side": "bid",
		"created_time": "2026-03-20T12:00:00Z",
		"is_block_trade": false
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
	if tr.IsBlockTrade {
		t.Fatalf("is_block_trade: %+v", tr)
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
		"created_time": "2026-03-20T12:00:00Z",
		"is_block_trade": true
	}`
	var tr Trade
	if err := json.Unmarshal([]byte(payload), &tr); err != nil {
		t.Fatal(err)
	}
	if tr.TakerOutcomeSide != "yes" || tr.TakerBookSide != "bid" || !tr.IsBlockTrade {
		t.Fatalf("direction: %+v", tr)
	}
}

func TestMarket_Unmarshal_OccurrenceDatetime(t *testing.T) {
	const payload = `{
		"ticker": "MKT",
		"event_ticker": "EVT",
		"market_type": "binary",
		"yes_sub_title": "y",
		"no_sub_title": "n",
		"created_time": "2026-01-01T00:00:00Z",
		"updated_time": "2026-01-01T00:00:00Z",
		"open_time": "2026-01-01T00:00:00Z",
		"close_time": "2026-01-02T00:00:00Z",
		"latest_expiration_time": "2026-01-02T00:00:00Z",
		"settlement_timer_seconds": 0,
		"status": "active",
		"notional_value_dollars": "1.0000",
		"yes_bid_dollars": "0.5000",
		"yes_ask_dollars": "0.5100",
		"no_bid_dollars": "0.4900",
		"no_ask_dollars": "0.5000",
		"yes_bid_size_fp": "0.00",
		"yes_ask_size_fp": "0.00",
		"last_price_dollars": "0.5000",
		"previous_yes_bid_dollars": "0.5000",
		"previous_yes_ask_dollars": "0.5100",
		"previous_price_dollars": "0.5000",
		"volume_fp": "0.00",
		"volume_24h_fp": "0.00",
		"open_interest_fp": "0.00",
		"result": "",
		"can_close_early": false,
		"expiration_value": "",
		"rules_primary": "",
		"rules_secondary": "",
		"price_level_structure": "linear_cent",
		"price_ranges": [],
		"occurrence_datetime": "2026-01-01T12:00:00Z",
		"exchange_index": 0
	}`
	var m Market
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatal(err)
	}
	if m.OccurrenceDatetime == nil || *m.OccurrenceDatetime != "2026-01-01T12:00:00Z" {
		t.Fatalf("occurrence_datetime: %+v", m.OccurrenceDatetime)
	}
	if m.ExchangeIndex != 0 {
		t.Fatalf("exchange_index: %d", m.ExchangeIndex)
	}
}

func TestGetMarketOrderbookResponse_Unmarshal_OrderbookFpOnly(t *testing.T) {
	const payload = `{
		"orderbook_fp": {
			"yes_dollars": [["0.5000", "10.00"]],
			"no_dollars": [["0.4500", "5.00"]]
		}
	}`
	var out GetMarketOrderbookResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.OrderbookFp.YesDollars) != 1 || out.OrderbookFp.YesDollars[0][0] != "0.5000" {
		t.Fatalf("unexpected: %+v", out)
	}
}
