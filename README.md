# Oddrip

[![CI](https://github.com/UTXOnly/oddrip/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/UTXOnly/oddrip/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/UTXOnly/oddrip/oddrip.svg)](https://pkg.go.dev/github.com/UTXOnly/oddrip/oddrip)

Go client for the [Kalshi Trade API](https://docs.kalshi.com/openapi.yaml): REST for orders, portfolio, markets, events, and exchange info, plus WebSocket for real-time market data (ticker, orderbook, trades, fills, and related channels). One library, same auth; use REST to trade and WebSocket to stream.

REST coverage: **Exchange** (status, schedule, user_data_timestamp, historical cutoff, series fee changes), **Markets** (list, get, orderbook, **orderbooks**, trades, **historical** list/get/trades/candlesticks), **Events** (list, list multivariate, get, get metadata per [Get Events](https://docs.kalshi.com/api-reference/events/get-events)), **Orders** (list, get, queue positions, **V2 event orders** create/cancel/cancel-all/amend/decrease/batch), **Portfolio** (balance, fills, positions, **settlements**, **deposits**, **withdrawals**, **intra-exchange transfers**, **target balance allocation**, **historical** fills/orders/positions), **Account** (API limits, **endpoint costs**), **Live data** (**weather index** and calibrations, event live data), **Series** (list, get, per-series **market and event candlesticks**, forecast percentile history), **Order groups** (list/create/get/delete/reset/trigger/limit), **Subaccounts** (create, balances, transfer, transfer history, netting). Also `Markets.GetCandlesticks` (batch), `Portfolio.GetTotalRestingOrderValue`, and multivariate event collections on `Events` (list/get/`CreateMarketInMultivariateCollection`). 66 of the 96 paths in the vendored spec are covered; communications (RFQ/quotes), milestones, API-key management, FCM, and search remain unimplemented. See `CHANGELOG.md` and [Kalshi changelog](https://docs.kalshi.com/changelog) for API-facing changes.

Module path: `github.com/UTXOnly/oddrip`. Import the client as `github.com/UTXOnly/oddrip/oddrip` and types as `github.com/UTXOnly/oddrip/oddrip/types`. Current release **v0.6.0** — pin with `go get github.com/UTXOnly/oddrip/oddrip@v0.6.0`; runtime string `oddrip.Version` matches the module tag.

---

## Install

```bash
go get github.com/UTXOnly/oddrip/oddrip@latest
# or pin: go get github.com/UTXOnly/oddrip/oddrip@v0.6.0
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

Non-2xx responses are returned as `*oddrip.APIError`. Use `errors.As` to inspect status, message, and body. This includes the case where every retry attempt was rate-limited or failed server-side: the last response is surfaced as an `APIError` with its real status code (e.g. 429), never as a nil response.

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

The client retries on 429 and 5xx with exponential backoff and jitter. It honors `Retry-After` in both delta-seconds and HTTP-date forms. Backoff waits are cancelled by the request context, so a cancelled or expired `ctx` returns promptly instead of sleeping out the delay. Tune with `RetryConfigOption`; `MaxAttempts` below 1 is treated as 1.

Retries apply to every method, including order creation. Kalshi deduplicates `POST /portfolio/events/orders` on `client_order_id`, so always set one on `CreateOrderV2Request` — it is what makes a retried create idempotent.

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

The client is safe for concurrent use. `DoConcurrent` fans out `n` calls with at most `maxInFlight` running at once (pass `0` for unbounded) and returns results in index order; each result carries its own error. If `ctx` is cancelled, the results collected so far are returned along with `ctx.Err()`.

```go
results, err := oddrip.DoConcurrent(ctx, len(tickers), 8, func(i int) (*types.GetMarketResponse, error) {
    return client.Markets.Get(ctx, tickers[i])
})
```

---

## Prices, counts, timestamps

The API emits prices as dollar strings (`"0.4500"`, up to 6 decimals in responses), contract counts as fixed-point strings (`"10.00"`), and times as RFC 3339 strings. The response structs keep those as `string` so nothing is lost; `types` provides lossless parsers when you need numbers.

```go
price, err := types.ParseDollars("0.4500") // Dollars, int64 scaled 1e-6
price.String()                             // "0.4500" — safe to send back in a request
price.Cents()                              // 45 (truncates toward zero)
price.Float64()                            // 0.45

qty, _ := types.ParseCount("10")           // Count, int64 scaled 1e-2
qty.String()                               // "10.00"

ts, _ := types.ParseTime("2022-11-22T20:44:01Z")
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
    case types.WSTypeTicker:
        var t types.TickerMsg
        if err := msg.Decode(&t); err != nil {
            return err
        }
        bid, _ := types.ParseDollars(t.YesBidDollars)
        fmt.Println(t.MarketTicker, bid.Cents())
    case types.WSTypeOrderbookSnapshot:
        var snap types.OrderbookSnapshotMsg
        _ = msg.Decode(&snap) // YesDollarsFp / NoDollarsFp are []OrderbookLevel{PriceDollars, CountFp}
    case types.WSTypeOrderbookDelta:
        var d types.OrderbookDeltaMsg
        _ = msg.Decode(&d)
    case types.WSTypeFill:
        var f types.FillMsg
        _ = msg.Decode(&f)
    }
}

// Messages() closes when the connection is gone. Err() says why.
if err := conn.Err(); !errors.Is(err, oddrip.ErrWSClosed) {
    // dead socket, slow consumer, or server close: reconnect and re-subscribe
}
```

Every server message type has a `types.WSType*` constant and a typed `*Msg` struct (`TickerMsg`, `OrderbookSnapshotMsg`, `OrderbookDeltaMsg`, `TradeMsg`, `FillMsg`, `MarketPositionMsg`, `UserOrderMsg`, `OrderGroupUpdatesMsg`, `MarketLifecycleV2Msg`, the RFQ/quote messages, `PythValueMsg`, `CFBenchmarksValueMsg`, ...). `msg.Decode(&v)` unmarshals the payload.

**Connection lifecycle.** The connection sends keepalive pings and enforces a read deadline, so a half-open socket is detected within `WSReadTimeout` (default 90s) instead of blocking forever. Messages are never dropped silently: if the consumer of `Messages()` falls behind and the buffer (`WSBufferSize`, default 4096) fills, the connection is failed with `ErrWSSlowConsumer` and closed, because a gap in an `orderbook_delta` stream would otherwise corrupt your local book without warning. Treat `Messages()` closing as "reconnect and re-subscribe"; `Err()` returns the terminal error (`ErrWSClosed` after a clean `Close`, `ErrWSSlowConsumer`, or the underlying read error) and `Done()` is closed when the read loop exits. Options: `WSBufferSize(n)`, `WSPingInterval(d)` (default 30s, `<= 0` disables), `WSReadTimeout(d)` (default 90s, `<= 0` disables). `Close` is idempotent; all commands are safe to call concurrently.

**Commands:** `Subscribe`, `Unsubscribe`, `ListSubscriptions`, `UpdateSubscription` (add/remove markets, underlyings, or CF Benchmarks indices on a subscription). **Channels** (see `types`): ticker, orderbook_delta, trade, fill, market_positions, market_lifecycle_v2, multivariate_market_lifecycle, communications, order_group_updates, user_orders, pyth_value, cfbenchmarks_value, cfbenchmarks_value_5hz. The cfbenchmarks channels take `IndexIDs` instead of market tickers (`[]string{"all"}` for every index). Server errors come back as `*oddrip.WSError` (Code and Message). Use `oddrip.WSHost`, `oddrip.WSPath`, and `oddrip.WSScheme` to point at a different host or path (e.g. demo).

---

## Package layout

- **`oddrip`** – REST client, `ConnectWS`, and service methods (`Exchange`, `Markets`, `Events`, `Orders`, `Portfolio`, `Account`, `LiveData`, `Series`, `OrderGroups`, `Subaccounts`).
- **`oddrip/types`** – Request/response and enum types for both REST and WebSocket (e.g. `CreateOrderV2Request`, `SubscribeParams`, `WSMessage`, channel and message-type constants), typed WebSocket payloads, and the `Dollars`/`Count`/`ParseTime` helpers.
- **`oddrip/internal/retry`** – Retry with backoff and retryable-status classification.
- **`oddrip/internal/auth`** – RSA-PSS request signer.

All public methods take `context.Context`. The client and WebSocket connection are safe for concurrent use.

---

## Development and releases

CI runs on every pull request and on `main`: `gofmt`, `go mod tidy` drift, `go vet`, `staticcheck`, `govulncheck`, and `go test -race -shuffle=on` on Go 1.24 and stable across Linux, macOS, and Windows. A `version` job checks that `oddrip/version.go`, the top `CHANGELOG.md` entry, and the README pin agree, and fails a code-changing PR whose version is already tagged.

Releases are cut by merging to `main`. To ship a version:

1. Bump `const Version` in `oddrip/version.go`.
2. Add a `## [X.Y.Z] — YYYY-MM-DD` section at the top of `CHANGELOG.md`; its body becomes the release notes.
3. Update the `@vX.Y.Z` pin in this README.

When the merge lands and all checks pass, the `release` job tags `vX.Y.Z`, publishes a GitHub Release with the CHANGELOG section, and warms `proxy.golang.org`. A merge whose version is already tagged (docs-only changes) is a no-op.

**Versioning.** Semver. While the module is at v0, a minor release may contain breaking changes; they are always listed first under `### Breaking` in the CHANGELOG with migration notes. CI runs `gorelease` against the previous tag and refuses a release that has API-incompatible changes without that section, or that declares one on a patch bump. Patch releases never break. Behavioral changes that `gorelease` cannot see (e.g. a connection now closing where it used to hang) are declared under the same heading.
