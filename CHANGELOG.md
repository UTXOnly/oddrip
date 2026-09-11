# Changelog

All notable changes to this project are documented here. The client tracks [Kalshi’s API changelog](https://docs.kalshi.com/changelog); repository root `openapi.yaml` / `asyncapi.yaml` are the source of truth for shapes and endpoints.

## [0.6.0] — 2026-09-11

Audit release. Every item under **Fixed** was reproduced with a failing test before the fix; the shipped test suite now exercises retry exhaustion, context cancellation during backoff, multi-channel subscribes, concurrent WebSocket writes, slow consumers, and dead connections.

### Fixed

- **Retry exhaustion panicked the caller.** When every attempt returned 429/5xx with no transport error, `retry.Do` returned a nil response and `client.do` dereferenced it. No `RetryConfig` avoided it (`MaxAttempts: 1` panicked on the first 429). The last response is now surfaced as `*APIError` with its real status, code, and message.
- **Retry backoff ignored the context.** Both waits used `time.Sleep`; a cancelled request could block for the full `Retry-After` or `MaxDelay` (30s by default). Waits now return `ctx.Err()` promptly.
- **Multi-channel `Subscribe` hung until the context deadline.** The read loop discarded the pending reply slot after the first `subscribed` message, so `Subscribe` with two or more channels (the README's own example) never completed. One `SubscribedResponse` per channel is now returned. Reply buffering is sized to the channel count, so subscribing to more than 8 channels at once also works.
- **Concurrent WebSocket commands raced.** `Subscribe`/`Unsubscribe`/`UpdateSubscription`/`Close` wrote to the socket without serialization, tripping gorilla's concurrent-writer check under `-race`. All writes are now serialized; `WSConn` is safe for concurrent use as documented.

### Changed

- **WebSocket messages are never dropped silently.** Previously a consumer that fell 256 messages behind lost messages with no signal. Now the buffer is `WSBufferSize(n)` (default 4096) and on overflow the connection fails with `ErrWSSlowConsumer` and closes — a gap in `orderbook_delta` is unrecoverable without a re-snapshot, so failing loudly is correct. Treat `Messages()` closing as "reconnect and re-subscribe".
- **WebSocket keepalive and dead-connection detection.** Client pings every `WSPingInterval` (default 30s) and enforces a read deadline of `WSReadTimeout` (default 90s), extended on every frame. A half-open socket now surfaces as a timeout error instead of blocking `Messages()` forever. `<= 0` disables either.
- **`Close()`** is idempotent and no longer writes a close frame on an already-dead connection. `Subscribe` and friends return `ErrWSClosed` (or the terminal error) immediately after close/disconnect instead of waiting on the context.
- **`Retry-After`** is honored in HTTP-date form as well as delta-seconds. `RetryConfig.MaxAttempts` below 1 is treated as 1.
- **`DoConcurrent`** now takes a `maxInFlight` argument — `DoConcurrent(ctx, n, maxInFlight, fn)` — and actually bounds concurrency with a semaphore (the README had claimed it did). `maxInFlight <= 0` is unbounded; workers blocked on the semaphore honor `ctx`.
- `http.Client` timeouts in `New()` use typed `time.Duration` constants.

### Added

- **WebSocket:** `WSConn.Err()` (terminal error: `ErrWSClosed`, `ErrWSSlowConsumer`, or the read error) and `WSConn.Done()`; options `WSBufferSize`, `WSPingInterval`, `WSReadTimeout`; error `ErrWSSlowConsumer`.
- **Types — typed WebSocket payloads** for every server message: `TickerMsg`, `OrderbookSnapshotMsg`, `OrderbookDeltaMsg` (levels as `OrderbookLevel{PriceDollars, CountFp}` decoded from the spec's `[price, count]` pairs), `TradeMsg`, `FillMsg`, `MarketPositionMsg`, `UserOrderMsg`, `OrderGroupUpdatesMsg`, `MultivariateMarketLifecycleMsg`, `EventLifecycleMsg`, `EventFeeUpdateMsg`, RFQ/quote messages; `WSType*` constants for each `type` string; `WSMessage.Decode(&v)`.
- **Types — helpers:** `Dollars` (int64, 1e-6 scale — lossless for the 6 decimals responses emit) with `ParseDollars`/`String`/`Float64`/`Cents`; `Count` (int64, 1e-2 scale) with `ParseCount`/`String`/`Float64`; `ParseTime` for the RFC 3339 layouts Kalshi emits.
- **REST — `SeriesService`:** `List`, `Get`, `GetMarketCandlesticks`, `GetEventCandlesticks`, `GetForecastPercentileHistory`.
- **REST — `OrderGroupsService`:** `List`, `Create`, `Get`, `Delete`, `Reset`, `Trigger`, `UpdateLimit`.
- **REST — `SubaccountsService`:** `Create`, `GetBalances`, `Transfer`, `ListTransfers`, `GetNetting`, `UpdateNetting`.
- **REST:** `Markets.GetCandlesticks` (batch), `Portfolio.GetTotalRestingOrderValue`, `Events.ListMultivariateCollections` / `GetMultivariateCollection` / `CreateMarketInMultivariateCollection`. Coverage is 66 of 96 spec paths. Mutating order-group and subaccount calls whose spec response is empty return `error` only.
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
