package oddrip

import "context"

type ConcurrentResult[T any] struct {
	Value T
	Err   error
}

func DoConcurrent[T any](ctx context.Context, n int, fn func(i int) (T, error)) ([]ConcurrentResult[T], error) {
	results := make([]ConcurrentResult[T], n)
	type pair struct {
		i int
		r ConcurrentResult[T]
	}
	ch := make(chan pair, n)
	for i := 0; i < n; i++ {
		go func(idx int) {
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
