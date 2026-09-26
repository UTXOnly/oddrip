package types

import (
	"encoding/json"
	"testing"
)

func TestPerpsMarketTypes_MarginExchangeStatusUnmarshal(t *testing.T) {
	var out MarginExchangeStatus
	if err := json.Unmarshal([]byte(`{"exchange_active":true,"trading_active":true}`), &out); err != nil {
		t.Fatal(err)
	}
	if !out.ExchangeActive || !out.TradingActive {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestPerpsMarketTypes_MarginEnabledResponseUnmarshal(t *testing.T) {
	var out MarginEnabledResponse
	if err := json.Unmarshal([]byte(`{"enabled":false}`), &out); err != nil {
		t.Fatal(err)
	}
	if out.Enabled {
		t.Fatalf("unexpected: %+v", out)
	}
}

// The perps limits endpoint reuses GetAccountApiLimitsResponse; the perps
// schema is identical to the event-contract one.
func TestPerpsMarketTypes_PerpsAccountApiLimitsUnmarshal(t *testing.T) {
	const payload = `{
		"usage_tier": "premier",
		"read": {"refill_rate": 100, "bucket_capacity": 100},
		"write": {"refill_rate": 50, "bucket_capacity": 100},
		"grants": [
			{"exchange_instance": "margined", "level": "premier", "expires_ts": 1800000000, "source": "volume"},
			{"exchange_instance": "event_contract", "level": "prime", "expires_ts": null, "source": "manual"}
		]
	}`
	var out GetAccountApiLimitsResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.UsageTier != "premier" || out.Read.RefillRate != 100 || out.Write.RefillRate != 50 || out.Write.BucketCapacity != 100 {
		t.Fatalf("buckets: %+v", out)
	}
	if len(out.Grants) != 2 {
		t.Fatalf("grants: %+v", out.Grants)
	}
	g0, g1 := out.Grants[0], out.Grants[1]
	if g0.ExchangeInstance != ExchangeInstanceMargined || g0.Level != "premier" || g0.ExpiresTs == nil || *g0.ExpiresTs != 1800000000 || g0.Source != "volume" {
		t.Fatalf("grant 0: %+v", g0)
	}
	if g1.ExchangeInstance != ExchangeInstanceEventContract || g1.ExpiresTs != nil || g1.Source != "manual" {
		t.Fatalf("grant 1: %+v", g1)
	}
}

func TestPerpsMarketTypes_MarginMarketUnmarshal(t *testing.T) {
	// Every MarginMarket property; product_metadata is the spec's example.
	const payload = `{
		"ticker": "KXBTCPERP",
		"title": "Bitcoin perpetual",
		"exchange_index": 1,
		"contract_size": "0.001000",
		"underlying_multiplier": "1",
		"tick_size": "0.10",
		"status": "active",
		"fractional_trading_enabled": true,
		"leverage_estimate": 20,
		"leverage_estimates": {"1000": 20, "10000": 18.5, "100000": 12, "1000000": 5.25},
		"long_leverage_estimates": {"1000": 19.5},
		"short_leverage_estimates": {"1000": 20.5},
		"price": "65000.10",
		"volume": "1250.00",
		"volume_notional_value_dollars": "81250125.00",
		"open_interest": "400.00",
		"open_interest_notional_value_dollars": "26000040.00",
		"volume_24h": "90.00",
		"volume_24h_notional_value_dollars": "5850009.00",
		"bid": "65000.00",
		"ask": "65000.20",
		"settlement_mark_price": {"price": "65000.05", "ts_ms": 1700000000100},
		"liquidation_mark_price": {"price": "65000.07", "ts_ms": 1700000000200},
		"reference_price": {"price": "65000.01", "ts_ms": 1700000000300},
		"asset_class": "Crypto",
		"product_metadata": {"important_info": {"markdown": "**Important information:** Review this market's trading schedule."}},
		"schedule": {"is_open": true, "next_close_ts": 1700086400, "next_open_ts": null}
	}`
	var m MarginMarket
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatal(err)
	}
	if m.Ticker != "KXBTCPERP" || m.Title != "Bitcoin perpetual" || m.ExchangeIndex != 1 || m.ContractSize != "0.001000" ||
		m.UnderlyingMultiplier != "1" || m.TickSize != "0.10" || m.Status != MarginMarketStatusActive || !m.FractionalTradingEnabled {
		t.Fatalf("required: %+v", m)
	}
	if m.LeverageEstimate == nil || *m.LeverageEstimate != 20 || len(m.LeverageEstimates) != 4 || m.LeverageEstimates["10000"] != 18.5 ||
		m.LongLeverageEstimates["1000"] != 19.5 || m.ShortLeverageEstimates["1000"] != 20.5 {
		t.Fatalf("leverage: %+v", m)
	}
	if m.Price != "65000.10" || m.Volume != "1250.00" || m.VolumeNotionalValueDollars != "81250125.00" ||
		m.OpenInterest != "400.00" || m.OpenInterestNotionalValueDollars != "26000040.00" || m.Volume24h != "90.00" ||
		m.Volume24hNotionalValueDollars != "5850009.00" || m.Bid != "65000.00" || m.Ask != "65000.20" || m.AssetClass != "Crypto" {
		t.Fatalf("stats: %+v", m)
	}
	if m.SettlementMarkPrice == nil || m.SettlementMarkPrice.Price != "65000.05" || m.SettlementMarkPrice.TsMs != 1700000000100 ||
		m.LiquidationMarkPrice == nil || m.LiquidationMarkPrice.Price != "65000.07" ||
		m.ReferencePrice == nil || m.ReferencePrice.TsMs != 1700000000300 {
		t.Fatalf("prices: %+v", m)
	}
	info, ok := m.ProductMetadata["important_info"].(map[string]any)
	if !ok || info["markdown"] != "**Important information:** Review this market's trading schedule." {
		t.Fatalf("product_metadata: %+v", m.ProductMetadata)
	}
	if m.Schedule == nil || !m.Schedule.IsOpen || m.Schedule.NextCloseTs == nil || *m.Schedule.NextCloseTs != 1700086400 || m.Schedule.NextOpenTs != nil {
		t.Fatalf("schedule: %+v", m.Schedule)
	}
}

func TestPerpsMarketTypes_MarginMarketUnmarshal_RequiredOnly(t *testing.T) {
	// A 24/7 market with only the required fields: schedule is null and the
	// optional stats, prices, and leverage estimates are absent.
	const payload = `{
		"ticker": "TEST-PERP",
		"title": "Test perpetual",
		"exchange_index": 0,
		"contract_size": "1.000000",
		"underlying_multiplier": "1",
		"tick_size": "0.01",
		"status": "inactive",
		"fractional_trading_enabled": false,
		"schedule": null
	}`
	var resp MarginMarketResponse
	if err := json.Unmarshal([]byte(`{"market":`+payload+`}`), &resp); err != nil {
		t.Fatal(err)
	}
	m := resp.Market
	if m.Ticker != "TEST-PERP" || m.Status != MarginMarketStatusInactive || m.FractionalTradingEnabled {
		t.Fatalf("required: %+v", m)
	}
	if m.Schedule != nil || m.LeverageEstimate != nil || m.LeverageEstimates != nil || m.SettlementMarkPrice != nil ||
		m.LiquidationMarkPrice != nil || m.ReferencePrice != nil || m.ProductMetadata != nil || m.Price != "" || m.Bid != "" {
		t.Fatalf("optional fields should be empty: %+v", m)
	}
}

func TestPerpsMarketTypes_MarginMarketScheduleClosed(t *testing.T) {
	var s MarginMarketSchedule
	if err := json.Unmarshal([]byte(`{"is_open":false,"next_close_ts":null,"next_open_ts":1700090000}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.IsOpen || s.NextCloseTs != nil || s.NextOpenTs == nil || *s.NextOpenTs != 1700090000 {
		t.Fatalf("unexpected: %+v", s)
	}
}

func TestPerpsMarketTypes_GetMarginMarketsResponseUnmarshal(t *testing.T) {
	const payload = `{"markets":[
		{"ticker":"KXBTCPERP","title":"A","exchange_index":1,"contract_size":"0.001000","underlying_multiplier":"1","tick_size":"0.10","status":"active","fractional_trading_enabled":true,"schedule":null},
		{"ticker":"TEST-PERP","title":"B","exchange_index":1,"contract_size":"1.000000","underlying_multiplier":"1","tick_size":"0.01","status":"closed","fractional_trading_enabled":false,"schedule":{"is_open":false,"next_close_ts":null,"next_open_ts":null}}
	]}`
	var out GetMarginMarketsResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Markets) != 2 || out.Markets[0].Schedule != nil || out.Markets[1].Status != MarginMarketStatusClosed ||
		out.Markets[1].Schedule == nil || out.Markets[1].Schedule.IsOpen {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestPerpsMarketTypes_MarginOrderbookResponseUnmarshal(t *testing.T) {
	// Levels use the spec's PriceLevelDollarsCountFp example ["0.1500","100.00"].
	const payload = `{"orderbook":{"bids":[["0.1500","100.00"],["0.1400","5.50"]],"asks":[]}}`
	var out MarginOrderbookResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	ob := out.Orderbook
	if len(ob.Bids) != 2 || ob.Bids[0].PriceDollars != "0.1500" || ob.Bids[0].CountFp != "100.00" ||
		ob.Bids[1].CountFp != "5.50" || ob.Asks == nil || len(ob.Asks) != 0 {
		t.Fatalf("unexpected: %+v", ob)
	}
}

func TestPerpsMarketTypes_GetMarginMarketCandlesticksResponseUnmarshal(t *testing.T) {
	const payload = `{"ticker":"KXBTCPERP","candlesticks":[
		{
			"end_period_ts": 1700000060,
			"bid": {"open":"64990.00","low":"64980.00","high":"65010.00","close":"65000.00"},
			"ask": {"open":"64990.20","low":"64980.20","high":"65010.20","close":"65000.20"},
			"price": {"open":"64995.00","low":"64985.00","high":"65005.00","close":"65000.10","mean":"64999.50","previous":"64994.90"},
			"volume": "3.25",
			"volume_notional_value_dollars": "211248.38",
			"open_interest": "400.00",
			"open_interest_notional_value_dollars": "26000040.00"
		},
		{
			"end_period_ts": 1700000120,
			"bid": {"open":"65000.00","low":"65000.00","high":"65000.00","close":"65000.00"},
			"ask": {"open":"65000.20","low":"65000.20","high":"65000.20","close":"65000.20"},
			"price": {"open":null,"low":null,"high":null,"close":null,"mean":null,"previous":"65000.10"},
			"volume": "0.00",
			"volume_notional_value_dollars": "0.00",
			"open_interest": "400.00",
			"open_interest_notional_value_dollars": "26000040.00"
		}
	]}`
	var out GetMarginMarketCandlesticksResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.Ticker != "KXBTCPERP" || len(out.Candlesticks) != 2 {
		t.Fatalf("unexpected: %+v", out)
	}
	c0, c1 := out.Candlesticks[0], out.Candlesticks[1]
	if c0.EndPeriodTs != 1700000060 || c0.Bid.Open != "64990.00" || c0.Ask.Low != "64980.20" ||
		c0.Price.Mean == nil || *c0.Price.Mean != "64999.50" || c0.Price.Close == nil || *c0.Price.Close != "65000.10" ||
		c0.Volume != "3.25" || c0.VolumeNotionalValueDollars != "211248.38" || c0.OpenInterest != "400.00" ||
		c0.OpenInterestNotionalValueDollars != "26000040.00" {
		t.Fatalf("candle 0: %+v", c0)
	}
	if c1.Price.Open != nil || c1.Price.Low != nil || c1.Price.High != nil || c1.Price.Close != nil || c1.Price.Mean != nil ||
		c1.Price.Previous == nil || *c1.Price.Previous != "65000.10" || c1.Volume != "0.00" {
		t.Fatalf("candle 1: %+v", c1)
	}
}

func TestPerpsMarketTypes_GetMarginTradesResponseUnmarshal(t *testing.T) {
	const payload = `{"trades":[
		{"trade_id":"tr1","ticker":"KXBTCPERP","count":"2.50","price":"65000.10","created_time":"2026-01-02T03:04:05Z","taker_side":"bid"}
	],"cursor":""}`
	var out GetMarginTradesResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Trades) != 1 || out.Cursor != "" {
		t.Fatalf("unexpected: %+v", out)
	}
	tr := out.Trades[0]
	if tr.TradeID != "tr1" || tr.Ticker != "KXBTCPERP" || tr.Count != "2.50" || tr.Price != "65000.10" ||
		tr.TakerSide != BookSideBid {
		t.Fatalf("trade: %+v", tr)
	}
	if ts, err := ParseTime(tr.CreatedTime); err != nil || ts.Unix() != 1767323045 {
		t.Fatalf("created_time: %v %v", ts, err)
	}
}
