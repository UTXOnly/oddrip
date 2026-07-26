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
		"updated_ts": 1716300000,
		"balance_breakdown": [{"exchange_index": 0, "balance": "100.0000"}]
	}`
	var out GetBalanceResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.BalanceDollars != "100.0000" || out.Balance != 10000 {
		t.Fatalf("unexpected: %+v", out)
	}
	if len(out.BalanceBreakdown) != 1 || out.BalanceBreakdown[0].ExchangeIndex != 0 {
		t.Fatalf("breakdown: %+v", out.BalanceBreakdown)
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
			"created_ts": 1716300000,
			"finalized_ts": 1716300100
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
	if out.Deposits[0].FinalizedTs == nil || *out.Deposits[0].FinalizedTs != 1716300100 {
		t.Fatalf("finalized_ts: %+v", out.Deposits[0].FinalizedTs)
	}
}

func TestMarketPosition_Unmarshal_FixedPointOnly(t *testing.T) {
	const payload = `{
		"ticker": "MKT",
		"total_traded_dollars": "10.0000",
		"position_fp": "2.00",
		"market_exposure_dollars": "1.0000",
		"realized_pnl_dollars": "0.5000",
		"fees_paid_dollars": "0.0100",
		"last_updated_ts": "2026-07-01T00:00:00Z"
	}`
	var p MarketPosition
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		t.Fatal(err)
	}
	if p.Ticker != "MKT" || p.PositionFp != "2.00" || p.FeesPaidDollars != "0.0100" {
		t.Fatalf("unexpected: %+v", p)
	}
}

func TestGetAccountApiLimitsResponse_Unmarshal_Grants(t *testing.T) {
	const payload = `{
		"usage_tier": "premier",
		"read": {"refill_rate": 20, "bucket_capacity": 20},
		"write": {"refill_rate": 10, "bucket_capacity": 10},
		"grants": [{"exchange_instance": "predictions", "level": "premier", "source": "volume"}]
	}`
	var out GetAccountApiLimitsResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.UsageTier != "premier" || len(out.Grants) != 1 || out.Grants[0].Source != "volume" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestExchangeStatus_Unmarshal_IndexStatuses(t *testing.T) {
	const payload = `{
		"exchange_active": true,
		"trading_active": true,
		"intra_exchange_transfers_active": true,
		"exchange_index_statuses": [{
			"exchange_index": 0,
			"exchange_active": true,
			"trading_active": true,
			"intra_exchange_transfers_active": false
		}]
	}`
	var out ExchangeStatus
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.IntraExchangeTransfersActive == nil || !*out.IntraExchangeTransfersActive {
		t.Fatalf("intra: %+v", out.IntraExchangeTransfersActive)
	}
	if len(out.ExchangeIndexStatuses) != 1 || out.ExchangeIndexStatuses[0].IntraExchangeTransfersActive {
		t.Fatalf("index statuses: %+v", out.ExchangeIndexStatuses)
	}
}

func TestGetHistoricalCutoffResponse_Unmarshal_PositionsCutoff(t *testing.T) {
	const payload = `{
		"market_settled_ts": "2026-01-01T00:00:00Z",
		"trades_created_ts": "2026-01-02T00:00:00Z",
		"orders_updated_ts": "2026-01-03T00:00:00Z",
		"market_positions_last_updated_ts": "2026-01-04T00:00:00Z"
	}`
	var out GetHistoricalCutoffResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.MarketPositionsLastUpdatedTs != "2026-01-04T00:00:00Z" {
		t.Fatalf("unexpected: %+v", out)
	}
}
