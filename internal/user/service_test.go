package user

import (
	"errors"
	"path/filepath"
	"strings"
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
	// Режим БЕЗ пароля (свободный вход)
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

	// Режим С паролем
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

// TestSetRankPermissions проверяет, что ранг меняет только тот, кто выше цели:
// иначе обычный пользователь понижает администратора.
func TestSetRankPermissions(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, _ := svc.Login("owner", "secret")

	admin, err := svc.CreateUser(owner, "rank_admin", "pass", RankAdmin, Limits{})
	if err != nil {
		t.Fatalf("CreateUser admin failed: %v", err)
	}
	regular, err := svc.CreateUser(owner, "rank_user", "pass", RankUser, Limits{})
	if err != nil {
		t.Fatalf("CreateUser regular failed: %v", err)
	}

	// Пользователь не может понизить администратора
	if err := svc.SetRank(regular, admin.ID, RankGuest); err != ErrPermissionDenied {
		t.Errorf("regular demoting admin = %v, want %v", err, ErrPermissionDenied)
	}

	// Администратор не может тронуть равного себе
	second, err := svc.CreateUser(owner, "rank_admin2", "pass", RankAdmin, Limits{})
	if err != nil {
		t.Fatalf("CreateUser second admin failed: %v", err)
	}
	if err := svc.SetRank(admin, second.ID, RankUser); err != ErrPermissionDenied {
		t.Errorf("admin demoting equal admin = %v, want %v", err, ErrPermissionDenied)
	}

	// Владелец понижает администратора
	if err := svc.SetRank(owner, admin.ID, RankUser); err != nil {
		t.Errorf("owner demoting admin failed: %v", err)
	}

	// Себе ранг понизить можно
	if err := svc.SetRank(admin, admin.ID, RankGuest); err != nil {
		t.Errorf("self demotion failed: %v", err)
	}

	// Повысить себя нельзя
	if err := svc.SetRank(regular, regular.ID, RankOwner); err != ErrPermissionDenied {
		t.Errorf("self promotion = %v, want %v", err, ErrPermissionDenied)
	}

	// Ранг Owner неизменяем
	if err := svc.SetRank(owner, owner.ID, RankUser); err != ErrOwnerProtected {
		t.Errorf("owner rank change = %v, want %v", err, ErrOwnerProtected)
	}
}

func TestUserTorrentsAndViewedFiles(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, _ := svc.Login("owner", "secret")
	u1, _ := svc.CreateUser(owner, "viewer1", "pass", RankUser, Limits{})
	u2, _ := svc.CreateUser(owner, "viewer2", "pass", RankUser, Limits{})

	hash := "abcd1234efgh"

	// Пользователь u1 добавляет раздачу со своим постером и названием (например, из Lampa)
	err := svc.AddTorrent(u1, hash, "[Lampa] Interstellar 4K", "http://lampa.jpg", "Movies")
	if err != nil {
		t.Fatalf("Failed to add torrent for u1: %v", err)
	}

	// Пользователь u2 добавляет ТОТ ЖЕ САМЫЙ торрент со своим названием и постером (например, из NUM)
	err = svc.AddTorrent(u2, hash, "[NUM] Dune & Interstellar", "http://num.jpg", "Sci-Fi")
	if err != nil {
		t.Fatalf("Failed to add torrent for u2: %v", err)
	}

	// Изоляция метаданных через GetUserTorrent
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

	// Пользователь u1 смотрит серию #0 и #2
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

	// Удаление торрента и счетчик ссылок;
	// u1 удаляет торрент - раздача должна остаться у u2
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

	// u2 удаляет торрент - теперь раздачу больше никто не держит
	err = svc.RemoveTorrent(u2, hash)
	if err != nil {
		t.Fatalf("Failed to remove torrent for u2: %v", err)
	}
}

// Проверяет, что после регенерации токена Owner логин отдаёт живой токен,
// а не копию из памяти.
func TestOwnerTokenStaysInSyncAfterRegenerate(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "supersecret")
	defer cleanup()

	_, oldToken, err := svc.Login("owner", "supersecret")
	if err != nil {
		t.Fatalf("Initial login failed: %v", err)
	}

	owner, err := svc.GetUserByID("owner")
	if err != nil {
		t.Fatalf("Failed to get owner: %v", err)
	}

	newToken, err := svc.RegenerateToken(owner, "owner")
	if err != nil {
		t.Fatalf("Failed to regenerate owner token: %v", err)
	}
	if newToken == oldToken {
		t.Fatalf("Token was not actually regenerated")
	}

	// Токен, который вернул RegenerateToken, должен работать сразу.
	if _, err := svc.Authenticate(newToken); err != nil {
		t.Fatalf("Freshly regenerated token is not authenticatable: %v", err)
	}

	// Логин обязан вернуть тот же новый токен, а не устаревшую копию.
	_, loginToken, err := svc.Login("owner", "supersecret")
	if err != nil {
		t.Fatalf("Login after regenerate failed: %v", err)
	}
	if loginToken != newToken {
		t.Fatalf("Login returned stale token:\n  got  %q\n  want %q", loginToken, newToken)
	}

	// И выданный логином токен тоже должен проходить аутентификацию.
	if _, err := svc.Authenticate(loginToken); err != nil {
		t.Fatalf("Token handed out by login is not authenticatable: %v", err)
	}

	// Старый токен после смены работать не должен.
	if _, err := svc.Authenticate(oldToken); err == nil {
		t.Errorf("Old token must stop working after regeneration")
	}
}

