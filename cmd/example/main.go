package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/oddrip/client/oddrip"
	"github.com/oddrip/client/oddrip/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.elections.kalshi.com/trade-api/v2"
	}

	opts := []oddrip.Option{oddrip.BaseURL(baseURL)}
	hasAuth := false
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

	fmt.Println("Exchange status:")
	status, err := client.Exchange.GetStatus(ctx)
	if err != nil {
		var apiErr *oddrip.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintf(os.Stderr, "api error %d: %s\n", apiErr.StatusCode, apiErr.Message)
		} else {
			fmt.Fprintf(os.Stderr, "exchange status: %v\n", err)
		}
		os.Exit(1)
	}
	fmt.Printf("  exchange_active=%v trading_active=%v\n", status.ExchangeActive, status.TradingActive)

	fmt.Println("Historical cutoff:")
	cutoff, err := client.Exchange.GetHistoricalCutoff(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "historical cutoff: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  market_settled_ts=%s\n", cutoff.MarketSettledTs)

	fmt.Println("Markets (first page, limit 3):")
	markets, err := client.Markets.List(ctx, &types.GetMarketsOpts{Limit: ptr(int64(3))})
	if err != nil {
		fmt.Fprintf(os.Stderr, "markets list: %v\n", err)
		os.Exit(1)
	}
	for _, m := range markets.Markets {
		fmt.Printf("  %s %s\n", m.Ticker, m.YesSubTitle)
	}
	fmt.Printf("  cursor=%q\n", markets.Cursor)

	if hasAuth {
		fmt.Println("Balance (authenticated):")
		bal, err := client.Portfolio.GetBalance(ctx, nil)
		if err != nil {
			var apiErr *oddrip.APIError
			if errors.As(err, &apiErr) {
				fmt.Fprintf(os.Stderr, "api error %d: %s\n", apiErr.StatusCode, apiErr.Message)
			} else {
				fmt.Fprintf(os.Stderr, "balance: %v\n", err)
			}
		} else {
			fmt.Printf("  balance=%d cents portfolio_value=%d\n", bal.Balance, bal.PortfolioValue)
		}
	}
}

func ptr[T any](v T) *T { return &v }
