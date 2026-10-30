package rutor

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"server/leakcheck"
	"server/rutor/models"
	"server/rutor/torrsearch"
	"server/settings"
)

func compress(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write(data)
	_ = w.Close()
	return buf.Bytes()
}

// setup points settings.Path at a temp dir, enables rutor search and clears
// the package state again on cleanup.
func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	oldPath, oldSets := settings.Path, settings.BTsets
	settings.Path = dir
	settings.BTsets = &settings.BTSets{EnableRutorSearch: true}
	oldURL := rutorDBURL
	t.Cleanup(func() {
		settings.Path, settings.BTsets, rutorDBURL = oldPath, oldSets, oldURL
		mu.Lock()
		torrs = nil
		torrsearch.NewIndex(nil)
		mu.Unlock()
	})
	return dir
}

var sample = []*models.TorrentDetails{
	{Title: "Матрица / The Matrix (1999) BDRip", Name: "Матрица", Names: []string{"The Matrix"}, Year: 1999,
		CreateDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
	{Title: "Матрица: Революция / The Matrix Revolutions (2003)", Name: "Матрица: Революция", Names: []string{"The Matrix Revolutions"}, Year: 2003,
		CreateDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
	{Title: "Матрица / The Matrix (1999) WEB-DL", Name: "Матрица", Names: []string{"The Matrix"}, Year: 1999,
		CreateDate: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
	{Title: "Ёлки (2010)", Name: "Ёлки", Year: 2010},
}

func TestSearchDisabledReturnsNil(t *testing.T) {
	setup(t)
	if err := os.WriteFile(filepath.Join(settings.Path, "rutor.ls"), compress(t, sample), 0o600); err != nil {
		t.Fatal(err)
	}
	loadDB()
	settings.BTsets.EnableRutorSearch = false
	if got := Search("matrix"); got != nil {
		t.Fatalf("disabled search returned %d results", len(got))
	}
}

func TestSearchRanksByDistanceThenNewest(t *testing.T) {
	dir := setup(t)
	if err := os.WriteFile(filepath.Join(dir, "rutor.ls"), compress(t, sample), 0o600); err != nil {
		t.Fatal(err)
	}
	loadDB()

	got := Search("Матрица 1999")
	if len(got) != 2 {
		t.Fatalf("got %d results, want the two 1999 releases", len(got))
	}
	// Equal Levenshtein distance → the newer CreateDate comes first.
	if !got[0].CreateDate.After(got[1].CreateDate) {
		t.Errorf("order = %v then %v, want newest first", got[0].CreateDate, got[1].CreateDate)
	}

	// A closer name beats a longer one.
	all := Search("matrix")
	if len(all) != 3 {
		t.Fatalf("matrix results = %d, want 3", len(all))
	}
	if all[2].Year != 2003 {
		t.Errorf("the Revolutions entry should rank last, got year %d", all[2].Year)
	}

	if Search("несуществующее") != nil {
		t.Error("miss must return nil")
	}
	if got := Search("елки"); len(got) != 1 || got[0].Name != "Ёлки" {
		t.Errorf("ё/е folding: %+v", got)
	}
}

func TestLoadDBMissingAndCorrupt(t *testing.T) {
	dir := setup(t)

	loadDB() // no file: must not panic, state stays empty
	if Search("anything") != nil {
		t.Fatal("search without a db")
	}

	// A valid db is loaded, then a corrupt one must not wipe it.
	if err := os.WriteFile(filepath.Join(dir, "rutor.ls"), compress(t, sample), 0o600); err != nil {
		t.Fatal(err)
	}
	loadDB()
	if err := os.WriteFile(filepath.Join(dir, "rutor.ls"), []byte("<html>502 bad gateway</html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	loadDB()
	if len(Search("matrix")) != 3 {
		t.Fatal("corrupt db replaced a good in-memory index")
	}
}

func TestLoadDBSkipsBrokenEntries(t *testing.T) {
	dir := setup(t)
	// Year is an int; the middle element can't decode and is dropped.
	raw := `[{"Title":"good one","Year":2000},{"Title":"bad","Year":"nope"},{"Title":"good two","Year":2001}]`
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	_, _ = w.Write([]byte(raw))
	_ = w.Close()
	if err := os.WriteFile(filepath.Join(dir, "rutor.ls"), buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	loadDB()
	mu.RLock()
	n := len(torrs)
	mu.RUnlock()
	if n < 2 {
		t.Fatalf("loaded %d entries, want at least the 2 good ones", n)
	}
	if len(Search("good")) != 2 {
		t.Fatalf("good entries not searchable")
	}
}

func serve(t *testing.T, status int, body []byte, hits *atomic.Int32) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	rutorDBURL = srv.URL
}

func TestUpdateDBDownloadsAndIndexes(t *testing.T) {
	dir := setup(t)
	var hits atomic.Int32
	serve(t, http.StatusOK, compress(t, sample), &hits)

	if !updateDB() {
		t.Fatal("first download must report an update")
	}
	if _, err := os.Stat(filepath.Join(dir, "rutor.tmp")); !os.IsNotExist(err) {
		t.Error("rutor.tmp left behind")
	}
	if len(Search("matrix")) != 3 {
		t.Fatal("downloaded db was not indexed")
	}

	// Fresh file (< 175 min): no request at all.
	if updateDB() {
		t.Fatal("update inside the freshness window")
	}
	if hits.Load() != 1 {
		t.Fatalf("server hit %d times, want 1", hits.Load())
	}

	// Stale file with identical content: downloaded, found equal, temp removed.
	old := time.Now().Add(-4 * time.Hour)
	if err := os.Chtimes(filepath.Join(dir, "rutor.ls"), old, old); err != nil {
		t.Fatal(err)
	}
	if updateDB() {
		t.Fatal("identical content must not report an update")
	}
	if hits.Load() != 2 {
		t.Fatalf("server hit %d times, want 2", hits.Load())
	}
	if _, err := os.Stat(filepath.Join(dir, "rutor.tmp")); !os.IsNotExist(err) {
		t.Error("rutor.tmp left behind after unchanged download")
	}
}

// An error page from the mirror used to be renamed over the good rutor.ls.
func TestUpdateDBKeepsGoodFileOnHTTPError(t *testing.T) {
	dir := setup(t)
	good := compress(t, sample)
	fn := filepath.Join(dir, "rutor.ls")
	if err := os.WriteFile(fn, good, 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-4 * time.Hour)
	if err := os.Chtimes(fn, old, old); err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	serve(t, http.StatusNotFound, []byte("<html>404</html>"), &hits)

	if updateDB() {
		t.Fatal("404 reported as an update")
	}
	after, err := os.ReadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, good) {
		t.Fatal("404 body replaced the good rutor.ls")
	}
	if _, err := os.Stat(filepath.Join(dir, "rutor.tmp")); !os.IsNotExist(err) {
		t.Error("rutor.tmp left behind after HTTP error")
	}
}

func TestUpdateDBUnreachable(t *testing.T) {
	dir := setup(t)
	srv := httptest.NewServer(http.NotFoundHandler())
	rutorDBURL = srv.URL
	srv.Close() // connection refused

	if updateDB() {
		t.Fatal("unreachable mirror reported an update")
	}
	if _, err := os.Stat(filepath.Join(dir, "rutor.tmp")); !os.IsNotExist(err) {
		t.Error("rutor.tmp left behind after dial error")
	}
}

func TestUpdateDBCannotCreateTmp(t *testing.T) {
	setup(t)
	settings.Path = filepath.Join(t.TempDir(), "missing", "dir")
	if updateDB() {
		t.Fatal("update without a writable data dir")
	}
}

func TestStartDisabledDoesNothing(t *testing.T) {
	setup(t)
	settings.BTsets.EnableRutorSearch = false
	base := runtime.NumGoroutine()
	Start()
	time.Sleep(50 * time.Millisecond)
	leakcheck.AssertNoGoroutineLeak(t, base)
}

func TestStartLoadsFreshDBAndStopEndsLoop(t *testing.T) {
	dir := setup(t)
	// Fresh file: updateDB returns false (no network), Start falls back to loadDB.
	if err := os.WriteFile(filepath.Join(dir, "rutor.ls"), compress(t, sample), 0o600); err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	serve(t, http.StatusInternalServerError, nil, &hits)

	base := runtime.NumGoroutine()
	Start()
	deadline := time.Now().Add(5 * time.Second)
	for len(Search("matrix")) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("Start did not load the db")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if hits.Load() != 0 {
		t.Errorf("fresh db still hit the network %d times", hits.Load())
	}

	Stop()
	if Search("matrix") != nil {
		t.Error("Stop must drop the in-memory index")
	}
	// The updater goroutine notices the generation change within its 1s tick.
	leakcheck.AssertNoGoroutineLeak(t, base)
}

// Stop right after Start, before the updater goroutine took its first step,
// must still end it: the goroutine used to reset the stop flag and run on.
func TestStopImmediatelyAfterStartEndsUpdater(t *testing.T) {
	dir := setup(t)
	if err := os.WriteFile(filepath.Join(dir, "rutor.ls"), compress(t, sample), 0o600); err != nil {
		t.Fatal(err)
	}
	serve(t, http.StatusInternalServerError, nil, new(atomic.Int32))

	for i := 0; i < 20; i++ {
		Start()
		gen.Add(1) // what Stop does under mu, without its 1.5s settle sleep
		done := make(chan struct{})
		go func() { updaters.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("updater still running after Stop")
		}
	}
}

// A load that finishes after Stop must not bring the index back.
func TestStaleLoadDoesNotRepublishIndex(t *testing.T) {
	dir := setup(t)
	if err := os.WriteFile(filepath.Join(dir, "rutor.ls"), compress(t, sample), 0o600); err != nil {
		t.Fatal(err)
	}
	g := gen.Load()
	gen.Add(1) // Stop landed while a load that began at g was decoding
	loadDBAt(g)
	if Search("matrix") != nil {
		t.Fatal("stale load republished the index")
	}
}
