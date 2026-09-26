package types

import (
	"encoding/json"
	"testing"
)

// The perps spec has no examples for these schemas; payloads list every
// property the schema defines.

func TestPerpsPortfolioTypes_GetMarginBalanceResponse_Unmarshal(t *testing.T) {
	const payload = `{"subaccount_balances":[
		{"subaccount":0,"position_value":"125.5000","account_equity":"1000.0000","maintenance_margin":"40.0000",
		 "initial_margin":"80.0000","resting_orders_margin":"12.0000","available_balance":"908.0000"}
	],"settled_funds":"1050.0000"}`
	var out GetMarginBalanceResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	want := MarginSubaccountBalance{
		Subaccount: 0, PositionValue: "125.5000", AccountEquity: "1000.0000", MaintenanceMargin: "40.0000",
		InitialMargin: "80.0000", RestingOrdersMargin: "12.0000", AvailableBalance: "908.0000",
	}
	if out.SettledFunds != "1050.0000" || len(out.SubaccountBalances) != 1 || out.SubaccountBalances[0] != want {
		t.Fatalf("got %+v", out)
	}
}

func TestPerpsPortfolioTypes_GetMarginFillsResponse_Unmarshal(t *testing.T) {
	const payload = `{"fills":[
		{"fill_id":"f1","order_id":"o1","is_taker":true,"side":"ask","count":"10.00","created_time":"2026-01-02T03:04:05Z",
		 "ticker":"KXBTCPERP","price":"0.5600","entry_price":"0.5500","fees":"0.0200","realized_pnl":"0.1000","order_source":"system"},
		{"fill_id":"f2","order_id":"o2","is_taker":false,"side":"bid","count":"0.50","created_time":"2026-01-02T03:04:06Z",
		 "ticker":"KXBTCPERP","price":"0.5700","entry_price":"0.5700","fees":"0.0000","realized_pnl":"0.0000"}
	],"cursor":"next"}`
	var out GetMarginFillsResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.Cursor != "next" || len(out.Fills) != 2 {
		t.Fatalf("got %+v", out)
	}
	f := out.Fills[0]
	if f.FillID != "f1" || f.OrderID != "o1" || !f.IsTaker || f.Side != BookSideAsk || f.Count != "10.00" ||
		f.Ticker != "KXBTCPERP" || f.Price != "0.5600" || f.EntryPrice != "0.5500" || f.Fees != "0.0200" ||
		f.RealizedPnl != "0.1000" || f.OrderSource != OrderSourceSystem {
		t.Fatalf("fill: %+v", f)
	}
	if ts, err := ParseTime(f.CreatedTime); err != nil || ts.Unix() != 1767323045 {
		t.Fatalf("created_time %q: %v %v", f.CreatedTime, ts, err)
	}
	if out.Fills[1].IsTaker || out.Fills[1].OrderSource != "" {
		t.Fatalf("fill 2: %+v", out.Fills[1])
	}
}

func TestPerpsPortfolioTypes_GetMarginPositionsResponse_Unmarshal(t *testing.T) {
	const payload = `{"positions":[
		{"subaccount":0,"market_ticker":"KXBTCPERP","position":"-3.00","entry_price":"0.5600","unrealized_pnl":"-0.1200",
		 "margin_used":"2.5000","fees":"0.0300","roe":-4.8,"is_portfolio":false},
		{"subaccount":1,"market_ticker":"TEST-PERP","position":"2.00","entry_price":"0.4000","unrealized_pnl":"0.0000",
		 "margin_used":null,"fees":"0.0100","roe":null,"is_portfolio":true}
	]}`
	var out GetMarginPositionsResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Positions) != 2 {
		t.Fatalf("got %+v", out)
	}
	p := out.Positions[0]
	if p.Subaccount != 0 || p.MarketTicker != "KXBTCPERP" || p.Position != "-3.00" || p.EntryPrice != "0.5600" ||
		p.UnrealizedPnl != "-0.1200" || p.MarginUsed == nil || *p.MarginUsed != "2.5000" || p.Fees != "0.0300" ||
		p.ROE == nil || *p.ROE != -4.8 || p.IsPortfolio {
		t.Fatalf("position 0: %+v", p)
	}
	p = out.Positions[1]
	if p.Subaccount != 1 || p.MarginUsed != nil || p.ROE != nil || !p.IsPortfolio {
		t.Fatalf("position 1: %+v", p)
	}
}

func TestPerpsPortfolioTypes_ApplyMarginSubaccountTransferRequest_Marshal(t *testing.T) {
	// All four fields are required, so zero subaccount numbers (the primary
	// account) must still be sent.
	data, err := json.Marshal(ApplyMarginSubaccountTransferRequest{
		ClientTransferID: "8c35ecb3-328f-4f52-8c7c-0f4b9862f8d1", FromSubaccount: 2, ToSubaccount: 0, AmountCents: 1234,
	})
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"client_transfer_id":"8c35ecb3-328f-4f52-8c7c-0f4b9862f8d1","from_subaccount":2,"to_subaccount":0,"amount_cents":1234}`
	if string(data) != want {
		t.Fatalf("marshal: %s", data)
	}
}
