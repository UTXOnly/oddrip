package oddrip

import (
	"context"
	"fmt"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type MarketsService struct {
	client *Client
}

func (s *MarketsService) Get(ctx context.Context, ticker string) (*types.GetMarketResponse, error) {
	var out types.GetMarketResponse
	if err := s.client.get(ctx, joinPath("markets", ticker), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MarketsService) List(ctx context.Context, opts *types.GetMarketsOpts) (*types.GetMarketsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "event_ticker", opts.EventTicker)
		encodeQuery(v, "series_ticker", opts.SeriesTicker)
		encodeQueryInt64(v, "min_created_ts", opts.MinCreatedTs)
		encodeQueryInt64(v, "max_created_ts", opts.MaxCreatedTs)
		encodeQueryInt64(v, "min_updated_ts", opts.MinUpdatedTs)
		encodeQueryInt64(v, "max_close_ts", opts.MaxCloseTs)
		encodeQueryInt64(v, "min_close_ts", opts.MinCloseTs)
		encodeQueryInt64(v, "min_settled_ts", opts.MinSettledTs)
		encodeQueryInt64(v, "max_settled_ts", opts.MaxSettledTs)
		encodeQuery(v, "status", opts.Status)
		encodeQuery(v, "tickers", opts.Tickers)
		encodeQuery(v, "mve_filter", opts.MveFilter)
	}
	var out types.GetMarketsResponse
	if err := s.client.get(ctx, joinPath("markets"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MarketsService) GetOrderbook(ctx context.Context, ticker string, opts *types.GetMarketOrderbookOpts) (*types.GetMarketOrderbookResponse, error) {
	v := url.Values{}
	if opts != nil && opts.Depth > 0 {
		v.Set("depth", fmt.Sprintf("%d", opts.Depth))
	}
	var out types.GetMarketOrderbookResponse
	if err := s.client.get(ctx, joinPath("markets", ticker, "orderbook"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MarketsService) GetTrades(ctx context.Context, opts *types.GetTradesOpts) (*types.GetTradesResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
	}
	var out types.GetTradesResponse
	if err := s.client.get(ctx, joinPath("markets", "trades"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MarketsService) ListHistorical(ctx context.Context, opts *types.GetHistoricalMarketsOpts) (*types.GetMarketsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "tickers", opts.Tickers)
		encodeQuery(v, "event_ticker", opts.EventTicker)
		encodeQuery(v, "mve_filter", opts.MveFilter)
	}
	var out types.GetMarketsResponse
	if err := s.client.get(ctx, joinPath("historical", "markets"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MarketsService) GetHistorical(ctx context.Context, ticker string) (*types.GetMarketResponse, error) {
	var out types.GetMarketResponse
	if err := s.client.get(ctx, joinPath("historical", "markets", ticker), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MarketsService) GetHistoricalTrades(ctx context.Context, opts *types.GetTradesOpts) (*types.GetTradesResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
	}
	var out types.GetTradesResponse
	if err := s.client.get(ctx, joinPath("historical", "trades"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MarketsService) GetHistoricalCandlesticks(ctx context.Context, ticker string, opts *types.GetHistoricalMarketCandlesticksOpts) (*types.GetMarketCandlesticksHistoricalResponse, error) {
	if opts == nil {
		return nil, fmt.Errorf("opts required")
	}
	switch opts.PeriodInterval {
	case 1, 60, 1440:
	default:
		return nil, fmt.Errorf("period_interval must be 1, 60, or 1440")
	}
	v := url.Values{}
	v.Set("start_ts", fmt.Sprintf("%d", opts.StartTs))
	v.Set("end_ts", fmt.Sprintf("%d", opts.EndTs))
	v.Set("period_interval", fmt.Sprintf("%d", opts.PeriodInterval))
	var out types.GetMarketCandlesticksHistoricalResponse
	if err := s.client.get(ctx, joinPath("historical", "markets", ticker, "candlesticks"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
