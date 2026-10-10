package profile

import (
	"errors"
	"path/filepath"
	"testing"

	"silo/internal/database"
)

// newTestStore создаёт хранилище профилей на временной базе.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "profile_test.db"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return NewStore(db)
}

// newTestService создаёт сервис профилей на временной базе.
func newTestService(t *testing.T) *Service {
	t.Helper()
	return NewService(newTestStore(t))
}

func TestStoreSaveAndGet(t *testing.T) {
	store := newTestStore(t)

	p := Default()
	p.ID = ""
	p.Name = "Phone"
	p.IsDefault = false
	p.OwnerID = "u_1"

	if err := store.Save(&p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if p.ID == "" {
		t.Fatal("id was not assigned")
	}

	got, err := store.Get(p.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Phone" {
		t.Errorf("name = %q, want Phone", got.Name)
	}
	if got.OwnerID != "u_1" {
		t.Errorf("owner = %q, want u_1", got.OwnerID)
	}
	if got.CreatedAt.IsZero() {
		t.Error("created_at was not set")
	}
}

func TestStoreGetUnknown(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.Get("missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want %v", err, ErrNotFound)
	}
}

func TestStoreNameMustBeUniquePerOwner(t *testing.T) {
	store := newTestStore(t)

	first := Default()
	first.ID = ""
	first.Name = "TV"
	first.IsDefault = false
	first.OwnerID = "u_1"
	if err := store.Save(&first); err != nil {
		t.Fatalf("Save first: %v", err)
	}

	second := Default()
	second.ID = ""
	second.Name = "TV"
	second.IsDefault = false
	second.OwnerID = "u_1"

	if err := store.Save(&second); !errors.Is(err, ErrNameTaken) {
		t.Errorf("error = %v, want %v", err, ErrNameTaken)
	}
}

func TestStoreSameNameForDifferentOwners(t *testing.T) {
	store := newTestStore(t)

	for _, owner := range []string{"u_1", "u_2"} {
		p := Default()
		p.ID = ""
		p.Name = "TV"
		p.IsDefault = false
		p.OwnerID = owner
		if err := store.Save(&p); err != nil {
			t.Fatalf("Save for %s: %v", owner, err)
		}
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("profiles = %d, want 2", len(list))
	}
}

func TestStoreEmptyNameRejected(t *testing.T) {
	store := newTestStore(t)

	p := Default()
	p.ID = ""
	p.Name = "   "

	if err := store.Save(&p); err == nil {
		t.Error("expected an error for an empty name")
	}
}

