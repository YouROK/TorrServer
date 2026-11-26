package utils

import (
	"slices"
	"sort"
	"testing"
)

func TestCompareStringsSortsNumbersNaturally(t *testing.T) {
	want := []string{
		"Show/Episode 1.mp4",
		"Show/Episode 2.mp4",
		"Show/Episode 3.mp4",
		"Show/Episode 9.mp4",
		"Show/Episode 10.mp4",
		"Show/Episode 11.mp4",
		"Show/Episode 12.mp4",
		"Show/Episode 20.mp4",
		"Show/Episode 21.mp4",
	}
	got := slices.Clone(want)
	slices.Reverse(got)
	sort.Slice(got, func(i, j int) bool { return CompareStrings(got[i], got[j]) })
	if !slices.Equal(got, want) {
		t.Fatalf("sorted:\n%q\nwant:\n%q", got, want)
	}
}

func TestCompareStrings(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"Episode 1.mp4", "Episode 10.mp4", true},
		{"Episode 10.mp4", "Episode 1.mp4", false},
		{"Episode 2.mp4", "Episode 10.mp4", true},
		{"Episode 10.mp4", "Episode 2.mp4", false},
		{"S01E01.mkv", "S01E02.mkv", true},
		{"1.mkv", "10.mkv", true},
		{"10.mkv", "1.mkv", false},
		{"Episode 1a.mkv", "Episode 1b.mkv", true},
		{"Episode 1b.mkv", "Episode 1a.mkv", false},
		{"a.mkv", "b.mkv", true},
	}
	for _, tt := range tests {
		if got := CompareStrings(tt.a, tt.b); got != tt.want {
			t.Errorf("CompareStrings(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
