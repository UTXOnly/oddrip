package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/UTXOnly/oddrip/oddrip"
	"github.com/UTXOnly/oddrip/oddrip/types"
)

const logFilename = "api_calls.log"

func main() {
	logFile, err := os.Create(logFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://demo-api.kalshi.co/trade-api/v2"
	}

	var hasAuth bool
	opts := []oddrip.Option{
		oddrip.BaseURL(baseURL),
		oddrip.HTTPClient(&http.Client{
			Transport: &loggingTransport{
				base:   http.DefaultTransport,
				log:    logFile,
				baseURL: baseURL,
			},
			Timeout: 30 * time.Second,
		}),
	}
	if keyID := os.Getenv("KALSHI_ACCESS_KEY"); keyID != "" {
		if keyPath := os.Getenv("KALSHI_PRIVATE_KEY_PATH"); keyPath != "" {
			pemBytes, err := os.ReadFile(keyPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read private key: %v\n", err)
				os.Exit(1)
			}
			priv, err := oddrip.ParsePrivateKeyFromPEM(pemBytes)
			if err != nil {
				fmt.Fprintf(os.Stderr, "parse private key: %v\n", err)
				os.Exit(1)
			}
			opts = append(opts, oddrip.Auth(oddrip.NewKalshiSigner(keyID, priv)))
			hasAuth = true
		} else {
			sig := os.Getenv("KALSHI_ACCESS_SIGNATURE")
			ts := os.Getenv("KALSHI_ACCESS_TIMESTAMP")
			if sig != "" && ts != "" {
				opts = append(opts, oddrip.Auth(&oddrip.StaticHeaders{
					Headers: map[string][]string{
						"KALSHI-ACCESS-KEY":        {keyID},
						"KALSHI-ACCESS-SIGNATURE":  {sig},
						"KALSHI-ACCESS-TIMESTAMP":  {ts},
					},
				}))
				hasAuth = true
			}
		}
	}

	client := oddrip.New(opts...)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	runAll(ctx, client, logFile)

	if os.Getenv("LIVE") == "1" && hasAuth {
		runLiveOrder(ctx, client, logFile)
	}
}

type loggingTransport struct {
	base    http.RoundTripper
	log     io.Writer
	baseURL string
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fullURL := req.URL.String()
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(t.log, "[%s] %s %s\nERROR: %v\n\n", time.Now().Format(time.RFC3339), req.Method, fullURL, err)
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	fmt.Fprintf(t.log, "[%s] %s %s\n", time.Now().Format(time.RFC3339), req.Method, fullURL)
	fmt.Fprintf(t.log, "Endpoint: %s %s\n", req.Method, req.URL.Path)
	if q := req.URL.RawQuery; q != "" {
		fmt.Fprintf(t.log, "Query: %s\n", q)
	}
	fmt.Fprintf(t.log, "Full URL: %s\n", fullURL)
	fmt.Fprintf(t.log, "Response status: %d\n", resp.StatusCode)
	fmt.Fprintf(t.log, "Response body:\n%s\n\n", indentJSON(body))

	return resp, nil
}

func indentJSON(b []byte) []byte {
	var buf bytes.Buffer
	if json.Indent(&buf, b, "", "  ") != nil {
		return b
	}
	return buf.Bytes()
}

