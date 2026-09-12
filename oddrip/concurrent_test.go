package oddrip

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoConcurrentBounded(t *testing.T) {
	const n, limit = 20, 3
	var inFlight, peak atomic.Int32
	results, err := DoConcurrent(context.Background(), n, limit, func(i int) (int, error) {
		cur := inFlight.Add(1)
		defer inFlight.Add(-1)
		for p := peak.Load(); cur > p && !peak.CompareAndSwap(p, cur); p = peak.Load() {
		}
		time.Sleep(10 * time.Millisecond)
		return i, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != n {
		t.Fatalf("len(results) = %d, want %d", len(results), n)
	}
	if got := peak.Load(); got != limit {
		t.Fatalf("peak in-flight = %d, want %d", got, limit)
	}
}

func TestDoConcurrentUnbounded(t *testing.T) {
	for _, limit := range []int{0, -1} {
		const n = 8
		var arrived atomic.Int32
		start := make(chan struct{})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := DoConcurrent(ctx, n, limit, func(i int) (int, error) {
			if arrived.Add(1) == n {
				close(start)
			}
			select {
			case <-start:
				return i, nil
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		})
		cancel()
		if err != nil {
			t.Fatalf("maxInFlight=%d: all %d calls should run at once: %v", limit, n, err)
		}
	}
}

func TestDoConcurrentOrderAndErrors(t *testing.T) {
	const n = 10
	errOdd := errors.New("odd")
	results, err := DoConcurrent(context.Background(), n, 4, func(i int) (int, error) {
		time.Sleep(time.Duration(n-i) * time.Millisecond)
		if i%2 == 1 {
			return 0, errOdd
		}
		return i * i, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != n {
		t.Fatalf("len(results) = %d, want %d", len(results), n)
	}
	for i, r := range results {
		if i%2 == 1 {
			if !errors.Is(r.Err, errOdd) {
				t.Errorf("results[%d].Err = %v, want %v", i, r.Err, errOdd)
			}
			continue
		}
		if r.Err != nil || r.Value != i*i {
			t.Errorf("results[%d] = {%d %v}, want {%d nil}", i, r.Value, r.Err, i*i)
		}
	}
}

func TestDoConcurrentCancel(t *testing.T) {
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const n, fast = 50, 3
	var entered atomic.Int32
	type out struct {
		results []ConcurrentResult[int]
		err     error
	}
	done := make(chan out, 1)
	go func() {
		r, err := DoConcurrent(ctx, n, 4, func(i int) (int, error) {
			if entered.Add(1) <= fast {
				return i + 1, nil
			}
			<-ctx.Done()
			return 0, ctx.Err()
		})
		done <- out{r, err}
	}()

	for entered.Load() <= fast {
		time.Sleep(time.Millisecond)
	}
	time.Sleep(20 * time.Millisecond)
	cancel()

	var got out
	select {
	case got = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("DoConcurrent did not return after cancel")
	}
	if !errors.Is(got.err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", got.err)
	}
	if len(got.results) != n {
		t.Fatalf("len(results) = %d, want %d", len(got.results), n)
	}
	filled := 0
	for i, r := range got.results {
		if r.Value == 0 {
			continue
		}
		filled++
		if r.Err != nil || r.Value != i+1 {
			t.Errorf("results[%d] = {%d %v}, want {%d nil}", i, r.Value, r.Err, i+1)
		}
	}
	if filled != fast {
		t.Errorf("partial results = %d, want %d", filled, fast)
	}

	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before+2 {
		if time.Now().After(deadline) {
			t.Fatalf("goroutines leaked: before=%d after=%d", before, runtime.NumGoroutine())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// waitUntil polls cond until it holds or the deadline passes.
func waitUntil(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

// TestDoConcurrentLargeN checks that a bounded maxInFlight bounds the worker
// goroutines as well as the active fn calls (#14): with n=10000 and a limit of
// 4, only the workers exist while fn is blocked, the peak in-flight count is
// the limit, and the dynamically claimed indices still land in order.
func TestDoConcurrentLargeN(t *testing.T) {
	before := runtime.NumGoroutine()
	const n, limit = 10000, 4
	var inFlight, peak atomic.Int32
	release := make(chan struct{})
	type out struct {
		results []ConcurrentResult[int]
		err     error
	}
	done := make(chan out, 1)
	go func() {
		r, err := DoConcurrent(context.Background(), n, limit, func(i int) (int, error) {
			cur := inFlight.Add(1)
			defer inFlight.Add(-1)
			for p := peak.Load(); cur > p && !peak.CompareAndSwap(p, cur); p = peak.Load() {
			}
			<-release
			return i, nil
		})
		done <- out{r, err}
	}()

	waitUntil(t, "workers to block in fn", func() bool { return inFlight.Load() == limit })
	if now := runtime.NumGoroutine(); now-before >= 100 {
		t.Fatalf("goroutines while fn blocked: before=%d now=%d; n=%d with maxInFlight=%d should not start n goroutines", before, now, n, limit)
	}
	close(release)

	var got out
	select {
	case got = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("DoConcurrent did not return")
	}
	if got.err != nil {
		t.Fatal(got.err)
	}
	if len(got.results) != n {
		t.Fatalf("len(results) = %d, want %d", len(got.results), n)
	}
	for i, r := range got.results {
		if r.Err != nil || r.Value != i {
			t.Fatalf("results[%d] = {%d %v}, want {%d nil}", i, r.Value, r.Err, i)
		}
	}
	if got := peak.Load(); got != limit {
		t.Fatalf("peak in-flight = %d, want %d", got, limit)
	}
	waitUntil(t, "workers to exit", func() bool { return runtime.NumGoroutine() <= before+2 })
}

// A limit above n starts n workers, not maxInFlight.
func TestDoConcurrentLimitAboveN(t *testing.T) {
	before := runtime.NumGoroutine()
	const n, limit = 2, 1000
	var inFlight atomic.Int32
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := DoConcurrent(context.Background(), n, limit, func(i int) (int, error) {
			inFlight.Add(1)
			<-release
			return i, nil
		})
		done <- err
	}()

	waitUntil(t, "workers to block in fn", func() bool { return inFlight.Load() == n })
	if now := runtime.NumGoroutine(); now-before > n+1 {
		t.Fatalf("goroutines while fn blocked: before=%d now=%d; want at most %d extra", before, now, n+1)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// Cancelling while every worker is blocked inside fn (on something other than
// ctx) must return promptly with ctx.Err(). Once fn is released the workers must
// neither block sending a result nobody reads nor take another index.
func TestDoConcurrentCancelWhileBlocked(t *testing.T) {
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const n, limit = 1000, 4
	var entered atomic.Int32
	release := make(chan struct{})
	type out struct {
		results []ConcurrentResult[int]
		err     error
	}
	done := make(chan out, 1)
	go func() {
		r, err := DoConcurrent(ctx, n, limit, func(i int) (int, error) {
			entered.Add(1)
			<-release
			return i + 1, nil
		})
		done <- out{r, err}
	}()

	waitUntil(t, "workers to block in fn", func() bool { return entered.Load() == limit })
	cancel()

	var got out
	select {
	case got = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("DoConcurrent did not return after cancel while fn was blocked")
	}
	if !errors.Is(got.err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", got.err)
	}
	if len(got.results) != n {
		t.Fatalf("len(results) = %d, want %d", len(got.results), n)
	}
	for i, r := range got.results {
		if r.Value != 0 || r.Err != nil {
			t.Fatalf("results[%d] = {%d %v}, want zero value: nothing completed before cancel", i, r.Value, r.Err)
		}
	}

	close(release)
	waitUntil(t, "workers to exit", func() bool { return runtime.NumGoroutine() <= before+2 })
	if got := entered.Load(); got != limit {
		t.Errorf("fn entered %d times, want %d: workers took new indices after cancel", got, limit)
	}
}

func TestDoConcurrentZero(t *testing.T) {
	results, err := DoConcurrent(context.Background(), 0, 4, func(i int) (int, error) {
		t.Errorf("fn called with i=%d for n=0", i)
		return 0, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}
