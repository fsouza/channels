package channels

import "context"

// Take takes an input channel and returns an output channel that will contain
// at most N elements from the input channel.
//
// The capacity of the output channel will be min(cap(inputChannel, n)).
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output channel is always closed on cancellation or after sending N
// elements, even if the input channel is never closed.
func Take[T any](ctx context.Context, in <-chan T, n uint) <-chan T {
	maxLen := int(n)
	out := make(chan T, min(maxLen, cap(in)))
	go func() {
		defer close(out)
		if maxLen == 0 {
			return
		}
		length := 0
		receiveLoop(ctx, in, func(v T) bool {
			if !trySend(ctx, out, v) {
				return false
			}
			length++
			return length < maxLen
		})
	}()
	return out
}

// TakeWhile takes an input channel and returns an output channel that will
// emit values until the provided function returns false. It discards the value
// of the first element for which the predicate function returns false.
//
// The capacity of the output channel will be cap(inputChannel).
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output channel is always closed on cancellation or after the provided
// function returns false for an element, even if the input channel is never
// closed.
func TakeWhile[T any](ctx context.Context, in <-chan T, f func(T) bool) <-chan T {
	out := make(chan T, cap(in))
	go func() {
		defer close(out)
		receiveLoop(ctx, in, func(v T) bool {
			if !f(v) {
				return false
			}
			return trySend(ctx, out, v)
		})
	}()
	return out
}

// TakeUntil takes an input channel and returns an output channel that will
// emit values until the provided function returns true. It discards the value
// of the first element for which the predicate function returns true.
//
// The capacity of the output channel will be cap(inputChannel).
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output channel is always closed on cancellation or after the provided
// function returns false for an element, even if the input channel is never
// closed.
func TakeUntil[T any](ctx context.Context, in <-chan T, f func(T) bool) <-chan T {
	return TakeWhile(ctx, in, func(v T) bool {
		return !f(v)
	})
}
