package rutor

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"server/rutor/models"
	"server/settings"
)

// TestConcurrentSearchAndLoadDB проверяет отсутствие гонки при одновременном
// обновлении индекса (loadDB) и поиске (Search). 
// !Запускать с -count=3
func TestConcurrentSearchAndLoadDB(t *testing.T) {
	if settings.BTsets == nil {
		settings.BTsets = &settings.BTSets{EnableRutorSearch: true}
		defer func() { settings.BTsets = nil }()
	} else {
		old := settings.BTsets.EnableRutorSearch
		settings.BTsets.EnableRutorSearch = true
		defer func() { settings.BTsets.EnableRutorSearch = old }()
	}

	dir := t.TempDir()
	oldPath := settings.Path
	settings.Path = dir
	defer func() { settings.Path = oldPath }()

	const numTorrents = 800
	seed := make([]*models.TorrentDetails, numTorrents)
	for i := 0; i < numTorrents; i++ {
		s := strconv.Itoa(i)
		seed[i] = &models.TorrentDetails{
			Title: "Test Film Number " + s + " Part One Two Three Year",
			Name:  "Film " + s,
			Year:  2015 + i%10,
		}
	}
	data, err := json.Marshal(seed)
	if err != nil {
		t.Fatal(err)
	}
	var compressed bytes.Buffer
	w, _ := flate.NewWriter(&compressed, flate.DefaultCompression)
	_, _ = w.Write(data)
	_ = w.Close()
	if err := os.WriteFile(filepath.Join(dir, "rutor.ls"), compressed.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	var wg sync.WaitGroup

	// Горутина: многократно перезагружает БД (долгая перезапись индекса)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			select {
			case <-done:
				return
			default:
				loadDB()
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Несколько горутин: постоянный поиск, пока идёт переиндексация
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			queries := []string{"Test", "Film", "Number", "Part", "Year", "xxx"}
			for j := 0; j < 200; j++ {
				select {
				case <-done:
					return
				default:
					_ = Search(queries[j%len(queries)])
				}
			}
		}()
	}

	// Даём время на пересечение loadDB и Search
	time.Sleep(800 * time.Millisecond)
	close(done)
	wg.Wait()
}
