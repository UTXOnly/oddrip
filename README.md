# Oddrip

Go client for the [Kalshi Trade API](https://docs.kalshi.com/openapi.yaml): REST for orders, portfolio, markets, events, and exchange info, plus WebSocket for real-time market data (ticker, orderbook, trades, fills, and related channels). One library, same auth; use REST to trade and WebSocket to stream.

REST coverage: **Exchange** (status, schedule, user_data_timestamp, historical cutoff, series fee changes), **Markets** (list, get, orderbook, **orderbooks**, trades, **historical** list/get/trades/candlesticks), **Events** (list, list multivariate, get, get metadata per [Get Events](https://docs.kalshi.com/api-reference/events/get-events)), **Orders** (list, get, queue positions, **V2 event orders** create/cancel/cancel-all/amend/decrease/batch), **Portfolio** (balance, fills, positions, **settlements**, **deposits**, **withdrawals**, **intra-exchange transfers**, **target balance allocation**, **historical** fills/orders/positions), **Account** (API limits, **endpoint costs**), **Live data** (**weather index** and calibrations, event live data). The OpenAPI spec also defines order groups, communications, milestones, and other endpoints; those can be added as needed. See `CHANGELOG.md` and [Kalshi changelog](https://docs.kalshi.com/changelog) for API-facing changes.

Module path: `github.com/UTXOnly/oddrip`. Import the client as `github.com/UTXOnly/oddrip/oddrip` and types as `github.com/UTXOnly/oddrip/oddrip/types`. Release **v0.5.0** — pin with `go get github.com/UTXOnly/oddrip/oddrip@v0.5.0` after tagging; runtime string `oddrip.Version` matches the module release.

---

## Install

```bash
go get github.com/UTXOnly/oddrip/oddrip@latest
# or pin: go get github.com/UTXOnly/oddrip/oddrip@v0.5.0
```

---

## Initialize client

```go
import (
    "github.com/UTXOnly/oddrip/oddrip"
    "github.com/UTXOnly/oddrip/oddrip/types"
)

// Public endpoints only (no auth)
client := oddrip.New()

// Authenticated: API key + RSA-PSS request signing (required for orders, portfolio, WebSocket)
key, _ := oddrip.ParsePrivateKeyFromPEM(privateKeyPEM)
client := oddrip.New(
    oddrip.Auth(oddrip.NewKalshiSigner(apiKeyID, key)),
    oddrip.BaseURL("https://api.elections.kalshi.com/trade-api/v2"),
)
```

Kalshi uses request signing: you sign each HTTP request (method + path + timestamp) with your private key. Use `ParsePrivateKeyFromPEM` for PKCS#8 or PKCS#1 PEM; pass the key and key ID to `NewKalshiSigner`. The same auth is used for REST and for the WebSocket handshake.

---

## REST: requests and services

The client exposes services that match the API: `Exchange`, `Markets`, `Events`, `Orders`, `Portfolio`, `Account`, `LiveData`. All calls take `context.Context` (for timeouts and cancellation).

```go
ctx := context.Background()

status, err := client.Exchange.GetStatus(ctx)
market, err := client.Markets.Get(ctx, "TICKER-24JAN01")
events, err := client.Events.List(ctx, &types.GetEventsOpts{Status: "open"})
orders, err := client.Orders.List(ctx, &types.GetOrdersOpts{Status: "resting"})
balance, err := client.Portfolio.GetBalance(ctx, nil)
```

Weather markets are priced off Kalshi's own city temperature index; `LiveData` serves that index and the calibration timeline behind it.

```go
// Last hour of the Miami index, with each station's reading and QC disposition.
index, err := client.LiveData.GetWeatherIndex(ctx, "miami", &types.GetWeatherIndexOpts{
    LastSec:  ptr(int64(3600)),
    Detailed: ptr(true),
})

// Station weights and offsets used to compute those values.
cal, err := client.LiveData.GetWeatherIndexCalibrations(ctx, "miami")
```

Minutes where the index quorum failed are absent from `Timeseries`, so gaps in the series are real gaps.

Optional parameters use pointer fields in opts structs (e.g. `Limit *int64`, `Cursor string`). Omit or set to `nil` what you don’t need.

---

## Pagination

List endpoints return a `Cursor` when there are more results. Pass it back on the next call.

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

---

## Error handling

Non-2xx responses are returned as `*oddrip.APIError`. Use `errors.As` to inspect status, message, and body.

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

---

## Retries

The client retries on 429 and 5xx with exponential backoff and jitter. It honors `Retry-After` when present. You can tune behavior with `RetryConfigOption`.

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

---

## Concurrent requests

The client is safe for concurrent use. For bounded concurrency (e.g. many tickers), use `DoConcurrent`:

```go
results, err := oddrip.DoConcurrent(ctx, 3, func(i int) (*types.GetMarketResponse, error) {
    return client.Markets.Get(ctx, tickers[i])
})
```

---

## WebSocket (market data)

The WebSocket API is **read-only**: subscribe to channels and receive streams. There is no order placement over WebSocket; use the REST client for that. Auth is required; the same signer used for REST is applied to the WebSocket handshake.

```go
conn, err := client.ConnectWS(ctx)
if err != nil {
    return err
}
defer conn.Close()

subs, err := conn.Subscribe(ctx, types.SubscribeParams{
    Channels:     []string{types.WSChannelTicker, types.WSChannelOrderbookDelta},
    MarketTicker: "FED-23DEC-T3.00",
})
if err != nil {
    return err
}

for msg := range conn.Messages() {
    switch msg.Type {
    case "ticker":
        // decode msg.Msg
    case "orderbook_snapshot", "orderbook_delta":
        // ...
    }
}
```

**Commands:** `Subscribe`, `Unsubscribe`, `ListSubscriptions`, `UpdateSubscription` (add/remove markets, underlyings, or CF Benchmarks indices on a subscription). **Channels** (see `types`): ticker, orderbook_delta, trade, fill, market_positions, market_lifecycle_v2, multivariate_market_lifecycle, communications, order_group_updates, user_orders, pyth_value, cfbenchmarks_value, cfbenchmarks_value_5hz. The cfbenchmarks channels take `IndexIDs` instead of market tickers (`[]string{"all"}` for every index). Server errors come back as `*oddrip.WSError` (Code and Message). Use `oddrip.WSHost`, `oddrip.WSPath`, and `oddrip.WSScheme` to point at a different host or path (e.g. demo).

---

## Package layout

- **`oddrip`** – REST client, `ConnectWS`, and service methods (`Exchange`, `Markets`, `Events`, `Orders`, `Portfolio`, `Account`, `LiveData`).
- **`oddrip/types`** – Request/response and enum types for both REST and WebSocket (e.g. `CreateOrderV2Request`, `SubscribeParams`, `WSMessage`, channel constants).
- **`oddrip/internal/errors`** – Parsing of API error responses.
- **`oddrip/internal/retry`** – Retry with backoff.
- **`oddrip/internal/auth`** – Auth provider interface and RSA-PSS signer.
- **`oddrip/internal/transport`** – Minimal HTTP `Doer` interface (not used directly by callers).

All public methods take `context.Context`. The client and WebSocket connection are safe for concurrent use.
