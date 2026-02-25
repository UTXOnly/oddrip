package oddrip

import (
	"context"
	"fmt"
	"net/url"

	"github.com/oddrip/client/oddrip/types"
)

type PortfolioService struct {
	client *Client
}

func (s *PortfolioService) GetBalance(ctx context.Context, opts *types.GetBalanceOpts) (*types.GetBalanceResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetBalanceResponse
	if err := s.client.get(ctx, joinPath("portfolio", "balance"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) GetFills(ctx context.Context, opts *types.GetFillsOpts) (*types.GetFillsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQuery(v, "order_id", opts.OrderID)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetFillsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "fills"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) GetPositions(ctx context.Context, opts *types.GetPositionsOpts) (*types.GetPositionsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "count_filter", opts.CountFilter)
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQuery(v, "event_ticker", opts.EventTicker)
		encodeQueryInt(v, "subaccount", opts.Subaccount)
		if opts.Limit != nil {
			v.Set("limit", fmt.Sprintf("%d", *opts.Limit))
		}
	}
	var out types.GetPositionsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "positions"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