func runAll(ctx context.Context, client *oddrip.Client, log io.Writer) {
	limit3 := int64(3)
	limit5 := int64(5)
	limit10 := int64(10)
	limit5Int := 5

	logCall := func(desc string, fn func()) {
		fmt.Fprintf(log, "=== %s ===\n", desc)
		fn()
	}

	logCall("Exchange.GetStatus", func() { client.Exchange.GetStatus(ctx) })
	logCall("Exchange.GetAnnouncements", func() { client.Exchange.GetAnnouncements(ctx) })
	logCall("Exchange.GetSchedule", func() { client.Exchange.GetSchedule(ctx) })
	logCall("Exchange.GetUserDataTimestamp", func() { client.Exchange.GetUserDataTimestamp(ctx) })
	logCall("Exchange.GetHistoricalCutoff", func() { client.Exchange.GetHistoricalCutoff(ctx) })
	logCall("Exchange.GetSeriesFeeChanges (no filter)", func() { client.Exchange.GetSeriesFeeChanges(ctx, "", false) })
	logCall("Exchange.GetSeriesFeeChanges (KXBTC, historical)", func() { client.Exchange.GetSeriesFeeChanges(ctx, "KXBTC", true) })

	logCall("Markets.List (limit=5)", func() { client.Markets.List(ctx, &types.GetMarketsOpts{Limit: &limit5}) })
	logCall("Markets.List (limit=10, status=open)", func() { client.Markets.List(ctx, &types.GetMarketsOpts{Limit: &limit10, Status: types.MarketStatusOpen}) })
	logCall("Markets.List (limit=3, event_ticker=KXBTC)", func() { client.Markets.List(ctx, &types.GetMarketsOpts{Limit: &limit3, EventTicker: "KXBTC"}) })
	var markets *types.GetMarketsResponse
	logCall("Markets.List (limit=5, for follow-up)", func() {
		markets, _ = client.Markets.List(ctx, &types.GetMarketsOpts{Limit: &limit5})
	})
	if markets != nil && len(markets.Markets) > 0 {
		ticker := markets.Markets[0].Ticker
		logCall("Markets.Get "+ticker, func() { client.Markets.Get(ctx, ticker) })
		logCall("Markets.GetOrderbook "+ticker+" default", func() { client.Markets.GetOrderbook(ctx, ticker, nil) })
		logCall("Markets.GetOrderbook "+ticker+" depth=5", func() { client.Markets.GetOrderbook(ctx, ticker, &types.GetMarketOrderbookOpts{Depth: 5}) })
	}
	logCall("Markets.GetTrades (no opts)", func() { client.Markets.GetTrades(ctx, nil) })
	logCall("Markets.GetTrades (limit=5)", func() { client.Markets.GetTrades(ctx, &types.GetTradesOpts{Limit: &limit5}) })
	if markets != nil && len(markets.Markets) > 0 {
		ticker := markets.Markets[0].Ticker
		logCall("Markets.GetTrades (limit=3, ticker="+ticker+")", func() { client.Markets.GetTrades(ctx, &types.GetTradesOpts{Limit: &limit3, Ticker: ticker}) })
	}

	logCall("Events.List (limit=5)", func() { client.Events.List(ctx, &types.GetEventsOpts{Limit: &limit5}) })
	logCall("Events.List (limit=3, status=open)", func() { client.Events.List(ctx, &types.GetEventsOpts{Limit: &limit3, Status: "open"}) })
	logCall("Events.List (limit=3, series_ticker=KXBTC)", func() { client.Events.List(ctx, &types.GetEventsOpts{Limit: &limit3, SeriesTicker: "KXBTC"}) })
	nestedTrue := true
	logCall("Events.List (limit=3, with_nested_markets=true)", func() { client.Events.List(ctx, &types.GetEventsOpts{Limit: &limit3, WithNestedMarkets: &nestedTrue}) })
	var eventsResp *types.GetEventsResponse
	logCall("Events.List (limit=5, for follow-up)", func() {
		eventsResp, _ = client.Events.List(ctx, &types.GetEventsOpts{Limit: &limit5})
	})
	if eventsResp != nil && len(eventsResp.Events) > 0 {
		eventTicker := eventsResp.Events[0].EventTicker
		logCall("Events.Get "+eventTicker, func() { client.Events.Get(ctx, eventTicker, nil) })
		logCall("Events.Get "+eventTicker+" (with_nested_markets=true)", func() { client.Events.Get(ctx, eventTicker, &types.GetEventOpts{WithNestedMarkets: &nestedTrue}) })
		logCall("Events.GetMetadata "+eventTicker, func() { client.Events.GetMetadata(ctx, eventTicker) })
	}
	logCall("Events.ListMultivariate (limit=3)", func() { client.Events.ListMultivariate(ctx, &types.GetMultivariateEventsOpts{Limit: &limit3}) })
	logCall("Events.ListMultivariate (limit=3, with_nested_markets=true)", func() { client.Events.ListMultivariate(ctx, &types.GetMultivariateEventsOpts{Limit: &limit3, WithNestedMarkets: &nestedTrue}) })

	logCall("Orders.List (no opts)", func() { client.Orders.List(ctx, nil) })
	logCall("Orders.List (limit=5)", func() { client.Orders.List(ctx, &types.GetOrdersOpts{Limit: &limit5}) })
	logCall("Orders.List (status=resting, limit=5)", func() { client.Orders.List(ctx, &types.GetOrdersOpts{Status: types.OrderStatusResting, Limit: &limit5}) })
	logCall("Orders.List (status=executed, limit=3)", func() { client.Orders.List(ctx, &types.GetOrdersOpts{Status: types.OrderStatusExecuted, Limit: &limit3}) })
	var ordersResp *types.GetOrdersResponse
	logCall("Orders.List (limit=5, for follow-up)", func() {
		ordersResp, _ = client.Orders.List(ctx, &types.GetOrdersOpts{Limit: &limit5})
	})
	if ordersResp != nil && len(ordersResp.Orders) > 0 {
		oid := ordersResp.Orders[0].OrderID
		logCall("Orders.Get "+oid, func() { client.Orders.Get(ctx, oid) })
		logCall("Orders.GetQueuePosition "+oid, func() { client.Orders.GetQueuePosition(ctx, oid) })
		logCall("Orders.GetQueuePositions (market_tickers="+ordersResp.Orders[0].Ticker+")", func() {
			client.Orders.GetQueuePositions(ctx, &types.GetOrderQueuePositionsOpts{MarketTickers: ordersResp.Orders[0].Ticker})
		})
	}
	if markets != nil && len(markets.Markets) > 0 {
		ticker := markets.Markets[0].Ticker
		logCall("Orders.GetQueuePositions (event_ticker)", func() {
			client.Orders.GetQueuePositions(ctx, &types.GetOrderQueuePositionsOpts{EventTicker: markets.Markets[0].EventTicker})
		})
		logCall("Orders.GetQueuePositions (market_tickers="+ticker+")", func() {
			client.Orders.GetQueuePositions(ctx, &types.GetOrderQueuePositionsOpts{MarketTickers: ticker})
		})
	}

	logCall("Portfolio.GetBalance", func() { client.Portfolio.GetBalance(ctx, nil) })
	logCall("Portfolio.GetFills (no opts)", func() { client.Portfolio.GetFills(ctx, nil) })
	logCall("Portfolio.GetFills (limit=5)", func() { client.Portfolio.GetFills(ctx, &types.GetFillsOpts{Limit: &limit5}) })
	logCall("Portfolio.GetPositions (no opts)", func() { client.Portfolio.GetPositions(ctx, nil) })
	logCall("Portfolio.GetPositions (limit=5)", func() { client.Portfolio.GetPositions(ctx, &types.GetPositionsOpts{Limit: &limit5Int}) })

	logCall("Account.GetAPILimits", func() { client.Account.GetAPILimits(ctx) })
}

