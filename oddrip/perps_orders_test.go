package oddrip

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

const perpsMarginOrderJSON = `{
	"order_id":"o1","user_id":"u1","client_order_id":"c1","ticker":"KXBTCPERP","side":"bid",
	"last_update_reason":"","price":"105.2500","fill_count":"0.00","remaining_count":"2.00",
	"expiration_time":null,"created_time":"2026-01-02T03:04:05Z","self_trade_prevention_type":"maker",
	"cancel_order_on_pause":false,"order_source":"user"
}`

func TestPerpsOrders_List(t *testing.T) {
	client, ct := newCaptureClient(200, `{"orders":[`+perpsMarginOrderJSON+`],"cursor":"next"}`)
	got, err := client.Perps.Orders.List(context.Background(), &types.GetMarginOrdersOpts{
		Ticker: "KXBTCPERP", MinTs: ptrOf[int64](100), MaxTs: ptrOf[int64](200), Status: "resting",
		Limit: ptrOf[int64](10000), Cursor: "c", Subaccount: ptrOf(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/orders")
	ct.assertQuery(t, map[string]string{
		"ticker": "KXBTCPERP", "min_ts": "100", "max_ts": "200", "status": "resting",
		"limit": "10000", "cursor": "c", "subaccount": "0",
	})
	if got.Cursor != "next" || len(got.Orders) != 1 || got.Orders[0].OrderID != "o1" || got.Orders[0].Side != types.BookSideBid {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsOrders_List_NilOpts(t *testing.T) {
	client, ct := newCaptureClient(200, `{"orders":[],"cursor":""}`)
	if _, err := client.Perps.Orders.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/orders")
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsOrders_Create(t *testing.T) {
	client, ct := newCaptureClient(201, `{"order_id":"o1","client_order_id":"cid-1","fill_count":"1.00","remaining_count":"1.00","average_fill_price":"105.2500","average_fee_paid":"0.0100"}`)
	got, err := client.Perps.Orders.Create(context.Background(), &types.CreateMarginOrderRequest{
		Ticker: "KXBTCPERP", ClientOrderID: "cid-1", Side: types.BookSideAsk, Count: "2.00", Price: "105.2500",
		TimeInForce: types.TimeInForceGTC, SelfTradePreventionType: types.SelfTradeMaker,
		ExpirationTime: 1700000000, PostOnly: ptrOf(false), CancelOrderOnPause: ptrOf(true), ReduceOnly: ptrOf(false),
		Subaccount: ptrOf(0), OrderGroupID: "og1",
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/margin/orders")
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{
		"ticker":"KXBTCPERP","client_order_id":"cid-1","side":"ask","count":"2.00","price":"105.2500",
		"time_in_force":"good_till_canceled","self_trade_prevention_type":"maker","expiration_time":1700000000,
		"post_only":false,"cancel_order_on_pause":true,"reduce_only":false,"subaccount":0,"order_group_id":"og1"
	}`)
	if got.OrderID != "o1" || got.FillCount != "1.00" || got.AverageFillPrice != "105.2500" || got.AverageFeePaid != "0.0100" {
		t.Fatalf("resp: %+v", got)
	}
}

func TestPerpsOrders_Create_NilRequest(t *testing.T) {
	client, ct := newCaptureClient(201, `{}`)
	if _, err := client.Perps.Orders.Create(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
	if ct.req != nil {
		t.Fatal("request was sent")
	}
}

// A keyed create is retried after a 5xx; an unkeyed one is not, because the
// first attempt may have placed the order.
func TestPerpsOrders_CreateRetry(t *testing.T) {
	newClient := func(t *testing.T) (*Client, *int32) {
		t.Helper()
		var calls int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/margin/orders" {
				t.Errorf("request: %s %s", r.Method, r.URL.Path)
			}
			if atomic.AddInt32(&calls, 1) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"code":"internal","message":"boom"}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"order_id":"o1","fill_count":"0.00","remaining_count":"1.00"}`))
		}))
		t.Cleanup(srv.Close)
		client := New(BaseURL(srv.URL), RetryConfigOption(RetryConfig{
			MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond,
		}))
		return client, &calls
	}
	order := func(clientOrderID string) *types.CreateMarginOrderRequest {
		return &types.CreateMarginOrderRequest{
			Ticker: "KXBTCPERP", ClientOrderID: clientOrderID, Side: types.BookSideBid, Count: "1", Price: "100.00",
			TimeInForce: types.TimeInForceGTC, SelfTradePreventionType: types.SelfTradeTakerAtCross,
		}
	}

	t.Run("with client_order_id retried", func(t *testing.T) {
		client, calls := newClient(t)
		got, err := client.Perps.Orders.Create(context.Background(), order("cid-1"))
		if err != nil {
			t.Fatal(err)
		}
		if got.OrderID != "o1" {
			t.Fatalf("resp: %+v", got)
		}
		if n := atomic.LoadInt32(calls); n != 2 {
			t.Fatalf("attempts = %d, want 2", n)
		}
	})
	t.Run("without client_order_id not retried", func(t *testing.T) {
		client, calls := newClient(t)
		_, err := client.Perps.Orders.Create(context.Background(), order(""))
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
			t.Fatalf("err = %v, want *APIError 500", err)
		}
		if n := atomic.LoadInt32(calls); n != 1 {
			t.Fatalf("attempts = %d, want 1", n)
		}
	})
}

func TestPerpsOrders_CancelAll(t *testing.T) {
	client, ct := newCaptureClient(204, ``)
	if err := client.Perps.Orders.CancelAll(context.Background(), &types.CancelAllOrdersOpts{Subaccount: ptrOf(2)}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/orders")
	ct.assertQuery(t, map[string]string{"subaccount": "2"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}

	if err := client.Perps.Orders.CancelAll(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/orders")
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsOrders_Get(t *testing.T) {
	client, ct := newCaptureClient(200, `{"order":`+perpsMarginOrderJSON+`}`)
	got, err := client.Perps.Orders.Get(context.Background(), "o1")
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/orders/o1")
	ct.assertQuery(t, map[string]string{})
	o := got.Order
	if o.OrderID != "o1" || o.Ticker != "KXBTCPERP" || o.ExpirationTime != nil || o.CreatedTime == nil || *o.CreatedTime != "2026-01-02T03:04:05Z" {
		t.Fatalf("resp: %+v", o)
	}
}

func TestPerpsOrders_Cancel(t *testing.T) {
	client, ct := newCaptureClient(200, `{"order_id":"o1","client_order_id":"c1","reduced_by":"2.00"}`)
	got, err := client.Perps.Orders.Cancel(context.Background(), "o1", &types.MarginOrderOpts{Subaccount: ptrOf(3)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/orders/o1")
	ct.assertQuery(t, map[string]string{"subaccount": "3"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}
	if got.OrderID != "o1" || got.ClientOrderID != "c1" || got.ReducedBy != "2.00" {
		t.Fatalf("resp: %+v", got)
	}

	if _, err := client.Perps.Orders.Cancel(context.Background(), "o1", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsOrders_Decrease(t *testing.T) {
	client, ct := newCaptureClient(200, `{"order_id":"o1","client_order_id":"c1","remaining_count":"1.00"}`)
	got, err := client.Perps.Orders.Decrease(context.Background(), "o1", &types.DecreaseMarginOrderRequest{ReduceTo: ptrOf("1.00")}, &types.MarginOrderOpts{Subaccount: ptrOf(1)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/margin/orders/o1/decrease")
	ct.assertQuery(t, map[string]string{"subaccount": "1"})
	ct.assertBody(t, `{"reduce_to":"1.00"}`)
	if got.OrderID != "o1" || got.RemainingCount != "1.00" {
		t.Fatalf("resp: %+v", got)
	}

	if _, err := client.Perps.Orders.Decrease(context.Background(), "o1", &types.DecreaseMarginOrderRequest{ReduceBy: ptrOf("1")}, nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"reduce_by":"1"}`)
}

func TestPerpsOrders_Decrease_Validation(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	for name, req := range map[string]*types.DecreaseMarginOrderRequest{
		"nil":     nil,
		"neither": {},
		"both":    {ReduceBy: ptrOf("1"), ReduceTo: ptrOf("1")},
	} {
		if _, err := client.Perps.Orders.Decrease(ctx, "o1", req, nil); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if ct.req != nil {
		t.Fatal("request was sent")
	}
}

func TestPerpsOrders_Amend(t *testing.T) {
	client, ct := newCaptureClient(200, `{"order_id":"o1","client_order_id":"c2","remaining_count":"3.00","fill_count":"0.00","average_fill_price":null,"average_fee_paid":null}`)
	got, err := client.Perps.Orders.Amend(context.Background(), "o1", &types.AmendMarginOrderRequest{
		Ticker: "KXBTCPERP", Side: types.BookSideBid, Price: "104.0000", Count: "3.00",
		ClientOrderID: "c1", UpdatedClientOrderID: "c2",
	}, &types.MarginOrderOpts{Subaccount: ptrOf(0)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/margin/orders/o1/amend")
	ct.assertQuery(t, map[string]string{"subaccount": "0"})
	ct.assertBody(t, `{"ticker":"KXBTCPERP","side":"bid","price":"104.0000","count":"3.00","client_order_id":"c1","updated_client_order_id":"c2"}`)
	if got.OrderID != "o1" || got.RemainingCount == nil || *got.RemainingCount != "3.00" || got.AverageFillPrice != nil {
		t.Fatalf("resp: %+v", got)
	}

	if _, err := client.Perps.Orders.Amend(context.Background(), "o1", &types.AmendMarginOrderRequest{Ticker: "KXBTCPERP", Side: types.BookSideAsk, Price: "1", Count: "1"}, nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"ticker":"KXBTCPERP","side":"ask","price":"1","count":"1"}`)

	ct.req = nil
	if _, err := client.Perps.Orders.Amend(context.Background(), "o1", nil, nil); err == nil {
		t.Fatal("expected error for nil request")
	}
	if ct.req != nil {
		t.Fatal("request was sent")
	}
}

// Decrease and amend cannot be deduplicated, so only 429 is retried; cancel
// is a DELETE and keeps the full policy.
func TestPerpsOrders_RetryPolicy(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name      string
		status    int
		call      func(*Client) error
		wantCalls int32
	}{
		{"create without client_order_id 429 retried", 429, func(c *Client) error {
			_, err := c.Perps.Orders.Create(ctx, &types.CreateMarginOrderRequest{Ticker: "KXBTCPERP"})
			return err
		}, 3},
		{"cancel 503 retried", 503, func(c *Client) error { _, err := c.Perps.Orders.Cancel(ctx, "o1", nil); return err }, 3},
		{"cancel all 503 retried", 503, func(c *Client) error { return c.Perps.Orders.CancelAll(ctx, nil) }, 3},
		{"decrease 502 NOT retried", 502, func(c *Client) error {
			_, err := c.Perps.Orders.Decrease(ctx, "o1", &types.DecreaseMarginOrderRequest{ReduceBy: ptrOf("1")}, nil)
			return err
		}, 1},
		{"decrease 429 retried", 429, func(c *Client) error {
			_, err := c.Perps.Orders.Decrease(ctx, "o1", &types.DecreaseMarginOrderRequest{ReduceBy: ptrOf("1")}, nil)
			return err
		}, 3},
		{"amend 500 NOT retried", 500, func(c *Client) error {
			_, err := c.Perps.Orders.Amend(ctx, "o1", &types.AmendMarginOrderRequest{}, nil)
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

// An empty order ID must not turn DELETE /margin/orders/{id} into CancelAll.
func TestPerpsOrders_EmptyPathParam(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	calls := map[string]func() error{
		"Get":    func() error { _, err := client.Perps.Orders.Get(ctx, ""); return err },
		"Cancel": func() error { _, err := client.Perps.Orders.Cancel(ctx, "", nil); return err },
		"Decrease": func() error {
			_, err := client.Perps.Orders.Decrease(ctx, "", &types.DecreaseMarginOrderRequest{ReduceBy: ptrOf("1")}, nil)
			return err
		},
		"Amend": func() error {
			_, err := client.Perps.Orders.Amend(ctx, "..", &types.AmendMarginOrderRequest{}, nil)
			return err
		},
		"Cancel(dot)":    func() error { _, err := client.Perps.Orders.Cancel(ctx, ".", nil); return err },
		"Cancel(dotdot)": func() error { _, err := client.Perps.Orders.Cancel(ctx, "..", nil); return err },
	}
	for name, call := range calls {
		ct.req = nil
		if err := call(); !errors.Is(err, ErrEmptyPathParam) {
			t.Errorf("%s: err = %v, want ErrEmptyPathParam", name, err)
		}
		if ct.req != nil {
			t.Errorf("%s: request was sent: %s %s", name, ct.req.Method, ct.req.URL.Path)
		}
	}
}

func TestPerpsOrderGroups_List(t *testing.T) {
	client, ct := newCaptureClient(200, `{"order_groups":[{"id":"og1","contracts_limit_fp":"10.00","is_auto_cancel_enabled":true,"exchange_index":0}]}`)
	got, err := client.Perps.OrderGroups.List(context.Background(), &types.GetOrderGroupsOpts{Subaccount: ptrOf(3)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/order_groups")
	ct.assertQuery(t, map[string]string{"subaccount": "3"})
	if len(got.OrderGroups) != 1 || got.OrderGroups[0].ID != "og1" || got.OrderGroups[0].ContractsLimitFp != "10.00" || !got.OrderGroups[0].IsAutoCancelEnabled {
		t.Fatalf("resp: %+v", got)
	}

	if _, err := client.Perps.OrderGroups.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsOrderGroups_Create(t *testing.T) {
	client, ct := newCaptureClient(201, `{"order_group_id":"og1","subaccount":2,"exchange_index":0}`)
	got, err := client.Perps.OrderGroups.Create(context.Background(), &types.CreateOrderGroupRequest{
		Subaccount: ptrOf(2), ContractsLimit: ptrOf[int64](10), ContractsLimitFp: ptrOf("10.00"), ExchangeIndex: ptrOf(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/margin/order_groups/create")
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"subaccount":2,"contracts_limit":10,"contracts_limit_fp":"10.00","exchange_index":0}`)
	if got.OrderGroupID != "og1" || got.Subaccount != 2 {
		t.Fatalf("resp: %+v", got)
	}

	ct.req = nil
	for name, req := range map[string]*types.CreateOrderGroupRequest{"nil": nil, "no limit": {Subaccount: ptrOf(1)}} {
		if _, err := client.Perps.OrderGroups.Create(context.Background(), req); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if ct.req != nil {
		t.Fatal("request was sent")
	}
}

func TestPerpsOrderGroups_Get(t *testing.T) {
	client, ct := newCaptureClient(200, `{"is_auto_cancel_enabled":false,"contracts_limit_fp":"5.00","orders":["o1","o2"],"exchange_index":0}`)
	got, err := client.Perps.OrderGroups.Get(context.Background(), "og1", &types.GetOrderGroupOpts{Subaccount: ptrOf(0)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/order_groups/og1")
	ct.assertQuery(t, map[string]string{"subaccount": "0"})
	if got.IsAutoCancelEnabled || got.ContractsLimitFp != "5.00" || len(got.Orders) != 2 || got.Orders[1] != "o2" {
		t.Fatalf("resp: %+v", got)
	}

	if _, err := client.Perps.OrderGroups.Get(context.Background(), "og1", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsOrderGroups_Delete(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.Perps.OrderGroups.Delete(context.Background(), "og1", &types.MarginOrderGroupOpts{Subaccount: ptrOf(1)}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/order_groups/og1")
	ct.assertQuery(t, map[string]string{"subaccount": "1"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}

	if err := client.Perps.OrderGroups.Delete(context.Background(), "og1", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsOrderGroups_Reset(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.Perps.OrderGroups.Reset(context.Background(), "og1", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/margin/order_groups/og1/reset")
	ct.assertQuery(t, map[string]string{})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}

	if err := client.Perps.OrderGroups.Reset(context.Background(), "og1", &types.MarginOrderGroupOpts{Subaccount: ptrOf(4)}); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{"subaccount": "4"})
}

func TestPerpsOrderGroups_Trigger(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.Perps.OrderGroups.Trigger(context.Background(), "og1", &types.MarginOrderGroupOpts{Subaccount: ptrOf(2)}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/margin/order_groups/og1/trigger")
	ct.assertQuery(t, map[string]string{"subaccount": "2"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}

	if err := client.Perps.OrderGroups.Trigger(context.Background(), "og1", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsOrderGroups_UpdateLimit(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.Perps.OrderGroups.UpdateLimit(context.Background(), "og1", &types.UpdateOrderGroupLimitRequest{ContractsLimitFp: ptrOf("25.00")}, &types.MarginOrderGroupOpts{Subaccount: ptrOf(0)}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/margin/order_groups/og1/limit")
	ct.assertQuery(t, map[string]string{"subaccount": "0"})
	ct.assertBody(t, `{"contracts_limit_fp":"25.00"}`)

	if err := client.Perps.OrderGroups.UpdateLimit(context.Background(), "og1", &types.UpdateOrderGroupLimitRequest{ContractsLimit: ptrOf[int64](25)}, nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"contracts_limit":25}`)

	ct.req = nil
	for name, req := range map[string]*types.UpdateOrderGroupLimitRequest{"nil": nil, "no limit": {}} {
		if err := client.Perps.OrderGroups.UpdateLimit(context.Background(), "og1", req, nil); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if ct.req != nil {
		t.Fatal("request was sent")
	}
}

func TestPerpsOrderGroups_RetryPolicy(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name      string
		status    int
		call      func(*Client) error
		wantCalls int32
	}{
		{"create 503 NOT retried", 503, func(c *Client) error {
			_, err := c.Perps.OrderGroups.Create(ctx, &types.CreateOrderGroupRequest{ContractsLimit: ptrOf[int64](10)})
			return err
		}, 1},
		{"create 429 retried", 429, func(c *Client) error {
			_, err := c.Perps.OrderGroups.Create(ctx, &types.CreateOrderGroupRequest{ContractsLimit: ptrOf[int64](10)})
			return err
		}, 3},
		{"reset 503 retried", 503, func(c *Client) error { return c.Perps.OrderGroups.Reset(ctx, "og1", nil) }, 3},
		{"delete 503 retried", 503, func(c *Client) error { return c.Perps.OrderGroups.Delete(ctx, "og1", nil) }, 3},
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

func TestPerpsOrderGroups_EmptyPathParam(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	limit := &types.UpdateOrderGroupLimitRequest{ContractsLimit: ptrOf[int64](1)}
	calls := map[string]func() error{
		"Get":         func() error { _, err := client.Perps.OrderGroups.Get(ctx, "", nil); return err },
		"Delete":      func() error { return client.Perps.OrderGroups.Delete(ctx, "", nil) },
		"Reset":       func() error { return client.Perps.OrderGroups.Reset(ctx, "", nil) },
		"Trigger":     func() error { return client.Perps.OrderGroups.Trigger(ctx, "..", nil) },
		"UpdateLimit": func() error { return client.Perps.OrderGroups.UpdateLimit(ctx, "", limit, nil) },
	}
	for name, call := range calls {
		ct.req = nil
		if err := call(); !errors.Is(err, ErrEmptyPathParam) {
			t.Errorf("%s: err = %v, want ErrEmptyPathParam", name, err)
		}
		if ct.req != nil {
			t.Errorf("%s: request was sent: %s %s", name, ct.req.Method, ct.req.URL.Path)
		}
	}
}
