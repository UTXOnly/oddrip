package types

import (
	"encoding/json"
	"testing"
)

func TestPerpsRiskTypes_GetMarginRiskParametersResponse(t *testing.T) {
	const payload = `{
		"liquidation_margin_ratio_threshold": 1.0,
		"queue_entry_margin_ratio_threshold": 1.1,
		"initial_margin_multiplier": {"KXBTCPERP": 2.0, "TEST-PERP": 1.5}
	}`
	var out GetMarginRiskParametersResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.LiquidationMarginRatioThreshold != 1.0 || out.QueueEntryMarginRatioThreshold != 1.1 ||
		len(out.InitialMarginMultiplier) != 2 || out.InitialMarginMultiplier["TEST-PERP"] != 1.5 {
		t.Fatalf("unexpected: %+v", out)
	}
}

// The payload uses the per-field examples from the spec.
func TestPerpsRiskTypes_NotionalRiskLimitResponse(t *testing.T) {
	const payload = `{
		"default_notional_value_risk_limit": "5000.0000",
		"notional_value_risk_limits_by_market_ticker": {"market-abc-123": "5000.0000"},
		"total_current_usage": "1250.0000",
		"current_usage_by_market_ticker": {"market-abc-123": "1250.0000"},
		"member_notional_value_risk_limit": "5000.0000",
		"effective_account_notional_value_risk_limit": "5000.0000"
	}`
	var out NotionalRiskLimitResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.DefaultNotionalValueRiskLimit != "5000.0000" ||
		out.NotionalValueRiskLimitsByMarketTicker["market-abc-123"] != "5000.0000" ||
		out.TotalCurrentUsage != "1250.0000" ||
		out.CurrentUsageByMarketTicker["market-abc-123"] != "1250.0000" ||
		out.MemberNotionalValueRiskLimit != "5000.0000" ||
		out.EffectiveAccountNotionalValueRiskLimit != "5000.0000" {
		t.Fatalf("unexpected: %+v", out)
	}

	var absent NotionalRiskLimitResponse
	if err := json.Unmarshal([]byte(`{
		"default_notional_value_risk_limit": "5000.0000",
		"notional_value_risk_limits_by_market_ticker": {},
		"total_current_usage": "0.0000",
		"current_usage_by_market_ticker": {}
	}`), &absent); err != nil {
		t.Fatal(err)
	}
	if absent.MemberNotionalValueRiskLimit != "" || absent.EffectiveAccountNotionalValueRiskLimit != "" {
		t.Fatalf("absent limits: %+v", absent)
	}
}

func TestPerpsRiskTypes_GetMarginRiskResponse(t *testing.T) {
	const payload = `{
		"account_leverage": 4.5,
		"total_position_notional": "650.0000",
		"total_maintenance_margin": "144.4444",
		"positions": [
			{
				"subaccount": 0,
				"market_ticker": "KXBTCPERP",
				"position": "-0.01",
				"mark_price": "65000.0000",
				"position_notional": "650.0000",
				"maintenance_margin_required": "144.4444",
				"position_leverage": 4.5,
				"estimated_liquidation_price": "78000.0000",
				"is_portfolio": false
			},
			{
				"subaccount": 2,
				"market_ticker": "TEST-PERP",
				"position": "3.00",
				"mark_price": "1.2500",
				"position_notional": "3.7500",
				"maintenance_margin_required": null,
				"position_leverage": null,
				"estimated_liquidation_price": null,
				"is_portfolio": true
			}
		]
	}`
	var out GetMarginRiskResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.AccountLeverage == nil || *out.AccountLeverage != 4.5 ||
		out.TotalPositionNotional != "650.0000" || out.TotalMaintenanceMargin != "144.4444" ||
		len(out.Positions) != 2 {
		t.Fatalf("unexpected: %+v", out)
	}
	p := out.Positions[0]
	if p.Subaccount != 0 || p.MarketTicker != "KXBTCPERP" || p.Position != "-0.01" || p.MarkPrice != "65000.0000" ||
		p.PositionNotional != "650.0000" || p.IsPortfolio ||
		p.MaintenanceMarginRequired == nil || *p.MaintenanceMarginRequired != "144.4444" ||
		p.PositionLeverage == nil || *p.PositionLeverage != 4.5 ||
		p.EstimatedLiquidationPrice == nil || *p.EstimatedLiquidationPrice != "78000.0000" {
		t.Fatalf("position 0: %+v", p)
	}
	p = out.Positions[1]
	if p.Subaccount != 2 || !p.IsPortfolio ||
		p.MaintenanceMarginRequired != nil || p.PositionLeverage != nil || p.EstimatedLiquidationPrice != nil {
		t.Fatalf("position 1: %+v", p)
	}

	var nullLev GetMarginRiskResponse
	if err := json.Unmarshal([]byte(`{"account_leverage":null,"total_position_notional":"0.0000","total_maintenance_margin":"0.0000","positions":[]}`), &nullLev); err != nil {
		t.Fatal(err)
	}
	if nullLev.AccountLeverage != nil || nullLev.Positions == nil || len(nullLev.Positions) != 0 {
		t.Fatalf("null leverage: %+v", nullLev)
	}
}

func TestPerpsRiskTypes_GetMarginFeeTiersResponse(t *testing.T) {
	const payload = `{
		"maker_fee_rates": {"KXBTCPERP": 0.0005, "TEST-PERP": 0},
		"taker_fee_rates": {"KXBTCPERP": 0.0012, "TEST-PERP": 0.001}
	}`
	var out GetMarginFeeTiersResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.MakerFeeRates["KXBTCPERP"] != 0.0005 || out.TakerFeeRates["KXBTCPERP"] != 0.0012 ||
		out.TakerFeeRates["TEST-PERP"] != 0.001 {
		t.Fatalf("unexpected: %+v", out)
	}
	if r, ok := out.MakerFeeRates["TEST-PERP"]; !ok || r != 0 {
		t.Fatalf("zero maker rate: %v %v", r, ok)
	}
}