func runLiveOrder(ctx context.Context, client *oddrip.Client, log io.Writer) {
	fmt.Fprintf(log, "=== Live: limit order (1 share yes, current 15m BTC market) ===\n")
	limit1 := int64(1)
	markets, err := client.Markets.List(ctx, &types.GetMarketsOpts{
		SeriesTicker: "KXBTC15M",
		Status:       types.MarketStatusOpen,
		Limit:        &limit1,
	})
	if err != nil {
		fmt.Fprintf(log, "Markets.List (KXBTC15M, open) error: %v\n\n", err)
		return
	}
	if len(markets.Markets) == 0 {
		fmt.Fprintf(log, "No open 15m BTC market found.\n\n")
		return
	}
	ticker := markets.Markets[0].Ticker
	yesPrice := 1
	req := &types.CreateOrderRequest{
		Ticker:      ticker,
		Side:        types.OrderSideYes,
		Action:      types.OrderActionBuy,
		Count:       ptr(1),
		YesPrice:    &yesPrice,
		TimeInForce: ptr(types.TimeInForceGTC),
	}
	createResp, err := client.Orders.Create(ctx, req)
	if err != nil {
		fmt.Fprintf(log, "Orders.Create (resting) error: %v\n\n", err)
		return
	}
	fmt.Fprintf(log, "Orders.Create (resting) response: order_id=%s — left resting for you to confirm.\n", createResp.Order.OrderID)

	createResp2, err := client.Orders.Create(ctx, req)
	if err != nil {
		fmt.Fprintf(log, "Orders.Create (to cancel) error: %v\n\n", err)
		return
	}
	fmt.Fprintf(log, "Orders.Create (to cancel) response: order_id=%s\n", createResp2.Order.OrderID)
	_, _ = client.Orders.Cancel(ctx, createResp2.Order.OrderID, nil)
	fmt.Fprintf(log, "Orders.Cancel called on second order. First order (%s) remains resting.\n\n", createResp.Order.OrderID)
}

func ptr[T any](v T) *T { return &v }
