package user

import (
	"path/filepath"
	"testing"

	"silo/internal/config"
	"silo/internal/database"
)

// setupTestEnv создает изолированную тестовую базу данных во временной папке
func setupTestEnv(t *testing.T, ownerPassword string) (*Service, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_user_vault.db")

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Auth.OwnerPassword = ownerPassword

	store := NewStore(db)
	svc := NewService(store, cfg)

	cleanup := func() {
		_ = db.Close()
	}

	return svc, cleanup
}

func TestOwnerBootstrap(t *testing.T) {
	// 1. Тестируем режим БЕЗ пароля (свободный вход)
	svcNoAuth, cleanup := setupTestEnv(t, "")
	defer cleanup()

	// В режиме без пароля пустой токен возвращает Owner
	u, err := svcNoAuth.Authenticate("")
	if err != nil {
		t.Fatalf("Expected successful auth without password, got: %v", err)
	}
	if !u.Rank.IsOwner() {
		t.Errorf("Expected user to be Owner, got rank: %d", u.Rank)
	}

	// 2. Тестируем режим С паролем
	svcWithAuth, cleanupAuth := setupTestEnv(t, "supersecret")
	defer cleanupAuth()

	// Пустой токен отбивается ошибкой
	_, err = svcWithAuth.Authenticate("")
	if err == nil {
		t.Errorf("Expected auth failure with empty token when password is set")
	}

	// Логин с неверным паролем
	_, _, err = svcWithAuth.Login("owner", "wrongpass")
	if err == nil {
		t.Errorf("Expected login failure with wrong password")
	}

	// Логин с верным паролем
	ownerUser, token, err := svcWithAuth.Login("owner", "supersecret")
	if err != nil {
		t.Fatalf("Failed to login as owner: %v", err)
	}

	// Проверяем авторизацию по токену
	authU, err := svcWithAuth.Authenticate(token)
	if err != nil {
		t.Fatalf("Failed to authenticate with owner token: %v", err)
	}
	if authU.ID != ownerUser.ID {
		t.Errorf("Authenticated user ID mismatch: got %s, want %s", authU.ID, ownerUser.ID)
	}

	// Проверяем метод GetUserByID для Owner
	ownerByID, err := svcWithAuth.GetUserByID("owner")
	if err != nil || ownerByID.ID != "owner" {
		t.Errorf("GetUserByID('owner') failed: %v", err)
	}
}

func TestRankHierarchy(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, _ := svc.Login("owner", "secret")

	// Owner создает двух админов (Ранг 50)
	admin1, err := svc.CreateUser(owner, "admin1", "pass1", RankAdmin, Limits{})
	if err != nil {
		t.Fatalf("Failed to create admin1: %v", err)
	}
	admin2, err := svc.CreateUser(owner, "admin2", "pass2", RankAdmin, Limits{})
	if err != nil {
		t.Fatalf("Failed to create admin2: %v", err)
	}

	// Admin1 создает обычного пользователя (Ранг 10)
	user1, err := svc.CreateUser(admin1, "user1", "pass1", RankUser, Limits{})
	if err != nil {
		t.Fatalf("Failed to create user1 by admin1: %v", err)
	}

	// ПРОВЕРКА 1: Admin1 НЕ МОЖЕТ создать пользователя с рангом Админ или выше!
	_, err = svc.CreateUser(admin1, "hacker_admin", "pass", RankAdmin, Limits{})
	if err == nil {
		t.Errorf("Admin should NOT be allowed to create another Admin")
	}

	// ПРОВЕРКА 2: Admin1 НЕ МОЖЕТ забанить Admin2 (равный ранг 50 == 50)!
	err = svc.BanUser(admin1, admin2.ID, true)
	if err == nil {
		t.Errorf("Admin1 should NOT be allowed to ban Admin2")
	}

	// ПРОВЕРКА 3: Admin1 МОЖЕТ забанить User1 (ранг 50 > 10)
	err = svc.BanUser(admin1, user1.ID, true)
	if err != nil {
		t.Errorf("Admin1 should be able to ban User1, got: %v", err)
	}

	// Забаненный User1 больше не может пройти аутентификацию
	_, err = svc.Authenticate(user1.APIToken)
	if err == nil {
		t.Errorf("Banned user must not be authenticated")
	}

	// ПРОВЕРКА 4: Никто не может забанить Owner
	err = svc.BanUser(admin1, owner.ID, true)
	if err == nil {
		t.Errorf("Admin must NOT be allowed to ban Owner")
	}

	// ПРОВЕРКА 5: Проверяем ListUsers
	usersList, err := svc.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	// Должно быть как минимум 3 созданных пользователя + owner
	if len(usersList) < 3 {
		t.Errorf("Expected at least 3 users in ListUsers, got: %d", len(usersList))
	}
}

