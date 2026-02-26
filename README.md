# Oddrip – Kalshi Go REST API Client

Production-quality Go client for the [Kalshi Trade API](https://docs.kalshi.com/openapi.yaml).

## Install

```bash
go get github.com/oddrip/client/oddrip
```

## Initialize client

```go
import "github.com/oddrip/client/oddrip"

// No auth (public endpoints only)
client := oddrip.New()

// With API key + RSA-PSS signing
key, _ := oddrip.ParsePrivateKeyFromPEM(privateKeyPEM)
client := oddrip.New(
    oddrip.Auth(oddrip.NewKalshiSigner(apiKeyID, key)),
    oddrip.BaseURL("https://api.elections.kalshi.com/trade-api/v2"),
)
```

## Make a request

```go
ctx := context.Background()

status, err := client.Exchange.GetStatus(ctx)
if err != nil {
    // handle error (see error handling)
    return
}

market, err := client.Markets.Get(ctx, "TICKER-24JAN01")
orders, err := client.Orders.List(ctx, &types.GetOrdersOpts{Status: "resting"})
balance, err := client.Portfolio.GetBalance(ctx, nil)
```

## Pagination

```go
var all []types.Market
opts := &types.GetMarketsOpts{Limit: ptr(int64(100))}
for {
    resp, err := client.Markets.List(ctx, opts)
    if err != nil { return err }
    all = append(all, resp.Markets...)
    if resp.Cursor == "" { break }
    opts.Cursor = resp.Cursor
}
```

## Error handling

Non-2xx responses are returned as `*oddrip.APIError`:

```go
if err != nil {
    var apiErr *oddrip.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.StatusCode, apiErr.Message, apiErr.RequestID)
        fmt.Println(apiErr.RawBody)
    }
    return err
}
```

## Retry configuration

Retries apply to 429 and 5xx with exponential backoff and jitter. Optional `Retry-After` is respected.

```go
client := oddrip.New(
    oddrip.RetryConfigOption(oddrip.RetryConfig{
        MaxAttempts:   5,
        InitialDelay:  1 * time.Second,
        MaxDelay:      60 * time.Second,
        JitterPercent: 0.2,
    }),
)
```

## Concurrent requests

```go
results, err := oddrip.DoConcurrent(ctx, 3, func(i int) (*types.GetMarketResponse, error) {
    return client.Markets.Get(ctx, tickers[i])
})
```

## WebSocket (market data)

The client supports the [Kalshi Market Data WebSocket API](https://docs.kalshi.com/getting_started/quick_start_websockets). The WebSocket API is **read-only**: subscribe to channels (ticker, orderbook, trades, fills, etc.) and manage subscriptions; there is no order placement. Place and cancel orders via the REST client.

```go
// Connect (uses same auth as client)
conn, err := client.ConnectWS(ctx)
if err != nil {
    return err
}
defer conn.Close()

// Subscribe to channels
subs, err := conn.Subscribe(ctx, types.SubscribeParams{
    Channels:     []string{types.WSChannelTicker, types.WSChannelOrderbookDelta},
    MarketTicker: "FED-23DEC-T3.00",
})
if err != nil {
    return err
}

// Receive messages (ticker, orderbook_snapshot, orderbook_delta, etc.)
for msg := range conn.Messages() {
    switch msg.Type {
    case "ticker":
        // decode msg.Msg into your ticker struct
    case "orderbook_snapshot", "orderbook_delta":
        // ...
    }
}
```

Commands: `Subscribe`, `Unsubscribe`, `ListSubscriptions`, `UpdateSubscription`. Server errors are returned as `*oddrip.WSError` (Code + Message). Optional `oddrip.WSHost`, `oddrip.WSPath` for non-default endpoints.

## Package layout

- `oddrip` – REST client, WebSocket client, service methods (`Markets`, `Orders`, `Exchange`, `Portfolio`, `Account`)
- `oddrip/types` – request/response and enum types (including WebSocket command/response types)
- `oddrip/internal/errors` – typed API errors
- `oddrip/internal/retry` – retry with backoff
- `oddrip/internal/auth` – auth provider interface and RSA-PSS signer
- `oddrip/internal/transport` – minimal `Doer` interface for shared HTTP execution

All methods take `context.Context` and support cancellation and timeouts. The client is safe for concurrent use.
