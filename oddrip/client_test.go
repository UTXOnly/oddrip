package oddrip

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
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
