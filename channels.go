package channels

import "context"

// Take takes at most n elements from the input channel and returns it in a
// slice.
//
// This is a blocking function that can be aborted via the provided context or
// by closing the input channel.
func Take[T any](ctx context.Context, in <-chan T, n uint) []T {
	var result []T
	for {
		if uint(len(result)) == n {
			return result
		}
		select {
		case v, ok := <-in:
			if !ok {
				return result
			}
			result = append(result, v)
		case <-ctx.Done():
			return result
		}
	}
}
