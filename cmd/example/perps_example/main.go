package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/UTXOnly/oddrip/oddrip"
	"github.com/UTXOnly/oddrip/oddrip/types"
)

const logFilename = "perps_calls.log"

const (
	perpsWSHostProduction = "external-api-margin-ws.kalshi.com"
	perpsWSHostDemo       = "external-api-margin-ws.demo.kalshi.co"
	perpsWSPath           = "/trade-api/ws/v2/margin"
)

const (
	wsMaxMessages = 20
	wsReadWindow  = 30 * time.Second
)

func main() {
	logFile, err := os.Create(logFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://external-api.demo.kalshi.co/trade-api/v2"
	}

	var hasAuth bool
	opts := []oddrip.Option{
		oddrip.BaseURL(baseURL),
		oddrip.HTTPClient(&http.Client{
			Transport: &loggingTransport{
				base: http.DefaultTransport,
				log:  logFile,
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
						"KALSHI-ACCESS-KEY":       {keyID},
						"KALSHI-ACCESS-SIGNATURE": {sig},
						"KALSHI-ACCESS-TIMESTAMP": {ts},
					},
				}))
				hasAuth = true
			}
		}
	}

	client := oddrip.New(opts...)

	restCtx, cancelREST := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelREST()
	ticker := runPublic(restCtx, client, logFile)
	if !hasAuth {
		logSection(logFile, "Authenticated REST and WebSocket (skipped)", "No credentials: set KALSHI_ACCESS_KEY and KALSHI_PRIVATE_KEY_PATH.")
		return
	}
	runAuthenticated(restCtx, client, logFile)

	wsCtx, cancelWS := context.WithTimeout(context.Background(), wsReadWindow+15*time.Second)
	defer cancelWS()
	runWS(wsCtx, client, logFile, perpsWSHostFromBase(baseURL), ticker)
}

type loggingTransport struct {
	base http.RoundTripper
	log  io.Writer
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

// perpsWSHostFromBase picks the perps WebSocket host for a REST base URL.
// Perps has its own WebSocket host, so no REST host is reused: demo hosts
// (kalshi.co) map to the demo margin host and everything else to production.
func perpsWSHostFromBase(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return perpsWSHostDemo
	}
	if strings.HasSuffix(u.Hostname(), ".kalshi.co") {
		return perpsWSHostDemo
	}
	return perpsWSHostProduction
}

func logSection(log io.Writer, title string, body string) {
	fmt.Fprintf(log, "=== %s ===\n", title)
	fmt.Fprintf(log, "%s\n\n", body)
}

func logJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

// logCall writes a section header, runs fn (the transport logs the HTTP
// exchange), and logs the returned error. It never stops the run.
func logCall(log io.Writer, desc string, fn func() error) {
	fmt.Fprintf(log, "=== %s ===\n", desc)
	if err := fn(); err != nil {
		fmt.Fprintf(log, "Error: %v\n\n", err)
	}
}

// runPublic makes the unauthenticated perps calls and returns the market
// ticker used for the per-market calls, or "" if there is none.
func runPublic(ctx context.Context, client *oddrip.Client, log io.Writer) string {
	logCall(log, "Perps.Exchange.GetStatus", func() error {
		_, err := client.Perps.Exchange.GetStatus(ctx)
		return err
	})

	var markets *types.GetMarginMarketsResponse
	logCall(log, "Perps.Markets.List", func() (err error) {
		markets, err = client.Perps.Markets.List(ctx, nil)
		return err
	})

	ticker := os.Getenv("PERPS_TICKER")
	source := "PERPS_TICKER"
	if ticker == "" && markets != nil {
		for _, m := range markets.Markets {
			if m.Status == types.MarginMarketStatusActive {
				ticker = m.Ticker
				source = "first active market"
				break
			}
		}
	}
	if ticker == "" {
		logSection(log, "Per-market calls (skipped)", "No active perps market found and PERPS_TICKER is unset.")
	} else {
		logSection(log, "Market", fmt.Sprintf("Using %s (%s)", ticker, source))
		runMarket(ctx, client, log, ticker)
	}

	logCall(log, "Perps.Risk.GetParameters", func() error {
		_, err := client.Perps.Risk.GetParameters(ctx)
		return err
	})
	return ticker
}

