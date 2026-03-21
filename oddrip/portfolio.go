package oddrip

import (
	"context"
	"fmt"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
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

func (s *PortfolioService) ListSettlements(ctx context.Context, opts *types.GetSettlementsOpts) (*types.GetSettlementsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQuery(v, "event_ticker", opts.EventTicker)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetSettlementsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "settlements"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) ListHistoricalFills(ctx context.Context, opts *types.GetHistoricalArchiveOpts) (*types.GetFillsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetFillsResponse
	if err := s.client.get(ctx, joinPath("historical", "fills"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) ListHistoricalOrders(ctx context.Context, opts *types.GetHistoricalArchiveOpts) (*types.GetOrdersResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetOrdersResponse
	if err := s.client.get(ctx, joinPath("historical", "orders"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
