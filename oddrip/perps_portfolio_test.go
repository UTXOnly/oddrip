package oddrip

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

func TestPerpsPortfolio_GetBalance(t *testing.T) {
	client, ct := newCaptureClient(200, `{"subaccount_balances":[
		{"subaccount":0,"position_value":"125.5000","account_equity":"1000.0000","maintenance_margin":"40.0000",
		 "initial_margin":"80.0000","resting_orders_margin":"12.0000","available_balance":"908.0000"},
		{"subaccount":1,"position_value":"0.0000","account_equity":"50.0000","maintenance_margin":"0.0000",
		 "initial_margin":"0.0000","resting_orders_margin":"0.0000","available_balance":"50.0000"}
	],"settled_funds":"1050.0000"}`)
	got, err := client.Perps.Portfolio.GetBalance(context.Background(), &types.GetMarginBalanceOpts{ComputeAvailableBalance: ptrOf(true)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/balance")
	ct.assertQuery(t, map[string]string{"compute_available_balance": "true"})
	if got.SettledFunds != "1050.0000" || len(got.SubaccountBalances) != 2 ||
		got.SubaccountBalances[0].AvailableBalance != "908.0000" || got.SubaccountBalances[1].Subaccount != 1 {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, `{"subaccount_balances":[],"settled_funds":"0.0000"}`)
	if _, err := client.Perps.Portfolio.GetBalance(context.Background(), &types.GetMarginBalanceOpts{ComputeAvailableBalance: ptrOf(false)}); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{"compute_available_balance": "false"})

	client, ct = newCaptureClient(200, `{"subaccount_balances":[],"settled_funds":"0.0000"}`)
	if _, err := client.Perps.Portfolio.GetBalance(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/balance")
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsPortfolio_GetFills(t *testing.T) {
	client, ct := newCaptureClient(200, `{"fills":[{
		"fill_id":"f1","order_id":"o1","is_taker":true,"side":"bid","count":"2.50","created_time":"2026-01-02T03:04:05Z",
		"ticker":"KXBTCPERP","price":"0.5600","entry_price":"0.5500","fees":"0.0100","realized_pnl":"0.0000","order_source":"user"
	}],"cursor":"c2"}`)
	got, err := client.Perps.Portfolio.GetFills(context.Background(), &types.GetMarginFillsOpts{
		Subaccount: ptrOf(0), Limit: ptrOf[int64](500), Cursor: "c1", MinTs: ptrOf[int64](1700000000), MaxTs: ptrOf[int64](1700086400),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/fills")
	ct.assertQuery(t, map[string]string{
		"subaccount": "0", "limit": "500", "cursor": "c1", "min_ts": "1700000000", "max_ts": "1700086400",
	})
	if len(got.Fills) != 1 || got.Cursor != "c2" || got.Fills[0].FillID != "f1" || got.Fills[0].Side != types.BookSideBid ||
		got.Fills[0].Count != "2.50" || got.Fills[0].OrderSource != types.OrderSourceUser {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, `{"fills":[],"cursor":""}`)
	if _, err := client.Perps.Portfolio.GetFills(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/fills")
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsPortfolio_GetPositions(t *testing.T) {
	client, ct := newCaptureClient(200, `{"positions":[
		{"subaccount":2,"market_ticker":"KXBTCPERP","position":"-3.00","entry_price":"0.5600","unrealized_pnl":"1.2000",
		 "margin_used":"10.0000","fees":"0.0300","roe":12.0,"is_portfolio":false},
		{"subaccount":2,"market_ticker":"TEST-PERP","position":"1.00","entry_price":"0.4000","unrealized_pnl":"0.0000",
		 "margin_used":null,"fees":"0.0100","roe":null,"is_portfolio":true}
	]}`)
	got, err := client.Perps.Portfolio.GetPositions(context.Background(), &types.GetMarginPositionsOpts{Subaccount: ptrOf(2), Ticker: "KXBTCPERP"})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/positions")
	ct.assertQuery(t, map[string]string{"subaccount": "2", "ticker": "KXBTCPERP"})
	if len(got.Positions) != 2 || got.Positions[0].Position != "-3.00" || got.Positions[0].MarginUsed == nil ||
		got.Positions[0].ROE == nil || *got.Positions[0].ROE != 12.0 || got.Positions[1].MarginUsed != nil || !got.Positions[1].IsPortfolio {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, `{"positions":[]}`)
	if _, err := client.Perps.Portfolio.GetPositions(context.Background(), &types.GetMarginPositionsOpts{Subaccount: ptrOf(0)}); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{"subaccount": "0"})

	client, ct = newCaptureClient(200, `{"positions":[]}`)
	if _, err := client.Perps.Portfolio.GetPositions(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/positions")
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsPortfolio_CreateSubaccount(t *testing.T) {
	client, ct := newCaptureClient(201, `{"subaccount_number":3}`)
	got, err := client.Perps.Portfolio.CreateSubaccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/portfolio/margin/subaccounts")
	ct.assertQuery(t, map[string]string{})
	if len(ct.sent) != 0 {
		t.Fatalf("body should be empty, got %q", ct.sent)
	}
	if got.SubaccountNumber != 3 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsPortfolio_TransferBetweenSubaccounts(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	err := client.Perps.Portfolio.TransferBetweenSubaccounts(context.Background(), &types.ApplyMarginSubaccountTransferRequest{
		ClientTransferID: "8c35ecb3-328f-4f52-8c7c-0f4b9862f8d1", FromSubaccount: 0, ToSubaccount: 1, AmountCents: 5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/portfolio/margin/subaccounts/transfer")
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"client_transfer_id":"8c35ecb3-328f-4f52-8c7c-0f4b9862f8d1","from_subaccount":0,"to_subaccount":1,"amount_cents":5000}`)

	client, ct = newCaptureClient(200, `{}`)
	if err := client.Perps.Portfolio.TransferBetweenSubaccounts(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

// Keyed transfers keep the full retry policy; unkeyed ones and subaccount
// creation must not be replayed on an ambiguous failure.
func TestPerpsPortfolio_WriteRetryPolicy(t *testing.T) {
	ctx := context.Background()
	transfer := func(clientTransferID string) *types.ApplyMarginSubaccountTransferRequest {
		return &types.ApplyMarginSubaccountTransferRequest{ClientTransferID: clientTransferID, FromSubaccount: 0, ToSubaccount: 1, AmountCents: 100}
	}
	cases := []struct {
		name      string
		status    int
		call      func(*Client) error
		wantCalls int32
	}{
		{"transfer with client_transfer_id 503 retried", 503, func(c *Client) error {
			return c.Perps.Portfolio.TransferBetweenSubaccounts(ctx, transfer("t1"))
		}, 3},
		{"transfer without client_transfer_id 503 NOT retried", 503, func(c *Client) error {
			return c.Perps.Portfolio.TransferBetweenSubaccounts(ctx, transfer(""))
		}, 1},
		{"transfer without client_transfer_id 429 retried", 429, func(c *Client) error {
			return c.Perps.Portfolio.TransferBetweenSubaccounts(ctx, transfer(""))
		}, 3},
		{"create subaccount 503 NOT retried", 503, func(c *Client) error {
			_, err := c.Perps.Portfolio.CreateSubaccount(ctx)
			return err
		}, 1},
		{"create subaccount 429 retried", 429, func(c *Client) error {
			_, err := c.Perps.Portfolio.CreateSubaccount(ctx)
			return err
		}, 3},
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
