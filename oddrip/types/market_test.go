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
