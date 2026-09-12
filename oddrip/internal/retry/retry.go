package retry

import (
	"context"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Config struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	JitterPercent float64
}

var DefaultConfig = Config{
	MaxAttempts:   4,
	InitialDelay:  500 * time.Millisecond,
	MaxDelay:      30 * time.Second,
	JitterPercent: 0.2,
}

func (c Config) Delay(attempt int, retryAfter time.Duration) time.Duration {
	var d time.Duration
	if retryAfter > 0 {
		d = retryAfter
		if d > c.MaxDelay {
			d = c.MaxDelay
		}
	} else {
		backoff := c.InitialDelay * time.Duration(math.Pow(2, float64(attempt)))
		if backoff > c.MaxDelay {
			backoff = c.MaxDelay
		}
		jitter := float64(backoff) * c.JitterPercent * (2*rand.Float64() - 1)
		d = backoff + time.Duration(jitter)
		if d < 0 {
			d = c.InitialDelay
		}
	}
	return d
}

// IsRetryable reports whether a status should be retried. 429 always is: the
// server rejected the request before acting on it. 5xx is ambiguous — a
// gateway timeout or a handler error after the write committed both look the
// same to the client — so it is retried only for idempotent requests.
func IsRetryable(statusCode int, idempotent bool) bool {
	if statusCode == 429 {
		return true
	}
	return idempotent && statusCode >= 500 && statusCode < 600
}

// Do never returns (nil, nil). When every attempt yields a retryable status
// the last response is returned with its body open so the caller can parse it.
//
// Transport errors (connection failures, resets, client timeouts) are retried
// only when idempotent is true: the server may have applied the request even
// though no response arrived, and replaying a non-idempotent write would apply
// it twice.
func Do(ctx context.Context, cfg Config, idempotent bool, fn func() (*http.Response, error)) (*http.Response, error) {
	attempts := cfg.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 0; ; attempt++ {
		resp, err := fn()
		last := attempt == attempts-1
		var retryAfter time.Duration
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if last || !idempotent {
				return nil, err
			}
		} else {
			if last || !IsRetryable(resp.StatusCode, idempotent) {
				return resp, nil
			}
			retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
			// Drain a bounded amount so the connection can be reused for the retry.
			io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
			resp.Body.Close()
		}
		if err := wait(ctx, cfg.Delay(attempt, retryAfter)); err != nil {
			return nil, err
		}
	}
}

func parseRetryAfter(s string) time.Duration {
	if s == "" {
		return 0
	}
	if sec, err := strconv.Atoi(s); err == nil {
		return time.Duration(sec) * time.Second
	}
	if t, err := http.ParseTime(s); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func wait(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
