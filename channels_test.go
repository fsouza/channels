package channels

import (
	"context"
	"fmt"
	"reflect"
	"sync"
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

func TestTake(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 9 {
			return p, false
		}
		return p + 1, true
	}, nil)

	t.Run("takes only the number of arguments", func(t *testing.T) {
		values := ToSlice(context.TODO(), Take(context.TODO(), ch, 5))
		expectedSlice := []int{1, 2, 3, 4, 5}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})

	t.Run("can take 0 elements", func(t *testing.T) {
		values := ToSlice(context.TODO(), Take(context.TODO(), ch, 0))
		if values != nil {
			t.Errorf("unexpected non-nil slice from Take(0): %#v", values)
		}
	})

	t.Run("can take less than N if channel is closed", func(t *testing.T) {
		values := ToSlice(context.TODO(), Take(context.TODO(), ch, 10))
		expectedSlice := []int{6, 7, 8, 9, 10}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})
}

func TestTakeWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "", func(v string) (string, bool) {
		return v, true
	}, func() { time.Sleep(time.Second) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	values := ToSlice(context.TODO(), Take(ctx, ch, 5))
	expectedSlice := []string(nil)
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestTakeWhile(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		return p + 1, true
	}, nil)

	t.Run("stops as soon as predicate returns false", func(t *testing.T) {
		values := ToSlice(context.TODO(), TakeWhile(context.TODO(), ch, func(v int) bool {
			return v < 7 || v%2 == 1
		}))
		expectedSlice := []int{1, 2, 3, 4, 5, 6, 7}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})

	t.Run("can continue from where it left, discarding the element for which it returned false", func(t *testing.T) {
		values := ToSlice(context.TODO(), TakeWhile(context.TODO(), ch, func(v int) bool {
			return v < 11
		}))
		expectedSlice := []int{9, 10}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})
}

func TestTakeWhileEmpty(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		return p + 1, true
	}, nil)

	t.Run("stops as soon as predicate returns false", func(t *testing.T) {
		values := ToSlice(context.TODO(), TakeWhile(context.TODO(), ch, func(v int) bool {
			return v < 0
		}))
		if values != nil {
			t.Errorf("unexpected non-nil slice: %#v", values)
		}
	})
}

func TestTakeWhileWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "", func(v string) (string, bool) {
		return v, true
	}, func() { time.Sleep(time.Second) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	values := ToSlice(context.TODO(), TakeWhile(ctx, ch, func(v string) bool { return true }))
	expectedSlice := []string(nil)
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestTakeUntil(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		return p + 1, true
	}, nil)

	t.Run("stops as soon as predicate returns true", func(t *testing.T) {
		values := ToSlice(context.TODO(), TakeUntil(context.TODO(), ch, func(v int) bool {
			return v > 10
		}))
		expectedSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})

	t.Run("can continue from where it left, discarding the element for which it returned true", func(t *testing.T) {
		values := ToSlice(context.TODO(), TakeUntil(context.TODO(), ch, func(v int) bool {
			return v%5 == 0
		}))
		expectedSlice := []int{12, 13, 14}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})
}

func TestTakeUntilEmpty(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		return p + 1, true
	}, nil)

	t.Run("stops as soon as predicate returns false", func(t *testing.T) {
		values := ToSlice(context.TODO(), TakeUntil(context.TODO(), ch, func(v int) bool {
			return v > 0
		}))
		if values != nil {
			t.Errorf("unexpected non-nil slice: %#v", values)
		}
	})
}

func TestTakeUntilWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "", func(v string) (string, bool) {
		return v, true
	}, func() { time.Sleep(time.Second) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	values := ToSlice(context.TODO(), TakeUntil(ctx, ch, func(v string) bool { return false }))
	expectedSlice := []string(nil)
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestMap(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		return p + 1, true
	}, nil)

	doubled := Map(context.TODO(), ch, func(v int) int { return v * 2 })

	expected := []int{2, 4, 6, 8, 10}
	got := ToSlice(context.TODO(), Take(context.TODO(), doubled, 5))
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expected, got)
	}
}

func TestMapWithClosedInputChannel(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 3 {
			return p, false
		}
		return p + 1, true
	}, nil)

	doubled := Map(context.TODO(), ch, func(v int) int { return v * 2 })

	expected := []int{2, 4, 6, 8}
	got := ToSlice(context.TODO(), doubled)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expected, got)
	}
}

func TestMapWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "foo", func(p string) (string, bool) {
		return p, true
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	lengths := Map(ctx, ch, func(v string) int { return len(v) })

	got := ToSlice(context.TODO(), lengths)
	if len(got) == 0 {
		t.Fatal("unexpected empty slice")
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

func TestFilterMap(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		return p + 1, true
	}, nil)

	doubledEvens := FilterMap(context.TODO(), ch, func(v int) (int, bool) { return v * 2, v%2 == 0 })

	expected := []int{4, 8, 12, 16, 20}
	got := ToSlice(context.TODO(), Take(context.TODO(), doubledEvens, 5))
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expected, got)
	}
}

func TestFilterMapWithClosedInputChannel(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 8 {
			return p, false
		}
		return p + 1, true
	}, nil)

	doubledEvens := FilterMap(context.TODO(), ch, func(v int) (int, bool) { return v * 2, v%2 == 0 })

	expected := []int{4, 8, 12, 16}
	got := ToSlice(context.TODO(), doubledEvens)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expected, got)
	}
}

func TestFilterMapWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "foo", func(p string) (string, bool) {
		return p, true
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	lengths := FilterMap(ctx, ch, func(v string) (int, bool) { return len(v), len(v) > 2 })

	got := ToSlice(context.TODO(), lengths)
	if len(got) == 0 {
		t.Fatal("unexpected empty slice")
	}
}

func TestMapError(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		return p + 1, true
	}, nil)

	doubledOdds, errs := MapError(context.TODO(), ch, func(v int) (int, error) {
		if v%2 == 0 {
			return 0, fmt.Errorf("%d is even, don't like that", v)
		}
		return v * 2, nil
	})

	var gotVals []int
	var gotErrs []string
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		gotVals = ToSlice(context.TODO(), Take(context.TODO(), doubledOdds, 5))
	}()

	go func() {
		defer wg.Done()
		gotErrs = ToSlice(context.TODO(), Map(context.TODO(), Take(context.TODO(), errs, 5), func(err error) string { return err.Error() }))
	}()
	wg.Wait()

	expectedVals := []int{2, 6, 10, 14, 18}
	if !reflect.DeepEqual(gotVals, expectedVals) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedVals, gotVals)
	}

	expectedErrs := []string{
		"2 is even, don't like that",
		"4 is even, don't like that",
		"6 is even, don't like that",
		"8 is even, don't like that",
		"10 is even, don't like that",
	}
	if !reflect.DeepEqual(gotErrs, expectedErrs) {
		t.Errorf("wrong errors returned\nwant %#v\ngot  %#v", expectedErrs, gotErrs)
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