func TestUserTorrentsAndViewedFiles(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, _ := svc.Login("owner", "secret")
	u1, _ := svc.CreateUser(owner, "viewer1", "pass", RankUser, Limits{})
	u2, _ := svc.CreateUser(owner, "viewer2", "pass", RankUser, Limits{})

	hash := "abcd1234efgh"

	// 1. Пользователь u1 добавляет раздачу со своим постером и названием (например, из Lampa)
	err := svc.AddTorrent(u1, hash, "[Lampa] Interstellar 4K", "http://lampa.jpg", "Movies")
	if err != nil {
		t.Fatalf("Failed to add torrent for u1: %v", err)
	}

	// 2. Пользователь u2 добавляет ТОТ ЖЕ САМЫЙ торрент со своим названием и постером (например, из NUM)
	err = svc.AddTorrent(u2, hash, "[NUM] Dune & Interstellar", "http://num.jpg", "Sci-Fi")
	if err != nil {
		t.Fatalf("Failed to add torrent for u2: %v", err)
	}

	// 3. Проверяем изоляцию метаданных через GetUserTorrent
	ut1, err := svc.GetUserTorrent(u1.ID, hash)
	if err != nil {
		t.Fatalf("GetUserTorrent failed for u1: %v", err)
	}
	if ut1.Title != "[Lampa] Interstellar 4K" || ut1.Poster != "http://lampa.jpg" {
		t.Errorf("Metadata mismatch for u1: %+v", ut1)
	}

	ut2, err := svc.GetUserTorrent(u2.ID, hash)
	if err != nil {
		t.Fatalf("GetUserTorrent failed for u2: %v", err)
	}
	if ut2.Title != "[NUM] Dune & Interstellar" || ut2.Poster != "http://num.jpg" {
		t.Errorf("Metadata mismatch for u2: %+v", ut2)
	}

	// 4. Пользователь u1 смотрит серию #0 и #2
	_ = svc.SetFileViewed(u1, hash, 0, true)
	_ = svc.SetFileViewed(u1, hash, 2, true)

	// Проверяем список торрентов u1
	list1, _ := svc.ListTorrents(u1)
	if len(list1) != 1 {
		t.Fatalf("Expected 1 torrent for u1, got %d", len(list1))
	}
	if !list1[0].IsFileViewed(0) || !list1[0].IsFileViewed(2) || list1[0].IsFileViewed(1) {
		t.Errorf("Viewed files mismatch for u1: got %v", list1[0].ViewedFiles)
	}

	// У пользователя u2 список просмотренных должен быть ПУСТЫМ (изоляция просмотров!)
	list2, _ := svc.ListTorrents(u2)
	if len(list2[0].ViewedFiles) != 0 {
		t.Errorf("Expected 0 viewed files for u2, got %v", list2[0].ViewedFiles)
	}

	// 5. Удаление торрента и счетчик ссылок
	// u1 удаляет торрент — раздача должна остаться у u2
	err = svc.RemoveTorrent(u1, hash)
	if err != nil {
		t.Fatalf("Failed to remove torrent for u1: %v", err)
	}

	// Проверяем, что у u1 торрента больше нет, а у u2 остался
	_, err = svc.GetUserTorrent(u1.ID, hash)
	if err == nil {
		t.Errorf("Torrent should have been removed for u1")
	}

	_, err = svc.GetUserTorrent(u2.ID, hash)
	if err != nil {
		t.Errorf("Torrent should still exist for u2: %v", err)
	}

	// u2 удаляет торрент — теперь раздачу больше никто не держит
	err = svc.RemoveTorrent(u2, hash)
	if err != nil {
		t.Fatalf("Failed to remove torrent for u2: %v", err)
	}
}
