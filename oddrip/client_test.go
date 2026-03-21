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