func TestPerpsRiskTypes_MarginFeeSchedule(t *testing.T) {
	for got, want := range map[string]string{
		MarginFeeScheduleSelfClearingMembers: "self_clearing_members",
		MarginFeeScheduleKalshiPrime:         "kalshi_prime",
		MarginFeeScheduleFCM:                 "fcm",
	} {
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestPerpsRiskTypes_GetMarginFeeTierRatesResponse(t *testing.T) {
	const payload = `{"fee_tier_rates": [
		{"fee_schedule": "self_clearing_members", "tier": 0, "maker_fee_rate": 0.0005, "taker_fee_rate": 0.0012},
		{"fee_schedule": "self_clearing_members", "tier": 1, "maker_fee_rate": 0.0003, "taker_fee_rate": 0.001},
		{"fee_schedule": "kalshi_prime", "tier": 3, "maker_fee_rate": -0.0001, "taker_fee_rate": 0.0008}
	]}`
	var out GetMarginFeeTierRatesResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.FeeTierRates) != 3 {
		t.Fatalf("unexpected: %+v", out)
	}
	r := out.FeeTierRates[1]
	if r.FeeSchedule != MarginFeeScheduleSelfClearingMembers || r.Tier != 1 || r.MakerFeeRate != 0.0003 || r.TakerFeeRate != 0.001 {
		t.Fatalf("tier 1: %+v", r)
	}
	r = out.FeeTierRates[2]
	if r.FeeSchedule != MarginFeeScheduleKalshiPrime || r.Tier != 3 || r.MakerFeeRate != -0.0001 {
		t.Fatalf("tier 3: %+v", r)
	}
}

func TestPerpsRiskTypes_GetMarginFundingHistoryResponse(t *testing.T) {
	const payload = `{"funding_history": [
		{
			"market_ticker": "KXBTCPERP",
			"funding_time": "2026-01-31T08:00:00Z",
			"funding_rate": 0.0001,
			"mark_price": "65000.0000",
			"funding_amount": "-0.0650",
			"quantity": "0.01",
			"subaccount_number": 3
		},
		{
			"market_ticker": "KXBTCPERP",
			"funding_time": "2026-01-31T16:00:00Z",
			"funding_rate": -0.0002,
			"mark_price": "64000.0000",
			"funding_amount": "0.1280",
			"quantity": "0.01",
			"subaccount_number": null
		}
	]}`
	var out GetMarginFundingHistoryResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.FundingHistory) != 2 {
		t.Fatalf("unexpected: %+v", out)
	}
	e := out.FundingHistory[0]
	if e.MarketTicker != "KXBTCPERP" || e.FundingRate != 0.0001 || e.MarkPrice != "65000.0000" ||
		e.FundingAmount != "-0.0650" || e.Quantity != "0.01" ||
		e.SubaccountNumber == nil || *e.SubaccountNumber != 3 {
		t.Fatalf("entry 0: %+v", e)
	}
	if ts, err := ParseTime(e.FundingTime); err != nil || ts.Unix() != 1769846400 {
		t.Fatalf("funding_time: %v %v", ts, err)
	}
	e = out.FundingHistory[1]
	if e.FundingRate != -0.0002 || e.FundingAmount != "0.1280" || e.SubaccountNumber != nil {
		t.Fatalf("entry 1: %+v", e)
	}
}

func TestPerpsRiskTypes_GetMarginHistoricalFundingRatesResponse(t *testing.T) {
	const payload = `{"funding_rates": [
		{"market_ticker": "KXBTCPERP", "funding_time": "2026-01-31T08:00:00Z", "funding_rate": 0.0001, "mark_price": "65000.0000"}
	]}`
	var out GetMarginHistoricalFundingRatesResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.FundingRates) != 1 {
		t.Fatalf("unexpected: %+v", out)
	}
	r := out.FundingRates[0]
	if r.MarketTicker != "KXBTCPERP" || r.FundingTime != "2026-01-31T08:00:00Z" || r.FundingRate != 0.0001 || r.MarkPrice != "65000.0000" {
		t.Fatalf("rate: %+v", r)
	}
}

func TestPerpsRiskTypes_GetMarginFundingRateEstimateResponse(t *testing.T) {
	const payload = `{
		"market_ticker": "KXBTCPERP",
		"computed_time": "2026-01-31T07:30:00Z",
		"funding_rate": 0,
		"mark_price": "65000.0000",
		"next_funding_time": "2026-01-31T08:00:00Z"
	}`
	var out GetMarginFundingRateEstimateResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.MarketTicker != "KXBTCPERP" || out.ComputedTime != "2026-01-31T07:30:00Z" || out.MarkPrice != "65000.0000" ||
		out.NextFundingTime != "2026-01-31T08:00:00Z" || out.FundingRate == nil || *out.FundingRate != 0 {
		t.Fatalf("unexpected: %+v", out)
	}

	var minimal GetMarginFundingRateEstimateResponse
	if err := json.Unmarshal([]byte(`{"next_funding_time": "2026-01-31T08:00:00Z"}`), &minimal); err != nil {
		t.Fatal(err)
	}
	if minimal.NextFundingTime != "2026-01-31T08:00:00Z" || minimal.FundingRate != nil || minimal.MarketTicker != "" {
		t.Fatalf("minimal: %+v", minimal)
	}
}
