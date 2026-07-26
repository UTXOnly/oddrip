package oddrip

import (
	"context"
	"errors"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type OrdersService struct {
	client *Client
}

func (s *OrdersService) Get(ctx context.Context, orderID string) (*types.GetOrderResponse, error) {
	var out types.GetOrderResponse
	if err := s.client.get(ctx, joinPath("portfolio", "orders", orderID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) List(ctx context.Context, opts *types.GetOrdersOpts) (*types.GetOrdersResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQuery(v, "event_ticker", opts.EventTicker)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQuery(v, "status", opts.Status)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetOrdersResponse
	if err := s.client.get(ctx, joinPath("portfolio", "orders"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) GetQueuePosition(ctx context.Context, orderID string) (*types.GetOrderQueuePositionResponse, error) {
	var out types.GetOrderQueuePositionResponse
	if err := s.client.get(ctx, joinPath("portfolio", "orders", orderID, "queue_position"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) GetQueuePositions(ctx context.Context, opts *types.GetOrderQueuePositionsOpts) (*types.GetOrderQueuePositionsResponse, error) {
	if opts == nil || (opts.MarketTickers == "" && opts.EventTicker == "") {
		return nil, errors.New("market_tickers or event_ticker required")
	}
	v := url.Values{}
	encodeQuery(v, "market_tickers", opts.MarketTickers)
	encodeQuery(v, "event_ticker", opts.EventTicker)
	encodeQueryInt(v, "subaccount", opts.Subaccount)
	var out types.GetOrderQueuePositionsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "orders", "queue_positions"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) CreateV2(ctx context.Context, req *types.CreateOrderV2Request) (*types.CreateOrderV2Response, error) {
	var out types.CreateOrderV2Response
	if err := s.client.post(ctx, joinPath("portfolio", "events", "orders"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) CancelV2(ctx context.Context, orderID string, subaccount, exchangeIndex *int) (*types.CancelOrderV2Response, error) {
	v := url.Values{}
	encodeQueryInt(v, "subaccount", subaccount)
	encodeQueryInt(v, "exchange_index", exchangeIndex)
	var out types.CancelOrderV2Response
	if err := s.client.delete(ctx, joinPath("portfolio", "events", "orders", orderID), v, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) AmendV2(ctx context.Context, orderID string, req *types.AmendOrderV2Request, subaccount *int) (*types.AmendOrderV2Response, error) {
	v := url.Values{}
	encodeQueryInt(v, "subaccount", subaccount)
	var out types.AmendOrderV2Response
	if err := s.client.postQuery(ctx, joinPath("portfolio", "events", "orders", orderID, "amend"), v, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) DecreaseV2(ctx context.Context, orderID string, req *types.DecreaseOrderV2Request, subaccount *int) (*types.DecreaseOrderV2Response, error) {
	v := url.Values{}
	encodeQueryInt(v, "subaccount", subaccount)
	var out types.DecreaseOrderV2Response
	if err := s.client.postQuery(ctx, joinPath("portfolio", "events", "orders", orderID, "decrease"), v, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) BatchCreateV2(ctx context.Context, req *types.BatchCreateOrdersV2Request) (*types.BatchCreateOrdersV2Response, error) {
	var out types.BatchCreateOrdersV2Response
	if err := s.client.post(ctx, joinPath("portfolio", "events", "orders", "batched"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) BatchCancelV2(ctx context.Context, req *types.BatchCancelOrdersV2Request) (*types.BatchCancelOrdersV2Response, error) {
	var out types.BatchCancelOrdersV2Response
	if err := s.client.delete(ctx, joinPath("portfolio", "events", "orders", "batched"), nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
