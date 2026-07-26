package oddrip

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type mockTransport struct {
	statusCode int
	body       []byte
	req        *http.Request
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.req = req
	resp := &http.Response{
		StatusCode: m.statusCode,
		Header:     make(http.Header),
		Body:       &mockBody{data: m.body},
		Request:    req,
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

type mockBody struct {
	data []byte
	pos  int
}

func (b *mockBody) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *mockBody) Close() error { return nil }

func TestExchange_GetStatus_Success(t *testing.T) {
	want := map[string]interface{}{
		"exchange_active": true,
		"trading_active":  true,
	}
	body, _ := json.Marshal(want)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(
		HTTPClient(&http.Client{Transport: mt}),
	)
	ctx := context.Background()

	got, err := client.Exchange.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if !got.ExchangeActive || !got.TradingActive {
		t.Errorf("GetStatus: got ExchangeActive=%v TradingActive=%v", got.ExchangeActive, got.TradingActive)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/exchange/status" {
		t.Errorf("request path: got %v", mt.req.URL.Path)
	}
}

func TestExchange_GetStatus_APIError(t *testing.T) {
	body := []byte(`{"code":"NOT_FOUND","message":"resource not found"}`)
	mt := &mockTransport{statusCode: 404, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Exchange.GetStatus(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 || apiErr.Message != "resource not found" {
		t.Errorf("APIError: status=%d message=%s", apiErr.StatusCode, apiErr.Message)
	}
}

func TestMarkets_ListHistorical_RequestPath(t *testing.T) {
	body := []byte(`{"markets":[],"cursor":""}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Markets.ListHistorical(ctx, &types.GetHistoricalMarketsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/historical/markets" {
		t.Fatalf("path: %v", mt.req)
	}
}

func TestMarkets_GetHistoricalTrades_RequestPath(t *testing.T) {
	body := []byte(`{"trades":[],"cursor":""}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Markets.GetHistoricalTrades(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/historical/trades" {
		t.Fatalf("path: %v", mt.req)
	}
}

func TestMarkets_GetHistoricalCandlesticks_QueryAndPath(t *testing.T) {
	body := []byte(`{"ticker":"X","candlesticks":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Markets.GetHistoricalCandlesticks(ctx, "X", &types.GetHistoricalMarketCandlesticksOpts{
		StartTs:        1,
		EndTs:          2,
		PeriodInterval: 60,
	})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/historical/markets/X/candlesticks" {
		t.Fatalf("path: %v", mt.req)
	}
	q := mt.req.URL.Query()
	if q.Get("start_ts") != "1" || q.Get("end_ts") != "2" || q.Get("period_interval") != "60" {
		t.Fatalf("query: %v", q)
	}
}

func TestMarkets_GetHistoricalCandlesticks_InvalidPeriod(t *testing.T) {
	client := New()
	ctx := context.Background()
	_, err := client.Markets.GetHistoricalCandlesticks(ctx, "X", &types.GetHistoricalMarketCandlesticksOpts{
		StartTs:        1,
		EndTs:          2,
		PeriodInterval: 99,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPortfolio_ListSettlements_RequestPath(t *testing.T) {
	body := []byte(`{"settlements":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Portfolio.ListSettlements(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/portfolio/settlements" {
		t.Fatalf("path: %v", mt.req)
	}
}

func TestPortfolio_ListHistoricalFills_RequestPath(t *testing.T) {
	body := []byte(`{"fills":[],"cursor":""}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Portfolio.ListHistoricalFills(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/historical/fills" {
		t.Fatalf("path: %v", mt.req)
	}
}

func TestAccount_GetEndpointCosts_RequestPath(t *testing.T) {
	body := []byte(`{"default_cost":10,"endpoint_costs":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Account.GetEndpointCosts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/account/endpoint_costs" {
		t.Fatalf("path: %v", mt.req)
	}
}

func TestMarkets_GetOrderbooks_QueryAndPath(t *testing.T) {
	body := []byte(`{"orderbooks":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Markets.GetOrderbooks(ctx, &types.GetMarketOrderbooksOpts{
		Tickers: []string{"A", "B"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/markets/orderbooks" {
		t.Fatalf("path: %v", mt.req)
	}
	q := mt.req.URL.Query()
	if len(q["tickers"]) != 2 || q["tickers"][0] != "A" || q["tickers"][1] != "B" {
		t.Fatalf("query: %v", q)
	}
}

func TestAccount_GetAPILimits_RequestPath(t *testing.T) {
	body := []byte(`{"usage_tier":"basic","read":{"refill_rate":20,"bucket_capacity":20},"write":{"refill_rate":10,"bucket_capacity":10},"grants":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	got, err := client.Account.GetAPILimits(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/account/limits" {
		t.Fatalf("path: %v", mt.req)
	}
	if got.Read.RefillRate != 20 || got.Write.BucketCapacity != 10 {
		t.Fatalf("limits: %+v", got)
	}
}

func TestPortfolio_ListDeposits_RequestPathAndQuery(t *testing.T) {
	body := []byte(`{"deposits":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()
	limit := int64(5)

	_, err := client.Portfolio.ListDeposits(ctx, &types.GetDepositsOpts{Limit: &limit, Cursor: "c1"})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/portfolio/deposits" {
		t.Fatalf("path: %v", mt.req)
	}
	q := mt.req.URL.Query()
	if q.Get("limit") != "5" || q.Get("cursor") != "c1" {
		t.Fatalf("query: %v", q)
	}
}

func TestPortfolio_ListWithdrawals_RequestPath(t *testing.T) {
	body := []byte(`{"withdrawals":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Portfolio.ListWithdrawals(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/portfolio/withdrawals" {
		t.Fatalf("path: %v", mt.req)
	}
}

func TestMarkets_ListHistorical_SeriesTickerQuery(t *testing.T) {
	body := []byte(`{"markets":[],"cursor":""}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Markets.ListHistorical(ctx, &types.GetHistoricalMarketsOpts{SeriesTicker: "SERIES-A"})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Query().Get("series_ticker") != "SERIES-A" {
		t.Fatalf("query: %v", mt.req.URL.Query())
	}
}

func TestMarkets_GetOrderbooks_EmptyTickers(t *testing.T) {
	client := New()
	ctx := context.Background()
	_, err := client.Markets.GetOrderbooks(ctx, &types.GetMarketOrderbooksOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOrders_CreateV2_RequestPath(t *testing.T) {
	body := []byte(`{"order_id":"x","fill_count":"0.00","remaining_count":"1.00","ts_ms":1}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Orders.CreateV2(ctx, &types.CreateOrderV2Request{
		Ticker:                  "MKT",
		ClientOrderID:           "c1",
		Side:                    types.BookSideBid,
		Count:                   "1.00",
		Price:                   "0.5000",
		TimeInForce:             types.TimeInForceGTC,
		SelfTradePreventionType: types.SelfTradeTakerAtCross,
	})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.Method != http.MethodPost || mt.req.URL.Path != "/trade-api/v2/portfolio/events/orders" {
		t.Fatalf("request: %v %v", mt.req.Method, mt.req.URL.Path)
	}
}

func TestOrders_CancelV2_RequestPathAndQuery(t *testing.T) {
	body := []byte(`{"order_id":"o1","reduced_by":"1.00","ts_ms":1}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()
	sub := 2
	ex := 1

	_, err := client.Orders.CancelV2(ctx, "o1", &sub, &ex)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.Method != http.MethodDelete || mt.req.URL.Path != "/trade-api/v2/portfolio/events/orders/o1" {
		t.Fatalf("request: %v %v", mt.req.Method, mt.req.URL.Path)
	}
	q := mt.req.URL.Query()
	if q.Get("subaccount") != "2" || q.Get("exchange_index") != "1" {
		t.Fatalf("query: %v", q)
	}
}

func TestOrders_AmendV2_RequestPathAndQuery(t *testing.T) {
	body := []byte(`{"order_id":"o1","ts_ms":1}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()
	sub := 1

	_, err := client.Orders.AmendV2(ctx, "o1", &types.AmendOrderV2Request{
		Ticker: "MKT",
		Side:   types.BookSideAsk,
		Price:  "0.6000",
		Count:  "2.00",
	}, &sub)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.Method != http.MethodPost || mt.req.URL.Path != "/trade-api/v2/portfolio/events/orders/o1/amend" {
		t.Fatalf("request: %v %v", mt.req.Method, mt.req.URL.Path)
	}
	if mt.req.URL.Query().Get("subaccount") != "1" {
		t.Fatalf("query: %v", mt.req.URL.Query())
	}
}

func TestOrders_DecreaseV2_RequestPathAndQuery(t *testing.T) {
	body := []byte(`{"order_id":"o1","remaining_count":"1.00","ts_ms":1}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()
	reduceBy := "1.00"

	_, err := client.Orders.DecreaseV2(ctx, "o1", &types.DecreaseOrderV2Request{ReduceBy: &reduceBy}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.Method != http.MethodPost || mt.req.URL.Path != "/trade-api/v2/portfolio/events/orders/o1/decrease" {
		t.Fatalf("request: %v %v", mt.req.Method, mt.req.URL.Path)
	}
}

func TestOrders_BatchCreateV2_RequestPath(t *testing.T) {
	body := []byte(`{"orders":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Orders.BatchCreateV2(ctx, &types.BatchCreateOrdersV2Request{Orders: []types.CreateOrderV2Request{}})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.Method != http.MethodPost || mt.req.URL.Path != "/trade-api/v2/portfolio/events/orders/batched" {
		t.Fatalf("request: %v %v", mt.req.Method, mt.req.URL.Path)
	}
}

func TestOrders_BatchCancelV2_RequestPath(t *testing.T) {
	body := []byte(`{"orders":[]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Orders.BatchCancelV2(ctx, &types.BatchCancelOrdersV2Request{
		Orders: []types.BatchCancelOrdersV2RequestOrder{{OrderID: "o1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.Method != http.MethodDelete || mt.req.URL.Path != "/trade-api/v2/portfolio/events/orders/batched" {
		t.Fatalf("request: %v %v", mt.req.Method, mt.req.URL.Path)
	}
}

func TestPortfolio_ListHistoricalPositions_RequestPathAndQuery(t *testing.T) {
	body := []byte(`{"market_positions":[],"event_positions":[],"cursor":""}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()
	limit := int64(10)

	_, err := client.Portfolio.ListHistoricalPositions(ctx, &types.GetHistoricalPositionsOpts{
		Ticker:      "MKT",
		EventTicker: "EVT",
		Limit:       &limit,
		Cursor:      "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req == nil || mt.req.URL.Path != "/trade-api/v2/historical/positions" {
		t.Fatalf("path: %v", mt.req)
	}
	q := mt.req.URL.Query()
	if q.Get("ticker") != "MKT" || q.Get("event_ticker") != "EVT" || q.Get("limit") != "10" || q.Get("cursor") != "c1" {
		t.Fatalf("query: %v", q)
	}
}

func TestMarkets_GetTrades_IsBlockTradeQuery(t *testing.T) {
	body := []byte(`{"trades":[],"cursor":""}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()
	block := true

	_, err := client.Markets.GetTrades(ctx, &types.GetTradesOpts{IsBlockTrade: &block})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Path != "/trade-api/v2/markets/trades" {
		t.Fatalf("path: %v", mt.req.URL.Path)
	}
	if mt.req.URL.Query().Get("is_block_trade") != "true" {
		t.Fatalf("query: %v", mt.req.URL.Query())
	}
}

func TestEvents_List_TickersQuery(t *testing.T) {
	body := []byte(`{"events":[],"cursor":""}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	_, err := client.Events.List(ctx, &types.GetEventsOpts{Tickers: "EVT-A,EVT-B"})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Path != "/trade-api/v2/events" {
		t.Fatalf("path: %v", mt.req.URL.Path)
	}
	if mt.req.URL.Query().Get("tickers") != "EVT-A,EVT-B" {
		t.Fatalf("query: %v", mt.req.URL.Query())
	}
}

func TestExchange_GetStatus_IndexFields(t *testing.T) {
	want := map[string]interface{}{
		"exchange_active":                 true,
		"trading_active":                  true,
		"intra_exchange_transfers_active": true,
		"exchange_index_statuses": []map[string]interface{}{
			{
				"exchange_index":                  0,
				"exchange_active":                 true,
				"trading_active":                  true,
				"intra_exchange_transfers_active": true,
			},
		},
	}
	body, _ := json.Marshal(want)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	got, err := client.Exchange.GetStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.IntraExchangeTransfersActive == nil || !*got.IntraExchangeTransfersActive {
		t.Fatalf("intra: %+v", got.IntraExchangeTransfersActive)
	}
	if len(got.ExchangeIndexStatuses) != 1 {
		t.Fatalf("statuses: %+v", got.ExchangeIndexStatuses)
	}
}
