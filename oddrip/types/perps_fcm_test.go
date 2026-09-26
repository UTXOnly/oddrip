package types

import (
	"encoding/json"
	"testing"
)

func TestPerpsFCMTypes_AssetClassConstants(t *testing.T) {
	if MarginAssetClassCrypto != "Crypto" || MarginAssetClassEquities != "Equities" || MarginAssetClassMetals != "Metals" {
		t.Fatalf("asset classes: %q %q %q", MarginAssetClassCrypto, MarginAssetClassEquities, MarginAssetClassMetals)
	}
}

// Field values are the spec's per-field examples; the spec has no whole-object
// example for these schemas.
func TestPerpsFCMTypes_GetSubtraderRiskControlsResponse(t *testing.T) {
	const payload = `{
		"risk_controls":[
			{"subtrader_id":"acct123_desk1","im_cap":"100.0000","current_im":"42.0000"},
			{"subtrader_id":"acct123_desk1","market_ticker":"KXBTCPERP","im_cap":"100.0000","current_im":"42.0000"},
			{"subtrader_id":"acct123_desk1","asset_class":"Equities","im_cap":"100.0000","current_im":"42.0000"}
		],
		"notional_limits":[
			{"subtrader_id":"acct123_desk1","notional_value_risk_limit":"5000.0000","current_notional":"1250.0000"},
			{"subtrader_id":"acct123_desk1","market_ticker":"KXBTCPERP","notional_value_risk_limit":"5000.0000","current_notional":"1250.0000"}
		]
	}`
	var r GetFCMSubtraderRiskControlsResponse
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		t.Fatal(err)
	}
	if len(r.RiskControls) != 3 || len(r.NotionalLimits) != 2 {
		t.Fatalf("resp: %+v", r)
	}
	all, market, class := r.RiskControls[0], r.RiskControls[1], r.RiskControls[2]
	if all.SubtraderID != "acct123_desk1" || all.ImCap != "100.0000" || all.CurrentIm != "42.0000" || all.MarketTicker != "" || all.AssetClass != "" {
		t.Fatalf("all-markets cap: %+v", all)
	}
	if market.MarketTicker != "KXBTCPERP" || market.AssetClass != "" {
		t.Fatalf("market cap: %+v", market)
	}
	if class.AssetClass != MarginAssetClassEquities || class.MarketTicker != "" {
		t.Fatalf("asset-class cap: %+v", class)
	}
	whole, perMarket := r.NotionalLimits[0], r.NotionalLimits[1]
	if whole.SubtraderID != "acct123_desk1" || whole.MarketTicker != "" || whole.NotionalValueRiskLimit != "5000.0000" || whole.CurrentNotional != "1250.0000" {
		t.Fatalf("whole-subtrader limit: %+v", whole)
	}
	if perMarket.MarketTicker != "KXBTCPERP" {
		t.Fatalf("per-market limit: %+v", perMarket)
	}
}

func TestPerpsFCMTypes_CreateSubtrader(t *testing.T) {
	b, err := json.Marshal(CreateMarginFCMSubtraderRequest{SubtraderSuffix: "desk1"})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"subtrader_suffix":"desk1"}` {
		t.Fatalf("marshal: %s", b)
	}
	var resp CreateMarginFCMSubtraderResponse
	if err := json.Unmarshal([]byte(`{"subtrader_id":"acct123_desk1"}`), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.SubtraderID != "acct123_desk1" {
		t.Fatalf("resp: %+v", resp)
	}
}

func TestPerpsFCMTypes_RequestMarshal(t *testing.T) {
	cases := []struct {
		name string
		v    any
		want string
	}{
		{"risk controls all markets", UpdateFCMSubtraderRiskControlsRequest{SubtraderID: "acct123_desk1", ImCap: "100.0000"},
			`{"subtrader_id":"acct123_desk1","im_cap":"100.0000"}`},
		{"risk controls market", UpdateFCMSubtraderRiskControlsRequest{SubtraderID: "acct123_desk1", MarketTicker: "KXBTCPERP", ImCap: "100.0000"},
			`{"subtrader_id":"acct123_desk1","market_ticker":"KXBTCPERP","im_cap":"100.0000"}`},
		{"risk controls asset class", UpdateFCMSubtraderRiskControlsRequest{SubtraderID: "acct123_desk1", AssetClass: MarginAssetClassCrypto, ImCap: "0"},
			`{"subtrader_id":"acct123_desk1","asset_class":"Crypto","im_cap":"0"}`},
		{"notional limit", UpdateFCMNotionalRiskLimitRequest{NotionalValueRiskLimit: "5000.0000"},
			`{"notional_value_risk_limit":"5000.0000"}`},
	}
	for _, tc := range cases {
		b, err := json.Marshal(tc.v)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if string(b) != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, b, tc.want)
		}
	}
}
