package oddrip

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

func TestPerpsFCM_CreateSubtrader(t *testing.T) {
	client, ct := newCaptureClient(201, `{"subtrader_id":"acct123_desk1"}`)
	got, err := client.Perps.FCM.CreateSubtrader(context.Background(), &types.CreateMarginFCMSubtraderRequest{SubtraderSuffix: "desk1"})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPost, "/margin/fcm/subtraders")
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"subtrader_suffix":"desk1"}`)
	if got.SubtraderID != "acct123_desk1" {
		t.Fatalf("resp: %+v", got)
	}

	client, ct = newCaptureClient(201, `{}`)
	if _, err := client.Perps.FCM.CreateSubtrader(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestPerpsFCM_GetSubtraderRiskControls(t *testing.T) {
	client, ct := newCaptureClient(200, `{
		"risk_controls":[{"subtrader_id":"acct123_desk1","asset_class":"Crypto","im_cap":"100.0000","current_im":"42.0000"}],
		"notional_limits":[
			{"subtrader_id":"acct123_desk1","notional_value_risk_limit":"5000.0000","current_notional":"1250.0000"},
			{"subtrader_id":"acct123_desk1","market_ticker":"KXBTCPERP","notional_value_risk_limit":"2000.0000","current_notional":"500.0000"}
		]
	}`)
	got, err := client.Perps.FCM.GetSubtraderRiskControls(context.Background(), &types.GetFCMSubtraderRiskControlsOpts{
		SubtraderID: "acct123_desk1", AssetClass: types.MarginAssetClassCrypto,
	})
	if err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodGet, "/margin/fcm/subtraders/risk_controls")
	ct.assertQuery(t, map[string]string{"subtrader_id": "acct123_desk1", "asset_class": "Crypto"})
	if len(got.RiskControls) != 1 || got.RiskControls[0].ImCap != "100.0000" || got.RiskControls[0].CurrentIm != "42.0000" ||
		got.RiskControls[0].AssetClass != types.MarginAssetClassCrypto {
		t.Fatalf("risk_controls: %+v", got.RiskControls)
	}
	if len(got.NotionalLimits) != 2 || got.NotionalLimits[0].MarketTicker != "" || got.NotionalLimits[1].MarketTicker != "KXBTCPERP" ||
		got.NotionalLimits[0].CurrentNotional != "1250.0000" {
		t.Fatalf("notional_limits: %+v", got.NotionalLimits)
	}

	client, ct = newCaptureClient(200, `{"risk_controls":[],"notional_limits":[]}`)
	if _, err := client.Perps.FCM.GetSubtraderRiskControls(context.Background(), &types.GetFCMSubtraderRiskControlsOpts{MarketTicker: "KXBTCPERP"}); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{"market_ticker": "KXBTCPERP"})

	client, ct = newCaptureClient(200, `{"risk_controls":[],"notional_limits":[]}`)
	if _, err := client.Perps.FCM.GetSubtraderRiskControls(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{})
}

func TestPerpsFCM_UpdateSubtraderRiskControls(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.Perps.FCM.UpdateSubtraderRiskControls(context.Background(), &types.UpdateFCMSubtraderRiskControlsRequest{
		SubtraderID: "acct123_desk1", MarketTicker: "KXBTCPERP", ImCap: "100.0000",
	}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/margin/fcm/subtraders/risk_controls")
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"subtrader_id":"acct123_desk1","market_ticker":"KXBTCPERP","im_cap":"100.0000"}`)

	client, ct = newCaptureClient(200, `{}`)
	if err := client.Perps.FCM.UpdateSubtraderRiskControls(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestPerpsFCM_DeleteSubtraderRiskControls(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.Perps.FCM.DeleteSubtraderRiskControls(context.Background(), &types.DeleteFCMSubtraderRiskControlsOpts{
		SubtraderID: "acct123_desk1", AssetClass: types.MarginAssetClassMetals,
	}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/fcm/subtraders/risk_controls")
	ct.assertQuery(t, map[string]string{"subtrader_id": "acct123_desk1", "asset_class": "Metals"})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}

	client, ct = newCaptureClient(200, `{}`)
	if err := client.Perps.FCM.DeleteSubtraderRiskControls(context.Background(), &types.DeleteFCMSubtraderRiskControlsOpts{
		SubtraderID: "acct123_desk1", MarketTicker: "KXBTCPERP",
	}); err != nil {
		t.Fatal(err)
	}
	ct.assertQuery(t, map[string]string{"subtrader_id": "acct123_desk1", "market_ticker": "KXBTCPERP"})
}

func TestPerpsFCM_DeleteSubtraderRiskControls_Validation(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	ctx := context.Background()
	if err := client.Perps.FCM.DeleteSubtraderRiskControls(ctx, nil); err == nil {
		t.Fatal("expected error for nil opts")
	}
	if err := client.Perps.FCM.DeleteSubtraderRiskControls(ctx, &types.DeleteFCMSubtraderRiskControlsOpts{MarketTicker: "KXBTCPERP"}); err == nil {
		t.Fatal("expected error for missing subtrader_id")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestPerpsFCM_UpdateNotionalRiskLimit(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.Perps.FCM.UpdateNotionalRiskLimit(context.Background(), &types.UpdateFCMNotionalRiskLimitRequest{
		NotionalValueRiskLimit: "5000.0000",
	}); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodPut, "/margin/fcm/notional_risk_limit")
	ct.assertQuery(t, map[string]string{})
	ct.assertBody(t, `{"notional_value_risk_limit":"5000.0000"}`)

	client, ct = newCaptureClient(200, `{}`)
	if err := client.Perps.FCM.UpdateNotionalRiskLimit(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
	if ct.req != nil {
		t.Fatalf("request should not be sent: %v", ct.req.URL)
	}
}

func TestPerpsFCM_DeleteNotionalRiskLimit(t *testing.T) {
	client, ct := newCaptureClient(200, `{}`)
	if err := client.Perps.FCM.DeleteNotionalRiskLimit(context.Background()); err != nil {
		t.Fatal(err)
	}
	ct.assertRequest(t, http.MethodDelete, "/margin/fcm/notional_risk_limit")
	ct.assertQuery(t, map[string]string{})
	if len(ct.sent) != 0 {
		t.Fatalf("unexpected body: %s", ct.sent)
	}
}

// CreateSubtrader is a POST the server cannot apply twice (a duplicate suffix
// is a 409), so it keeps the full retry policy, as do the PUT/DELETE writes.
func TestPerpsFCM_RetryPolicy(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		call func(*Client) error
	}{
		{"CreateSubtrader", func(c *Client) error {
			_, err := c.Perps.FCM.CreateSubtrader(ctx, &types.CreateMarginFCMSubtraderRequest{SubtraderSuffix: "desk1"})
			return err
		}},
		{"UpdateSubtraderRiskControls", func(c *Client) error {
			return c.Perps.FCM.UpdateSubtraderRiskControls(ctx, &types.UpdateFCMSubtraderRiskControlsRequest{SubtraderID: "s", ImCap: "1"})
		}},
		{"DeleteNotionalRiskLimit", func(c *Client) error { return c.Perps.FCM.DeleteNotionalRiskLimit(ctx) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, calls := countingServer(t, 503)
			err := tc.call(client)
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != 503 {
				t.Fatalf("err = %v, want *APIError 503", err)
			}
			if got := atomic.LoadInt32(calls); got != 3 {
				t.Fatalf("attempts = %d, want 3", got)
			}
		})
	}

	// A 409 is final: the subtrader already exists, and retrying cannot help.
	client, calls := countingServer(t, 409)
	_, err := client.Perps.FCM.CreateSubtrader(ctx, &types.CreateMarginFCMSubtraderRequest{SubtraderSuffix: "desk1"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusConflict {
		t.Fatalf("err = %v, want *APIError 409", err)
	}
	if got := atomic.LoadInt32(calls); got != 1 {
		t.Fatalf("attempts = %d, want 1", got)
	}
}
