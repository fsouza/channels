package channels

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestTake(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 9 {
			return p, false
		}
		return p + 1, true
	}, nil)

	t.Run("takes only the number of arguments", func(t *testing.T) {
		values := Take(context.TODO(), ch, 5)
		expectedSlice := []int{1, 2, 3, 4, 5}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong slice returned by Take(5)\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})

	t.Run("can take less than N if channel is closed", func(t *testing.T) {
		values := Take(context.TODO(), ch, 10)
		expectedSlice := []int{6, 7, 8, 9, 10}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong slice returned by Take(5)\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})
}

func TestTakeWithContextCancellation(t *testing.T) {
	ch := startGenerator(t, "", func(v string) (string, bool) {
		return v, true
	}, func() { time.Sleep(time.Second) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	values := Take(ctx, ch, 5)
	expectedSlice := []string(nil)
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong slice returned by Take(5)\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func startGenerator[T any](t *testing.T, init T, gen func(prev T) (T, bool), cb func()) <-chan T {
	t.Helper()
	abort := make(chan struct{})
	t.Cleanup(func() { close(abort) })

	v := init
	ch := make(chan T)
	go func() {
		defer close(ch)
		cont := true
		for {
			v, cont = gen(v)
			if !cont {
				return
			}

			if cb != nil {
				cb()
			}
			select {
			case ch <- v:
			case <-abort:
				return
			}
		}
	}()
	return ch
}
