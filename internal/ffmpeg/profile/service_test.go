package profile

import (
	"errors"
	"testing"
)

func TestServiceListFiltersByAccess(t *testing.T) {
	svc := newTestService(t)

	owner := testUser("owner", RankOwner)
	admin := testUser("admin", RankAdmin)
	regular := testUser("user", RankUser)

	// Приватный профиль админа
	private := Default()
	private.ID = ""
	private.Name = "Private phone"
	private.OwnerID = admin.ID
	private.Visibility = VisibilityPrivate
	private.IsDefault = false
	if err := svc.store.Save(&private); err != nil {
		t.Fatalf("Save private: %v", err)
	}

	// Публичный профиль для админов и выше
	public := Default()
	public.ID = ""
	public.Name = "Public TV"
	public.OwnerID = admin.ID
	public.Visibility = VisibilityPublic
	public.RankRequired = int(RankAdmin)
	public.IsDefault = false
	if err := svc.store.Save(&public); err != nil {
		t.Fatalf("Save public: %v", err)
	}

	list, err := svc.List(regular)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("regular user sees %d profiles, want 0", len(list))
	}

	list, err = svc.List(admin)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("admin sees %d profiles, want 2", len(list))
	}

	list, err = svc.List(owner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("owner sees %d profiles, want only the public one", len(list))
	}
}

