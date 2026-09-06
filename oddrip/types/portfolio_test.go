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

func TestExchangeIndexFieldsOnPortfolioShapes(t *testing.T) {
	const fillPayload = `{"fills":[{"fill_id":"f1","order_id":"o1","ticker":"T","market_ticker":"T","count_fp":"1.0000","yes_price_dollars":"0.5600","no_price_dollars":"0.4400","is_taker":true,"fee_cost":"0.0100","outcome_side":"yes","book_side":"bid","exchange_index":1}],"cursor":""}`
	var fills GetFillsResponse
	if err := json.Unmarshal([]byte(fillPayload), &fills); err != nil {
		t.Fatal(err)
	}
	if len(fills.Fills) != 1 || fills.Fills[0].ExchangeIndex == nil || *fills.Fills[0].ExchangeIndex != 1 {
		t.Fatalf("fill: %+v", fills.Fills)
	}

	const positionsPayload = `{"cursor":"","market_positions":[{"ticker":"T","total_traded_dollars":"1.0000","position_fp":"2.0000","market_exposure_dollars":"1.0000","realized_pnl_dollars":"0.0000","fees_paid_dollars":"0.0100","last_updated_ts":"2026-09-06T00:00:00Z","exchange_index":2}],"event_positions":[]}`
	var positions GetPositionsResponse
	if err := json.Unmarshal([]byte(positionsPayload), &positions); err != nil {
		t.Fatal(err)
	}
	if positions.MarketPositions[0].ExchangeIndex == nil || *positions.MarketPositions[0].ExchangeIndex != 2 {
		t.Fatalf("position: %+v", positions.MarketPositions)
	}

	const settlementsPayload = `{"settlements":[{"ticker":"T","event_ticker":"E","market_result":"yes","yes_count_fp":"1.0000","yes_total_cost_dollars":"0.5600","no_count_fp":"0.0000","no_total_cost_dollars":"0.0000","revenue":100,"settled_time":"2026-09-06T00:00:00Z","fee_cost":"0.0100","exchange_index":0}]}`
	var settlements GetSettlementsResponse
	if err := json.Unmarshal([]byte(settlementsPayload), &settlements); err != nil {
		t.Fatal(err)
	}
	if settlements.Settlements[0].ExchangeIndex == nil || *settlements.Settlements[0].ExchangeIndex != 0 {
		t.Fatalf("settlement: %+v", settlements.Settlements)
	}
}

func TestGetIntraExchangeTransfersResponse_Unmarshal(t *testing.T) {
	const payload = `{
		"transfers": [{
			"transfer_id": "tr1",
			"source": "event_contract",
			"destination": "margined",
			"source_exchange_shard": 0,
			"destination_exchange_shard": 1,
			"amount": "25.0000",
			"status": "complete",
			"created_ts": 1716300000
		}],
		"cursor": "next"
	}`
	var out GetIntraExchangeTransfersResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Transfers) != 1 {
		t.Fatalf("transfers: %+v", out.Transfers)
	}
	tr := out.Transfers[0]
	if tr.Source != ExchangeInstanceEventContract || tr.Destination != ExchangeInstanceMargined {
		t.Fatalf("instances: %+v", tr)
	}
	if tr.Status != IntraExchangeInstanceTransferStatusComplete || tr.Amount != "25.0000" {
		t.Fatalf("transfer: %+v", tr)
	}
}

func TestSetTargetBalanceAllocationRequest_Marshal(t *testing.T) {
	req := SetTargetBalanceAllocationRequest{
		Allocations:              []TargetBalanceAllocation{{ExchangeIndex: 0, Percent: 70}, {ExchangeIndex: 1, Percent: 30}},
		RestingMarginReservation: RestingMarginReservationSum,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"allocations":[{"exchange_index":0,"percent":70},{"exchange_index":1,"percent":30}],"resting_margin_reservation":"sum"}`
	if string(data) != want {
		t.Fatalf("marshal: %s", data)
	}

	empty, err := json.Marshal(SetTargetBalanceAllocationRequest{Allocations: []TargetBalanceAllocation{}})
	if err != nil {
		t.Fatal(err)
	}
	if string(empty) != `{"allocations":[]}` {
		t.Fatalf("empty marshal: %s", empty)
	}
}
