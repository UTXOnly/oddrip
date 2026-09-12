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

// nestedErrorBody returns a production-shape error body padded to exactly n
// bytes with a long details string.
func nestedErrorBody(n int) string {
	const head = `{"error":{"code":"invalid_parameters","message":"invalid parameters","details":"`
	const tail = `"}}`
	return head + strings.Repeat("d", n-len(head)-len(tail)) + tail
}

func TestNewAPIError_LongBodyDecodesStructuredFields(t *testing.T) {
	long := strings.Repeat("d", maxBodySnippet*2)
	cases := []struct {
		name        string
		body        string
		wantCode    string
		wantMessage string
		wantDetails string
	}{
		{
			name:        "spec flat shape",
			body:        `{"code":"invalid_parameters","message":"invalid parameters","details":"` + long + `"}`,
			wantCode:    "invalid_parameters",
			wantMessage: "invalid parameters",
			wantDetails: long,
		},
		{
			name:        "production nested under error",
			body:        `{"error":{"code":"invalid_parameters","message":"invalid parameters","details":"` + long + `"}}`,
			wantCode:    "invalid_parameters",
			wantMessage: "invalid parameters",
			wantDetails: long,
		},
		{
			// Gateway metadata ahead of the error object pushes code and
			// message past the RawBody snippet entirely.
			name:        "nested after long metadata",
			body:        `{"trace":"` + long + `","error":{"code":"not_found","message":"not found"}}`,
			wantCode:    "not_found",
			wantMessage: "not found",
		},
		{
			name:        "parameter binding msg shape",
			body:        `{"msg":"` + long + `"}`,
			wantMessage: long,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.body) <= maxBodySnippet {
				t.Fatalf("fixture is %d bytes, must exceed %d", len(tc.body), maxBodySnippet)
			}
			resp := &http.Response{
				StatusCode: 400,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(tc.body)),
			}
			e := newAPIError(resp)
			if e.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", e.Code, tc.wantCode)
			}
			if e.Message != tc.wantMessage {
				t.Errorf("Message = %q, want %q", e.Message, tc.wantMessage)
			}
			if e.Details != tc.wantDetails {
				t.Errorf("Details length = %d, want %d", len(e.Details), len(tc.wantDetails))
			}
			if len(e.RawBody) != maxBodySnippet {
				t.Errorf("RawBody length = %d, want %d", len(e.RawBody), maxBodySnippet)
			}
			if !strings.HasPrefix(tc.body, e.RawBody) {
				t.Errorf("RawBody is not a prefix of the body")
			}
		})
	}
}

func TestNewAPIError_MalformedLongBody(t *testing.T) {
	full := nestedErrorBody(maxBodySnippet * 2)
	cases := []struct {
		name string
		body string
	}{
		{name: "truncated json", body: full[:len(full)-3]},
		{name: "not json", body: "<html>" + strings.Repeat("x", maxBodySnippet*2) + "</html>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: 502,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(tc.body)),
			}
			e := newAPIError(resp)
			if e.StatusCode != 502 {
				t.Errorf("StatusCode = %d, want 502", e.StatusCode)
			}
			if e.Code != "" || e.Message != "" || e.Details != "" || e.Service != "" {
				t.Errorf("structured fields set from malformed body: %+v", e)
			}
			if len(e.RawBody) != maxBodySnippet {
				t.Errorf("RawBody length = %d, want %d", len(e.RawBody), maxBodySnippet)
			}
			if !strings.HasPrefix(tc.body, e.RawBody) {
				t.Errorf("RawBody is not a prefix of the body")
			}
			if got := e.Error(); got != "api error 502" {
				t.Errorf("Error() = %q, want %q", got, "api error 502")
			}
		})
	}
}

// countingReader records how many bytes newAPIError pulls from the body.
type countingReader struct {
	r io.Reader
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

func TestNewAPIError_ErrorBodyReadIsBounded(t *testing.T) {
	cases := []struct {
		name        string
		size        int
		wantMessage string
	}{
		{name: "at limit decodes", size: maxErrorBody, wantMessage: "invalid parameters"},
		{name: "one byte over is not decoded", size: maxErrorBody + 1},
		{name: "far over is not decoded", size: 4 * maxErrorBody},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := nestedErrorBody(tc.size)
			if len(body) != tc.size {
				t.Fatalf("fixture is %d bytes, want %d", len(body), tc.size)
			}
			cr := &countingReader{r: strings.NewReader(body)}
			resp := &http.Response{
				StatusCode: 500,
				Header:     http.Header{},
				Body:       io.NopCloser(cr),
			}
			e := newAPIError(resp)
			if cr.n > maxErrorBody {
				t.Errorf("read %d bytes from the body, want at most %d", cr.n, maxErrorBody)
			}
			if e.Message != tc.wantMessage {
				t.Errorf("Message = %q, want %q", e.Message, tc.wantMessage)
			}
			if len(e.RawBody) != maxBodySnippet {
				t.Errorf("RawBody length = %d, want %d", len(e.RawBody), maxBodySnippet)
			}
			if !strings.HasPrefix(body, e.RawBody) {
				t.Errorf("RawBody is not a prefix of the body")
			}
		})
	}
}