func runMarket(ctx context.Context, client *oddrip.Client, log io.Writer, ticker string) {
	depth10 := 10
	limit10 := int64(10)
	now := time.Now()
	hourAgo := now.Add(-time.Hour).Unix()
	dayAgo := now.Add(-24 * time.Hour).Unix()
	nowTs := now.Unix()

	logCall(log, "Perps.Markets.Get "+ticker, func() error {
		_, err := client.Perps.Markets.Get(ctx, ticker)
		return err
	})
	logCall(log, "Perps.Markets.GetOrderbook "+ticker+" depth=10", func() error {
		_, err := client.Perps.Markets.GetOrderbook(ctx, ticker, &types.GetMarginMarketOrderbookOpts{Depth: &depth10})
		return err
	})
	logCall(log, "Perps.Markets.GetTrades (limit=10, ticker="+ticker+")", func() error {
		_, err := client.Perps.Markets.GetTrades(ctx, &types.GetMarginTradesOpts{Ticker: ticker, Limit: &limit10})
		return err
	})
	logCall(log, "Perps.Markets.GetCandlesticks "+ticker+" (last hour, 1-minute)", func() error {
		_, err := client.Perps.Markets.GetCandlesticks(ctx, ticker, &types.GetMarginMarketCandlesticksOpts{
			StartTs:        hourAgo,
			EndTs:          nowTs,
			PeriodInterval: types.PeriodInterval1Min,
		})
		return err
	})
	logCall(log, "Perps.Funding.GetRateEstimate "+ticker, func() error {
		_, err := client.Perps.Funding.GetRateEstimate(ctx, &types.GetMarginFundingRateEstimateOpts{Ticker: ticker})
		return err
	})
	logCall(log, "Perps.Funding.GetHistoricalRates "+ticker+" (last 24h)", func() error {
		_, err := client.Perps.Funding.GetHistoricalRates(ctx, &types.GetMarginHistoricalFundingRatesOpts{
			Ticker:  ticker,
			StartTs: &dayAgo,
			EndTs:   &nowTs,
		})
		return err
	})
}

func runAuthenticated(ctx context.Context, client *oddrip.Client, log io.Writer) {
	limit5 := int64(5)

	logCall(log, "Perps.Exchange.GetEnabled", func() error {
		_, err := client.Perps.Exchange.GetEnabled(ctx)
		return err
	})
	logCall(log, "Perps.Account.GetAPILimits", func() error {
		_, err := client.Perps.Account.GetAPILimits(ctx)
		return err
	})
	logCall(log, "Perps.Portfolio.GetBalance", func() error {
		_, err := client.Perps.Portfolio.GetBalance(ctx, nil)
		return err
	})
	logCall(log, "Perps.Portfolio.GetPositions", func() error {
		_, err := client.Perps.Portfolio.GetPositions(ctx, nil)
		return err
	})
	logCall(log, "Perps.Portfolio.GetFills (limit=5)", func() error {
		_, err := client.Perps.Portfolio.GetFills(ctx, &types.GetMarginFillsOpts{Limit: &limit5})
		return err
	})
	logCall(log, "Perps.Orders.List (limit=5)", func() error {
		_, err := client.Perps.Orders.List(ctx, &types.GetMarginOrdersOpts{Limit: &limit5})
		return err
	})
	logCall(log, "Perps.OrderGroups.List", func() error {
		_, err := client.Perps.OrderGroups.List(ctx, nil)
		return err
	})
	logCall(log, "Perps.Risk.Get", func() error {
		_, err := client.Perps.Risk.Get(ctx)
		return err
	})
	logCall(log, "Perps.Risk.GetNotionalRiskLimit", func() error {
		_, err := client.Perps.Risk.GetNotionalRiskLimit(ctx)
		return err
	})
	logCall(log, "Perps.Fees.GetTiers", func() error {
		_, err := client.Perps.Fees.GetTiers(ctx)
		return err
	})
	logCall(log, "Perps.Fees.GetTierRates", func() error {
		_, err := client.Perps.Fees.GetTierRates(ctx)
		return err
	})
}

