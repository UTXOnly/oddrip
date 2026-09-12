package oddrip

import (
	"context"
	"sync/atomic"
)

type ConcurrentResult[T any] struct {
	Value T
	Err   error
}

// DoConcurrent runs fn for i in [0, n) with at most maxInFlight calls active at
// once (maxInFlight <= 0 means unbounded). A positive maxInFlight bounds the
// goroutines too: at most min(n, maxInFlight) workers are started, each taking
// the next index when its current call returns, so a large n does not create n
// goroutines. Results are index-ordered. If ctx is cancelled, the results
// collected so far are returned along with ctx.Err(); workers stop taking new
// indices, and a call already inside fn finishes in the background (have fn
// honor ctx to cut it short).
func DoConcurrent[T any](ctx context.Context, n, maxInFlight int, fn func(i int) (T, error)) ([]ConcurrentResult[T], error) {
	results := make([]ConcurrentResult[T], n)
	type pair struct {
		i int
		r ConcurrentResult[T]
	}
	workers := n
	if maxInFlight > 0 && maxInFlight < n {
		workers = maxInFlight
	}
	ch := make(chan pair, workers)
	var next atomic.Int64
	for w := 0; w < workers; w++ {
		go func() {
			for {
				idx := next.Add(1) - 1
				if idx >= int64(n) || ctx.Err() != nil {
					return
				}
				val, err := fn(int(idx))
				select {
				case ch <- pair{int(idx), ConcurrentResult[T]{Value: val, Err: err}}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case p := <-ch:
			results[p.i] = p.r
		}
	}
	return results, nil
}
