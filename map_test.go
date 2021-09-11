package channels

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

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
