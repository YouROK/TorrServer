package tgbot

import (
	"strings"
	"sync"
	"testing"

	"server/torr/state"
	cacheSt "server/torr/storage/state"
)

func TestSnakeFormatAndParse(t *testing.T) {
	if !strings.Contains(formatSnake(1, &cacheSt.CacheState{}, "h", 0, 1), "snake_no_data") && formatSnake(1, &cacheSt.CacheState{}, "h", 0, 1) == "" {
		t.Fatal("zero blocks")
	}
	empty := formatSnake(1, &cacheSt.CacheState{Hash: "abcdef0123456789", Capacity: 100, Torrent: &state.TorrentStatus{Title: "Show <x>"}}, "h", 4, 2)
	if !strings.Contains(empty, "Show") || strings.Contains(empty, "<x>") {
		t.Fatalf("empty pieces %q", empty)
	}

	st := &cacheSt.CacheState{
		Hash:         strings.Repeat("ab", 20),
		Capacity:     8,
		PiecesLength: 1,
		PiecesCount:  40,
		Filled:       3,
		Torrent:      &state.TorrentStatus{Title: "Film"},
		Pieces: map[int]cacheSt.ItemState{
			1: {Size: 1},
			5: {Size: 1},
		},
		Readers: []*cacheSt.ReaderState{{Start: 2, End: 6, Reader: 4}},
	}
	txt := formatSnake(1, st, st.Hash, 4, 2)
	if !strings.Contains(txt, "🟩") && !strings.Contains(txt, "🔵") && !strings.Contains(txt, "🟦") {
		t.Fatalf("blocks %q", txt)
	}
	// Second call scrolls the window once the reader is past the threshold.
	st.Readers[0].Reader = 30
	txt = formatSnake(1, st, st.Hash, 4, 2)
	if txt == "" {
		t.Fatal("scrolled")
	}
	filledOnly := &cacheSt.CacheState{
		PiecesCount: 10, PiecesLength: 1, Capacity: 4,
		Pieces: map[int]cacheSt.ItemState{2: {Size: 1}, 8: {Size: 1}},
	}
	if formatSnake(1, filledOnly, "only", 2, 2) == "" {
		t.Fatal("filled window")
	}

	hash, cols, rows := parseSnakeData("abc|8|2")
	if hash != "abc" || cols != 8 || rows != 2 {
		t.Fatalf("%s %d %d", hash, cols, rows)
	}
	hash, cols, rows = parseSnakeData("abc|0|x")
	if hash != "abc" || cols != 20 || rows != 3 {
		t.Fatalf("defaults %s %d %d", hash, cols, rows)
	}
	if snakeData("abc", 8, 2) != "abc|8|2" {
		t.Fatal(snakeData("abc", 8, 2))
	}
	if snakeKeyboard(1, "abc", 8, 2, true) == nil || snakeKeyboard(1, "abc", 8, 2, false) == nil {
		t.Fatal("keyboard")
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = formatSnake(1, st, st.Hash, 5, 2)
			_, _, _ = parseSnakeData(snakeData(st.Hash, 5, 2))
		}()
	}
	wg.Wait()
}
