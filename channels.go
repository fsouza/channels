package channels

import "context"

// ToSlice converts the provided channel to a slice.
//
// This is a blocking function that can be aborted via the provided context or
// by closing the input channel.
func ToSlice[T any](ctx context.Context, in <-chan T) []T {
	var result []T
	for {
		if v, ok := tryRead(ctx, in); ok {
			result = append(result, v)
		} else {
			return result
		}
	}
}

// Take takes an input channel and returns an output channel that will contain
// at most N elements from the input channel.
//
// The capacity of the output channel will be max(cap(inputChannel, n)).
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output channel is always closed on cancellation or after sending N
// elements, even if the input channel is never closed.
func Take[T any](ctx context.Context, in <-chan T, n uint) <-chan T {
	maxLen := int(n)
	out := make(chan T, max(maxLen, cap(in)))
	go func() {
		defer close(out)
		for i := 0; i < maxLen; i++ {
			if v, ok := tryRead(ctx, in); ok {
				if !trySend(ctx, out, v) {
					return
				}
			} else {
				return
			}
		}
	}()
	return out
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
			if v, ok := tryRead(ctx, in); ok {
				if !trySend(ctx, out, f(v)) {
					return
				}
			} else {
				return
			}
		}
	}()
	return out
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
		defer close(out)
		for {
			if v, ok := tryRead(ctx, in); ok {
				if predicate(v) {
					if !trySend(ctx, out, v) {
						return
					}
				}
			} else {
				return
			}
		}
	}()
	return out
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func tryRead[T any](ctx context.Context, ch <-chan T) (T, bool) {
	select {
	case v, ok := <-ch:
		return v, ok
	case <-ctx.Done():
		var zero T
		return zero, false
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
