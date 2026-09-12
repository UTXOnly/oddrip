# Changelog

All notable changes to this project are documented here. The client tracks [Kalshi’s API changelog](https://docs.kalshi.com/changelog); repository root `openapi.yaml` / `asyncapi.yaml` are the source of truth for shapes and endpoints.

**Versioning.** The module follows semver. While it is at major version 0, a **minor** release may contain breaking changes; when it does, they are listed first under a `### Breaking` heading with migration notes, and CI refuses a release that has API-incompatible changes (per `gorelease`) without that section, or that has one on a patch bump. Patch releases never break. From v1.0.0 on, breaking changes require a major bump.

## [0.6.2] — 2026-09-12

Default hosts move to Kalshi's recommended `external-api` endpoints; WebSocket writes and `Close` are bounded; `DoConcurrent` no longer spawns a goroutine per index; long error bodies decode; the repository has a license. No API-incompatible changes.

### Fixed

- **A WebSocket write could block past the caller's deadline and hang `Close`** ([#15](https://github.com/UTXOnly/oddrip/issues/15)). `Subscribe` / `Unsubscribe` / `ListSubscriptions` / `UpdateSubscription` wrote to the socket with no deadline while holding the write lock, so on a wedged or backpressured socket one command blocked indefinitely, every later command queued behind it, context deadlines had no effect, and `Close` never reached its bounded wait. Each frame is now written with the sooner of the context deadline and `WSWriteTimeout` (default 10s), and waiting for the write slot honors the context. A write that times out fails the connection with an error wrapping the new `ErrWSWriteTimeout` (check with `errors.Is`): `Messages()` closes and `Err()` reports it, the same path as `ErrWSSlowConsumer`; the caller whose own deadline cut the write gets `context.DeadlineExceeded`. A caller that gives up waiting for the slot gets `ctx.Err()` and the connection stays healthy. `Close` now sends the close frame as a bounded control write, skips it when a command write is in flight, and returns within about `WSWriteTimeout` plus five seconds even when the peer is not reading.
- **`DoConcurrent` started one goroutine per index even with a bounded `maxInFlight`** ([#14](https://github.com/UTXOnly/oddrip/issues/14)). The limit gated the active `fn` calls behind a semaphore, but every goroutine was created up front and parked on it, so a large `n` with a small limit still cost `n` goroutine stacks. A positive `maxInFlight` now starts at most `min(n, maxInFlight)` workers that take the next index as their current call returns, and the result buffer is sized to the workers rather than `n`. Results are still index-ordered, and a cancelled context still returns the results collected so far plus `ctx.Err()`, after which workers take no further indices and never block on the result channel. `maxInFlight <= 0` is unchanged (all `n` calls at once), except that a context already cancelled on entry no longer invokes `fn` — matching what bounded mode already did.
- **`APIError` lost `Code` / `Message` / `Details` when a valid error body exceeded 512 bytes** ([#16](https://github.com/UTXOnly/oddrip/issues/16)). Only the 512-byte `RawBody` snippet was read, and the structured fields were decoded from that truncated buffer, so long validation details or gateway metadata reduced the error to `api error STATUS`. The structured fields are now decoded from up to 64 KiB of the body; `RawBody` is still the first 512 bytes, and bodies over 64 KiB are read no further and are not decoded.

### Changed

- **Default hosts are now `external-api.kalshi.com` (REST) and `external-api-ws.kalshi.com` (WebSocket)** ([#18](https://github.com/UTXOnly/oddrip/issues/18)), the production endpoints Kalshi recommends and the primary hosts in both specs. The shared `api.elections.kalshi.com` host that was the default through 0.6.1 remains supported for both protocols; pass `oddrip.BaseURL("https://api.elections.kalshi.com/trade-api/v2")` and `oddrip.WSHost("api.elections.kalshi.com")` to keep using it. Request signing is unaffected — the signed message is `timestamp + METHOD + path` and excludes the host. If you allowlist egress hosts, add the new ones. The WebSocket example maps `external-api.kalshi.com` in `BASE_URL` to the `-ws` host.

### Added

- `WSWriteTimeout(d)` option (default 10s; `<= 0` uses the default) and `ErrWSWriteTimeout` ([#15](https://github.com/UTXOnly/oddrip/issues/15)).
- `LICENSE` — MIT ([#17](https://github.com/UTXOnly/oddrip/issues/17)). The vendored `openapi.yaml` / `asyncapi.yaml` are Kalshi's published specifications and are not covered by it; the README says so.
- **Tests:** `DoConcurrent` goroutine bound at large `n`, cancellation while workers are blocked, `n == 0`; WebSocket write-slot cancellation, cancelled and live callers contending, write timeout as a terminal error, a caller deadline cutting a write, a blocked writer not holding other callers, bounded `Close` with a stuck writer and with a full socket; default REST request destination and WebSocket dial URL; the signed path is identical across the external-api, shared, and demo base URLs; long, malformed, and oversized error bodies through `newAPIError`.

## [0.6.1] — 2026-09-12

Error bodies from production now decode; specs synced to OpenAPI 3.30.0; four missing query filters; malformed WebSocket frames fail the connection. No API-incompatible changes.

### Fixed

- **`APIError.Code` / `Message` were empty for every production error** ([#7](https://github.com/UTXOnly/oddrip/issues/7)). The body was decoded as the spec's flat `ErrorResponse`, but production returns `{"error":{"code":...,"message":...}}` for most errors and `{"msg":"..."}` for parameter-binding 400s, so `Error()` printed only `api error 404`. All three shapes are parsed now; the flat shape still wins when present. `RawBody` is unchanged.
- **A WebSocket text frame that was not valid JSON was skipped silently** ([#10](https://github.com/UTXOnly/oddrip/issues/10)). The connection now fails with an error wrapping the new `ErrWSMalformedFrame` (check with `errors.Is`), `Messages()` closes, and pending commands return that error — the same path as `ErrWSSlowConsumer`. Frames received before the bad one are still delivered.

### Added

- Vendored Kalshi OpenAPI **3.30.0** (AsyncAPI remains **2.0.0**; its content change is CF Benchmarks index-ID documentation) ([#8](https://github.com/UTXOnly/oddrip/issues/8)).
- **Types:** `Series.Categories` — the full discovery-category list; `Series.Category` is now the primary one and the `category` filter on `Series.List` matches any entry in `Categories`. `GetTargetBalanceAllocationResponse.RestingMarginReservation`.
- **Query filters** ([#9](https://github.com/UTXOnly/oddrip/issues/9)): `ExchangeIndex` on `GetOrdersOpts`, `GetFillsOpts`, `GetPositionsOpts`; `Subaccount` on `GetHistoricalPositionsOpts`. All optional; nil omits the parameter as before.
- `ErrWSMalformedFrame`.
- **Tests:** error-body shapes observed from production; `Series.categories` and `resting_margin_reservation` unmarshal; the new filters (set and omitted); malformed-frame failure; `Unsubscribe` one-reply-per-sid and `ListSubscriptions` success paths ([#12](https://github.com/UTXOnly/oddrip/issues/12)).

### Changed

- `APIError` documents which body shapes populate `Code` / `Message`. `RequestID` is still read from `Request-Id`, which production does not currently send.

## [0.6.0] — 2026-09-12

Fixes retry panics and WebSocket hangs, adds Series / OrderGroups / Subaccounts REST, and aligns typed WS payloads with AsyncAPI 2.0.0.

### Breaking

Two source-incompatible API changes (the only ones `gorelease -base=v0.5.0` reports) and three behavioral changes that a consumer must account for.

- **`DoConcurrent` signature.** `DoConcurrent(ctx, n, fn)` is now `DoConcurrent(ctx, n, maxInFlight, fn)`. The old function was documented as bounded but ran all `n` calls at once; the new argument makes it true. Migrate by inserting a limit — `0` reproduces the old unbounded behavior exactly:
  ```go
  // before
  oddrip.DoConcurrent(ctx, len(tickers), fn)
  // after
  oddrip.DoConcurrent(ctx, len(tickers), 8, fn) // or 0 for unbounded
  ```
- **`types.CFBenchmarksAvgData` fields renamed to match the AsyncAPI schema.** The struct shipped in 0.5.0 with tags (`index_id`, `value_usd`, `source_ts_ms`, `window_sec`) that the server never sends, so `Avg60sData` and `Last60sWindowedAverage15Min` always decoded empty. The fields are now `Value`, `WindowSize`, `WindowStartTsMs`, `WindowEndTsExclusive` (`value`, `window_size`, `window_start_ts_ms`, `window_end_ts_exclusive`). Code that read the old fields was reading zero values; switch to the new names.
- **Non-idempotent writes are no longer retried on 5xx or transport errors.** Previously every request was retried on 429, 5xx, and connection errors alike, so a `DecreaseV2` whose connection dropped after the server applied it was replayed and reduced the order twice. Now GET/PUT/DELETE and POSTs the server deduplicates (`CreateV2`/`BatchCreateV2` with `client_order_id` on every order, `Subaccounts.Transfer`, `SetTargetBalanceAllocation`) keep the full policy; `AmendV2`, `DecreaseV2`, creates without a `client_order_id`, `OrderGroups.Create`, `Subaccounts.Create`, and `CreateMarketInMultivariateCollection` are retried on 429 only and surface the 5xx `*APIError` or transport error on the first occurrence. Code that relied on those being retried through transient 5xx must now handle the error (reconcile with `Orders.Get`, then resend). Setting `client_order_id` on creates restores full retries for them.
- **WebSocket slow consumer now closes the connection.** Previously, if the reader of `Messages()` fell more than 256 messages behind, messages were dropped with no signal. Now the buffer is 4096 (`WSBufferSize`) and on overflow the connection fails with `ErrWSSlowConsumer`, `Messages()` closes, and `Err()` reports why. A consumer that tolerated silent gaps must now reconnect (and re-snapshot any local book) when `Messages()` closes. Consumers that already treat a closed `Messages()` as a disconnect need no change.
- **WebSocket read deadline.** Connections now enforce `WSReadTimeout` (default 90s, extended by every frame including keepalive pongs). A half-open socket that previously left `Messages()` open forever now closes it with a timeout error in `Err()`. Live connections are unaffected — the client pings every `WSPingInterval` (30s), so an idle-but-healthy subscription stays up. Pass `WSReadTimeout(0)` / `WSPingInterval(0)` to restore the old behavior.

### Fixed

- **Retry exhaustion panicked the caller.** When every attempt returned 429/5xx with no transport error, `retry.Do` returned a nil response and `client.do` dereferenced it. No `RetryConfig` avoided it (`MaxAttempts: 1` panicked on the first 429). The last response is now surfaced as `*APIError` with its real status, code, and message.
- **Retry backoff ignored the context.** Both waits used `time.Sleep`; a cancelled request could block for the full `Retry-After` or `MaxDelay` (30s by default). Waits now return `ctx.Err()` promptly.
- **Multi-channel `Subscribe` hung until the context deadline.** The read loop discarded the pending reply slot after the first `subscribed` message, so `Subscribe` with two or more channels (the README's own example) never completed. One `SubscribedResponse` per channel is now returned. Reply buffering is sized to the channel count, so subscribing to more than 8 channels at once also works.
- **Concurrent WebSocket commands raced.** `Subscribe`/`Unsubscribe`/`UpdateSubscription`/`Close` wrote to the socket without serialization, tripping gorilla's concurrent-writer check under `-race`. All writes are now serialized; `WSConn` is safe for concurrent use as documented.
- **`UpdateSubscription` with `get_snapshot` never returned.** The spec answers `get_snapshot` with `orderbook_snapshot` frames, which carry no command `id`, but the client waited for an id-matched reply and blocked until the context expired. The call now completes on the first `orderbook_snapshot` for the subscription (or an id-matched `ok`/`error` if the server sends one), returning `Type: "orderbook_snapshot"` with the frame's `SID`/`Seq`; the snapshots themselves arrive on `Messages()` as before. `get_snapshot` now requires `SID` or a single-element `Sids`, matching the command schema.
- **CF Benchmarks averages decoded empty.** See **Breaking** above: `CFBenchmarksAvgData` used field names that are not in the schema, so the typed 60-second and quarter-hour averages were always zero. The unmarshal test now uses the AsyncAPI example payload.
- **An empty path parameter routed the call to a different endpoint.** Paths were built with `path.Join`, which drops empty segments, so `Orders.CancelV2(ctx, "", nil)` sent `DELETE /portfolio/events/orders` — the **CancelAll** endpoint — and `Markets.Get(ctx, "")` quietly called the list endpoint. Every path segment is now required to be non-empty (and not `.`/`..`) and is path-escaped; an offending call returns `ErrEmptyPathParam` before any request is sent.
- **`Subscribe` discarded accepted channels on a partial failure.** The server confirms each channel separately; when it rejected one channel after accepting others, the call returned only the error while the accepted subscriptions stayed live on the server. The accepted `SubscribedResponse`s are now returned alongside the `*WSError`.
- **`MarketLifecycleV2Msg` dropped `exchange_index`.** `created` events carry the shard the market lives on; the field was missing from the struct and silently discarded. Added as `ExchangeIndex *int` (nil on every other event type, so shard 0 is distinguishable from absent).

### Changed

- **WebSocket messages are never dropped for a slow consumer.** Previously a consumer that fell 256 messages behind lost messages with no signal. Now the buffer is `WSBufferSize(n)` (default 4096) and on overflow the connection fails with `ErrWSSlowConsumer` and closes — a gap in `orderbook_delta` is unrecoverable without a re-snapshot, so failing loudly is correct. Treat `Messages()` closing as "reconnect and re-subscribe".
- **WebSocket keepalive and dead-connection detection.** Client pings every `WSPingInterval` (default 30s) and enforces a read deadline of `WSReadTimeout` (default 90s), extended on every frame. A half-open socket now surfaces as a timeout error instead of blocking `Messages()` forever. `<= 0` disables either.
- **`Close()`** is idempotent and no longer writes a close frame on an already-dead connection. `Subscribe` and friends return `ErrWSClosed` (or the terminal error) immediately after close/disconnect instead of waiting on the context.
- **`Retry-After`** is honored in HTTP-date form as well as delta-seconds. `RetryConfig.MaxAttempts` below 1 is treated as 1.
- **`DoConcurrent`** now takes a `maxInFlight` argument — `DoConcurrent(ctx, n, maxInFlight, fn)` — and actually bounds concurrency with a semaphore (the README had claimed it did). `maxInFlight <= 0` is unbounded; workers blocked on the semaphore honor `ctx`.
- **Malformed command replies are errors.** `Subscribe`, `ListSubscriptions`, and `UpdateSubscription` previously ignored a JSON decode failure on the reply's `msg` and returned zero values (`sid: 0`) with a nil error; they now return the decode error.
- Retried responses are drained before being closed so the connection is reused for the next attempt.
- `http.Client` timeouts in `New()` use typed `time.Duration` constants.

### Added

- `ErrEmptyPathParam`, returned by any call whose ticker / ID path parameter is empty.
- **WebSocket:** `WSConn.Err()` (terminal error: `ErrWSClosed`, `ErrWSSlowConsumer`, or the read error) and `WSConn.Done()`; options `WSBufferSize`, `WSPingInterval`, `WSReadTimeout`; error `ErrWSSlowConsumer`.
- **Types — typed WebSocket payloads** for every server message: `TickerMsg`, `OrderbookSnapshotMsg`, `OrderbookDeltaMsg` (levels as `OrderbookLevel{PriceDollars, CountFp}` decoded from the spec's `[price, count]` pairs), `TradeMsg`, `FillMsg`, `MarketPositionMsg`, `UserOrderMsg`, `OrderGroupUpdatesMsg`, `MultivariateMarketLifecycleMsg`, `EventLifecycleMsg`, `EventFeeUpdateMsg`, RFQ/quote messages; `WSType*` constants for each `type` string; `WSMessage.Decode(&v)`.
- **Types — helpers:** `Dollars` (int64, 1e-6 scale — lossless for the 6 decimals responses emit) with `ParseDollars`/`String`/`Float64`/`Cents`; `Count` (int64, 1e-2 scale) with `ParseCount`/`String`/`Float64`; `ParseTime` for the RFC 3339 layouts Kalshi emits.
- **REST — `SeriesService`:** `List`, `Get`, `GetMarketCandlesticks`, `GetEventCandlesticks`, `GetForecastPercentileHistory`.
- **REST — `OrderGroupsService`:** `List`, `Create`, `Get`, `Delete`, `Reset`, `Trigger`, `UpdateLimit`.
- **REST — `SubaccountsService`:** `Create`, `GetBalances`, `Transfer`, `ListTransfers`, `GetNetting`, `UpdateNetting`.
- **REST:** `Markets.GetCandlesticks` (batch), `Portfolio.GetTotalRestingOrderValue`, `Events.ListMultivariateCollections` / `GetMultivariateCollection` / `CreateMarketInMultivariateCollection`. Coverage is 64 of 96 spec paths. Mutating order-group and subaccount calls whose spec response is empty return `error` only.
- **Types:** `Series`, `MarketCandlestick`, `BidAskDistribution`, `PriceDistribution`, `ForecastPercentilesPoint`, `OrderGroup`, `SubaccountBalance`, `SubaccountTransfer`, `SubaccountNettingConfig`, `MultivariateEventCollection`, `AssociatedEvent`, `TickerPair`; constants `FeeType*`, `CollectionStatus*`.
- **Tests:** `internal/retry` and `internal/auth` (signature verified with `rsa.VerifyPSS`, query string excluded from the signed path, PKCS#1/PKCS#8 parsing) had none; both are covered now.

### Removed

- The 8.5 MB compiled `example` binary and the empty `cmd/example/key_id` / `cmd/example/private_key.pem` placeholders are no longer tracked; `/example`, `*.pem`, and `cmd/example/key_id` are gitignored. See `cmd/example/README.md` for where to put credentials.
- Unused `internal/transport` and `internal/errors` packages, and the unused `BearerToken` / duplicate `StaticHeaders` from `internal/auth`. The public `oddrip.APIError` and `oddrip.StaticHeaders` are unchanged.
- `gorilla/websocket` is no longer marked `// indirect` in `go.mod`.

## [0.5.0] — 2026-09-06

### Added

- Vendored Kalshi OpenAPI **3.29.0** (AsyncAPI remains **2.0.0** with new CF Benchmarks content).
- **Live data (new `LiveData` service):** `GetWeatherIndex` and `GetWeatherIndexCalibrations` for the Kalshi-computed city temperature index behind hourly temperature markets (`GET /live_data/weather/{city}`, `.../calibrations`), and `GetEvent` for event-keyed live data (`GET /live_data/events/{event_ticker}`).
- **Orders:** `CancelAll` for `DELETE /portfolio/events/orders` (cancels every resting event-market order across shards).
- **Portfolio:** `ListIntraExchangeTransfers` / `GetIntraExchangeTransfer` for `/portfolio/intra_exchange_instance_transfers`, and `GetTargetBalanceAllocation` / `SetTargetBalanceAllocation` for `/portfolio/target_balance_allocation`.
- **Types:** weather index and calibration payloads, `EventLiveData`, `IntraExchangeInstanceTransfer` (+ status/instance constants), `TargetBalanceAllocation` and `SetTargetBalanceAllocationRequest` (+ `RestingMarginReservation` constants).
- **Types:** `exchange_index` on `Fill`, `MarketPosition`, and `Settlement`; `description` on `ExchangeIndexStatus`; `market_ticker` on `DecreaseOrderV2Request` and batch-cancel entries (required for auto-routing when `exchange_index` is omitted or `-1`); `exchange_index` on `GetBalanceOpts`.
- **WebSocket:** `cfbenchmarks_value` and `cfbenchmarks_value_5hz` channels with `CFBenchmarksValueMsg`, `CFBenchmarksValue5HzMsg`, and `CFBenchmarksIndexListMsg`; `index_ids` on subscribe/update params; `subscribe_indices` / `unsubscribe_indices` / `indexlist` actions; `use_yes_price` on subscribe params.
- Tests for the new REST paths/queries, JSON shapes, and WebSocket payloads above.

### Changed

- **Orders:** `CancelV2` takes `*types.CancelOrderV2Opts` instead of positional `subaccount, exchangeIndex` pointers, so the new `market_ticker` auto-routing parameter can be passed.
- **`oddrip.Version`** is `0.5.0`; publish with git tag **`v0.5.0`**.

### Removed

- **Types:** `EventData.available_on_brokers` (dropped from the OpenAPI `EventData` schema).
- **Types:** `ErrorResponse.Service` is no longer part of the published error schema; the field is retained (now `omitempty`) so older payloads still decode.

## [0.4.0] — 2026-07-26

### Added

- Vendored Kalshi OpenAPI **3.26.0** (AsyncAPI remains **2.0.0** with expanded channel content).
- **Portfolio:** `ListHistoricalPositions` for `GET /historical/positions`.
- **Types:** `ExchangeIndexStatus`; `market_positions_last_updated_ts` on historical cutoff; `balance_breakdown` / `IndexedBalance`; `ApiUsageLevelGrant` on account limits; `is_block_trade` on trades; `exchange_index` on markets/orders; deposit/withdrawal `finalized_ts`; event `settlement_sources` and related fee/index fields; events list `tickers` filter; trades `is_block_trade` query.
- **WebSocket:** `WSChannelPythValue`, underlying subscribe/update actions, `PythValueMsg` / `PythUnderlyingListMsg`, and `MarketLifecycleV2Msg` (including `price_ranges` and metadata strike fields).
- Tests for the new REST paths/queries and JSON shapes above.

### Changed

- **Orders:** Legacy write methods (`Create`, `Cancel`, `Amend`, `Decrease`, `BatchCreate`, `BatchCancel`) removed — OpenAPI 3.26 only publishes V2 `/portfolio/events/orders*` mutations. Use `CreateV2` / `CancelV2` / `AmendV2` / `DecreaseV2` / `BatchCreateV2` / `BatchCancelV2`. GET list/get/queue-position APIs remain.
- **Types:** Market, MarketPosition, EventPosition, Order, Fill, and Trade aligned to fixed-point / dollars fields; removed fields Kalshi dropped (`response_price_units`, `fractional_trading_enabled`, `resting_orders_count`, legacy integer prices/counts, etc.).
- **Types:** `GetMarketOrderbookResponse` is `orderbook_fp` only.
- **WebSocket:** `UpdateSubscription` accepts `get_snapshot` and pyth underlying actions.
- **`oddrip.Version`** is `0.4.0`; publish with git tag **`v0.4.0`**.

### Removed

- **Exchange:** `GetAnnouncements` and announcement types (`GET /exchange/announcements` removed from Predictions REST).

## [0.3.0] — 2026-05-23

### Added

- Vendored Kalshi OpenAPI **3.19.0** and AsyncAPI **2.0.0** (published at docs.kalshi.com).
- **Orders (V2):** `CreateV2`, `CancelV2`, `AmendV2`, `DecreaseV2`, `BatchCreateV2`, and `BatchCancelV2` for `/portfolio/events/orders*`.
- **Markets:** `GetOrderbooks` for `GET /markets/orderbooks`.
- **Portfolio:** `ListDeposits` and `ListWithdrawals` for deposit/withdrawal history.
- **Account:** `GetEndpointCosts` for `GET /account/endpoint_costs`.
- **Types:** V2 order request/response structs; `BucketLimit` rate-limit buckets; `balance_dollars`; `outcome_side` / `book_side` on orders and fills; `taker_outcome_side` / `taker_book_side` on trades; `occurrence_datetime` on markets; deposit/withdrawal types; `WSChannelMultivariateLifecycle` and `WSUpdateSubscriptionGetSnapshot`.
- **`oddrip.Version`** constant (`0.3.0`) for the client module; use git tag **`v0.3.0`** when publishing.

### Changed

- **Types:** `GetAccountApiLimitsResponse` now uses nested `read` / `write` `BucketLimit` objects (replacing flat `read_limit` / `write_limit`) per Apr 2026 rate-limit API.
- **Types:** `GetHistoricalMarketsOpts` adds `series_ticker` filter.

## [0.2.0] — 2026-03-21

### Added

- **Portfolio:** `ListSettlements`, `ListHistoricalFills`, and `ListHistoricalOrders` for `GET /portfolio/settlements`, `GET /historical/fills`, and `GET /historical/orders`.
- **Markets:** `ListHistorical`, `GetHistorical`, `GetHistoricalTrades`, and `GetHistoricalCandlesticks` for historical market data and archived trades.
- **Types:** `Settlement`, `GetSettlementsOpts`, historical candlestick payloads (`GetMarketCandlesticksHistoricalResponse`, nested distributions), and additional `Market` / `Fill` / `Order` fields aligned with OpenAPI 3.10 (e.g. `yes_price_dollars` / `no_price_dollars` on fills, `taker_fees_dollars` / `maker_fees_dollars` on orders, settlement and lifecycle fields on markets).
- **`oddrip.Version`** constant (`0.2.0`) for the client module; use git tag **`v0.2.0`** when publishing.

### Changed

- **Types:** `Trade` and queue-position types adjusted for current spec (e.g. `Trade.created_time` and dollar price fields; queue position responses emphasize `queue_position_fp`).
- **Types:** `MarketPosition.last_updated_ts` is always unmarshaled when present (required in the published contract).
