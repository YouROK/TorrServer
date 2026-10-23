package user

import (
	"sync"
	"testing"
)

// TestSetFileViewedConcurrent проверяет, что параллельные отметки просмотра
// не теряются: правка идёт в одной транзакции с чтением.
func TestSetFileViewedConcurrent(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, _ := svc.Login("owner", "secret")
	u, err := svc.CreateUser(owner, "viewer_race", "pass", RankUser, Limits{})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	const hash = "0123456789abcdef0123456789abcdef01234567"
	if err := svc.AddTorrent(u, hash, "Test", "", ""); err != nil {
		t.Fatalf("AddTorrent: %v", err)
	}

	const total = 20
	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := svc.SetFileViewed(u, hash, idx, true); err != nil {
				t.Errorf("SetFileViewed(%d): %v", idx, err)
			}
		}(i)
	}
	wg.Wait()

	ut, err := svc.GetUserTorrent(u.ID, hash)
	if err != nil {
		t.Fatalf("GetUserTorrent: %v", err)
	}
	if len(ut.ViewedFiles) != total {
		t.Errorf("saved %d of %d marks: %v", len(ut.ViewedFiles), total, ut.ViewedFiles)
	}
}

// TestSetFileViewedKeepsMetaConcurrent проверяет, что правка карточки
// и отметка просмотра не затирают друг друга.
func TestSetFileViewedKeepsMetaConcurrent(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, _ := svc.Login("owner", "secret")
	u, err := svc.CreateUser(owner, "viewer_meta", "pass", RankUser, Limits{})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	const hash = "abcdefabcdefabcdefabcdefabcdefabcdefabcd"
	if err := svc.AddTorrent(u, hash, "Original", "", ""); err != nil {
		t.Fatalf("AddTorrent: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = svc.UpdateTorrentMeta(u, hash, "Renamed", "poster.jpg", "movies")
	}()
	go func() {
		defer wg.Done()
		_ = svc.SetFileViewed(u, hash, 1, true)
	}()
	wg.Wait()

	ut, err := svc.GetUserTorrent(u.ID, hash)
	if err != nil {
		t.Fatalf("GetUserTorrent: %v", err)
	}
	if ut.Title != "Renamed" {
		t.Errorf("title = %q, want Renamed", ut.Title)
	}
	if !ut.IsFileViewed(1) {
		t.Error("view mark was lost after meta update")
	}
}

// TestMaxTorrentsConcurrent проверяет, что лимит торрентов нельзя обойти
// параллельными запросами.
func TestMaxTorrentsConcurrent(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, _ := svc.Login("owner", "secret")
	u, err := svc.CreateUser(owner, "viewer_limit", "pass", RankUser, Limits{MaxTorrents: 3})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	const attempts = 12
	var wg sync.WaitGroup
	var added int64
	var mu sync.Mutex

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			hash := hashForIndex(n)
			if err := svc.AddTorrent(u, hash, "T", "", ""); err == nil {
				mu.Lock()
				added++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if added > 3 {
		t.Errorf("limit bypassed: added %d torrents with max 3", added)
	}

	count, err := svc.store.CountUserTorrents(u.ID)
	if err != nil {
		t.Fatalf("CountUserTorrents: %v", err)
	}
	if count > 3 {
		t.Errorf("stored %d torrents, limit is 3", count)
	}
}

// hashForIndex строит уникальный info-hash для проверки лимита.
func hashForIndex(n int) string {
	const hexDigits = "0123456789abcdef"
	out := make([]byte, 40)
	for i := range out {
		out[i] = hexDigits[(n+i)%len(hexDigits)]
	}
	return string(out)
}