// Режим без пароля: сервер открыт, подходит пустой или произвольный токен.
func TestOpenModeAcceptsAnyToken(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "")
	defer cleanup()

	for _, token := range []string{"", "garbage", "ts_deadbeef"} {
		u, err := svc.Authenticate(token)
		if err != nil {
			t.Fatalf("Open mode must accept token %q, got error: %v", token, err)
		}
		if !u.Rank.IsOwner() {
			t.Errorf("Token %q: expected Owner rank, got %d", token, u.Rank)
		}
	}

	if _, token, err := svc.Login("owner", ""); err != nil {
		t.Fatalf("Open mode login failed: %v", err)
	} else if _, err := svc.Authenticate(token); err != nil {
		t.Fatalf("Open mode: token from login was rejected: %v", err)
	}
}

// Owner защищён от бана: и админом, и самим собой.
func TestOwnerCannotBeBanned(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, err := svc.Login("owner", "secret")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if err := svc.BanUser(owner, owner.ID, true); !errors.Is(err, ErrOwnerProtected) {
		t.Errorf("Owner must not be able to ban himself, got: %v", err)
	}

	admin, err := svc.CreateUser(owner, "admin_x", "pass", RankAdmin, Limits{})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}
	if err := svc.BanUser(admin, owner.ID, true); !errors.Is(err, ErrOwnerProtected) {
		t.Errorf("Admin must not be able to ban Owner, got: %v", err)
	}
}

// ChangePassword: пользователь может сменить пароль себе, вышестоящий - подчинённому.
func TestChangePassword(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, err := svc.Login("owner", "secret")
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}

	bob, err := svc.CreateUser(owner, "bob", "oldpass", RankUser, Limits{})
	if err != nil {
		t.Fatalf("create bob: %v", err)
	}

	admin, err := svc.CreateUser(owner, "carol", "adminpass", RankAdmin, Limits{})
	if err != nil {
		t.Fatalf("create carol: %v", err)
	}

	// Старый пароль работает, новый - ещё нет.
	if _, _, err := svc.Login("bob", "oldpass"); err != nil {
		t.Fatalf("login with old password failed: %v", err)
	}
	if _, _, err := svc.Login("bob", "newpass"); err == nil {
		t.Errorf("new password must not work before change")
	}

	// Админ меняет пароль подчинённому.
	if err := svc.ChangePassword(admin, bob.ID, "newpass"); err != nil {
		t.Fatalf("admin failed to change password: %v", err)
	}

	if _, _, err := svc.Login("bob", "newpass"); err != nil {
		t.Errorf("login with new password failed: %v", err)
	}
	if _, _, err := svc.Login("bob", "oldpass"); err == nil {
		t.Errorf("old password must stop working")
	}

	// Пользователь меняет пароль себе.
	if err := svc.ChangePassword(bob, bob.ID, "selfpass"); err != nil {
		t.Fatalf("self change failed: %v", err)
	}
	if _, _, err := svc.Login("bob", "selfpass"); err != nil {
		t.Errorf("login after self change failed: %v", err)
	}

	// Пользователь не может менять пароль другому.
	if err := svc.ChangePassword(bob, admin.ID, "hacked"); !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("user must not change another password, got: %v", err)
	}

	// Пароль Owner задаётся в конфиге, менять его нельзя.
	if err := svc.ChangePassword(owner, "owner", "whatever"); !errors.Is(err, ErrOwnerProtected) {
		t.Errorf("owner password must be protected, got: %v", err)
	}

	// Пустой пароль отклоняется.
	if err := svc.ChangePassword(admin, bob.ID, ""); !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("empty password must be rejected, got: %v", err)
	}
}