func TestServiceListRejectsMissingActor(t *testing.T) {
	svc := newTestService(t)

	// Пустой Actor не имеет прав: сервис вернёт только доступное ему, то есть ничего
	list, err := svc.List(Actor{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("anonymous actor sees %d profiles, want 0", len(list))
	}
}

func TestServiceCreateRequiresAdmin(t *testing.T) {
	svc := newTestService(t)

	regular := testUser("user", RankUser)
	p := Default()
	p.Name = "Mine"

	if _, err := svc.Create(regular, &p); !errors.Is(err, ErrPermission) {
		t.Errorf("error = %v, want %v", err, ErrPermission)
	}
}

func TestServiceCreateSetsOwner(t *testing.T) {
	svc := newTestService(t)

	admin := testUser("admin", RankAdmin)
	p := Default()
	p.Name = "My profile"
	p.IsDefault = false

	created, err := svc.Create(admin, &p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.OwnerID != admin.ID {
		t.Errorf("owner = %q, want %q", created.OwnerID, admin.ID)
	}
	if created.ID == "" || created.ID == DefaultID {
		t.Errorf("id = %q, want a generated one", created.ID)
	}
	if created.CreatedAt.IsZero() {
		t.Error("created_at was not set")
	}
}

func TestServiceCreateDefaultRequiresOwner(t *testing.T) {
	svc := newTestService(t)

	admin := testUser("admin", RankAdmin)
	p := Default()
	p.Name = "Default attempt"
	p.IsDefault = true

	if _, err := svc.Create(admin, &p); !errors.Is(err, ErrPermission) {
		t.Errorf("error = %v, want %v for an admin", err, ErrPermission)
	}

	owner := testUser("owner", RankOwner)
	p.Name = "Owner default"
	created, err := svc.Create(owner, &p)
	if err != nil {
		t.Fatalf("Create by owner: %v", err)
	}
	if !created.IsDefault {
		t.Error("profile was not marked as default")
	}
}

func TestServiceUpdateOwnProfile(t *testing.T) {
	svc := newTestService(t)

	admin := testUser("admin", RankAdmin)
	p := Default()
	p.Name = "Before"
	p.IsDefault = false

	created, err := svc.Create(admin, &p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	patch := *created
	patch.Name = "After"
	patch.Video.MaxHeight = 720

	updated, err := svc.Update(admin, created.ID, &patch)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "After" {
		t.Errorf("name = %q, want After", updated.Name)
	}
	if updated.Video.MaxHeight != 720 {
		t.Errorf("max height = %d, want 720", updated.Video.MaxHeight)
	}
	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Error("created_at must be preserved")
	}
}

func TestServiceUpdateOtherProfileDenied(t *testing.T) {
	svc := newTestService(t)

	first := testUser("admin1", RankAdmin)
	second := testUser("admin2", RankAdmin)

	p := Default()
	p.Name = "First profile"
	p.IsDefault = false

	created, err := svc.Create(first, &p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	patch := *created
	patch.Name = "Hijacked"

	if _, err := svc.Update(second, created.ID, &patch); !errors.Is(err, ErrPermission) {
		t.Errorf("error = %v, want %v", err, ErrPermission)
	}
}

func TestServiceUpdateByOwnerAllowed(t *testing.T) {
	svc := newTestService(t)

	admin := testUser("admin", RankAdmin)
	owner := testUser("owner", RankOwner)

	p := Default()
	p.Name = "Admin profile"
	p.IsDefault = false

	created, err := svc.Create(admin, &p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	patch := *created
	patch.Name = "Renamed by owner"

	updated, err := svc.Update(owner, created.ID, &patch)
	if err != nil {
		t.Fatalf("Update by owner: %v", err)
	}
	if updated.Name != "Renamed by owner" {
		t.Errorf("name = %q, want Renamed by owner", updated.Name)
	}
}

func TestServiceDelete(t *testing.T) {
	svc := newTestService(t)

	admin := testUser("admin", RankAdmin)
	other := testUser("other", RankAdmin)

	p := Default()
	p.Name = "Temp"
	p.IsDefault = false

	created, err := svc.Create(admin, &p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(other, created.ID); !errors.Is(err, ErrPermission) {
		t.Errorf("error = %v, want %v", err, ErrPermission)
	}

	if err := svc.Delete(admin, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := svc.store.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("profile still exists: %v", err)
	}
}

func TestServiceDeleteBuiltin(t *testing.T) {
	svc := newTestService(t)

	owner := testUser("owner", RankOwner)
	if err := svc.Delete(owner, DefaultID); !errors.Is(err, ErrBuiltin) {
		t.Errorf("error = %v, want %v", err, ErrBuiltin)
	}
}

func TestServiceSetDefaultRequiresOwner(t *testing.T) {
	svc := newTestService(t)

	admin := testUser("admin", RankAdmin)
	owner := testUser("owner", RankOwner)

	p := Default()
	p.Name = "Candidate"
	p.IsDefault = false

	created, err := svc.Create(admin, &p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := svc.SetDefault(admin, created.ID); !errors.Is(err, ErrPermission) {
		t.Errorf("error = %v, want %v for an admin", err, ErrPermission)
	}

	updated, err := svc.SetDefault(owner, created.ID)
	if err != nil {
		t.Fatalf("SetDefault: %v", err)
	}
	if !updated.IsDefault {
		t.Error("profile was not marked as default")
	}

	// Второй профиль становится единственным по умолчанию
	second, err := svc.Create(admin, &Profile{
		Name:      "Second",
		Protocol:  "progressive",
		Container: "mp4",
		Video:     Default().Video,
		Audio:     Default().Audio,
	})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}
	if _, err := svc.SetDefault(owner, second.ID); err != nil {
		t.Fatalf("SetDefault second: %v", err)
	}

	defaults := 0
	all, _ := svc.store.List()
	for _, item := range all {
		if item.IsDefault {
			defaults++
		}
	}
	if defaults != 1 {
		t.Errorf("default profiles = %d, want 1", defaults)
	}
}

func TestServiceResolveProfile(t *testing.T) {
	svc := newTestService(t)

	regular := testUser("user", RankUser)

	// Пустой идентификатор даёт встроенный профиль
	p, err := svc.ResolveProfile(regular, "")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}
	if p.ID != DefaultID {
		t.Errorf("id = %q, want %q", p.ID, DefaultID)
	}

	// Явный идентификатор встроенного профиля тоже работает
	if _, err := svc.ResolveProfile(regular, DefaultID); err != nil {
		t.Fatalf("ResolveProfile builtin: %v", err)
	}

	// Неизвестный идентификатор даёт ошибку
	if _, err := svc.ResolveProfile(regular, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want %v", err, ErrNotFound)
	}
}

func TestServiceResolveProfileUsesStoredDefault(t *testing.T) {
	svc := newTestService(t)

	owner := testUser("owner", RankOwner)
	regular := testUser("user", RankUser)

	p := Default()
	p.Name = "Global default"
	p.Visibility = VisibilityPublic
	p.IsDefault = true

	created, err := svc.Create(owner, &p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	resolved, err := svc.ResolveProfile(regular, "")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}
	if resolved.ID != created.ID {
		t.Errorf("id = %q, want the stored default %q", resolved.ID, created.ID)
	}
}

func TestServiceGetDeniedForForeignPrivate(t *testing.T) {
	svc := newTestService(t)

	admin := testUser("admin", RankAdmin)
	regular := testUser("user", RankUser)

	p := Default()
	p.Name = "Secret"
	p.Visibility = VisibilityPrivate
	p.IsDefault = false

	created, err := svc.Create(admin, &p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := svc.Get(regular, created.ID); !errors.Is(err, ErrPermission) {
		t.Errorf("error = %v, want %v", err, ErrPermission)
	}
	if _, err := svc.Get(admin, created.ID); err != nil {
		t.Errorf("owner must read own profile: %v", err)
	}
}

func TestValidateForActor(t *testing.T) {
	regular := testUser("user", RankUser)

	p := Default()
	p.IsDefault = false
	if err := ValidateForActor(regular, &p); err != nil {
		t.Fatalf("ValidateForActor: %v", err)
	}

	p.RankRequired = int(RankOwner)
	if err := ValidateForActor(regular, &p); !errors.Is(err, ErrPermission) {
		t.Errorf("error = %v, want %v", err, ErrPermission)
	}
}

// Ранги повторяют значения пакета пользователей, чтобы не тянуть его зависимость.
const (
	RankUser  = 10
	RankAdmin = 50
	RankOwner = 100
)

// testUser создаёт права пользователя с указанным рангом.
func testUser(id string, rank int) Actor {
	return Actor{
		ID:    id,
		Rank:  rank,
		Admin: rank >= RankAdmin,
		Owner: rank >= RankOwner,
	}
}
