package types

import (
	"encoding/json"
	"testing"
)

func TestCreateOrderV2Response_Unmarshal(t *testing.T) {
	const payload = `{
		"order_id": "ord-1",
		"client_order_id": "cli-1",
		"fill_count": "0.00",
		"remaining_count": "10.00",
		"ts_ms": 1716300000000
	}`
	var out CreateOrderV2Response
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.OrderID != "ord-1" || out.FillCount != "0.00" || out.TsMs != 1716300000000 {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestGetAccountApiLimitsResponse_Unmarshal(t *testing.T) {
	const payload = `{
		"usage_tier": "advanced",
		"read": {"refill_rate": 200, "bucket_capacity": 200},
		"write": {"refill_rate": 100, "bucket_capacity": 200}
	}`
	var out GetAccountApiLimitsResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.Read.RefillRate != 200 || out.Write.BucketCapacity != 200 {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestCancelOrderV2Response_Unmarshal(t *testing.T) {
	const payload = `{"order_id":"o1","client_order_id":"c1","reduced_by":"3.00","ts_ms":99}`
	var out CancelOrderV2Response
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.OrderID != "o1" || out.ReducedBy != "3.00" || out.TsMs != 99 {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestOrder_Unmarshal_OutcomeSide(t *testing.T) {
	const payload = `{
		"order_id": "o1",
		"user_id": "u1",
		"client_order_id": "c1",
		"ticker": "MKT",
		"side": "yes",
		"action": "buy",
		"type": "limit",
		"status": "resting",
		"yes_price": 50,
		"no_price": 50,
		"yes_price_dollars": "0.5000",
		"no_price_dollars": "0.5000",
		"fill_count_fp": "0.00",
		"remaining_count_fp": "1.00",
		"initial_count_fp": "1.00",
		"taker_fees_dollars": "0.0000",
		"maker_fees_dollars": "0.0000",
		"taker_fill_cost_dollars": "0.0000",
		"maker_fill_cost_dollars": "0.0000",
		"outcome_side": "yes",
		"book_side": "bid"
	}`
	var o Order
	if err := json.Unmarshal([]byte(payload), &o); err != nil {
		t.Fatal(err)
	}
	if o.OutcomeSide != "yes" || o.BookSide != "bid" {
		t.Fatalf("direction: %+v", o)
	}
}

func TestFill_Unmarshal_OutcomeSide(t *testing.T) {
	const payload = `{
		"fill_id": "f1",
		"trade_id": "t1",
		"order_id": "o1",
		"ticker": "MKT",
		"market_ticker": "MKT",
		"side": "yes",
		"action": "buy",
		"count_fp": "1.00",
		"yes_price_dollars": "0.5500",
		"no_price_dollars": "0.4500",
		"outcome_side": "yes",
		"book_side": "bid",
		"is_taker": true,
		"fee_cost": "0.0100"
	}`
	var f Fill
	if err := json.Unmarshal([]byte(payload), &f); err != nil {
		t.Fatal(err)
	}
	if f.OutcomeSide != "yes" || f.BookSide != "bid" {
		t.Fatalf("direction: %+v", f)
	}
}
