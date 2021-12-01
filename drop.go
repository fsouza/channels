package channels

import "context"

// Drop takes an input channel and returns an output channel that will contain
// all elements from the input channel, except for the first N.
//
// The capacity of the output channel will be cap(inputChannel).
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output channel is always closed on cancellation, even if the input
// channel is never closed.
func Drop[T any](ctx context.Context, in <-chan T, n uint) <-chan T {
	return DropWhile(ctx, in, func(T) bool {
		if n > 0 {
			n--
			return true
		}
		return false
	})
}

// DropWhile takes an input channel and returns an output channel that will
// skip values from the input channel until the provided function returns
// false.
//
// The capacity of the output channel will be cap(inputChannel).
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output channel is always closed on cancellation, even if the input
// channel is never closed.
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
// skip values from the input channel until the provided function returns
// true.
//
// The capacity of the output channel will be cap(inputChannel).
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output channel is always closed on cancellation, even if the input
// channel is never closed.
func DropUntil[T any](ctx context.Context, in <-chan T, f func(T) bool) <-chan T {
	return DropWhile(ctx, in, func(v T) bool {
		return !f(v)
	})
}