func TestStoreListSortedWithDefaultFirst(t *testing.T) {
	store := newTestStore(t)

	for _, name := range []string{"Zulu", "alpha"} {
		p := Default()
		p.ID = ""
		p.Name = name
		p.IsDefault = false
		p.OwnerID = "u_1"
		if err := store.Save(&p); err != nil {
			t.Fatalf("Save %s: %v", name, err)
		}
	}

	marked := Default()
	marked.ID = ""
	marked.Name = "Middle"
	marked.OwnerID = "u_1"
	marked.IsDefault = true
	if err := store.Save(&marked); err != nil {
		t.Fatalf("Save marked: %v", err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("profiles = %d, want 3", len(list))
	}
	if !list[0].IsDefault {
		t.Errorf("first profile is %q, want the default one", list[0].Name)
	}
	if list[1].Name != "alpha" || list[2].Name != "Zulu" {
		t.Errorf("order = %q, %q, want alpha, Zulu", list[1].Name, list[2].Name)
	}
}

func TestStoreOnlyOneDefault(t *testing.T) {
	store := newTestStore(t)

	first := Default()
	first.ID = ""
	first.Name = "First"
	first.OwnerID = "u_1"
	first.IsDefault = true
	if err := store.Save(&first); err != nil {
		t.Fatalf("Save first: %v", err)
	}

	second := Default()
	second.ID = ""
	second.Name = "Second"
	second.OwnerID = "u_1"
	second.IsDefault = true
	if err := store.Save(&second); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	defaults := 0
	for _, p := range list {
		if p.IsDefault {
			defaults++
			if p.ID != second.ID {
				t.Errorf("default is %q, want %q", p.Name, second.Name)
			}
		}
	}
	if defaults != 1 {
		t.Errorf("default profiles = %d, want 1", defaults)
	}
}

func TestStoreAvailable(t *testing.T) {
	store := newTestStore(t)

	private := Default()
	private.ID = ""
	private.Name = "Private"
	private.OwnerID = "u_admin"
	private.Visibility = VisibilityPrivate
	private.IsDefault = false
	if err := store.Save(&private); err != nil {
		t.Fatalf("Save private: %v", err)
	}

	public := Default()
	public.ID = ""
	public.Name = "Public"
	public.OwnerID = "u_admin"
	public.Visibility = VisibilityPublic
	public.RankRequired = 50
	public.IsDefault = false
	if err := store.Save(&public); err != nil {
		t.Fatalf("Save public: %v", err)
	}

	// Владелец профиля видит свой приватный профиль
	own, err := store.Available("u_admin", 50)
	if err != nil {
		t.Fatalf("Available: %v", err)
	}
	if len(own) != 2 {
		t.Errorf("owner sees %d profiles, want 2", len(own))
	}

	// Обычный пользователь не видит приватный и не проходит по рангу
	regular, err := store.Available("u_user", 10)
	if err != nil {
		t.Fatalf("Available: %v", err)
	}
	if len(regular) != 0 {
		t.Errorf("regular user sees %d profiles, want 0", len(regular))
	}

	// Админ другого аккаунта проходит по рангу для публичного
	admin, err := store.Available("u_other", 50)
	if err != nil {
		t.Fatalf("Available: %v", err)
	}
	if len(admin) != 1 || admin[0].Name != "Public" {
		t.Errorf("other admin sees %d profiles, want only Public", len(admin))
	}
}

func TestStoreDelete(t *testing.T) {
	store := newTestStore(t)

	p := Default()
	p.ID = ""
	p.Name = "Temp"
	p.OwnerID = "u_1"
	p.IsDefault = false
	if err := store.Save(&p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := store.Delete(p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(p.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("profile still exists: %v", err)
	}
}

func TestStoreDeleteUnknown(t *testing.T) {
	store := newTestStore(t)

	if err := store.Delete("missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want %v", err, ErrNotFound)
	}
}

func TestStoreDeleteBuiltin(t *testing.T) {
	store := newTestStore(t)

	if err := store.Delete(DefaultID); !errors.Is(err, ErrBuiltin) {
		t.Errorf("error = %v, want %v", err, ErrBuiltin)
	}
}

func TestStoreDefaultFallback(t *testing.T) {
	store := newTestStore(t)

	// В базе пусто, поэтому возвращается встроенный профиль
	p := store.Default()
	if p == nil || p.ID != DefaultID {
		t.Fatalf("default profile = %+v, want builtin", p)
	}
}

func TestStoreDefaultFromDatabase(t *testing.T) {
	store := newTestStore(t)

	custom := Default()
	custom.ID = ""
	custom.Name = "Custom"
	custom.OwnerID = "u_1"
	custom.IsDefault = true
	if err := store.Save(&custom); err != nil {
		t.Fatalf("Save: %v", err)
	}

	p := store.Default()
	if p.ID != custom.ID {
		t.Errorf("default id = %q, want %q", p.ID, custom.ID)
	}
	if p.Name != "Custom" {
		t.Errorf("default name = %q, want Custom", p.Name)
	}
}

func TestNormalize(t *testing.T) {
	p := Profile{Protocol: "weird", Visibility: "nope", RankRequired: -5}
	p.Normalize()

	if p.Protocol != "progressive" {
		t.Errorf("protocol = %q, want progressive", p.Protocol)
	}
	if p.Visibility != VisibilityPrivate {
		t.Errorf("visibility = %q, want private", p.Visibility)
	}
	if p.RankRequired != 0 {
		t.Errorf("rank = %d, want 0", p.RankRequired)
	}
	if p.HLS.SegmentLength <= 0 {
		t.Error("segment length was not filled")
	}
}

func TestAvailableRules(t *testing.T) {
	cases := []struct {
		name    string
		profile Profile
		owned   bool
		rank    int
		want    bool
	}{
		{name: "default always", profile: Profile{IsDefault: true}, rank: 1, want: true},
		{name: "own private", profile: Profile{Visibility: VisibilityPrivate}, owned: true, want: true},
		{name: "other private", profile: Profile{Visibility: VisibilityPrivate, RankRequired: 1}, rank: 100, want: false},
		{name: "public enough rank", profile: Profile{Visibility: VisibilityPublic, RankRequired: 50}, rank: 50, want: true},
		{name: "public low rank", profile: Profile{Visibility: VisibilityPublic, RankRequired: 50}, rank: 10, want: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.profile.Available(c.owned, c.rank); got != c.want {
				t.Errorf("Available = %v, want %v", got, c.want)
			}
		})
	}
}
