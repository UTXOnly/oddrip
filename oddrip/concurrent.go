package oddrip

import "context"

type ConcurrentResult[T any] struct {
	Value T
	Err   error
}

// DoConcurrent runs fn for i in [0, n) with at most maxInFlight calls active at
// once (maxInFlight <= 0 means unbounded). Results are index-ordered. If ctx is
// cancelled, the results collected so far are returned along with ctx.Err().
func DoConcurrent[T any](ctx context.Context, n, maxInFlight int, fn func(i int) (T, error)) ([]ConcurrentResult[T], error) {
	results := make([]ConcurrentResult[T], n)
	type pair struct {
		i int
		r ConcurrentResult[T]
	}
	ch := make(chan pair, n)
	var sem chan struct{}
	if maxInFlight > 0 {
		sem = make(chan struct{}, maxInFlight)
	}
	for i := 0; i < n; i++ {
		go func(idx int) {
			if sem != nil {
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-ctx.Done():
					return
				}
			}
			val, err := fn(idx)
			ch <- pair{idx, ConcurrentResult[T]{Value: val, Err: err}}
		}(i)
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
