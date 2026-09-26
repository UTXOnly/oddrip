package oddrip

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

const perpsMarketJSON = `{
	"ticker":"KXBTCPERP","title":"Bitcoin perpetual","exchange_index":1,
	"contract_size":"0.001000","underlying_multiplier":"1","tick_size":"0.10",
	"status":"active","fractional_trading_enabled":true,
	"leverage_estimate":10.5,"leverage_estimates":{"1000":10.5,"1000000":4.2},
	"price":"65000.10","bid":"65000.00","ask":"65000.20",
	"settlement_mark_price":{"price":"65000.05","ts_ms":1700000000123},
	"schedule":null
}`

func TestPerpsExchange_GetStatus(t *testing.T) {
	client, ct := newCaptureClient(200, `{"exchange_active":true,"trading_active":false}`)
	got, err := client.Perps.Exchange.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/exchange/status")
	ct.assertQuery(t, map[string]string{})
	if !got.ExchangeActive || got.TradingActive {
		t.Fatalf("status: %+v", got)
	}
}

func TestPerpsExchange_GetEnabled(t *testing.T) {
	client, ct := newCaptureClient(200, `{"enabled":true}`)
	got, err := client.Perps.Exchange.GetEnabled(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/enabled")
	ct.assertQuery(t, map[string]string{})
	if !got.Enabled {
		t.Fatalf("enabled: %+v", got)
	}
}

func TestPerpsAccount_GetAPILimits(t *testing.T) {
	client, ct := newCaptureClient(200, `{
		"usage_tier":"basic",
		"read":{"refill_rate":20,"bucket_capacity":20},
		"write":{"refill_rate":10,"bucket_capacity":10},
		"grants":[{"exchange_instance":"margined","level":"premier","expires_ts":null,"source":"manual"}]
	}`)
	got, err := client.Perps.Account.GetAPILimits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/account/limits/perps")
	ct.assertQuery(t, map[string]string{})
	if got.UsageTier != "basic" || got.Read.RefillRate != 20 || got.Write.BucketCapacity != 10 ||
		len(got.Grants) != 1 || got.Grants[0].ExchangeInstance != types.ExchangeInstanceMargined || got.Grants[0].ExpiresTs != nil {
		t.Fatalf("limits: %+v", got)
	}
}

func TestPerpsMarkets_List(t *testing.T) {
	client, ct := newCaptureClient(200, `{"markets":[`+perpsMarketJSON+`]}`)
	got, err := client.Perps.Markets.List(context.Background(), &types.GetMarginMarketsOpts{Status: types.MarginMarketStatusActive})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/markets")
	ct.assertQuery(t, map[string]string{"status": "active"})
	if len(got.Markets) != 1 {
		t.Fatalf("markets: %+v", got.Markets)
	}
	m := got.Markets[0]
	if m.Ticker != "KXBTCPERP" || m.ExchangeIndex != 1 || m.TickSize != "0.10" || !m.FractionalTradingEnabled ||
		m.LeverageEstimate == nil || *m.LeverageEstimate != 10.5 || m.LeverageEstimates["1000000"] != 4.2 ||
		m.SettlementMarkPrice == nil || m.SettlementMarkPrice.TsMs != 1700000000123 || m.Schedule != nil {
		t.Fatalf("market: %+v", m)
	}

	if _, err := client.Perps.Markets.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/markets")
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsMarkets_Get(t *testing.T) {
	client, ct := newCaptureClient(200, `{"market":`+perpsMarketJSON+`}`)
	got, err := client.Perps.Markets.Get(context.Background(), "KXBTCPERP")
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/markets/KXBTCPERP")
	ct.assertQuery(t, map[string]string{})
	if got.Market.Ticker != "KXBTCPERP" || got.Market.Price != "65000.10" || got.Market.Status != types.MarginMarketStatusActive {
		t.Fatalf("market: %+v", got.Market)
	}
}

func TestPerpsMarkets_GetOrderbook(t *testing.T) {
	client, ct := newCaptureClient(200, `{"orderbook":{"bids":[["65000.00","1.50"],["64999.90","3.00"]],"asks":[["65000.20","0.75"]]}}`)
	got, err := client.Perps.Markets.GetOrderbook(context.Background(), "KXBTCPERP", &types.GetMarginMarketOrderbookOpts{
		Depth: ptrOf(0), AggregationTickSize: "0.10",
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/markets/KXBTCPERP/orderbook")
	ct.assertQuery(t, map[string]string{"depth": "0", "aggregation_tick_size": "0.10"})
	ob := got.Orderbook
	if len(ob.Bids) != 2 || ob.Bids[0].PriceDollars != "65000.00" || ob.Bids[1].CountFp != "3.00" ||
		len(ob.Asks) != 1 || ob.Asks[0].PriceDollars != "65000.20" || ob.Asks[0].CountFp != "0.75" {
		t.Fatalf("orderbook: %+v", ob)
	}

	if _, err := client.Perps.Markets.GetOrderbook(context.Background(), "KXBTCPERP", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/markets/KXBTCPERP/orderbook")
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsMarkets_GetCandlesticks(t *testing.T) {
	client, ct := newCaptureClient(200, `{"ticker":"KXBTCPERP","candlesticks":[{
		"end_period_ts":1700003600,
		"bid":{"open":"64990.00","low":"64900.00","high":"65100.00","close":"65000.00"},
		"ask":{"open":"64990.20","low":"64900.20","high":"65100.20","close":"65000.20"},
		"price":{"open":null,"low":null,"high":null,"close":null,"mean":null,"previous":"64995.00"},
		"volume":"12.50","volume_notional_value_dollars":"812500.00",
		"open_interest":"40.00","open_interest_notional_value_dollars":"2600000.00"
	}]}`)
	got, err := client.Perps.Markets.GetCandlesticks(context.Background(), "KXBTCPERP", &types.GetMarginMarketCandlesticksOpts{
		StartTs: 1700000000, EndTs: 1700003600, PeriodInterval: types.PeriodInterval1Hour, IncludeLatestBeforeStart: ptrOf(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/markets/KXBTCPERP/candlesticks")
	ct.assertQuery(t, map[string]string{
		"start_ts": "1700000000", "end_ts": "1700003600", "period_interval": "60", "include_latest_before_start": "true",
	})
	if got.Ticker != "KXBTCPERP" || len(got.Candlesticks) != 1 {
		t.Fatalf("resp: %+v", got)
	}
	c := got.Candlesticks[0]
	if c.EndPeriodTs != 1700003600 || c.Bid.Close != "65000.00" || c.Ask.High != "65100.20" ||
		c.Price.Open != nil || c.Price.Previous == nil || *c.Price.Previous != "64995.00" ||
		c.Volume != "12.50" || c.OpenInterestNotionalValueDollars != "2600000.00" {
		t.Fatalf("candlestick: %+v", c)
	}

	if _, err := client.Perps.Markets.GetCandlesticks(context.Background(), "KXBTCPERP", &types.GetMarginMarketCandlesticksOpts{
		StartTs: 1, EndTs: 2, PeriodInterval: types.PeriodInterval1Day,
	}); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{"start_ts": "1", "end_ts": "2", "period_interval": "1440"})
}

func TestPerpsMarkets_GetCandlesticks_Validation(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	if _, err := client.Perps.Markets.GetCandlesticks(ctx, "KXBTCPERP", nil); err == nil {
		t.Fatal("expected error for nil opts")
	}
	for _, pi := range []int{0, 5, 1441} {
		if _, err := client.Perps.Markets.GetCandlesticks(ctx, "KXBTCPERP", &types.GetMarginMarketCandlesticksOpts{StartTs: 1, EndTs: 2, PeriodInterval: pi}); err == nil {
			t.Fatalf("expected error for period_interval %d", pi)
		}
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestPerpsMarkets_GetTrades(t *testing.T) {
	client, ct := newCaptureClient(200, `{"trades":[{
		"trade_id":"tr1","ticker":"KXBTCPERP","count":"2.50","price":"65000.10",
		"created_time":"2026-01-02T03:04:05Z","taker_side":"ask"
	}],"cursor":"next"}`)
	got, err := client.Perps.Markets.GetTrades(context.Background(), &types.GetMarginTradesOpts{
		Ticker: "KXBTCPERP", Limit: ptrOf[int64](500), Cursor: "c1", MinTs: ptrOf[int64](1700000000), MaxTs: ptrOf[int64](1700003600),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/trades")
	ct.assertQuery(t, map[string]string{
		"ticker": "KXBTCPERP", "limit": "500", "cursor": "c1", "min_ts": "1700000000", "max_ts": "1700003600",
	})
	if got.Cursor != "next" || len(got.Trades) != 1 {
		t.Fatalf("resp: %+v", got)
	}
	tr := got.Trades[0]
	if tr.TradeID != "tr1" || tr.Count != "2.50" || tr.Price != "65000.10" || tr.TakerSide != types.BookSideAsk ||
		tr.CreatedTime != "2026-01-02T03:04:05Z" {
		t.Fatalf("trade: %+v", tr)
	}

	if _, err := client.Perps.Markets.GetTrades(context.Background(), &types.GetMarginTradesOpts{Ticker: "TEST-PERP"}); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{"ticker": "TEST-PERP"})
}

func TestPerpsMarkets_GetTrades_Validation(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	if _, err := client.Perps.Markets.GetTrades(ctx, nil); err == nil {
		t.Fatal("expected error for nil opts")
	}
	if _, err := client.Perps.Markets.GetTrades(ctx, &types.GetMarginTradesOpts{Limit: ptrOf[int64](10)}); err == nil {
		t.Fatal("expected error for empty ticker")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestPerpsMarkets_EmptyPathParam(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	candleOpts := &types.GetMarginMarketCandlesticksOpts{StartTs: 1, EndTs: 2, PeriodInterval: 60}
	for _, ticker := range []string{"", ".."} {
		calls := map[string]func() error{
			"Get":             func() error { _, err := client.Perps.Markets.Get(ctx, ticker); return err },
			"GetOrderbook":    func() error { _, err := client.Perps.Markets.GetOrderbook(ctx, ticker, nil); return err },
			"GetCandlesticks": func() error { _, err := client.Perps.Markets.GetCandlesticks(ctx, ticker, candleOpts); return err },
		}
		for name, call := range calls {
			ct.req = nil
			if err := call(); !errors.Is(err, ErrEmptyPathParam) {
				t.Errorf("%s(%q): err = %v, want ErrEmptyPathParam", name, ticker, err)
			}
			if ct.req != nil {
				t.Errorf("%s(%q): request was sent: %s %s", name, ticker, ct.req.Method, ct.req.URL.Path)
			}
		}
	}
}