func runWS(ctx context.Context, client *oddrip.Client, log io.Writer, wsHost, ticker string) {
	if ticker == "" {
		logSection(log, "WebSocket (skipped)", "No market to subscribe to.")
		return
	}
	logSection(log, "WebSocket Connect", fmt.Sprintf("Full URL: wss://%s%s", wsHost, perpsWSPath))
	conn, err := client.ConnectPerpsWS(ctx, oddrip.WSHost(wsHost))
	if err != nil {
		logSection(log, "WebSocket Connect (error)", err.Error())
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logSection(log, "Close (error)", err.Error())
			return
		}
		logSection(log, "Close", "Connection closed.")
	}()

	sendSnapshot := true
	params := types.SubscribeParams{
		Channels:            []string{types.WSChannelTicker, types.WSChannelOrderbookDelta},
		MarketTicker:        ticker,
		SendInitialSnapshot: &sendSnapshot,
	}
	fmt.Fprintf(log, "=== Subscribe (ticker + orderbook_delta, %s) ===\n", ticker)
	fmt.Fprintf(log, "Sent:\n%s\n", logJSON(types.SubscribeCommand{Cmd: "subscribe", Params: params}))
	subs, err := conn.Subscribe(ctx, params)
	if len(subs) > 0 {
		fmt.Fprintf(log, "Response:\n%s\n", logJSON(subs))
	}
	if err != nil {
		fmt.Fprintf(log, "Response (error): %v\n", err)
	}
	fmt.Fprintln(log)
	if len(subs) == 0 {
		return
	}

	fmt.Fprintf(log, "=== Receive messages (up to %d or %s) ===\n", wsMaxMessages, wsReadWindow)
	deadline := time.After(wsReadWindow)
	received := 0
loop:
	for received < wsMaxMessages {
		select {
		case <-ctx.Done():
			break loop
		case <-deadline:
			break loop
		case msg, ok := <-conn.Messages():
			if !ok {
				fmt.Fprintf(log, "Connection ended: %v\n", conn.Err())
				break loop
			}
			received++
			logWSMessage(log, received, msg)
		}
	}
	fmt.Fprintf(log, "Received %d messages.\n\n", received)
}

// logWSMessage decodes the perps payloads this example subscribes to and
// logs anything else (such as the subscribed replies) raw.
func logWSMessage(log io.Writer, n int, msg *types.WSMessage) {
	fmt.Fprintf(log, "Message #%d type=%s sid=%d seq=%d\n", n, msg.Type, msg.SID, msg.Seq)
	var decoded interface{}
	switch msg.Type {
	case types.WSTypeTicker:
		decoded = &types.MarginTickerMsg{}
	case types.WSTypeOrderbookSnapshot:
		decoded = &types.MarginOrderbookSnapshotMsg{}
	case types.WSTypeOrderbookDelta:
		decoded = &types.MarginOrderbookDeltaMsg{}
	case types.WSTypeError:
		decoded = &types.MarginErrorMsg{}
	default:
		fmt.Fprintf(log, "Raw msg:\n%s\n\n", string(msg.Msg))
		return
	}
	if err := json.Unmarshal(msg.Msg, decoded); err != nil {
		fmt.Fprintf(log, "Decode error: %v\nRaw msg:\n%s\n\n", err, string(msg.Msg))
		return
	}
	fmt.Fprintf(log, "Decoded %T:\n%s\n\n", decoded, logJSON(decoded))
}
