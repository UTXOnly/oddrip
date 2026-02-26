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

func (s *OrdersService) Create(ctx context.Context, req *types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
	var out types.CreateOrderResponse
	if err := s.client.post(ctx, joinPath("portfolio", "orders"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
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

func (s *OrdersService) Cancel(ctx context.Context, orderID string, subaccount *int) (*types.CancelOrderResponse, error) {
	v := url.Values{}
	if subaccount != nil {
		encodeQueryInt(v, "subaccount", subaccount)
	}
	var out types.CancelOrderResponse
	if err := s.client.delete(ctx, joinPath("portfolio", "orders", orderID), v, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) Amend(ctx context.Context, orderID string, req *types.AmendOrderRequest) (*types.AmendOrderResponse, error) {
	var out types.AmendOrderResponse
	if err := s.client.post(ctx, joinPath("portfolio", "orders", orderID, "amend"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) Decrease(ctx context.Context, orderID string, req *types.DecreaseOrderRequest) (*types.DecreaseOrderResponse, error) {
	var out types.DecreaseOrderResponse
	if err := s.client.post(ctx, joinPath("portfolio", "orders", orderID, "decrease"), req, &out); err != nil {
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

func (s *OrdersService) BatchCreate(ctx context.Context, req *types.BatchCreateOrdersRequest) (*types.BatchCreateOrdersResponse, error) {
	var out types.BatchCreateOrdersResponse
	if err := s.client.post(ctx, joinPath("portfolio", "orders", "batched"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrdersService) BatchCancel(ctx context.Context, req *types.BatchCancelOrdersRequest) (*types.BatchCancelOrdersResponse, error) {
	var out types.BatchCancelOrdersResponse
	if err := s.client.delete(ctx, joinPath("portfolio", "orders", "batched"), nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
