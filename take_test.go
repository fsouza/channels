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
