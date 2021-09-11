package channels

import "context"

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
