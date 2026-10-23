package torrstor

import "testing"

func TestRanges(t *testing.T) {
	if inRanges(nil, 0) || inRanges([]Range{{Start: 2, End: 4}}, 1) || inRanges([]Range{{Start: 2, End: 4}}, 5) {
		t.Fatal("outside")
	}
	if !inRanges([]Range{{Start: 2, End: 4}, {Start: 8, End: 8}}, 2) || !inRanges([]Range{{Start: 2, End: 4}}, 4) || !inRanges([]Range{{Start: 8, End: 8}}, 8) {
		t.Fatal("inside")
	}
	merged := mergeRange([]Range{{Start: 5, End: 6}, {Start: 1, End: 3}, {Start: 3, End: 4}, {Start: 10, End: 12}})
	if len(merged) != 3 || merged[0].Start != 1 || merged[0].End != 4 || merged[1].Start != 5 || merged[2].End != 12 {
		t.Fatalf("%#v", merged)
	}
	touch := mergeRange([]Range{{Start: 1, End: 4}, {Start: 4, End: 9}})
	if len(touch) != 1 || touch[0].Start != 1 || touch[0].End != 9 {
		t.Fatalf("%#v", touch)
	}
	one := mergeRange([]Range{{Start: 1, End: 1}})
	if len(one) != 1 || one[0].End != 1 {
		t.Fatalf("%#v", one)
	}
	if mergeRange(nil) != nil {
		t.Fatal("nil")
	}
}
