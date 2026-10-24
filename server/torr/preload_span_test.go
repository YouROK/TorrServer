package torr

import "testing"

func TestPreloadSpansKeepsBothEdges(t *testing.T) {
	const mb int64 = 1 << 20
	const gb int64 = 1 << 30
	tests := []struct {
		name             string
		size, file, edge int64
		head, tail       int64
	}{
		{"long file", 32 * mb, 2 * gb, 8 * mb, 24 * mb, 2*gb - 8*mb},
		{"budget equals window, file shorter than two windows", 16 * mb, 20 * mb, 16 * mb, 16 * mb, 16 * mb},
		{"budget equals window, long file", 16 * mb, 2 * gb, 16 * mb, 16 * mb, 2*gb - 16*mb},
		{"file fits in the budget", 32 * mb, 30 * mb, 8 * mb, 22 * mb, 22 * mb},
		{"file shorter than the window", 64 << 10, 64 << 10, 8 * mb, 64 << 10, 64 << 10},
		{"budget smaller than window, long file", 4 * mb, 2 * gb, 8 * mb, 4 * mb, 2*gb - 8*mb},
		{"budget smaller than window, short file", 4 * mb, 10 * mb, 8 * mb, 4 * mb, 4 * mb},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head, tail := preloadSpans(tt.size, tt.file, tt.edge)
			if head != tt.head || tail != tt.tail {
				t.Fatalf("head, tail = %d, %d; want %d, %d", head, tail, tt.head, tt.tail)
			}
			if head > 0 && head > tt.file {
				t.Fatal("head past EOF")
			}
			if tail > tt.file || tail < head {
				t.Fatalf("tail %d outside [%d, %d]", tail, head, tt.file)
			}
		})
	}
}
