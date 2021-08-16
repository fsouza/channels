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
		if maxLen == 0 {
			return
		}
		length := 0
		receiveLoop(ctx, in, func(v T) bool {
			if !trySend(ctx, out, v) {
				return false
			}
			length += 1
			return length < maxLen
		})
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
	return FilterMap(ctx, in, func(v InputType) (OutputType, bool) {
		return f(v), true
	})
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

// FilterMap takes an input channel and a function that maps the input type to
// an output type and a boolean, and returns a channel of OutputType that will
// only include items for which the function return true.
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
func FilterMap[InputType, OutputType any](ctx context.Context, in <-chan InputType, f func(InputType) (OutputType, bool)) <-chan OutputType {
	out := make(chan OutputType, cap(in))
	go func() {
		receiveLoop(ctx, in, func(v InputType) bool {
			if outValue, ok := f(v); ok {
				return trySend(ctx, out, outValue)
			}
			return true
		})
		close(out)
	}()
	return out
}

// MapError takes an input channel and a function to transform values of the
// input type to some other type or an error, and returns two channels: one
// with the output type and another one with errors. For each value consumed in
// the input channel, the function is invoked and its results are processed as
// follows:
//
//  - if the function does not return an error, sends the output of the
//    function in the output channel.
//  - if the function returns an error, send it to the error channel and
//    ignores the other value.
//
// The capacity of the output channel will be same as the capacity of the input
// channel. The capacity of the error channel will always be 0.
//
// This is a non-blocking function: it launches a goroutine and returns the
// channel for consumption. In order to stop the inner goroutine, one can close
// the input channel or cancel the provided context.
//
// The output and errors channels is always closed on cancellation, even if the
// input channel is never closed.
func MapError[InputType, OutputType any](ctx context.Context, in <-chan InputType, f func(InputType) (OutputType, error)) (<-chan OutputType, <-chan error) {
	out := make(chan OutputType, cap(in))
	errs := make(chan error)
	go func() {
		receiveLoop(ctx, in, func(v InputType) bool {
			if outValue, err := f(v); err != nil {
				return trySend(ctx, errs, err)
			} else {
				return trySend(ctx, out, outValue)
			}
		})
		close(out)
		close(errs)
	}()
	return out, errs
}

func max(x, y int) int {
	if x > y {
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
