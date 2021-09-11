package channels

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestDrop(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 9 {
			return p, false
		}
		return p + 1, true
	}, nil)

	values := ToSlice(context.TODO(), Drop(context.TODO(), ch, 5))
	expectedSlice := []int{6, 7, 8, 9, 10}
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestDropNothing(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 9 {
			return p, false
		}
		return p + 1, true
	}, nil)

	values := ToSlice(context.TODO(), Drop(context.TODO(), ch, 0))
	expectedSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestDropAllElements(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 20 {
			return p, false
		}
		return p + 1, true
	}, nil)

	t.Run("drops only the number of arguments", func(t *testing.T) {
		values := ToSlice(context.TODO(), Drop(context.TODO(), ch, 100))
		if values != nil {
			t.Errorf("unexpected non-nil slice: %#v", values)
		}
	})
}

func TestDropWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "", func(v string) (string, bool) {
		return v, true
	}, func() { time.Sleep(time.Second) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	values := ToSlice(context.TODO(), Drop(ctx, ch, 5))
	expectedSlice := []string(nil)
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestDropWhile(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 9 {
			return p, false
		}
		return p + 1, true
	}, nil)

	t.Run("stops as soon as predicate returns false", func(t *testing.T) {
		values := ToSlice(context.TODO(), DropWhile(context.TODO(), ch, func(v int) bool {
			return v < 7 || v%2 == 1
		}))
		expectedSlice := []int{8, 9, 10}
		if !reflect.DeepEqual(values, expectedSlice) {
			t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
		}
	})
}

func TestDropWhileAll(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 9 {
			return p, false
		}
		return p + 1, true
	}, nil)

	values := ToSlice(context.TODO(), DropWhile(context.TODO(), ch, func(v int) bool {
		return v < 20
	}))
	if values != nil {
		t.Errorf("unexpected non-nil slice: %#v", values)
	}
}

func TestDropWhileWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "", func(v string) (string, bool) {
		return v, true
	}, func() { time.Sleep(time.Second) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	values := ToSlice(context.TODO(), DropWhile(ctx, ch, func(v string) bool { return false }))
	expectedSlice := []string(nil)
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestDropUntil(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 19 {
			return p, false
		}
		return p + 1, true
	}, nil)

	values := ToSlice(context.TODO(), DropUntil(context.TODO(), ch, func(v int) bool {
		return v > 10
	}))
	expectedSlice := []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}

func TestDropUntilEmpty(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, 0, func(p int) (int, bool) {
		if p > 19 {
			return p, false
		}
		return p + 1, true
	}, nil)

	values := ToSlice(context.TODO(), DropUntil(context.TODO(), ch, func(v int) bool {
		return v < 0
	}))
	if values != nil {
		t.Errorf("unexpected non-nil slice: %#v", values)
	}
}

func TestDropUntilWithContextCancellation(t *testing.T) {
	t.Parallel()
	ch := startGenerator(t, "", func(v string) (string, bool) {
		return v, true
	}, func() { time.Sleep(time.Second) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	values := ToSlice(context.TODO(), DropUntil(ctx, ch, func(v string) bool { return true }))
	expectedSlice := []string(nil)
	if !reflect.DeepEqual(values, expectedSlice) {
		t.Errorf("wrong values returned\nwant %#v\ngot  %#v", expectedSlice, values)
	}
}
