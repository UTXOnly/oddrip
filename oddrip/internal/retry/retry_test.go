package retry

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

var fastConfig = Config{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: 2 * time.Millisecond}

type body struct {
	io.Reader
	closed bool
}

func (b *body) Close() error {
	b.closed = true
	return nil
}

func newResp(status int, text string, header http.Header) *http.Response {
	if header == nil {
		header = make(http.Header)
	}
	return &http.Response{StatusCode: status, Header: header, Body: &body{Reader: strings.NewReader(text)}}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

func TestDo_ExhaustedReturnsLastResponse(t *testing.T) {
	var calls int
	var bodies []*body
	resp, err := Do(context.Background(), fastConfig, true, func() (*http.Response, error) {
		calls++
		r := newResp(429, "rate limited", nil)
		bodies = append(bodies, r.Body.(*body))
		return r, nil
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp == nil || resp.StatusCode != 429 {
		t.Fatalf("resp: %+v", resp)
	}
	if got := readBody(t, resp); got != "rate limited" {
		t.Fatalf("body: %q", got)
	}
	if calls != fastConfig.MaxAttempts {
		t.Fatalf("calls: %d", calls)
	}
	for i, b := range bodies[:len(bodies)-1] {
		if !b.closed {
			t.Fatalf("attempt %d body not closed", i)
		}
	}
	if bodies[len(bodies)-1].closed {
		t.Fatal("final body closed")
	}
}

func TestDo_TransportErrorThenSuccess(t *testing.T) {
	var calls int
	resp, err := Do(context.Background(), fastConfig, true, func() (*http.Response, error) {
		calls++
		if calls < fastConfig.MaxAttempts {
			return nil, errors.New("dial")
		}
		return newResp(200, "ok", nil), nil
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.StatusCode != 200 || readBody(t, resp) != "ok" {
		t.Fatalf("resp: %+v", resp)
	}
	if calls != fastConfig.MaxAttempts {
		t.Fatalf("calls: %d", calls)
	}
}

func TestDo_TransportErrorExhausted(t *testing.T) {
	sentinel := errors.New("dial")
	var calls int
	resp, err := Do(context.Background(), fastConfig, true, func() (*http.Response, error) {
		calls++
		return nil, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err: %v", err)
	}
	if resp != nil {
		t.Fatalf("resp: %+v", resp)
	}
	if calls != fastConfig.MaxAttempts {
		t.Fatalf("calls: %d", calls)
	}
}

func TestDo_NonRetryableReturnedImmediately(t *testing.T) {
	var calls int
	resp, err := Do(context.Background(), fastConfig, true, func() (*http.Response, error) {
		calls++
		return newResp(400, "bad request", nil), nil
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.StatusCode != 400 || readBody(t, resp) != "bad request" {
		t.Fatalf("resp: %+v", resp)
	}
	if calls != 1 {
		t.Fatalf("calls: %d", calls)
	}
}

func TestDo_ContextCancelledWhileWaiting(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialDelay: time.Second, MaxDelay: 30 * time.Second}
	h := http.Header{"Retry-After": []string{"5"}}

	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		start := time.Now()
		resp, err := Do(ctx, cfg, true, func() (*http.Response, error) {
			return newResp(503, "", h), nil
		})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err: %v", err)
		}
		if resp != nil {
			t.Fatalf("resp: %+v", resp)
		}
		if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
			t.Fatalf("took %v", elapsed)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var calls int
		start := time.Now()
		_, err := Do(ctx, cfg, true, func() (*http.Response, error) {
			calls++
			cancel()
			return newResp(503, "", h), nil
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err: %v", err)
		}
		if calls != 1 {
			t.Fatalf("calls: %d", calls)
		}
		if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
			t.Fatalf("took %v", elapsed)
		}
	})
}

func TestParseRetryAfter(t *testing.T) {
	if d := parseRetryAfter("2"); d != 2*time.Second {
		t.Fatalf("seconds: %v", d)
	}
	future := time.Now().Add(2 * time.Second).UTC().Format(http.TimeFormat)
	if d := parseRetryAfter(future); d <= 500*time.Millisecond || d > 2*time.Second {
		t.Fatalf("http-date: %v", d)
	}
	past := time.Now().Add(-time.Minute).UTC().Format(http.TimeFormat)
	if d := parseRetryAfter(past); d != 0 {
		t.Fatalf("past http-date: %v", d)
	}
	for _, s := range []string{"", "soon", "-3"} {
		if d := parseRetryAfter(s); d > 0 {
			t.Fatalf("%q: %v", s, d)
		}
	}
}

func TestDelay_RetryAfter(t *testing.T) {
	cfg := Config{InitialDelay: time.Millisecond, MaxDelay: 30 * time.Second}
	if d := cfg.Delay(0, 2*time.Second); d != 2*time.Second {
		t.Fatalf("delay: %v", d)
	}
	cfg.MaxDelay = time.Second
	if d := cfg.Delay(0, 2*time.Second); d != time.Second {
		t.Fatalf("clamped delay: %v", d)
	}
}

func TestDo_ZeroMaxAttempts(t *testing.T) {
	cfg := Config{MaxAttempts: 0, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}
	var calls int
	resp, err := Do(context.Background(), cfg, true, func() (*http.Response, error) {
		calls++
		return newResp(429, "rate limited", nil), nil
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp == nil || resp.StatusCode != 429 || readBody(t, resp) != "rate limited" {
		t.Fatalf("resp: %+v", resp)
	}
	if calls != 1 {
		t.Fatalf("calls: %d", calls)
	}
}

// A non-idempotent request must not be replayed after a transport error: the
// server may have applied it even though no response arrived.
func TestDo_NonIdempotent_TransportErrorNotRetried(t *testing.T) {
	var calls int
	want := errors.New("connection reset")
	_, err := Do(context.Background(), fastConfig, false, func() (*http.Response, error) {
		calls++
		return nil, want
	})
	if !errors.Is(err, want) || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestDo_NonIdempotent_5xxNotRetried(t *testing.T) {
	var calls int
	resp, err := Do(context.Background(), fastConfig, false, func() (*http.Response, error) {
		calls++
		return newResp(504, "gateway timeout", nil), nil
	})
	if err != nil || resp.StatusCode != 504 || calls != 1 {
		t.Fatalf("err=%v status=%d calls=%d", err, resp.StatusCode, calls)
	}
}

// 429 means the server rejected the request before acting on it, so it is
// safe to retry regardless of idempotency.
func TestDo_NonIdempotent_429Retried(t *testing.T) {
	var calls int
	resp, err := Do(context.Background(), fastConfig, false, func() (*http.Response, error) {
		calls++
		if calls < 3 {
			return newResp(429, "slow down", nil), nil
		}
		return newResp(200, "ok", nil), nil
	})
	if err != nil || resp.StatusCode != 200 || calls != 3 {
		t.Fatalf("err=%v status=%d calls=%d", err, resp.StatusCode, calls)
	}
}

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		status     int
		idempotent bool
		want       bool
	}{
		{429, false, true}, {429, true, true},
		{500, false, false}, {500, true, true},
		{503, false, false}, {503, true, true},
		{400, true, false}, {404, true, false}, {200, true, false},
	}
	for _, c := range cases {
		if got := IsRetryable(c.status, c.idempotent); got != c.want {
			t.Errorf("IsRetryable(%d, %v) = %v, want %v", c.status, c.idempotent, got, c.want)
		}
	}
}
