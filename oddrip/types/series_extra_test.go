package types

import (
	"encoding/json"
	"testing"
)

func TestMarketCandlestick_NullPrices(t *testing.T) {
	const payload = `{
		"end_period_ts": 1700003600,
		"yes_bid": {"open_dollars":"0.5500","low_dollars":"0.5400","high_dollars":"0.5700","close_dollars":"0.5600"},
		"yes_ask": {"open_dollars":"0.5700","low_dollars":"0.5600","high_dollars":"0.5900","close_dollars":"0.5800"},
		"price": {"open_dollars":null,"low_dollars":null,"high_dollars":null,"close_dollars":null,"mean_dollars":null,"previous_dollars":"0.5600","min_dollars":"0.1000","max_dollars":"0.9000"},
		"volume_fp": "10.00",
		"open_interest_fp": "100.00"
	}`
	var c MarketCandlestick
	if err := json.Unmarshal([]byte(payload), &c); err != nil {
		t.Fatal(err)
	}
	if c.Price.OpenDollars != nil || c.Price.CloseDollars != nil || c.Price.MeanDollars != nil {
		t.Fatalf("null prices should stay nil: %+v", c.Price)
	}
	if c.Price.PreviousDollars == nil || *c.Price.PreviousDollars != "0.5600" ||
		c.Price.MinDollars == nil || *c.Price.MinDollars != "0.1000" ||
		c.Price.MaxDollars == nil || *c.Price.MaxDollars != "0.9000" {
		t.Fatalf("price: %+v", c.Price)
	}
	out, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var round map[string]json.RawMessage
	if err := json.Unmarshal(out, &round); err != nil {
		t.Fatal(err)
	}
	var price map[string]string
	if err := json.Unmarshal(round["price"], &price); err != nil {
		t.Fatal(err)
	}
	if _, ok := price["open_dollars"]; ok {
		t.Fatalf("nil prices should be omitted on marshal: %s", round["price"])
	}
	if price["previous_dollars"] != "0.5600" {
		t.Fatalf("price: %s", round["price"])
	}
}

func TestSeries_NullableArrays(t *testing.T) {
	const payload = `{"series":{"ticker":"S","frequency":"daily","title":"T","category":"C","tags":null,"settlement_sources":null,"contract_url":"","contract_terms_url":"","fee_type":"flat","fee_multiplier":1,"additional_prohibitions":null}}`
	var out GetSeriesResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.Series.Tags != nil || out.Series.SettlementSources != nil || out.Series.AdditionalProhibitions != nil {
		t.Fatalf("null arrays should be nil: %+v", out.Series)
	}
	if out.Series.FeeType != FeeTypeFlat || out.Series.VolumeFp != "" || out.Series.ProductMetadata != nil {
		t.Fatalf("series: %+v", out.Series)
	}
}

func TestSeries_Categories(t *testing.T) {
	// OpenAPI 3.30.0: category is the primary category, categories is the full
	// discovery list the Series.List category filter matches against.
	const payload = `{"series":{"ticker":"KXHIGHNY","frequency":"daily","title":"NYC high temp","category":"Climate and Weather","categories":["Climate and Weather","Science"],"tags":["Weather"],"settlement_sources":[],"contract_url":"","contract_terms_url":"","fee_type":"quadratic","fee_multiplier":1,"additional_prohibitions":[]}}`
	var out GetSeriesResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.Series.Category != "Climate and Weather" {
		t.Fatalf("category: %q", out.Series.Category)
	}
	if len(out.Series.Categories) != 2 || out.Series.Categories[0] != "Climate and Weather" || out.Series.Categories[1] != "Science" {
		t.Fatalf("categories: %v", out.Series.Categories)
	}

	var empty GetSeriesResponse
	if err := json.Unmarshal([]byte(`{"series":{"ticker":"S","categories":[]}}`), &empty); err != nil {
		t.Fatal(err)
	}
	if empty.Series.Categories == nil || len(empty.Series.Categories) != 0 {
		t.Fatalf("empty categories should be an empty slice: %#v", empty.Series.Categories)
	}
}

func TestCreateOrderGroupRequest_OmitsUnset(t *testing.T) {
	limit := "10.00"
	out, err := json.Marshal(CreateOrderGroupRequest{ContractsLimitFp: &limit})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"contracts_limit_fp":"10.00"}` {
		t.Fatalf("got %s", out)
	}
}

func TestApplySubaccountTransferRequest_KeepsZeroSubaccount(t *testing.T) {
	out, err := json.Marshal(ApplySubaccountTransferRequest{ClientTransferID: "id", FromSubaccount: 0, ToSubaccount: 1, AmountCents: 100})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"client_transfer_id":"id","from_subaccount":0,"to_subaccount":1,"amount_cents":100}` {
		t.Fatalf("got %s", out)
	}
}
