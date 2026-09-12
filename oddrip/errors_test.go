package oddrip

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewAPIError_BodyShapes(t *testing.T) {
	cases := []struct {
		name          string
		status        int
		body          string
		wantCode      string
		wantMessage   string
		wantDetails   string
		wantErrString string
	}{
		{
			name:          "spec flat shape",
			status:        404,
			body:          `{"code":"NOT_FOUND","message":"resource not found","details":"ticker X"}`,
			wantCode:      "NOT_FOUND",
			wantMessage:   "resource not found",
			wantDetails:   "ticker X",
			wantErrString: "api error 404: resource not found",
		},
		{
			// GET /markets/{ticker} for an unknown ticker, as returned by production.
			name:          "production nested under error",
			status:        404,
			body:          `{"error":{"code":"not_found","message":"not found"}}`,
			wantCode:      "not_found",
			wantMessage:   "not found",
			wantErrString: "api error 404: not found",
		},
		{
			// Unauthenticated GET /portfolio/balance, as returned by production.
			name:          "production nested auth failure",
			status:        401,
			body:          `{"error":{"code":"token_authentication_failure","message":"token authentication failure"}}`,
			wantCode:      "token_authentication_failure",
			wantMessage:   "token authentication failure",
			wantErrString: "api error 401: token authentication failure",
		},
		{
			// GET /markets?limit=abc, as returned by production.
			name:          "parameter binding msg shape",
			status:        400,
			body:          `{"msg":"Invalid format for parameter limit: error binding string parameter"}`,
			wantMessage:   "Invalid format for parameter limit: error binding string parameter",
			wantErrString: "api error 400: Invalid format for parameter limit: error binding string parameter",
		},
		{
			name:          "flat wins over nested when both present",
			status:        400,
			body:          `{"code":"flat","message":"flat msg","error":{"code":"nested","message":"nested msg"}}`,
			wantCode:      "flat",
			wantMessage:   "flat msg",
			wantErrString: "api error 400: flat msg",
		},
		{
			name:          "not json",
			status:        502,
			body:          `<html>bad gateway</html>`,
			wantErrString: "api error 502",
		},
		{
			name:          "empty body",
			status:        500,
			body:          ``,
			wantErrString: "api error 500",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tc.status,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(tc.body)),
			}
			e := newAPIError(resp)
			if e.StatusCode != tc.status {
				t.Errorf("StatusCode = %d, want %d", e.StatusCode, tc.status)
			}
			if e.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", e.Code, tc.wantCode)
			}
			if e.Message != tc.wantMessage {
				t.Errorf("Message = %q, want %q", e.Message, tc.wantMessage)
			}
			if e.Details != tc.wantDetails {
				t.Errorf("Details = %q, want %q", e.Details, tc.wantDetails)
			}
			if e.RawBody != tc.body {
				t.Errorf("RawBody = %q, want %q", e.RawBody, tc.body)
			}
			if got := e.Error(); got != tc.wantErrString {
				t.Errorf("Error() = %q, want %q", got, tc.wantErrString)
			}
		})
	}
}

func TestNewAPIError_RawBodyTruncatedAndRequestID(t *testing.T) {
	long := strings.Repeat("x", maxBodySnippet*2)
	resp := &http.Response{
		StatusCode: 503,
		Header:     http.Header{"Request-Id": []string{"req-123"}},
		Body:       io.NopCloser(strings.NewReader(long)),
	}
	e := newAPIError(resp)
	if len(e.RawBody) != maxBodySnippet {
		t.Errorf("RawBody length = %d, want %d", len(e.RawBody), maxBodySnippet)
	}
	if e.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want req-123", e.RequestID)
	}
}
