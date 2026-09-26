# Oddrip

[![CI](https://github.com/UTXOnly/oddrip/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/UTXOnly/oddrip/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/UTXOnly/oddrip/oddrip.svg)](https://pkg.go.dev/github.com/UTXOnly/oddrip/oddrip)

Go client for the [Kalshi Trade API](https://docs.kalshi.com/): REST plus WebSocket market data, for event contracts and [perps](#perps). Tracks vendored OpenAPI **3.30.0** / AsyncAPI **2.0.0** and perps OpenAPI **0.0.1** / AsyncAPI **2.0.0**.

```bash
go get github.com/UTXOnly/oddrip/oddrip@v0.7.0
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

Default base URL is `https://external-api.kalshi.com/trade-api/v2`, the production host Kalshi recommends. The older shared host `https://api.elections.kalshi.com/trade-api/v2` (the default before v0.6.2) is still supported; pass `oddrip.BaseURL(...)` to use it, or for demo (`https://demo-api.kalshi.co/trade-api/v2`). Switching hosts does not affect signing — the signed message covers the path only. `oddrip.HTTPClient(...)` swaps the transport. Auth is RSA-PSS (PKCS#8 or PKCS#1 PEM) over `timestamp + METHOD + path` (query string excluded); the same signer is used for REST and the WebSocket handshake.

## REST

Services: `Exchange`, `Markets`, `Events`, `Series`, `Orders`, `OrderGroups`, `Portfolio`, `Subaccounts`, `Account`, `LiveData`, and `Perps` for perpetual futures ([below](#perps)). Every call takes `context.Context`. Optional query params live on `*Opts` structs: strings are sent only when non-empty, numbers and bools are pointers — leave nil to omit. A nil `*Opts` is fine.

```go
status, err := client.Exchange.GetStatus(ctx)
market, err := client.Markets.Get(ctx, "TICKER-24JAN01")
events, err := client.Events.List(ctx, &types.GetEventsOpts{Status: "open"})

_, err = client.Orders.CreateV2(ctx, &types.CreateOrderV2Request{
    Ticker:                  "TICKER-24JAN01",
    ClientOrderID:           "cli-1", // set this so retries cannot double-place
    Side:                    types.BookSideBid,
    Count:                   "1.00",
    Price:                   "0.4500",
    TimeInForce:             types.TimeInForceGTC,
    SelfTradePreventionType: types.SelfTradeTakerAtCross, // required by the API
})
```

65 of 96 OpenAPI paths. Not implemented: communications (RFQs, quotes, block-trade proposals), FCM, API keys, milestones and milestone live data (`/live_data/batch`, `/live_data/milestone/*`), search, structured targets, incentive programs, `GET /events/fee_changes`, and `/account/api_usage_level/*`. Method list: [pkg.go.dev](https://pkg.go.dev/github.com/UTXOnly/oddrip/oddrip).

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

Non-2xx responses are `*oddrip.APIError` with `StatusCode`, `Code`, `Message`, and `RawBody` (first 512 bytes). `Code`, `Message`, and `Details` are decoded from up to 64 KiB of the body; a larger body is not decoded and leaves them empty. Kalshi's production error bodies nest `code` / `message` under `"error"` (the spec shows them flat) and parameter-binding 400s use `{"msg": ...}`; all three shapes are parsed. `RequestID` is read from a `Request-Id` header Kalshi does not currently send. An empty ticker or ID path parameter returns `oddrip.ErrEmptyPathParam` without sending — an empty order ID would otherwise hit CancelAll.

## Retries

Default 4 attempts (500ms initial, ×2, ±20% jitter, 30s cap). `Retry-After` (delta-seconds or HTTP-date) replaces the backoff for that attempt, still capped at `MaxDelay`. A cancelled `ctx` aborts the wait. Tune with `RetryConfigOption`; `MaxAttempts` below 1 is treated as 1.

| Request | 429 | 5xx / timeout |
|---|---|---|
| Idempotent — GET, PUT, DELETE; POSTs the server deduplicates (`CreateV2` / `BatchCreateV2` / `Perps.Orders.Create` with `client_order_id` on every order, `Subaccounts.Transfer` / `Perps.Portfolio.TransferBetweenSubaccounts` via `client_transfer_id`, `Perps.FCM.CreateSubtrader`, whose ID comes from your suffix); `SetTargetBalanceAllocation`, which sets absolute state | retried | retried |
| Non-idempotent — `AmendV2`, `DecreaseV2`, `Perps.Orders.Amend` / `Decrease`, creates without `client_order_id`, `OrderGroups.Create` / `Perps.OrderGroups.Create`, `Subaccounts.Create` / `Perps.Portfolio.CreateSubaccount`, `Portfolio.CreateIntraExchangeTransfer`, `CreateMarketInMultivariateCollection` | retried | not retried |

A 429 means the server rejected the request before acting. A 5xx or dropped connection is ambiguous — the write may already be applied. After an ambiguous failure of a non-idempotent write, check `Orders.Get` before resending. A replayed `CreateV2` whose first attempt did land is rejected with `409` ("order with this `client_order_id` already exists"), so treat a 409 `*APIError` after a retry as success and look the order up. The same holds for `Perps.Orders.Create` (check `Perps.Orders.Get` / `List`) and for `Perps.FCM.CreateSubtrader`, where a 409 means the subtrader exists.

```go
client := oddrip.New(oddrip.RetryConfigOption(oddrip.RetryConfig{MaxAttempts: 1})) // disable retries
```

## Concurrent requests

The client is safe for concurrent use. `DoConcurrent(ctx, n, maxInFlight, fn)` runs at most `maxInFlight` calls at once (`0` is unbounded) and returns results in index order. A positive `maxInFlight` also caps the worker goroutines at `min(n, maxInFlight)`, so `n` can be large without creating `n` goroutines. Cancel returns the results collected so far plus `ctx.Err()`.

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

Read-only streams: public market data plus your own fills, orders, positions, order-group and RFQ activity. Every connection needs auth, including for public channels. There are no order commands over WebSocket; place orders over REST.

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
    // dead socket, slow consumer, malformed frame, or server close: reconnect and re-subscribe
}
```

Commands: `Subscribe`, `Unsubscribe`, `ListSubscriptions`, `UpdateSubscription`. Channel names and `WSType*` constants are in `types`. CF Benchmarks channels take `IndexIDs` (`[]string{"all"}` for every index). Command rejections are returned as `*oddrip.WSError`. Default endpoint is `wss://external-api-ws.kalshi.com/trade-api/ws/v2`, the production host Kalshi recommends; the older shared `api.elections.kalshi.com` (the default before v0.6.2) still accepts connections. Point elsewhere with `WSHost` / `WSPath` / `WSScheme` — note the recommended REST and WebSocket hosts differ (`external-api` vs `external-api-ws`), so a `BaseURL` override does not imply a `WSHost` one.

- If `Messages()` falls behind, the connection fails with `ErrWSSlowConsumer` (buffer default 4096) rather than dropping deltas. Reconnect and re-snapshot any local book.
- Errors scoped to a subscription arrive on `Messages()` as `Type: "error"` with a `SID`, not as a returned `*WSError`. Codes 10 (channel error) and 25 (subscription buffer overflow) are terminal for that subscription — resubscribe. Decode into `types.ErrorMsg`.
- Client keepalive ping every 30s and a 90s read deadline (Kalshi also pings every 10s; any frame extends the deadline). `WSReadTimeout(0)` / `WSPingInterval(0)` disable either.
- A text frame that is not valid JSON fails the connection: `Err()` wraps `ErrWSMalformedFrame` and `Messages()` closes, same as a slow consumer.
- `get_snapshot` needs `SID` or a one-element `Sids`. It returns when the first `orderbook_snapshot` for that subscription arrives (`Type` is `"orderbook_snapshot"`); the frames also go to `Messages()`.
- `indexlist` / `underlying_list` replies are not `Type: "ok"`.
- A multi-channel `Subscribe` that fails partway returns the accepted SIDs alongside the `*WSError`.
- Every socket write is bounded by `WSWriteTimeout` (default 10s) or the command's context deadline, whichever is sooner. A write that times out fails the connection: `Err()` wraps `ErrWSWriteTimeout` and `Messages()` closes, same as a slow consumer; the command whose own deadline cut the write returns `context.DeadlineExceeded`. A command that gives up while waiting its turn to write returns `ctx.Err()` and leaves the connection healthy.
- `Close` is idempotent and returns within about `WSWriteTimeout` plus five seconds even if the peer has stopped reading. Commands are safe to call concurrently.

## Perps

Kalshi's perpetual futures ("perps"; the API itself says "margin") hang off `client.Perps`, one service per area: `Exchange`, `Account`, `Markets`, `Orders`, `OrderGroups`, `Portfolio`, `Risk`, `Fees`, `Funding`, `ExitTriggers`, `FCM`. They share the client's base URL, signer, and retry policy — perps REST is served under `/trade-api/v2/margin/...` on the same host. Kalshi's perps spec lists only the `external-api` hosts (demo: `https://external-api.demo.kalshi.co/trade-api/v2`), not the shared `api.elections.kalshi.com` / `demo-api.kalshi.co`, and Kalshi describes production access as rolling out member by member; `Perps.Exchange.GetEnabled` reports whether margin trading is on for your account. Types keep the spec's names, which are mostly `Margin`-prefixed (`types.MarginMarket`, `types.CreateMarginOrderRequest`).

```go
markets, err := client.Perps.Markets.List(ctx, nil)
book, err := client.Perps.Markets.GetOrderbook(ctx, "TICKER-PERP", nil)

_, err = client.Perps.Orders.Create(ctx, &types.CreateMarginOrderRequest{
    Ticker:                  "TICKER-PERP",
    ClientOrderID:           "cli-1", // required by the API; also makes retries safe
    Side:                    types.BookSideBid,
    Count:                   "1.00",
    Price:                   "65000.0000",
    TimeInForce:             types.TimeInForceGTC,
    SelfTradePreventionType: types.SelfTradeTakerAtCross,
})
```

38 of 38 perps OpenAPI paths. Where perps differ from event contracts:

- Sides are `bid` / `ask`. The order book is `Orderbook.Bids` / `Orderbook.Asks`, each a `[]types.OrderbookLevel`.
- Positions are signed (`MarginPosition.Position` is negative when short). Margin used, ROE, leverage, and liquidation price are `nil` when the exchange cannot attribute them to a single position (`IsPortfolio`).
- `Perps.Markets.List` and `Perps.Portfolio.GetPositions` are not paginated. `Perps.Markets.GetTrades` requires a ticker.
- `Perps.Funding.GetHistory` takes inclusive UTC dates (`YYYY-MM-DD`); `Perps.Funding.GetHistoricalRates` takes Unix seconds.
- What a nil `Subaccount` means depends on the endpoint. It means every subaccount on `Orders.List`, `Orders.CancelAll`, `OrderGroups.List` / `Get`, and `Funding.GetHistory`. It means the primary account (0) on single-order and order-group writes and on every cross-margin exit-trigger call.
- Exit triggers (stop-loss, take-profit, trailing stop) are PUTs that replace what is set: a leg you omit is cleared. Cross-margin creates that append (`Count` or `AnchorOrderID`) need a `ClientTriggerID`; the server returns the original trigger when that key is replayed, which is what makes the PUT safe to retry.
- `Portfolio.CreateIntraExchangeTransfer` (on the event-contract `Portfolio` service, next to `ListIntraExchangeTransfers`) moves funds between event-contract and margin balances. `Amount` is in centicents and the transfer completes asynchronously. Kalshi documents it as unavailable until the production rollout.

### Perps WebSocket

`client.ConnectPerpsWS(ctx)` dials `wss://external-api-margin-ws.kalshi.com/trade-api/ws/v2/margin` and returns the same `*WSConn` as `ConnectWS`: same commands, lifecycle, and errors as above. For demo pass `oddrip.WSHost("external-api-margin-ws.demo.kalshi.co")`. Only `orderbook_delta`, `ticker`, `trade`, `fill`, `user_orders`, and `order_group_updates` exist; markets are selected by ticker; `UpdateSubscription` supports `add_markets`, `delete_markets`, and `get_snapshot`.

Channel names and `msg.Type` values are the event-contract ones, but the payloads are not. Decode into `types.MarginOrderbookSnapshotMsg`, `MarginOrderbookDeltaMsg`, `MarginTickerMsg`, `MarginTradeMsg`, `MarginFillMsg`, or `MarginUserOrderMsg`, and error frames into `types.MarginErrorMsg`; `order_group_updates` uses `types.OrderGroupUpdatesMsg`. Timestamps are Unix milliseconds (`TsMs`, `CreatedTsMs`, ...). Perps tickers arrive at most once per market per second and carry optional mark, reference, and funding-rate fields.

```go
conn, err := client.ConnectPerpsWS(ctx)
if err != nil {
    return err
}
defer conn.Close()
if _, err := conn.Subscribe(ctx, types.SubscribeParams{
    Channels:     []string{types.WSChannelTicker},
    MarketTicker: "TICKER-PERP",
}); err != nil {
    return err
}
for msg := range conn.Messages() {
    if msg.Type == types.WSTypeTicker {
        var t types.MarginTickerMsg
        if err := msg.Decode(&t); err != nil {
            return err
        }
        // t.Price, t.FundingRate, t.LiquidationMarkPrice ...
    }
}
```

## Examples

`cmd/example` (REST) and `cmd/example/websocket_example` log request URLs and raw responses. Demo is the default and is incomplete; set `BASE_URL` to production. `cmd/example/perps_example` does the same for perps, REST and WebSocket, and is read-only.

## Releases

CI runs `gofmt`, `go mod tidy`, `go vet`, `staticcheck`, `govulncheck`, and `go test -race -shuffle=on` on Go 1.24 and stable (Linux, macOS, Windows). Merging to `main` tags `vX.Y.Z` and publishes a GitHub Release when that version is not already tagged.

1. Bump `const Version` in `oddrip/version.go`.
2. Add a `## [X.Y.Z] — YYYY-MM-DD` section at the top of `CHANGELOG.md`.
3. Update the `@vX.Y.Z` pin in this README.

Semver. While at v0, a minor release may break; those changes go first under `### Breaking`. CI runs `gorelease` against the previous tag and refuses a release that has API-incompatible changes without that heading, or that declares one on a patch bump.

## License

MIT — see [LICENSE](LICENSE). `openapi.yaml`, `asyncapi.yaml`, `perps_openapi.yaml`, and `perps_asyncapi.yaml` are Kalshi's published API specifications ([OpenAPI](https://docs.kalshi.com/openapi.yaml), [AsyncAPI](https://docs.kalshi.com/asyncapi.yaml), [perps OpenAPI](https://docs.kalshi.com/perps_openapi.yaml), [perps AsyncAPI](https://docs.kalshi.com/perps_asyncapi.yaml)), vendored unmodified as the contract this client is built against. They are Kalshi's documents and are not covered by this repository's license; Kalshi's own terms apply to them and to use of the API.
