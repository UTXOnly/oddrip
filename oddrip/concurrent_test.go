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
