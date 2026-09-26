package oddrip

import (
	"context"
	"errors"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

// GetBalance returns settled perps funds and a per-subaccount breakdown.
func (s *PerpsPortfolioService) GetBalance(ctx context.Context, opts *types.GetMarginBalanceOpts) (*types.GetMarginBalanceResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryBool(v, "compute_available_balance", opts.ComputeAvailableBalance)
	}
	var out types.GetMarginBalanceResponse
	if err := s.client.get(ctx, joinPath("margin", "balance"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PerpsPortfolioService) GetFills(ctx context.Context, opts *types.GetMarginFillsOpts) (*types.GetMarginFillsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
	}
	var out types.GetMarginFillsResponse
	if err := s.client.get(ctx, joinPath("margin", "fills"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PerpsPortfolioService) GetPositions(ctx context.Context, opts *types.GetMarginPositionsOpts) (*types.GetMarginPositionsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
		encodeQuery(v, "ticker", opts.Ticker)
	}
	var out types.GetMarginPositionsResponse
	if err := s.client.get(ctx, joinPath("margin", "positions"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateSubaccount adds the next numbered perps subaccount (1-63). There is no
// deduplication key, so the call is retried on 429 only: a replay after a 5xx
// or dropped connection could create a second subaccount.
func (s *PerpsPortfolioService) CreateSubaccount(ctx context.Context) (*types.CreateSubaccountResponse, error) {
	var out types.CreateSubaccountResponse
	if err := s.client.post(ctx, joinPath("portfolio", "margin", "subaccounts"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TransferBetweenSubaccounts moves funds between perps subaccounts. Kalshi
// deduplicates on ClientTransferID; when it is set the call is retried on 5xx
// and transport errors like any idempotent request, otherwise only on 429 so a
// dropped connection cannot apply the transfer twice.
func (s *PerpsPortfolioService) TransferBetweenSubaccounts(ctx context.Context, req *types.ApplyMarginSubaccountTransferRequest) error {
	if req == nil {
		return errors.New("request required")
	}
	post := s.client.post
	if req.ClientTransferID != "" {
		post = s.client.postIdempotent
	}
	return post(ctx, joinPath("portfolio", "margin", "subaccounts", "transfer"), req, nil)
}