// Пароль длиннее 72 байт bcrypt не принимает: это ошибка ввода, а не сбой сервиса.
func TestChangePasswordTooLong(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, err := svc.Login("owner", "secret")
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}
	bob, err := svc.CreateUser(owner, "bob", "pass", RankUser, Limits{})
	if err != nil {
		t.Fatalf("create bob: %v", err)
	}

	long := strings.Repeat("x", 73)
	if err := svc.ChangePassword(owner, bob.ID, long); !errors.Is(err, ErrPasswordTooLong) {
		t.Errorf("expected ErrPasswordTooLong, got: %v", err)
	}

	// Пароль в пределах лимита по-прежнему принимается.
	ok := strings.Repeat("y", 72)
	if err := svc.ChangePassword(owner, bob.ID, ok); err != nil {
		t.Errorf("72-byte password must be accepted, got: %v", err)
	}
}

// TestUserTorrentSetAllFilesViewed проверяет массовую отметку на уровне модели:
// повторный вызов не должен терять уже отмеченные файлы.
func TestUserTorrentSetAllFilesViewed(t *testing.T) {
	ut := &UserTorrent{ViewedFiles: []int{5}}

	// Отмечаем файлы 0..2: индекс 5 не входит в список, но должен сохраниться
	ut.SetAllFilesViewed([]int{0, 1, 2}, true)

	want := map[int]bool{0: true, 1: true, 2: true, 5: true}
	if len(ut.ViewedFiles) != len(want) {
		t.Fatalf("ViewedFiles = %v, want %d элементов", ut.ViewedFiles, len(want))
	}
	for _, idx := range ut.ViewedFiles {
		if !want[idx] {
			t.Errorf("лишний индекс %d в %v", idx, ut.ViewedFiles)
		}
	}

	// Повторный вызов не дублирует и не теряет
	before := len(ut.ViewedFiles)
	ut.SetAllFilesViewed([]int{0, 1, 2}, true)
	if len(ut.ViewedFiles) != before {
		t.Errorf("повторный вызов изменил список: %v", ut.ViewedFiles)
	}
	if !ut.IsFileViewed(5) {
		t.Error("индекс 5 потерян при повторном вызове")
	}

	// Снятие очищает всё
	ut.SetAllFilesViewed([]int{0, 1, 2}, false)
	if len(ut.ViewedFiles) != 0 {
		t.Errorf("после снятия ViewedFiles = %v, want пусто", ut.ViewedFiles)
	}
}

func TestUserTorrentFileProgress(t *testing.T) {
	ut := &UserTorrent{}

	ut.SetFileProgress(3, 125.5, 2700)
	if !ut.IsFileViewed(3) {
		t.Error("file 3 must be viewed after saving progress")
	}
	if got := ut.Progress[3]; got.Position != 125.5 || got.Duration != 2700 {
		t.Errorf("Progress[3] = %+v, want {125.5 2700}", got)
	}

	ut.SetFileProgress(3, 130, 0)
	if got := ut.Progress[3]; got.Position != 130 || got.Duration != 2700 {
		t.Errorf("Progress[3] = %+v, want {130 2700}", got)
	}

	ut.SetFileProgress(3, 0, 0)
	if _, ok := ut.Progress[3]; ok {
		t.Error("file 3 progress must be cleared")
	}
	if !ut.IsFileViewed(3) {
		t.Error("file 3 must stay viewed")
	}

	ut.SetFileProgress(1, 10, 0)
	ut.UnmarkFileViewed(1)
	if _, ok := ut.Progress[1]; ok {
		t.Error("file 1 progress must be dropped with the viewed mark")
	}

	ut.SetFileProgress(2, 20, 0)
	ut.SetAllFilesViewed([]int{1, 2, 3}, false)
	if len(ut.Progress) != 0 {
		t.Errorf("Progress = %v after unmarking all, want empty", ut.Progress)
	}
}

