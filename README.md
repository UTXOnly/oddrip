# Oddrip

[![CI](https://github.com/UTXOnly/oddrip/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/UTXOnly/oddrip/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/UTXOnly/oddrip/oddrip.svg)](https://pkg.go.dev/github.com/UTXOnly/oddrip/oddrip)

Go client for the [Kalshi Trade API](https://docs.kalshi.com/): REST plus WebSocket market data. Tracks vendored OpenAPI **3.29.0** / AsyncAPI **2.0.0**.

```bash
go get github.com/UTXOnly/oddrip/oddrip@v0.6.0
```

Go 1.24+. Import the client as `github.com/UTXOnly/oddrip/oddrip` and types as `github.com/UTXOnly/oddrip/oddrip/types`. `oddrip.Version` matches the module tag.

## Client

```go
client := oddrip.New() // public endpoints
```

```go
key, err := oddrip.ParsePrivateKeyFromPEM(privateKeyPEM)
if err != nil {
    return err
}
client := oddrip.New(
    oddrip.Auth(oddrip.NewKalshiSigner(apiKeyID, key)),
)
```

Default base URL is `https://api.elections.kalshi.com/trade-api/v2`. Pass `oddrip.BaseURL(...)` for demo (`https://demo-api.kalshi.co/trade-api/v2`) or `oddrip.HTTPClient(...)` for a custom transport. Auth is RSA-PSS (PKCS#8 or PKCS#1 PEM); the same signer is used for REST and the WebSocket handshake.

## REST

Services: `Exchange`, `Markets`, `Events`, `Series`, `Orders`, `OrderGroups`, `Portfolio`, `Subaccounts`, `Account`, `LiveData`. Every call takes `context.Context`. Optional query params are pointer fields on `*Opts` structs — omit or leave nil.

```go
status, err := client.Exchange.GetStatus(ctx)
market, err := client.Markets.Get(ctx, "TICKER-24JAN01")
events, err := client.Events.List(ctx, &types.GetEventsOpts{Status: "open"})

_, err = client.Orders.CreateV2(ctx, &types.CreateOrderV2Request{
    Ticker:        "TICKER-24JAN01",
    ClientOrderID: "cli-1", // set this so retries cannot double-place
    Side:          types.BookSideBid,
    Count:         "1.00",
    Price:         "0.4500",
    TimeInForce:   types.TimeInForceGTC,
})
```

64 of 96 OpenAPI paths. Not implemented: RFQ/quotes, FCM, API keys, milestones, search, structured targets, incentive programs. Method list: [pkg.go.dev](https://pkg.go.dev/github.com/UTXOnly/oddrip/oddrip).

List responses include `Cursor` when there is another page:

```go
limit := int64(100)
opts := &types.GetMarketsOpts{Limit: &limit}
for {
    resp, err := client.Markets.List(ctx, opts)
    if err != nil {
        return err
    }
    // ...
    if resp.Cursor == "" {
        break
    }
    opts.Cursor = resp.Cursor
}
```

Non-2xx responses are `*oddrip.APIError` (status, message, request ID, body). An empty ticker or ID path parameter returns `oddrip.ErrEmptyPathParam` without sending — an empty order ID would otherwise hit CancelAll.

## Retries

Default 4 attempts, exponential backoff with jitter, honors `Retry-After` (delta-seconds or HTTP-date). A cancelled `ctx` aborts the wait. Tune with `RetryConfigOption`; `MaxAttempts` below 1 is treated as 1.

| Request | 429 | 5xx / timeout |
|---|---|---|
| Idempotent — GET, PUT, DELETE, and POSTs the server deduplicates (`CreateV2` / `BatchCreateV2` with `client_order_id` on every order, `Subaccounts.Transfer`, `SetTargetBalanceAllocation`) | retried | retried |
| Non-idempotent — `AmendV2`, `DecreaseV2`, creates without `client_order_id`, `OrderGroups.Create`, `Subaccounts.Create`, `CreateMarketInMultivariateCollection` | retried | not retried |

A 429 means the server rejected the request before acting. A 5xx or dropped connection is ambiguous — the write may already be applied. After an ambiguous failure of a non-idempotent write, check `Orders.Get` before resending.

```go
client := oddrip.New(oddrip.RetryConfigOption(oddrip.RetryConfig{MaxAttempts: 1})) // disable retries
```

## Concurrent requests

The client is safe for concurrent use. `DoConcurrent(ctx, n, maxInFlight, fn)` runs at most `maxInFlight` calls at once (`0` is unbounded) and returns results in index order. Cancel returns the results collected so far plus `ctx.Err()`.

```go
results, err := oddrip.DoConcurrent(ctx, len(tickers), 8, func(i int) (*types.GetMarketResponse, error) {
    return client.Markets.Get(ctx, tickers[i])
})
```

## Prices, counts, timestamps

The API emits dollar strings (`"0.4500"`, up to 6 decimals), contract counts as fixed-point strings (`"10.00"`), and RFC 3339 times. Response structs keep those as `string`; parse when you need numbers:

```go
price, err := types.ParseDollars("0.4500") // int64 scaled 1e-6
price.String()                             // "0.4500"
price.Cents()                              // 45 (truncates toward zero)

qty, _ := types.ParseCount("10")           // int64 scaled 1e-2
ts, _ := types.ParseTime("2022-11-22T20:44:01Z")
```

## WebSocket

Read-only market data. Auth is required. Place orders over REST.

```go
conn, err := client.ConnectWS(ctx)
if err != nil {
    return err
}
defer conn.Close()

if _, err := conn.Subscribe(ctx, types.SubscribeParams{
    Channels:     []string{types.WSChannelTicker, types.WSChannelOrderbookDelta},
    MarketTicker: "FED-23DEC-T3.00",
}); err != nil {
    return err
}

for msg := range conn.Messages() {
    switch msg.Type {
    case types.WSTypeTicker:
        var t types.TickerMsg
        if err := msg.Decode(&t); err != nil {
            return err
        }
        // ...
    }
}
if err := conn.Err(); !errors.Is(err, oddrip.ErrWSClosed) {
    // dead socket, slow consumer, or server close: reconnect and re-subscribe
}
```

Commands: `Subscribe`, `Unsubscribe`, `ListSubscriptions`, `UpdateSubscription`. Channel names and `WSType*` constants are in `types`. CF Benchmarks channels take `IndexIDs` (`[]string{"all"}` for every index). Server errors are `*oddrip.WSError`. Point at demo with `WSHost` / `WSPath` / `WSScheme`.

- If `Messages()` falls behind, the connection fails with `ErrWSSlowConsumer` (buffer default 4096) rather than dropping deltas. Reconnect and re-snapshot any local book.
- Keepalive ping every 30s and a 90s read deadline. `WSReadTimeout(0)` / `WSPingInterval(0)` disable either.
- `get_snapshot` needs `SID` or a one-element `Sids`. It returns when the first `orderbook_snapshot` for that subscription arrives (`Type` is `"orderbook_snapshot"`); the frames also go to `Messages()`.
- `indexlist` / `underlying_list` replies are not `Type: "ok"`.
- A multi-channel `Subscribe` that fails partway returns the accepted SIDs alongside the `*WSError`.
- `Close` is idempotent. Commands are safe to call concurrently.

## Examples

`cmd/example` (REST) and `cmd/example/websocket_example` log request URLs and raw responses. Demo is the default and is incomplete; set `BASE_URL` to production.

## Releases

CI runs `gofmt`, `go mod tidy`, `go vet`, `staticcheck`, `govulncheck`, and `go test -race -shuffle=on` on Go 1.24 and stable (Linux, macOS, Windows). Merging to `main` tags `vX.Y.Z` and publishes a GitHub Release when that version is not already tagged.

1. Bump `const Version` in `oddrip/version.go`.
2. Add a `## [X.Y.Z] — YYYY-MM-DD` section at the top of `CHANGELOG.md`.
3. Update the `@vX.Y.Z` pin in this README.

Semver. While at v0, a minor release may break; those changes go first under `### Breaking`. CI runs `gorelease` against the previous tag and refuses a release that has API-incompatible changes without that heading, or that declares one on a patch bump.
