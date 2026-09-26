package oddrip

import (
	"context"
	"errors"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

// GetParameters returns the exchange-wide liquidation thresholds and
// per-market initial margin multipliers.
func (s *PerpsRiskService) GetParameters(ctx context.Context) (*types.GetMarginRiskParametersResponse, error) {
	var out types.GetMarginRiskParametersResponse
	if err := s.client.get(ctx, joinPath("margin", "risk_parameters"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNotionalRiskLimit returns the account's notional value risk limits and
// current usage.
func (s *PerpsRiskService) GetNotionalRiskLimit(ctx context.Context) (*types.NotionalRiskLimitResponse, error) {
	var out types.NotionalRiskLimitResponse
	if err := s.client.get(ctx, joinPath("margin", "notional_risk_limit"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns account leverage plus per-position leverage and estimated
// liquidation prices.
func (s *PerpsRiskService) Get(ctx context.Context) (*types.GetMarginRiskResponse, error) {
	var out types.GetMarginRiskResponse
	if err := s.client.get(ctx, joinPath("margin", "risk"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTiers returns the account's effective maker and taker fee rate per market.
func (s *PerpsFeesService) GetTiers(ctx context.Context) (*types.GetMarginFeeTiersResponse, error) {
	var out types.GetMarginFeeTiersResponse
	if err := s.client.get(ctx, joinPath("margin", "fee_tiers"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTierRates returns the fee schedule that applies to the account.
func (s *PerpsFeesService) GetTierRates(ctx context.Context) (*types.GetMarginFeeTierRatesResponse, error) {
	var out types.GetMarginFeeTierRatesResponse
	if err := s.client.get(ctx, joinPath("margin", "fee_tier_rates"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHistory returns the account's funding payments over an inclusive UTC
// date range. opts.StartDate and opts.EndDate are required.
func (s *PerpsFundingService) GetHistory(ctx context.Context, opts *types.GetMarginFundingHistoryOpts) (*types.GetMarginFundingHistoryResponse, error) {
	if opts == nil || opts.StartDate == "" || opts.EndDate == "" {
		return nil, errors.New("start_date and end_date required")
	}
	v := url.Values{}
	encodeQuery(v, "ticker", opts.Ticker)
	v.Set("start_date", opts.StartDate)
	v.Set("end_date", opts.EndDate)
	encodeQueryInt(v, "subaccount", opts.Subaccount)
	var out types.GetMarginFundingHistoryResponse
	if err := s.client.get(ctx, joinPath("margin", "funding_history"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHistoricalRates returns past funding rates for one market, or all markets
// when opts is nil or opts.Ticker is empty.
func (s *PerpsFundingService) GetHistoricalRates(ctx context.Context, opts *types.GetMarginHistoricalFundingRatesOpts) (*types.GetMarginHistoricalFundingRatesResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQueryInt64(v, "start_ts", opts.StartTs)
		encodeQueryInt64(v, "end_ts", opts.EndTs)
	}
	var out types.GetMarginHistoricalFundingRatesResponse
	if err := s.client.get(ctx, joinPath("margin", "funding_rates", "historical"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRateEstimate returns the running funding rate estimate for the
// in-progress period. opts.Ticker is required.
func (s *PerpsFundingService) GetRateEstimate(ctx context.Context, opts *types.GetMarginFundingRateEstimateOpts) (*types.GetMarginFundingRateEstimateResponse, error) {
	if opts == nil || opts.Ticker == "" {
		return nil, errors.New("ticker required")
	}
	v := url.Values{}
	v.Set("ticker", opts.Ticker)
	var out types.GetMarginFundingRateEstimateResponse
	if err := s.client.get(ctx, joinPath("margin", "funding_rates", "estimate"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
