package torrfs

import (
	"path/filepath"
	"testing"

	"silo/internal/config"
	"silo/internal/database"
	"silo/internal/torrent"
	"silo/internal/user"
)

// setupTorrFSTest собирает TorrFS с реальной базой, сервисом пользователей и хранилищем
func setupTorrFSTest(t *testing.T) (*TorrFS, *user.Service, *torrent.Store, *user.User) {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "torrfs_test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cfg := config.DefaultConfig()
	userSvc := user.NewService(user.NewStore(db), cfg)

	owner, err := userSvc.Authenticate("")
	if err != nil {
		t.Fatalf("authenticate owner: %v", err)
	}

	store := torrent.NewStore(db)
	fs := New(userSvc, store, nil)
	return fs, userSvc, store, owner
}

// addCategorized добавляет раздачу с готовой физической карточкой (файлы есть сразу)
func addCategorized(t *testing.T, userSvc *user.Service, store *torrent.Store, owner *user.User, hash, title, category string) {
	t.Helper()

	if err := userSvc.AddTorrent(owner, hash, title, "", category); err != nil {
		t.Fatalf("add user torrent: %v", err)
	}
	rec := &torrent.TorrentRecord{
		Hash: hash,
		Name: title,
		Files: []*torrent.TorrentFileStat{
			{Id: 0, Path: "movie.mkv", Name: "movie.mkv", Length: 1024},
		},
	}
	if err := store.Save(rec); err != nil {
		t.Fatalf("save record: %v", err)
	}
}

// rootNames возвращает имена дочерних узлов корня
func rootNames(t *testing.T, fs *TorrFS, u *user.User) []string {
	t.Helper()

	children, err := fs.Root(u).Children()
	if err != nil {
		t.Fatalf("root children: %v", err)
	}

	names := make([]string, 0, len(children))
	for _, c := range children {
		names = append(names, c.Name())
	}
	return names
}

// TestRootGroupsByCategory проверяет группировку раздач по папкам категорий
func TestRootGroupsByCategory(t *testing.T) {
	fs, userSvc, store, owner := setupTorrFSTest(t)

	addCategorized(t, userSvc, store, owner, "hash-movie", "Some Movie", torrent.CategoryMovie)
	addCategorized(t, userSvc, store, owner, "hash-tv", "Some Series", torrent.CategoryTV)

	names := rootNames(t, fs, owner)

	if len(names) != 2 {
		t.Fatalf("expected 2 category folders, got: %v", names)
	}
	// Папки сортируются по алфавиту: Movies, Series
	if names[0] != "Movies" || names[1] != "Series" {
		t.Errorf("unexpected category folders: %v", names)
	}
}

// TestRootKeepsUncategorizedAtTop проверяет, что раздачи без категории лежат прямо в корне
func TestRootKeepsUncategorizedAtTop(t *testing.T) {
	fs, userSvc, store, owner := setupTorrFSTest(t)

	addCategorized(t, userSvc, store, owner, "hash-plain", "Plain Torrent", "")
	addCategorized(t, userSvc, store, owner, "hash-movie", "Some Movie", torrent.CategoryMovie)

	names := rootNames(t, fs, owner)

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}

	if !found["Plain Torrent"] {
		t.Errorf("uncategorized torrent should be at root, got: %v", names)
	}
	if !found["Movies"] {
		t.Errorf("category folder missing, got: %v", names)
	}
	if found["Other"] {
		t.Errorf("empty category should not create an Other folder, got: %v", names)
	}
}

// TestRootWithoutCategories проверяет, что без категорий дерево плоское
func TestRootWithoutCategories(t *testing.T) {
	fs, userSvc, store, owner := setupTorrFSTest(t)

	addCategorized(t, userSvc, store, owner, "hash-one", "First", "")
	addCategorized(t, userSvc, store, owner, "hash-two", "Second", "")

	names := rootNames(t, fs, owner)

	if len(names) != 2 {
		t.Fatalf("expected 2 torrents at root, got: %v", names)
	}
	for _, n := range names {
		if n == "Other" {
			t.Errorf("no category folders expected, got: %v", names)
		}
	}
}
