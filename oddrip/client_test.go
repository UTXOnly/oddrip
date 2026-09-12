package oddrip

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/UTXOnly/oddrip/oddrip/internal/auth"
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

	_, err := client.Orders.CancelV2(ctx, "o1", &types.CancelOrderV2Opts{Subaccount: &sub, ExchangeIndex: &ex, MarketTicker: "TICKER-24JAN01"})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.Method != http.MethodDelete || mt.req.URL.Path != "/trade-api/v2/portfolio/events/orders/o1" {
		t.Fatalf("request: %v %v", mt.req.Method, mt.req.URL.Path)
	}
	q := mt.req.URL.Query()
	if q.Get("subaccount") != "2" || q.Get("exchange_index") != "1" || q.Get("market_ticker") != "TICKER-24JAN01" {
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

	sub := 2

	_, err := client.Portfolio.ListHistoricalPositions(ctx, &types.GetHistoricalPositionsOpts{
		Ticker:      "MKT",
		EventTicker: "EVT",
		Subaccount:  &sub,
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
	if q.Get("ticker") != "MKT" || q.Get("event_ticker") != "EVT" || q.Get("subaccount") != "2" || q.Get("limit") != "10" || q.Get("cursor") != "c1" {
		t.Fatalf("query: %v", q)
	}
}

func TestExchangeIndexFilters(t *testing.T) {
	// exchange_index is an optional shard filter on these list endpoints; nil
	// omits it and the server returns every shard.
	cases := []struct {
		name string
		body string
		path string
		call func(ctx context.Context, c *Client, idx *int) error
	}{
		{
			name: "Orders.List",
			body: `{"orders":[],"cursor":""}`,
			path: "/trade-api/v2/portfolio/orders",
			call: func(ctx context.Context, c *Client, idx *int) error {
				_, err := c.Orders.List(ctx, &types.GetOrdersOpts{ExchangeIndex: idx})
				return err
			},
		},
		{
			name: "Portfolio.GetFills",
			body: `{"fills":[],"cursor":""}`,
			path: "/trade-api/v2/portfolio/fills",
			call: func(ctx context.Context, c *Client, idx *int) error {
				_, err := c.Portfolio.GetFills(ctx, &types.GetFillsOpts{ExchangeIndex: idx})
				return err
			},
		},
		{
			name: "Portfolio.GetPositions",
			body: `{"market_positions":[],"event_positions":[],"cursor":""}`,
			path: "/trade-api/v2/portfolio/positions",
			call: func(ctx context.Context, c *Client, idx *int) error {
				_, err := c.Portfolio.GetPositions(ctx, &types.GetPositionsOpts{ExchangeIndex: idx})
				return err
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			idx := 1
			mt := &mockTransport{statusCode: 200, body: []byte(tc.body)}
			client := New(HTTPClient(&http.Client{Transport: mt}))
			if err := tc.call(context.Background(), client, &idx); err != nil {
				t.Fatal(err)
			}
			if mt.req.URL.Path != tc.path {
				t.Fatalf("path: %s", mt.req.URL.Path)
			}
			if got := mt.req.URL.Query().Get("exchange_index"); got != "1" {
				t.Fatalf("exchange_index: %q (query %v)", got, mt.req.URL.Query())
			}

			mt = &mockTransport{statusCode: 200, body: []byte(tc.body)}
			client = New(HTTPClient(&http.Client{Transport: mt}))
			if err := tc.call(context.Background(), client, nil); err != nil {
				t.Fatal(err)
			}
			if _, ok := mt.req.URL.Query()["exchange_index"]; ok {
				t.Fatalf("nil ExchangeIndex should be omitted: %v", mt.req.URL.Query())
			}
		})
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

func TestOrders_CancelAll_RequestPathAndQuery(t *testing.T) {
	mt := &mockTransport{statusCode: 204, body: nil}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()
	sub := 3

	if err := client.Orders.CancelAll(ctx, &types.CancelAllOrdersOpts{Subaccount: &sub}); err != nil {
		t.Fatal(err)
	}
	if mt.req.Method != http.MethodDelete || mt.req.URL.Path != "/trade-api/v2/portfolio/events/orders" {
		t.Fatalf("request: %v %v", mt.req.Method, mt.req.URL.Path)
	}
	if got := mt.req.URL.Query().Get("subaccount"); got != "3" {
		t.Fatalf("subaccount: %q", got)
	}
}

func TestPortfolio_GetBalance_ExchangeIndexQuery(t *testing.T) {
	body := []byte(`{"balance":1,"balance_dollars":"0.0001","portfolio_value":1,"updated_ts":1}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	idx := 1

	if _, err := client.Portfolio.GetBalance(context.Background(), &types.GetBalanceOpts{ExchangeIndex: &idx}); err != nil {
		t.Fatal(err)
	}
	if got := mt.req.URL.Query().Get("exchange_index"); got != "1" {
		t.Fatalf("exchange_index: %q", got)
	}
}

func TestPortfolio_ListIntraExchangeTransfers_Path(t *testing.T) {
	body := []byte(`{"transfers":[{"transfer_id":"t1","source":"event_contract","destination":"margined","source_exchange_shard":0,"destination_exchange_shard":1,"amount":"10.0000","status":"pending","created_ts":1716300000}],"cursor":"next"}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	limit := int64(10)

	got, err := client.Portfolio.ListIntraExchangeTransfers(context.Background(), &types.GetIntraExchangeTransfersOpts{Limit: &limit, Cursor: "c1"})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Path != "/trade-api/v2/portfolio/intra_exchange_instance_transfers" {
		t.Fatalf("path: %v", mt.req.URL.Path)
	}
	q := mt.req.URL.Query()
	if q.Get("limit") != "10" || q.Get("cursor") != "c1" {
		t.Fatalf("query: %v", q)
	}
	if len(got.Transfers) != 1 || got.Transfers[0].Status != types.IntraExchangeInstanceTransferStatusPending {
		t.Fatalf("transfers: %+v", got.Transfers)
	}
}

func TestPortfolio_GetIntraExchangeTransfer_Path(t *testing.T) {
	body := []byte(`{"transfer":{"transfer_id":"t1","source":"event_contract","destination":"margined","source_exchange_shard":0,"destination_exchange_shard":1,"amount":"10.0000","status":"complete","created_ts":1}}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))

	got, err := client.Portfolio.GetIntraExchangeTransfer(context.Background(), "t1")
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Path != "/trade-api/v2/portfolio/intra_exchange_instance_transfers/t1" {
		t.Fatalf("path: %v", mt.req.URL.Path)
	}
	if got.Transfer.Amount != "10.0000" {
		t.Fatalf("transfer: %+v", got.Transfer)
	}
}

func TestPortfolio_TargetBalanceAllocation(t *testing.T) {
	body := []byte(`{"allocations":[{"exchange_index":0,"percent":60},{"exchange_index":1,"percent":40}]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	ctx := context.Background()

	got, err := client.Portfolio.GetTargetBalanceAllocation(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Path != "/trade-api/v2/portfolio/target_balance_allocation" {
		t.Fatalf("path: %v", mt.req.URL.Path)
	}
	if len(got.Allocations) != 2 || got.Allocations[1].Percent != 40 {
		t.Fatalf("allocations: %+v", got.Allocations)
	}

	mt2 := &mockTransport{statusCode: 200, body: []byte(`{}`)}
	client2 := New(HTTPClient(&http.Client{Transport: mt2}))
	err = client2.Portfolio.SetTargetBalanceAllocation(ctx, &types.SetTargetBalanceAllocationRequest{
		Allocations:              []types.TargetBalanceAllocation{{ExchangeIndex: 0, Percent: 100}},
		RestingMarginReservation: types.RestingMarginReservationMax,
	})
	if err != nil {
		t.Fatal(err)
	}
	if mt2.req.Method != http.MethodPost || mt2.req.URL.Path != "/trade-api/v2/portfolio/target_balance_allocation" {
		t.Fatalf("request: %v %v", mt2.req.Method, mt2.req.URL.Path)
	}
}

func TestPortfolio_SetTargetBalanceAllocation_PercentValidation(t *testing.T) {
	mt := &mockTransport{statusCode: 200, body: []byte(`{}`)}
	client := New(HTTPClient(&http.Client{Transport: mt}))

	err := client.Portfolio.SetTargetBalanceAllocation(context.Background(), &types.SetTargetBalanceAllocationRequest{
		Allocations: []types.TargetBalanceAllocation{{ExchangeIndex: 0, Percent: 60}},
	})
	if err == nil {
		t.Fatal("expected error for allocations that do not total 100")
	}
	if mt.req != nil {
		t.Fatalf("request should not be sent: %v", mt.req.URL)
	}
}

func TestLiveData_GetWeatherIndex_RequestPathAndQuery(t *testing.T) {
	body := []byte(`{"city":"miami","config_version":"v3","units":"F","timeseries":[{"t":1716300000000,"v":81.25,"status":"ok","contributors":3}]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	from := int64(1716300000000)
	to := int64(1716303600000)
	detailed := true

	got, err := client.LiveData.GetWeatherIndex(context.Background(), "miami", &types.GetWeatherIndexOpts{From: &from, To: &to, Detailed: &detailed})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Path != "/trade-api/v2/live_data/weather/miami" {
		t.Fatalf("path: %v", mt.req.URL.Path)
	}
	q := mt.req.URL.Query()
	if q.Get("from") != "1716300000000" || q.Get("to") != "1716303600000" || q.Get("detailed") != "true" {
		t.Fatalf("query: %v", q)
	}
	if len(got.Timeseries) != 1 || got.Timeseries[0].V == nil || *got.Timeseries[0].V != 81.25 {
		t.Fatalf("timeseries: %+v", got.Timeseries)
	}
}

func TestLiveData_GetWeatherIndexCalibrations_Path(t *testing.T) {
	body := []byte(`{"city":"miami","units":"F","calibrations":[{"config_version":"v3","effective_at_ms":1716300000000,"city_reference_c":27.5,"stations":[{"station_id":"KMIA","weight":0.5,"offset_c":-0.12}]}]}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))

	got, err := client.LiveData.GetWeatherIndexCalibrations(context.Background(), "miami")
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Path != "/trade-api/v2/live_data/weather/miami/calibrations" {
		t.Fatalf("path: %v", mt.req.URL.Path)
	}
	if len(got.Calibrations) != 1 || got.Calibrations[0].Stations[0].StationID != "KMIA" {
		t.Fatalf("calibrations: %+v", got.Calibrations)
	}
}

func TestLiveData_GetEvent_RequestPathAndQuery(t *testing.T) {
	body := []byte(`{"live_data":{"type":"weather_observations","details":{"city":"miami"},"default_range":"1h","range_options":["15min","1h"]}}`)
	mt := &mockTransport{statusCode: 200, body: body}
	client := New(HTTPClient(&http.Client{Transport: mt}))

	got, err := client.LiveData.GetEvent(context.Background(), "KXHIGHMIA-25SEP06", &types.GetEventLiveDataOpts{Range: "1h"})
	if err != nil {
		t.Fatal(err)
	}
	if mt.req.URL.Path != "/trade-api/v2/live_data/events/KXHIGHMIA-25SEP06" {
		t.Fatalf("path: %v", mt.req.URL.Path)
	}
	if got := mt.req.URL.Query().Get("range"); got != "1h" {
		t.Fatalf("range: %q", got)
	}
	if got.LiveData.Type != "weather_observations" || string(got.LiveData.Details) != `{"city":"miami"}` {
		t.Fatalf("live_data: %+v", got.LiveData)
	}
}

func TestClient_RetryExhaustedReturnsAPIError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"code":"rate_limited","message":"slow down"}`))
	}))
	defer srv.Close()
	client := New(BaseURL(srv.URL), RetryConfigOption(RetryConfig{
		MaxAttempts:  2,
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Millisecond,
	}))

	_, err := client.Exchange.GetStatus(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 429 || apiErr.Code != "rate_limited" || apiErr.Message != "slow down" {
		t.Fatalf("APIError: %+v", apiErr)
	}
	if n := atomic.LoadInt32(&calls); n != 2 {
		t.Fatalf("calls: %d", n)
	}
}

func TestClient_RetryWaitHonorsContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	client := New(BaseURL(srv.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := client.Exchange.GetStatus(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("took %v", elapsed)
	}
}

// path.Join would drop an empty segment and route the call to a different
// endpoint; the worst case is CancelV2 with an empty order ID becoming
// DELETE /portfolio/events/orders, which is CancelAll.
func TestEmptyPathParam_Refused(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	calls := map[string]func() error{
		"Orders.CancelV2":    func() error { _, err := client.Orders.CancelV2(ctx, "", nil); return err },
		"Orders.Get":         func() error { _, err := client.Orders.Get(ctx, ""); return err },
		"Markets.Get":        func() error { _, err := client.Markets.Get(ctx, ""); return err },
		"Events.Get":         func() error { _, err := client.Events.Get(ctx, "", nil); return err },
		"Series.Get":         func() error { _, err := client.Series.Get(ctx, "", nil); return err },
		"OrderGroups.Delete": func() error { return client.OrderGroups.Delete(ctx, "", nil) },
		"Series.GetMarketCandlesticks(dotdot)": func() error {
			_, err := client.Series.GetMarketCandlesticks(ctx, "KXHIGHNY", "..", &types.GetMarketCandlesticksOpts{StartTs: 1, EndTs: 2, PeriodInterval: 60})
			return err
		},
	}
	for name, call := range calls {
		ct.req = nil
		err := call()
		if !errors.Is(err, ErrEmptyPathParam) {
			t.Errorf("%s: err = %v, want ErrEmptyPathParam", name, err)
		}
		if ct.req != nil {
			t.Errorf("%s: request was sent: %s %s", name, ct.req.Method, ct.req.URL.Path)
		}
	}
}

func TestJoinPath_EscapesSegments(t *testing.T) {
	if got := joinPath("markets", "FED-23DEC-T3.00", "orderbook"); got != "/markets/FED-23DEC-T3.00/orderbook" {
		t.Errorf("joinPath = %q", got)
	}
	if got := joinPath("live_data", "weather", "a/b c"); got != "/live_data/weather/a%2Fb%20c" {
		t.Errorf("joinPath escaped = %q", got)
	}
}

// countingServer answers every request with status until the caller's ctx
// ends, recording how many attempts the client made.
func countingServer(t *testing.T, status int) (*Client, *int32) {
	t.Helper()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(status)
		w.Write([]byte(`{"code":"boom","message":"boom"}`))
	}))
	t.Cleanup(srv.Close)
	client := New(BaseURL(srv.URL), RetryConfigOption(RetryConfig{
		MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond,
	}))
	return client, &calls
}

// Writes that Kalshi cannot deduplicate must not be replayed on an ambiguous
// failure; writes it can (or reads) keep the full retry policy.
func TestRetryPolicy_ByIdempotency(t *testing.T) {
	ctx := context.Background()
	order := func(clientOrderID string) *types.CreateOrderV2Request {
		return &types.CreateOrderV2Request{Ticker: "T", ClientOrderID: clientOrderID, Side: "yes", Count: "1", Price: "0.50"}
	}
	cases := []struct {
		name      string
		status    int
		call      func(*Client) error
		wantCalls int32
	}{
		{"GET 503 retried", 503, func(c *Client) error { _, err := c.Exchange.GetStatus(ctx); return err }, 3},
		{"DELETE cancel 503 retried", 503, func(c *Client) error { _, err := c.Orders.CancelV2(ctx, "o1", nil); return err }, 3},
		{"PUT reset 503 retried", 503, func(c *Client) error { return c.OrderGroups.Reset(ctx, "g1", nil) }, 3},
		{"create with client_order_id 503 retried", 503, func(c *Client) error { _, err := c.Orders.CreateV2(ctx, order("cid-1")); return err }, 3},
		{"create without client_order_id 503 NOT retried", 503, func(c *Client) error { _, err := c.Orders.CreateV2(ctx, order("")); return err }, 1},
		{"create without client_order_id 429 retried", 429, func(c *Client) error { _, err := c.Orders.CreateV2(ctx, order("")); return err }, 3},
		{"decrease 502 NOT retried", 502, func(c *Client) error {
			rb := "1"
			_, err := c.Orders.DecreaseV2(ctx, "o1", &types.DecreaseOrderV2Request{ReduceBy: &rb}, nil)
			return err
		}, 1},
		{"amend 500 NOT retried", 500, func(c *Client) error {
			_, err := c.Orders.AmendV2(ctx, "o1", &types.AmendOrderV2Request{}, nil)
			return err
		}, 1},
		{"batch create all keyed 503 retried", 503, func(c *Client) error {
			_, err := c.Orders.BatchCreateV2(ctx, &types.BatchCreateOrdersV2Request{Orders: []types.CreateOrderV2Request{*order("a"), *order("b")}})
			return err
		}, 3},
		{"batch create one unkeyed 503 NOT retried", 503, func(c *Client) error {
			_, err := c.Orders.BatchCreateV2(ctx, &types.BatchCreateOrdersV2Request{Orders: []types.CreateOrderV2Request{*order("a"), *order("")}})
			return err
		}, 1},
		{"transfer 503 retried", 503, func(c *Client) error {
			return c.Subaccounts.Transfer(ctx, &types.ApplySubaccountTransferRequest{ClientTransferID: "t1", FromSubaccount: 0, ToSubaccount: 1, AmountCents: 100})
		}, 3},
		{"create order group 503 NOT retried", 503, func(c *Client) error {
			lim := int64(10)
			_, err := c.OrderGroups.Create(ctx, &types.CreateOrderGroupRequest{ContractsLimit: &lim})
			return err
		}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, calls := countingServer(t, tc.status)
			err := tc.call(client)
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status {
				t.Fatalf("err = %v, want *APIError %d", err, tc.status)
			}
			if got := atomic.LoadInt32(calls); got != tc.wantCalls {
				t.Fatalf("attempts = %d, want %d", got, tc.wantCalls)
			}
		})
	}
}

// A dropped connection with no response is the replay case that matters most.
func TestRetryPolicy_TransportError(t *testing.T) {
	ctx := context.Background()
	newClient := func(t *testing.T) (*Client, *int32) {
		t.Helper()
		var calls int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("no hijacker")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatal(err)
			}
			conn.Close() // reset with no response
		}))
		t.Cleanup(srv.Close)
		return New(BaseURL(srv.URL), RetryConfigOption(RetryConfig{
			MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond,
		})), &calls
	}

	client, calls := newClient(t)
	if _, err := client.Exchange.GetStatus(ctx); err == nil {
		t.Fatal("expected transport error")
	}
	if got := atomic.LoadInt32(calls); got != 3 {
		t.Fatalf("GET attempts = %d, want 3", got)
	}

	client, calls = newClient(t)
	rb := "1"
	_, err := client.Orders.DecreaseV2(ctx, "o1", &types.DecreaseOrderV2Request{ReduceBy: &rb}, nil)
	if err == nil {
		t.Fatal("expected transport error")
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		t.Fatalf("transport error must not be an *APIError: %v", err)
	}
	if got := atomic.LoadInt32(calls); got != 1 {
		t.Fatalf("DecreaseV2 attempts = %d, want 1 (no replay)", got)
	}
}

func TestDefaultBaseURL(t *testing.T) {
	mt := &mockTransport{statusCode: 200, body: []byte(`{"exchange_active":true,"trading_active":true}`)}
	client := New(HTTPClient(&http.Client{Transport: mt}))
	if _, err := client.Exchange.GetStatus(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := mt.req.URL.String(), "https://external-api.kalshi.com/trade-api/v2/exchange/status"; got != want {
		t.Fatalf("request URL = %q, want %q", got, want)
	}
}

// The signed message is timestamp + METHOD + path, so switching hosts (the
// recommended external-api host, the shared api.elections host, demo) must
// not change what is signed.
func TestSignedPathExcludesHost(t *testing.T) {
	var signed []string
	signer := &auth.KalshiSigner{KeyID: "k", SignRequest: func(method, path string, _ int64) (string, error) {
		signed = append(signed, method+" "+path)
		return "sig", nil
	}}
	for _, base := range []string{"", "https://api.elections.kalshi.com/trade-api/v2", "https://demo-api.kalshi.co/trade-api/v2/"} {
		mt := &mockTransport{statusCode: 200, body: []byte(`{"balance":0}`)}
		opts := []Option{HTTPClient(&http.Client{Transport: mt}), Auth(signer)}
		if base != "" {
			opts = append(opts, BaseURL(base))
		}
		client := New(opts...)
		if _, err := client.Portfolio.GetBalance(context.Background(), nil); err != nil {
			t.Fatalf("base %q: unexpected error: %v", base, err)
		}
		if got := mt.req.Header.Get("KALSHI-ACCESS-SIGNATURE"); got != "sig" {
			t.Fatalf("base %q: signature header = %q", base, got)
		}
	}
	for i, got := range signed {
		if want := "GET /trade-api/v2/portfolio/balance"; got != want {
			t.Fatalf("signed[%d] = %q, want %q", i, got, want)
		}
	}
	if len(signed) != 3 {
		t.Fatalf("signed %d requests, want 3", len(signed))
	}
}
