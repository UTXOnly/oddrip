package oddrip

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type captureTransport struct {
	status int
	body   string
	req    *http.Request
	sent   []byte
}

func (c *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.req = req
	if req.Body != nil {
		c.sent, _ = io.ReadAll(req.Body)
	}
	resp := &http.Response{
		StatusCode: c.status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(c.body)),
		Request:    req,
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

func newCaptureClient(status int, body string) (*Client, *captureTransport) {
	ct := &captureTransport{status: status, body: body}
	return New(HTTPClient(&http.Client{Transport: ct})), ct
}

func (c *captureTransport) assertRequest(t *testing.T, method, path string) {
	t.Helper()
	if c.req == nil {
		t.Fatal("no request sent")
	}
	if c.req.Method != method || c.req.URL.Path != "/trade-api/v2"+path {
		t.Fatalf("request: got %s %s, want %s %s", c.req.Method, c.req.URL.Path, method, "/trade-api/v2"+path)
	}
}

func (c *captureTransport) assertQuery(t *testing.T, want map[string]string) {
	t.Helper()
	q := c.req.URL.Query()
	for k, v := range want {
		if q.Get(k) != v {
			t.Fatalf("query %s: got %q, want %q (all: %v)", k, q.Get(k), v, q)
		}
	}
	if len(q) != len(want) {
		t.Fatalf("query has extra keys: %v", q)
	}
}

func (c *captureTransport) assertBody(t *testing.T, want string) {
	t.Helper()
	var got, exp interface{}
	if err := json.Unmarshal(c.sent, &got); err != nil {
		t.Fatalf("sent body %q: %v", c.sent, err)
	}
	if err := json.Unmarshal([]byte(want), &exp); err != nil {
		t.Fatal(err)
	}
	g, _ := json.Marshal(got)
	e, _ := json.Marshal(exp)
	if string(g) != string(e) {
		t.Fatalf("body: got %s, want %s", g, e)
	}
}

func ptrOf[T any](v T) *T { return &v }

const seriesJSON = `{
	"ticker":"KXHIGHNY","frequency":"daily","title":"Highest temperature in NYC","category":"Climate and Weather",
	"tags":["Weather"],"settlement_sources":[{"name":"NWS","url":"https://weather.gov"}],
	"contract_url":"https://kalshi.com/c","contract_terms_url":"https://kalshi.com/t",
	"product_metadata":{"k":"v"},"fee_type":"quadratic","fee_multiplier":1.5,
	"additional_prohibitions":["none"],"volume_fp":"10.00","last_updated_ts":"2024-01-01T00:00:00Z","exchange_index":1
}`

func TestSeries_List(t *testing.T) {
	client, ct := newCaptureClient(200, `{"series":[`+seriesJSON+`]}`)
	got, err := client.Series.List(context.Background(), &types.GetSeriesListOpts{
		Category:               "Climate and Weather",
		Tags:                   "Weather",
		IncludeProductMetadata: ptrOf(true),
		IncludeVolume:          ptrOf(false),
		MinUpdatedTs:           ptrOf[int64](1700000000),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/series")
	ct.assertQuery(t, map[string]string{
		"category": "Climate and Weather", "tags": "Weather",
		"include_product_metadata": "true", "include_volume": "false", "min_updated_ts": "1700000000",
	})
	if len(got.Series) != 1 {
		t.Fatalf("series: %+v", got.Series)
	}
	s := got.Series[0]
	if s.Ticker != "KXHIGHNY" || s.FeeType != types.FeeTypeQuadratic || s.FeeMultiplier != 1.5 ||
		s.VolumeFp != "10.00" || s.ExchangeIndex != 1 || s.ProductMetadata["k"] != "v" ||
		len(s.SettlementSources) != 1 || s.SettlementSources[0].Name != "NWS" {
		t.Fatalf("series: %+v", s)
	}
}

func TestSeries_Get(t *testing.T) {
	client, ct := newCaptureClient(200, `{"series":`+seriesJSON+`}`)
	got, err := client.Series.Get(context.Background(), "KXHIGHNY", &types.GetSeriesOpts{IncludeVolume: ptrOf(true)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/series/KXHIGHNY")
	ct.assertQuery(t, map[string]string{"include_volume": "true"})
	if got.Series.Title != "Highest temperature in NYC" || got.Series.Tags[0] != "Weather" {
		t.Fatalf("series: %+v", got.Series)
	}
}

const candlestickJSON = `{
	"end_period_ts":1700003600,
	"yes_bid":{"open_dollars":"0.5500","low_dollars":"0.5400","high_dollars":"0.5700","close_dollars":"0.5600"},
	"yes_ask":{"open_dollars":"0.5700","low_dollars":"0.5600","high_dollars":"0.5900","close_dollars":"0.5800"},
	"price":{"open_dollars":null,"low_dollars":null,"high_dollars":null,"close_dollars":null,"previous_dollars":"0.5600","mean_dollars":null},
	"volume_fp":"10.00","open_interest_fp":"100.00"
}`

func TestSeries_GetMarketCandlesticks(t *testing.T) {
	client, ct := newCaptureClient(200, `{"ticker":"KXHIGHNY-24JAN01-T60","candlesticks":[`+candlestickJSON+`]}`)
	got, err := client.Series.GetMarketCandlesticks(context.Background(), "KXHIGHNY", "KXHIGHNY-24JAN01-T60", &types.GetMarketCandlesticksOpts{
		StartTs: 1700000000, EndTs: 1700003600, PeriodInterval: types.PeriodInterval1Hour, IncludeLatestBeforeStart: ptrOf(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/series/KXHIGHNY/markets/KXHIGHNY-24JAN01-T60/candlesticks")
	ct.assertQuery(t, map[string]string{
		"start_ts": "1700000000", "end_ts": "1700003600", "period_interval": "60", "include_latest_before_start": "true",
	})
	if got.Ticker != "KXHIGHNY-24JAN01-T60" || len(got.Candlesticks) != 1 {
		t.Fatalf("resp: %+v", got)
	}
	c := got.Candlesticks[0]
	if c.EndPeriodTs != 1700003600 || c.YesBid.CloseDollars != "0.5600" || c.YesAsk.HighDollars != "0.5900" ||
		c.Price.OpenDollars != nil || c.Price.PreviousDollars == nil || *c.Price.PreviousDollars != "0.5600" ||
		c.VolumeFp != "10.00" || c.OpenInterestFp != "100.00" {
		t.Fatalf("candlestick: %+v", c)
	}
}

func TestSeries_GetMarketCandlesticks_Validation(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	if _, err := client.Series.GetMarketCandlesticks(ctx, "S", "M", nil); err == nil {
		t.Fatal("expected error for nil opts")
	}
	if _, err := client.Series.GetMarketCandlesticks(ctx, "S", "M", &types.GetMarketCandlesticksOpts{PeriodInterval: 5}); err == nil {
		t.Fatal("expected error for bad period_interval")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestSeries_GetEventCandlesticks(t *testing.T) {
	client, ct := newCaptureClient(200, `{"market_tickers":["A","B"],"market_candlesticks":[[`+candlestickJSON+`],[]],"adjusted_end_ts":1700003600}`)
	got, err := client.Series.GetEventCandlesticks(context.Background(), "KXHIGHNY", "KXHIGHNY-24JAN01", &types.GetEventCandlesticksOpts{
		StartTs: 1, EndTs: 2, PeriodInterval: types.PeriodInterval1Min,
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/series/KXHIGHNY/events/KXHIGHNY-24JAN01/candlesticks")
	ct.assertQuery(t, map[string]string{"start_ts": "1", "end_ts": "2", "period_interval": "1"})
	if len(got.MarketTickers) != 2 || len(got.MarketCandlesticks) != 2 || len(got.MarketCandlesticks[0]) != 1 ||
		got.MarketCandlesticks[0][0].YesBid.OpenDollars != "0.5500" || got.AdjustedEndTs != 1700003600 {
		t.Fatalf("resp: %+v", got)
	}
	if _, err := client.Series.GetEventCandlesticks(context.Background(), "S", "E", &types.GetEventCandlesticksOpts{PeriodInterval: 0}); err == nil {
		t.Fatal("expected error for bad period_interval")
	}
}

func TestSeries_GetForecastPercentileHistory(t *testing.T) {
	client, ct := newCaptureClient(200, `{"forecast_history":[{
		"event_ticker":"KXHIGHNY-24JAN01","end_period_ts":1700003600,"period_interval":60,
		"percentile_points":[{"percentile":5000,"raw_numerical_forecast":61.2,"numerical_forecast":61,"formatted_forecast":"61°F"}]
	}]}`)
	got, err := client.Series.GetForecastPercentileHistory(context.Background(), "KXHIGHNY", "KXHIGHNY-24JAN01", &types.GetEventForecastPercentilesHistoryOpts{
		Percentiles: []int{500, 5000, 9500}, StartTs: 1700000000, EndTs: 1700003600, PeriodInterval: 60,
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/series/KXHIGHNY/events/KXHIGHNY-24JAN01/forecast_percentile_history")
	q := ct.req.URL.Query()
	if p := q["percentiles"]; len(p) != 3 || p[0] != "500" || p[1] != "5000" || p[2] != "9500" {
		t.Fatalf("percentiles: %v", p)
	}
	if q.Get("start_ts") != "1700000000" || q.Get("end_ts") != "1700003600" || q.Get("period_interval") != "60" {
		t.Fatalf("query: %v", q)
	}
	if len(got.ForecastHistory) != 1 {
		t.Fatalf("resp: %+v", got)
	}
	fp := got.ForecastHistory[0]
	if fp.EventTicker != "KXHIGHNY-24JAN01" || fp.PeriodInterval != 60 || len(fp.PercentilePoints) != 1 ||
		fp.PercentilePoints[0].Percentile != 5000 || fp.PercentilePoints[0].RawNumericalForecast != 61.2 ||
		fp.PercentilePoints[0].FormattedForecast != "61°F" {
		t.Fatalf("point: %+v", fp)
	}
}

func TestSeries_GetForecastPercentileHistory_Validation(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	if _, err := client.Series.GetForecastPercentileHistory(ctx, "S", "E", nil); err == nil {
		t.Fatal("expected error for nil opts")
	}
	if _, err := client.Series.GetForecastPercentileHistory(ctx, "S", "E", &types.GetEventForecastPercentilesHistoryOpts{Percentiles: make([]int, 11)}); err == nil {
		t.Fatal("expected error for >10 percentiles")
	}
	if _, err := client.Series.GetForecastPercentileHistory(ctx, "S", "E", &types.GetEventForecastPercentilesHistoryOpts{Percentiles: []int{1}, PeriodInterval: 5}); err == nil {
		t.Fatal("expected error for bad period_interval")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestMarkets_GetCandlesticks(t *testing.T) {
	client, ct := newCaptureClient(200, `{"markets":[{"market_ticker":"INXD-24JAN01","candlesticks":[`+candlestickJSON+`]},{"market_ticker":"B","candlesticks":[]}]}`)
	got, err := client.Markets.GetCandlesticks(context.Background(), &types.BatchGetMarketCandlesticksOpts{
		MarketTickers: "INXD-24JAN01,B", StartTs: 1, EndTs: 2, PeriodInterval: 1440, IncludeLatestBeforeStart: ptrOf(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/markets/candlesticks")
	ct.assertQuery(t, map[string]string{
		"market_tickers": "INXD-24JAN01,B", "start_ts": "1", "end_ts": "2", "period_interval": "1440", "include_latest_before_start": "false",
	})
	if len(got.Markets) != 2 || got.Markets[0].MarketTicker != "INXD-24JAN01" || len(got.Markets[0].Candlesticks) != 1 ||
		got.Markets[0].Candlesticks[0].OpenInterestFp != "100.00" {
		t.Fatalf("resp: %+v", got)
	}

	if _, err := client.Markets.GetCandlesticks(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil opts")
	}
	if _, err := client.Markets.GetCandlesticks(context.Background(), &types.BatchGetMarketCandlesticksOpts{MarketTickers: "A"}); err == nil {
		t.Fatal("expected error for period_interval 0")
	}
}

func TestOrderGroups_List(t *testing.T) {
	client, ct := newCaptureClient(200, `{"order_groups":[{"id":"og1","contracts_limit_fp":"10.00","is_auto_cancel_enabled":true,"exchange_index":0}]}`)
	got, err := client.OrderGroups.List(context.Background(), &types.GetOrderGroupsOpts{Subaccount: ptrOf(3)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/portfolio/order_groups")
	ct.assertQuery(t, map[string]string{"subaccount": "3"})
	if len(got.OrderGroups) != 1 || got.OrderGroups[0].ID != "og1" || got.OrderGroups[0].ContractsLimitFp != "10.00" || !got.OrderGroups[0].IsAutoCancelEnabled {
		t.Fatalf("resp: %+v", got)
	}
}

func TestOrderGroups_Create(t *testing.T) {
	client, ct := newCaptureClient(201, `{"order_group_id":"og1","subaccount":2,"exchange_index":0}`)
	got, err := client.OrderGroups.Create(context.Background(), &types.CreateOrderGroupRequest{
		Subaccount: ptrOf(2), ContractsLimitFp: ptrOf("10.00"), ExchangeIndex: ptrOf(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/portfolio/order_groups/create")
	ct.assertBody(t, `{"subaccount":2,"contracts_limit_fp":"10.00","exchange_index":0}`)
	if got.OrderGroupID != "og1" || got.Subaccount != 2 {
		t.Fatalf("resp: %+v", got)
	}

	if _, err := client.OrderGroups.Create(context.Background(), &types.CreateOrderGroupRequest{}); err == nil {
		t.Fatal("expected error when no limit given")
	}
}

func TestOrderGroups_Get(t *testing.T) {
	client, ct := newCaptureClient(200, `{"is_auto_cancel_enabled":false,"contracts_limit_fp":"5.00","orders":["o1","o2"],"exchange_index":1}`)
	got, err := client.OrderGroups.Get(context.Background(), "og1", &types.GetOrderGroupOpts{Subaccount: ptrOf(0)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/portfolio/order_groups/og1")
	ct.assertQuery(t, map[string]string{"subaccount": "0"})
	if got.IsAutoCancelEnabled || got.ContractsLimitFp != "5.00" || len(got.Orders) != 2 || got.Orders[1] != "o2" || got.ExchangeIndex != 1 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestOrderGroups_Delete(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.OrderGroups.Delete(context.Background(), "og1", &types.OrderGroupOpts{Subaccount: ptrOf(1), ExchangeIndex: ptrOf(2)}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/portfolio/order_groups/og1")
	ct.assertQuery(t, map[string]string{"subaccount": "1", "exchange_index": "2"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}
}

func TestOrderGroups_Reset(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.OrderGroups.Reset(context.Background(), "og1", nil); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/portfolio/order_groups/og1/reset")
	ct.assertQuery(t, map[string]string{})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}
}

func TestOrderGroups_Trigger(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.OrderGroups.Trigger(context.Background(), "og1", &types.OrderGroupOpts{ExchangeIndex: ptrOf(1)}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/portfolio/order_groups/og1/trigger")
	ct.assertQuery(t, map[string]string{"exchange_index": "1"})
}

func TestOrderGroups_UpdateLimit(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.OrderGroups.UpdateLimit(context.Background(), "og1", &types.UpdateOrderGroupLimitRequest{ContractsLimit: ptrOf[int64](25)}, &types.OrderGroupOpts{Subaccount: ptrOf(0)}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/portfolio/order_groups/og1/limit")
	ct.assertQuery(t, map[string]string{"subaccount": "0"})
	ct.assertBody(t, `{"contracts_limit":25}`)

	if err := client.OrderGroups.UpdateLimit(context.Background(), "og1", nil, nil); err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestSubaccounts_Create(t *testing.T) {
	client, ct := newCaptureClient(201, `{"subaccount_number":1}`)
	got, err := client.Subaccounts.Create(context.Background(), &types.CreateSubaccountRequest{ExchangeIndex: ptrOf(1)})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/portfolio/subaccounts")
	ct.assertBody(t, `{"exchange_index":1}`)
	if got.SubaccountNumber != 1 {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(201, `{"subaccount_number":2}`)
	if _, err := client.Subaccounts.Create(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertBody(t, `{}`)
}

func TestSubaccounts_GetBalances(t *testing.T) {
	client, ct := newCaptureClient(200, `{"subaccount_balances":[
		{"subaccount_number":0,"exchange_index":0,"balance":"100.0000","updated_ts":1716300000},
		{"subaccount_number":1,"exchange_index":0,"balance":"0.5600","updated_ts":1716300001}
	]}`)
	got, err := client.Subaccounts.GetBalances(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/portfolio/subaccounts/balances")
	if len(got.SubaccountBalances) != 2 || got.SubaccountBalances[1].SubaccountNumber != 1 ||
		got.SubaccountBalances[1].Balance != "0.5600" || got.SubaccountBalances[0].UpdatedTs != 1716300000 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestSubaccounts_Transfer(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	err := client.Subaccounts.Transfer(context.Background(), &types.ApplySubaccountTransferRequest{
		ClientTransferID: "8c35ecb3-328f-4f52-8c7c-0f4b9862f8d1", FromSubaccount: 0, ToSubaccount: 1, AmountCents: 5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/portfolio/subaccounts/transfer")
	ct.assertBody(t, `{"client_transfer_id":"8c35ecb3-328f-4f52-8c7c-0f4b9862f8d1","from_subaccount":0,"to_subaccount":1,"amount_cents":5000}`)

	client, ct = newCaptureClient(200, `{}`)
	if err := client.Subaccounts.Transfer(context.Background(), &types.ApplySubaccountTransferRequest{AmountCents: 1}); err == nil {
		t.Fatal("expected error for missing client_transfer_id")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestSubaccounts_ListTransfers(t *testing.T) {
	client, ct := newCaptureClient(200, `{"transfers":[{"transfer_id":"t1","from_subaccount":0,"to_subaccount":2,"amount_cents":250,"created_ts":1716300000,"exchange_index":0}],"cursor":"next"}`)
	got, err := client.Subaccounts.ListTransfers(context.Background(), &types.GetSubaccountTransfersOpts{Limit: ptrOf[int64](50), Cursor: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/portfolio/subaccounts/transfers")
	ct.assertQuery(t, map[string]string{"limit": "50", "cursor": "abc"})
	if len(got.Transfers) != 1 || got.Transfers[0].TransferID != "t1" || got.Transfers[0].ToSubaccount != 2 ||
		got.Transfers[0].AmountCents != 250 || got.Cursor != "next" {
		t.Fatalf("resp: %+v", got)
	}
}

func TestSubaccounts_GetNetting(t *testing.T) {
	client, ct := newCaptureClient(200, `{"netting_configs":[{"subaccount_number":0,"enabled":true,"exchange_index":0},{"subaccount_number":1,"enabled":false,"exchange_index":0}]}`)
	got, err := client.Subaccounts.GetNetting(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/portfolio/subaccounts/netting")
	if len(got.NettingConfigs) != 2 || !got.NettingConfigs[0].Enabled || got.NettingConfigs[1].Enabled || got.NettingConfigs[1].SubaccountNumber != 1 {
		t.Fatalf("resp: %+v", got)
	}
}

func TestSubaccounts_UpdateNetting(t *testing.T) {
	client, ct := newCaptureClient(200, ``)
	if err := client.Subaccounts.UpdateNetting(context.Background(), &types.UpdateSubaccountNettingRequest{SubaccountNumber: 0, Enabled: false}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/portfolio/subaccounts/netting")
	ct.assertBody(t, `{"subaccount_number":0,"enabled":false}`)

	if err := client.Subaccounts.UpdateNetting(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestPortfolio_GetTotalRestingOrderValue(t *testing.T) {
	client, ct := newCaptureClient(200, `{"total_resting_order_value":12345,"resting_order_value_breakdown":[{"exchange_index":0,"balance":"100.0000"},{"exchange_index":1,"balance":"23.4500"}]}`)
	got, err := client.Portfolio.GetTotalRestingOrderValue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/portfolio/summary/total_resting_order_value")
	if got.TotalRestingOrderValue != 12345 || len(got.RestingOrderValueBreakdown) != 2 ||
		got.RestingOrderValueBreakdown[1].ExchangeIndex != 1 || got.RestingOrderValueBreakdown[1].Balance != "23.4500" {
		t.Fatalf("resp: %+v", got)
	}
}

const collectionJSON = `{
	"collection_ticker":"KXNBAPARLAY","series_ticker":"KXNBA","exchange_index":0,"title":"NBA Parlay","description":"d",
	"open_date":"2024-01-01T00:00:00Z","close_date":"2024-12-31T00:00:00Z",
	"associated_events":[{"ticker":"KXNBA-24JAN01","is_yes_only":true,"size_max":null,"size_min":1,"active_quoters":["q1"]}],
	"associated_event_tickers":["KXNBA-24JAN01"],"is_ordered":false,"is_single_market_per_event":true,"is_all_yes":true,
	"size_min":2,"size_max":5,"functional_description":"all legs must hit"
}`

func TestEvents_ListMultivariateCollections(t *testing.T) {
	client, ct := newCaptureClient(200, `{"multivariate_contracts":[`+collectionJSON+`],"cursor":"c2"}`)
	got, err := client.Events.ListMultivariateCollections(context.Background(), &types.GetMultivariateEventCollectionsOpts{
		Status: types.CollectionStatusOpen, AssociatedEventTicker: "KXNBA-24JAN01", SeriesTicker: "KXNBA", Limit: ptrOf[int64](20), Cursor: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/multivariate_event_collections")
	ct.assertQuery(t, map[string]string{
		"status": "open", "associated_event_ticker": "KXNBA-24JAN01", "series_ticker": "KXNBA", "limit": "20", "cursor": "c1",
	})
	if len(got.MultivariateContracts) != 1 || got.Cursor != "c2" {
		t.Fatalf("resp: %+v", got)
	}
	c := got.MultivariateContracts[0]
	if c.CollectionTicker != "KXNBAPARLAY" || c.SizeMin != 2 || c.SizeMax != 5 || !c.IsAllYes || c.IsOrdered ||
		len(c.AssociatedEvents) != 1 || !c.AssociatedEvents[0].IsYesOnly || c.AssociatedEvents[0].SizeMax != nil ||
		c.AssociatedEvents[0].SizeMin == nil || *c.AssociatedEvents[0].SizeMin != 1 || c.AssociatedEvents[0].ActiveQuoters[0] != "q1" {
		t.Fatalf("collection: %+v", c)
	}
}

func TestEvents_GetMultivariateCollection(t *testing.T) {
	client, ct := newCaptureClient(200, `{"multivariate_contract":`+collectionJSON+`}`)
	got, err := client.Events.GetMultivariateCollection(context.Background(), "KXNBAPARLAY")
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/multivariate_event_collections/KXNBAPARLAY")
	ct.assertQuery(t, map[string]string{})
	if got.MultivariateContract.SeriesTicker != "KXNBA" || got.MultivariateContract.FunctionalDescription != "all legs must hit" {
		t.Fatalf("resp: %+v", got)
	}
}

func TestEvents_CreateMarketInMultivariateCollection(t *testing.T) {
	client, ct := newCaptureClient(200, `{"event_ticker":"KXNBAPARLAY-24JAN01","market_ticker":"KXNBAPARLAY-24JAN01-ABC","market":{"ticker":"KXNBAPARLAY-24JAN01-ABC","event_ticker":"KXNBAPARLAY-24JAN01","market_type":"binary","status":"open"}}`)
	got, err := client.Events.CreateMarketInMultivariateCollection(context.Background(), "KXNBAPARLAY", &types.CreateMarketInMultivariateEventCollectionRequest{
		SelectedMarkets: []types.TickerPair{
			{MarketTicker: "KXNBA-24JAN01-LAL", EventTicker: "KXNBA-24JAN01", Side: types.OrderSideYes},
			{MarketTicker: "KXNBA-24JAN01-BOS", EventTicker: "KXNBA-24JAN01", Side: types.OrderSideNo},
		},
		WithMarketPayload: ptrOf(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/multivariate_event_collections/KXNBAPARLAY")
	ct.assertBody(t, `{"selected_markets":[
		{"market_ticker":"KXNBA-24JAN01-LAL","event_ticker":"KXNBA-24JAN01","side":"yes"},
		{"market_ticker":"KXNBA-24JAN01-BOS","event_ticker":"KXNBA-24JAN01","side":"no"}
	],"with_market_payload":true}`)
	if got.EventTicker != "KXNBAPARLAY-24JAN01" || got.MarketTicker != "KXNBAPARLAY-24JAN01-ABC" ||
		got.Market == nil || got.Market.Status != "open" {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(200, `{}`)
	if _, err := client.Events.CreateMarketInMultivariateCollection(context.Background(), "X", &types.CreateMarketInMultivariateEventCollectionRequest{}); err == nil {
		t.Fatal("expected error for empty selected_markets")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}
