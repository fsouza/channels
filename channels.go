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

// Map takes an input channel and a function to transform values of the input
// type to some other type, and returns a channel from the output type.
//
// The capacity of the output channel will be same as the capacity of the input
// channel.
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output channel is always closed on cancellation, even if the input
// channel is never closed.
func Map[InputType, OutputType any](ctx context.Context, in <-chan InputType, f func(InputType) OutputType) <-chan OutputType {
	out := make(chan OutputType, cap(in))
	go func() {
		defer close(out)
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}

				select {
				case out <- f(v):
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
