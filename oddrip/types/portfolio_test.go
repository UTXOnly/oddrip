package types

import (
	"encoding/json"
	"testing"
)

func TestGetBalanceResponse_Unmarshal(t *testing.T) {
	const payload = `{
		"balance": 10000,
		"balance_dollars": "100.0000",
		"portfolio_value": 15000,
		"updated_ts": 1716300000
	}`
	var out GetBalanceResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.BalanceDollars != "100.0000" || out.Balance != 10000 {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestGetDepositsResponse_Unmarshal(t *testing.T) {
	const payload = `{
		"deposits": [{
			"id": "d1",
			"status": "applied",
			"type": "ach",
			"amount_cents": 5000,
			"fee_cents": 0,
			"created_ts": 1716300000
		}],
		"cursor": "next"
	}`
	var out GetDepositsResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Deposits) != 1 || out.Deposits[0].ID != "d1" || out.Cursor != "next" {
		t.Fatalf("unexpected: %+v", out)
	}
}