func TestSetFileProgressPersists(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "secret")
	defer cleanup()

	owner, _, _ := svc.Login("owner", "secret")
	u1, _ := svc.CreateUser(owner, "viewer1", "pass", RankUser, Limits{})
	u2, _ := svc.CreateUser(owner, "viewer2", "pass", RankUser, Limits{})

	hash := "abcd1234efgh"
	if err := svc.AddTorrent(u1, hash, "Movie", "", ""); err != nil {
		t.Fatalf("AddTorrent u1: %v", err)
	}
	if err := svc.AddTorrent(u2, hash, "Movie", "", ""); err != nil {
		t.Fatalf("AddTorrent u2: %v", err)
	}

	svc.trackProgress.Store(true)
	tracked, err := svc.SetFileProgress(u1, hash, 0, 61, 3600)
	if err != nil {
		t.Fatalf("SetFileProgress: %v", err)
	}
	if !tracked {
		t.Error("tracked = false with tracking on")
	}

	ut1, _ := svc.GetUserTorrent(u1.ID, hash)
	if got := ut1.Progress[0]; got.Position != 61 || got.Duration != 3600 {
		t.Errorf("u1 Progress[0] = %+v, want {61 3600}", got)
	}
	if !ut1.IsFileViewed(0) {
		t.Error("file 0 must be viewed for u1")
	}

	ut2, _ := svc.GetUserTorrent(u2.ID, hash)
	if len(ut2.Progress) != 0 || len(ut2.ViewedFiles) != 0 {
		t.Errorf("u2 must have no progress: %+v", ut2)
	}

	if _, err := svc.SetFileProgress(u1, hash, 0, -1, 0); !errors.Is(err, ErrInvalidProgress) {
		t.Errorf("negative position: err = %v, want ErrInvalidProgress", err)
	}
}

func TestSetFileProgressDisabled(t *testing.T) {
	svc, cleanup := setupTestEnv(t, "")
	defer cleanup()

	owner, _ := svc.Authenticate("")
	hash := "abcd1234efgh"
	if err := svc.AddTorrent(owner, hash, "Movie", "", ""); err != nil {
		t.Fatalf("AddTorrent: %v", err)
	}

	svc.trackProgress.Store(true)
	if _, err := svc.SetFileProgress(owner, hash, 0, 30, 0); err != nil {
		t.Fatalf("SetFileProgress: %v", err)
	}

	svc.trackProgress.Store(false)
	tracked, err := svc.SetFileProgress(owner, hash, 0, 90, 0)
	if err != nil {
		t.Fatalf("SetFileProgress: %v", err)
	}
	if tracked {
		t.Error("tracked = true with tracking off")
	}

	ut, _ := svc.GetUserTorrent(owner.ID, hash)
	if got := ut.Progress[0].Position; got != 30 {
		t.Errorf("Position = %v, want 30", got)
	}

	if _, err := svc.SetFileProgress(owner, hash, 1, 15, 0); err != nil {
		t.Fatalf("SetFileProgress: %v", err)
	}
	ut, _ = svc.GetUserTorrent(owner.ID, hash)
	if !ut.IsFileViewed(1) {
		t.Error("file 1 must be viewed with tracking off")
	}
	if _, ok := ut.Progress[1]; ok {
		t.Error("file 1 position must not be stored")
	}
}

func TestSetProgressTracking(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "progress_setting.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	cfg := config.DefaultConfig()
	cfg.Auth.OwnerPassword = "secret"
	svc := NewService(NewStore(db), cfg)

	if svc.ProgressTracking() {
		t.Fatal("tracking must be off by default")
	}

	owner, _, _ := svc.Login("owner", "secret")
	admin, _ := svc.CreateUser(owner, "admin", "pass", RankAdmin, Limits{})

	if err := svc.SetProgressTracking(admin, true); !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("admin: err = %v, want ErrPermissionDenied", err)
	}
	if err := svc.SetProgressTracking(owner, true); err != nil {
		t.Fatalf("SetProgressTracking: %v", err)
	}
	if !svc.ProgressTracking() {
		t.Error("toggle must apply immediately")
	}

	if !NewService(NewStore(db), cfg).ProgressTracking() {
		t.Error("saved value must override config.yaml")
	}

	if err := svc.SetProgressTracking(owner, false); err != nil {
		t.Fatalf("SetProgressTracking: %v", err)
	}
	cfg.Library.TrackProgress = true
	if NewService(NewStore(db), cfg).ProgressTracking() {
		t.Error("saved false must override track_progress: true")
	}
}

func TestUserTorrentSetFileDuration(t *testing.T) {
	ut := &UserTorrent{}

	ut.SetFileDuration(0, 100)
	if len(ut.Progress) != 0 {
		t.Errorf("Progress = %v, want empty", ut.Progress)
	}

	ut.SetFileProgress(0, 10, 0)
	ut.SetFileDuration(0, 1500)
	if got := ut.Progress[0]; got.Position != 10 || got.Duration != 1500 {
		t.Errorf("Progress[0] = %+v, want {10 1500}", got)
	}

	ut.SetFileProgress(1, 20, 600)
	ut.SetFileDuration(1, 700)
	if got := ut.Progress[1].Duration; got != 600 {
		t.Errorf("Duration = %v, want 600 from the player", got)
	}
}
