# Changelog

All notable changes to this project are documented here. The client tracks [Kalshi’s API changelog](https://docs.kalshi.com/changelog); repository root `openapi.yaml` / `asyncapi.yaml` are the source of truth for shapes and endpoints.

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
