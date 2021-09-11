package channels

import "context"

// ToSlice converts the provided channel to a slice.
//
// This is a blocking function that can be aborted via the provided context or
// by closing the input channel.
func ToSlice[T any](ctx context.Context, in <-chan T) []T {
	var result []T
	receiveLoop(ctx, in, func(v T) bool {
		result = append(result, v)
		return true
	})
	return result
}

// Filter takes an input channel and a function to filter values from the input
// channel and returns a channel from the input type that will only emit values
// for which the predicate function returns true.
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
func Filter[T any](ctx context.Context, in <-chan T, predicate func(T) bool) <-chan T {
	out := make(chan T, cap(in))
	go func() {
		receiveLoop(ctx, in, func(v T) bool {
			if predicate(v) {
				return trySend(ctx, out, v)
			}
			return true
		})
		close(out)
	}()
	return out
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// receiveLoop loops on the input channel, calling f with values it reads from the
// input channel.
//
// It exits the loop if the context is cancelled, if the input channel is
// closed or if f returns false.
func receiveLoop[T any](ctx context.Context, in <-chan T, f func(T) bool) {
	loop := true
	for loop {
		select {
		case v, ok := <-in:
			loop = ok && f(v)
		case <-ctx.Done():
			loop = false
		}
	}
}

func trySend[T any](ctx context.Context, ch chan<- T, v T) bool {
	select {
	case ch <- v:
		return true
	case <-ctx.Done():
		return false
	}
}
