package torrstor

import (
	"reflect"
	"testing"
)

// These tests exercise mergeRange and inRanges which are pure helpers
// used by the cache to decide which pieces are "interesting" for an
// active reader. They do NOT need a torrent.Torrent / torrent.File,
// because mergeRange/inRanges only look at the Start/End fields.

func TestInRanges(t *testing.T) {
	rs := []Range{{Start: 5, End: 10}, {Start: 20, End: 25}}

	cases := []struct {
		idx  int
		want bool
	}{
		{4, false},
		{5, true},
		{7, true},
		{10, true},
		{11, false},
		{19, false},
		{20, true},
		{25, true},
		{26, false},
	}
	for _, c := range cases {
		got := inRanges(rs, c.idx)
		if got != c.want {
			t.Errorf("inRanges(%d) = %v, want %v", c.idx, got, c.want)
		}
	}
}

func TestInRangesEmpty(t *testing.T) {
	if inRanges(nil, 5) {
		t.Errorf("inRanges(nil, 5) = true, want false")
	}
	if inRanges([]Range{}, 5) {
		t.Errorf("inRanges([], 5) = true, want false")
	}
}

func TestMergeRangeEmpty(t *testing.T) {
	got := mergeRange(nil)
	if len(got) != 0 {
		t.Errorf("mergeRange(nil) returned %v, want empty", got)
	}
}

func TestMergeRangeSingle(t *testing.T) {
	in := []Range{{Start: 5, End: 10}}
	got := mergeRange(in)
	if !reflect.DeepEqual(got, in) {
		t.Errorf("mergeRange(single) = %v, want %v", got, in)
	}
}

func TestMergeRangeNonOverlapping(t *testing.T) {
	in := []Range{
		{Start: 20, End: 25},
		{Start: 5, End: 10},
		{Start: 30, End: 35},
	}
	got := mergeRange(in)
	want := []Range{
		{Start: 5, End: 10},
		{Start: 20, End: 25},
		{Start: 30, End: 35},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeRange = %v, want %v", got, want)
	}
}

func TestMergeRangeOverlapping(t *testing.T) {
	in := []Range{
		{Start: 5, End: 15},
		{Start: 10, End: 20},
		{Start: 25, End: 30},
	}
	got := mergeRange(in)
	want := []Range{
		{Start: 5, End: 20},
		{Start: 25, End: 30},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeRange = %v, want %v", got, want)
	}
}

func TestMergeRangeAdjacent(t *testing.T) {
	// merged because cache.go uses inclusive End and treats touching ranges as overlap:
	// "if merged[j].End >= merged[i].Start"
	in := []Range{
		{Start: 5, End: 10},
		{Start: 10, End: 15},
	}
	got := mergeRange(in)
	want := []Range{{Start: 5, End: 15}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeRange(adjacent) = %v, want %v", got, want)
	}
}

func TestMergeRangeFullyContained(t *testing.T) {
	in := []Range{
		{Start: 5, End: 100},
		{Start: 20, End: 30},
		{Start: 50, End: 60},
	}
	got := mergeRange(in)
	want := []Range{{Start: 5, End: 100}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeRange(contained) = %v, want %v", got, want)
	}
}

func TestMergeRangeDoesNotMutateInput(t *testing.T) {
	in := []Range{
		{Start: 10, End: 20},
		{Start: 5, End: 15},
	}
	inCopy := append([]Range(nil), in...)
	_ = mergeRange(in)
	if !reflect.DeepEqual(in, inCopy) {
		t.Errorf("mergeRange mutated input: got %v, was %v", in, inCopy)
	}
}
