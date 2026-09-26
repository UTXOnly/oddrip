package oddrip

import (
	"context"
	"errors"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

// CreateSubtrader creates a perps subtrader for an FCM member. The server
// composes the ID as {user_id}_{SubtraderSuffix}, so a replay cannot create a
// second subtrader and the call is retried on 5xx and transport errors; a 409
// *APIError after a retry means the first attempt created it.
func (s *PerpsFCMService) CreateSubtrader(ctx context.Context, req *types.CreateMarginFCMSubtraderRequest) (*types.CreateMarginFCMSubtraderResponse, error) {
	if req == nil {
		return nil, errors.New("request required")
	}
	var out types.CreateMarginFCMSubtraderResponse
	if err := s.client.postIdempotent(ctx, joinPath("margin", "fcm", "subtraders"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSubtraderRiskControls returns a subtrader's initial margin caps and its
// admin-set notional value risk limits.
func (s *PerpsFCMService) GetSubtraderRiskControls(ctx context.Context, opts *types.GetFCMSubtraderRiskControlsOpts) (*types.GetFCMSubtraderRiskControlsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "subtrader_id", opts.SubtraderID)
		encodeQuery(v, "market_ticker", opts.MarketTicker)
		encodeQuery(v, "asset_class", opts.AssetClass)
	}
	var out types.GetFCMSubtraderRiskControlsResponse
	if err := s.client.get(ctx, joinPath("margin", "fcm", "subtraders", "risk_controls"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSubtraderRiskControls sets an initial margin cap for a subtrader.
func (s *PerpsFCMService) UpdateSubtraderRiskControls(ctx context.Context, req *types.UpdateFCMSubtraderRiskControlsRequest) error {
	if req == nil {
		return errors.New("request required")
	}
	return s.client.put(ctx, joinPath("margin", "fcm", "subtraders", "risk_controls"), nil, req, nil)
}

// DeleteSubtraderRiskControls removes an initial margin cap for a subtrader.
// opts.SubtraderID is required.
func (s *PerpsFCMService) DeleteSubtraderRiskControls(ctx context.Context, opts *types.DeleteFCMSubtraderRiskControlsOpts) error {
	if opts == nil || opts.SubtraderID == "" {
		return errors.New("subtrader_id required")
	}
	v := url.Values{}
	v.Set("subtrader_id", opts.SubtraderID)
	encodeQuery(v, "market_ticker", opts.MarketTicker)
	encodeQuery(v, "asset_class", opts.AssetClass)
	return s.client.delete(ctx, joinPath("margin", "fcm", "subtraders", "risk_controls"), v, nil, nil)
}

// UpdateNotionalRiskLimit sets the FCM member's own account-level notional
// value risk limit. The exchange enforces the smaller of this and the
// Kalshi-set limit, and rejects a value above the Kalshi-set limit with 400.
// Lowering the limit cancels resting orders until the account is within it.
func (s *PerpsFCMService) UpdateNotionalRiskLimit(ctx context.Context, req *types.UpdateFCMNotionalRiskLimitRequest) error {
	if req == nil {
		return errors.New("request required")
	}
	return s.client.put(ctx, joinPath("margin", "fcm", "notional_risk_limit"), nil, req, nil)
}

// DeleteNotionalRiskLimit clears the FCM member's own notional value risk
// limit. Any Kalshi-set limit stays in force; no orders are canceled.
func (s *PerpsFCMService) DeleteNotionalRiskLimit(ctx context.Context) error {
	return s.client.delete(ctx, joinPath("margin", "fcm", "notional_risk_limit"), nil, nil, nil)
}
