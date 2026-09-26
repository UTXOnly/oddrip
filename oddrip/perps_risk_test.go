package oddrip

import (
	"context"
	"net/http"
	"testing"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

func TestPerpsRisk_GetParameters(t *testing.T) {
	client, ct := newCaptureClient(200, `{
		"liquidation_margin_ratio_threshold":1.0,"queue_entry_margin_ratio_threshold":1.25,
		"initial_margin_multiplier":{"KXBTCPERP":2.0}
	}`)
	got, err := client.Perps.Risk.GetParameters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/risk_parameters")
	ct.assertQuery(t, map[string]string{})
	if got.QueueEntryMarginRatioThreshold != 1.25 || got.InitialMarginMultiplier["KXBTCPERP"] != 2.0 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsRisk_GetNotionalRiskLimit(t *testing.T) {
	client, ct := newCaptureClient(200, `{
		"default_notional_value_risk_limit":"5000.0000",
		"notional_value_risk_limits_by_market_ticker":{"market-abc-123":"5000.0000"},
		"total_current_usage":"1250.0000",
		"current_usage_by_market_ticker":{"market-abc-123":"1250.0000"}
	}`)
	got, err := client.Perps.Risk.GetNotionalRiskLimit(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/notional_risk_limit")
	ct.assertQuery(t, map[string]string{})
	if got.DefaultNotionalValueRiskLimit != "5000.0000" || got.TotalCurrentUsage != "1250.0000" ||
		got.CurrentUsageByMarketTicker["market-abc-123"] != "1250.0000" {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsRisk_Get(t *testing.T) {
	client, ct := newCaptureClient(200, `{
		"account_leverage":null,"total_position_notional":"0.0000","total_maintenance_margin":"0.0000","positions":[]
	}`)
	got, err := client.Perps.Risk.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/risk")
	ct.assertQuery(t, map[string]string{})
	if got.AccountLeverage != nil || got.TotalMaintenanceMargin != "0.0000" || got.Positions == nil {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsFees_GetTiers(t *testing.T) {
	client, ct := newCaptureClient(200, `{"maker_fee_rates":{"KXBTCPERP":0.0005},"taker_fee_rates":{"KXBTCPERP":0.0012}}`)
	got, err := client.Perps.Fees.GetTiers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/fee_tiers")
	ct.assertQuery(t, map[string]string{})
	if got.MakerFeeRates["KXBTCPERP"] != 0.0005 || got.TakerFeeRates["KXBTCPERP"] != 0.0012 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsFees_GetTierRates(t *testing.T) {
	client, ct := newCaptureClient(200, `{"fee_tier_rates":[
		{"fee_schedule":"fcm","tier":0,"maker_fee_rate":0.0005,"taker_fee_rate":0.0012}
	]}`)
	got, err := client.Perps.Fees.GetTierRates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/fee_tier_rates")
	ct.assertQuery(t, map[string]string{})
	if len(got.FeeTierRates) != 1 || got.FeeTierRates[0].FeeSchedule != types.MarginFeeScheduleFCM {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsFunding_GetHistory(t *testing.T) {
	client, ct := newCaptureClient(200, `{"funding_history":[{
		"market_ticker":"KXBTCPERP","funding_time":"2026-01-31T08:00:00Z","funding_rate":0.0001,
		"mark_price":"65000.0000","funding_amount":"-0.6500","quantity":"1.00","subaccount_number":0
	}]}`)
	got, err := client.Perps.Funding.GetHistory(context.Background(), &types.GetMarginFundingHistoryOpts{
		Ticker: "KXBTCPERP", StartDate: "2026-01-01", EndDate: "2026-01-31", Subaccount: ptrOf(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/funding_history")
	ct.assertQuery(t, map[string]string{
		"ticker": "KXBTCPERP", "start_date": "2026-01-01", "end_date": "2026-01-31", "subaccount": "0",
	})
	if len(got.FundingHistory) != 1 || got.FundingHistory[0].FundingAmount != "-0.6500" {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsFunding_GetHistory_RequiredOnly(t *testing.T) {
	client, ct := newCaptureClient(200, `{"funding_history":[]}`)
	if _, err := client.Perps.Funding.GetHistory(context.Background(), &types.GetMarginFundingHistoryOpts{
		StartDate: "2026-01-01", EndDate: "2026-01-01",
	}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/funding_history")
	ct.assertQuery(t, map[string]string{"start_date": "2026-01-01", "end_date": "2026-01-01"})
}

func TestPerpsFunding_GetHistory_Validation(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	for name, opts := range map[string]*types.GetMarginFundingHistoryOpts{
		"nil opts":       nil,
		"no start_date":  {EndDate: "2026-01-31"},
		"no end_date":    {StartDate: "2026-01-01"},
		"no dates":       {Ticker: "KXBTCPERP"},
		"subaccount set": {Subaccount: ptrOf(1)},
	} {
		if _, err := client.Perps.Funding.GetHistory(ctx, opts); err == nil {
			t.Fatalf("%s: expected error", name)
		}
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestPerpsFunding_GetHistoricalRates(t *testing.T) {
	client, ct := newCaptureClient(200, `{"funding_rates":[
		{"market_ticker":"KXBTCPERP","funding_time":"2026-01-31T08:00:00Z","funding_rate":-0.0002,"mark_price":"65000.0000"}
	]}`)
	got, err := client.Perps.Funding.GetHistoricalRates(context.Background(), &types.GetMarginHistoricalFundingRatesOpts{
		Ticker: "KXBTCPERP", StartTs: ptrOf[int64](1700000000), EndTs: ptrOf[int64](1700086400),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/funding_rates/historical")
	ct.assertQuery(t, map[string]string{"ticker": "KXBTCPERP", "start_ts": "1700000000", "end_ts": "1700086400"})
	if len(got.FundingRates) != 1 || got.FundingRates[0].FundingRate != -0.0002 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsFunding_GetHistoricalRates_NilOpts(t *testing.T) {
	client, ct := newCaptureClient(200, `{"funding_rates":[]}`)
	if _, err := client.Perps.Funding.GetHistoricalRates(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/funding_rates/historical")
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsFunding_GetRateEstimate(t *testing.T) {
	client, ct := newCaptureClient(200, `{
		"market_ticker":"KXBTCPERP","computed_time":"2026-01-31T07:30:00Z","funding_rate":0.00005,
		"mark_price":"65000.0000","next_funding_time":"2026-01-31T08:00:00Z"
	}`)
	got, err := client.Perps.Funding.GetRateEstimate(context.Background(), &types.GetMarginFundingRateEstimateOpts{Ticker: "KXBTCPERP"})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/funding_rates/estimate")
	ct.assertQuery(t, map[string]string{"ticker": "KXBTCPERP"})
	if got.NextFundingTime != "2026-01-31T08:00:00Z" || got.FundingRate == nil || *got.FundingRate != 0.00005 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsFunding_GetRateEstimate_Validation(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	if _, err := client.Perps.Funding.GetRateEstimate(ctx, nil); err == nil {
		t.Fatal("expected error for nil opts")
	}
	if _, err := client.Perps.Funding.GetRateEstimate(ctx, &types.GetMarginFundingRateEstimateOpts{}); err == nil {
		t.Fatal("expected error for empty ticker")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}
