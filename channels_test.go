package channels

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestToSlice(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 4 {
			return p, false
		}
		return p + 1, true
	}, nil)

	values := ToSlice(context.TODO(), ch)
	expectedSlice := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong slice returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestToSliceWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "", func(v string) (string, bool) {
		return v, true
	}, func() { time.Sleep(time.Second) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	values := ToSlice(ctx, ch)
	expectedSlice := []string(nil)
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		return p + 1, true
	}, nil)

	evens := Filter(context.TODO(), ch, func(v int) bool { return v%2 == 0 })

	expected := []int{2, 4, 6, 8, 10}
	got := ToSlice(context.TODO(), Take(context.TODO(), evens, 5))
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expected, got)
	}
}

func TestFilterWithClosedInputChannel(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 3 {
			return p, false
		}
		return p + 1, true
	}, nil)

	evens := Filter(context.TODO(), ch, func(v int) bool { return v%2 == 0 })

	expected := []int{2, 4}
	got := ToSlice(context.TODO(), evens)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expected, got)
	}
}

func TestFilterWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(i int) (int, bool) {
		return i + 1, true
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	evens := Filter(ctx, ch, func(v int) bool { return v%2 == 0 })

	got := ToSlice(context.TODO(), evens)
	if len(got) == 0 {
		t.Fatal("unexpected empty slice")
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
