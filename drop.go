package channels

import "context"

// Drop takes an input channel and returns an output channel that will contain
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
func Drop[T any](ctx context.Context, in <-chan T, n uint) <-chan T {
	maxLen := int(n)
	out := make(chan T, min(maxLen, cap(in)))
	go func() {
		defer close(out)
		length := 0
		receiveLoop(ctx, in, func(v T) bool {
			length++
			if length < maxLen+1 {
				return true
			}

			if !trySend(ctx, out, v) {
				return false
			}
			return true
		})
	}()
	return out
}

// DropWhile takes an input channel and returns an output channel that will
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
func DropWhile[T any](ctx context.Context, in <-chan T, f func(T) bool) <-chan T {
	out := make(chan T, cap(in))
	go func() {
		defer close(out)
		dropping := true
		receiveLoop(ctx, in, func(v T) bool {
			if dropping && f(v) {
				return true
			}
			dropping = false
			return trySend(ctx, out, v)
		})
	}()
	return out
}

// DropUntil takes an input channel and returns an output channel that will
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
func DropUntil[T any](ctx context.Context, in <-chan T, f func(T) bool) <-chan T {
	return DropWhile(ctx, in, func(v T) bool {
		return !f(v)
	})
}
