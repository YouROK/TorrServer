package source

import (
	"fmt"
	"testing"
	"time"
)

// newTestRegistry создаёт реестр с управляемым временем и предсказуемыми id.
func newTestRegistry(ttl time.Duration) (*Registry, func(time.Duration)) {
	r := NewRegistry(ttl)

	now := time.Now()
	var counter int

	r.now = func() time.Time { return now }
	r.newID = func() string {
		counter++
		return fmt.Sprintf("lease-%d", counter)
	}

	advance := func(d time.Duration) { now = now.Add(d) }
	return r, advance
}

func TestIssueAndGet(t *testing.T) {
	r, _ := newTestRegistry(10 * time.Minute)

	lease := r.Issue("owner", "abc123", 2)
	if lease.ID == "" {
		t.Fatal("lease id is empty")
	}
	if lease.UserID != "owner" || lease.Hash != "abc123" || lease.FileIdx != 2 {
		t.Errorf("lease fields are wrong: %+v", lease)
	}

	got, ok := r.Get(lease.ID)
	if !ok {
		t.Fatal("issued lease was not found")
	}
	if got.ID != lease.ID {
		t.Errorf("got lease %q, want %q", got.ID, lease.ID)
	}
}

func TestGetUnknown(t *testing.T) {
	r, _ := newTestRegistry(time.Minute)

	if _, ok := r.Get("missing"); ok {
		t.Error("unknown lease must not be returned")
	}
	if _, ok := r.Get(""); ok {
		t.Error("empty lease id must not be returned")
	}
}

func TestLeaseExpires(t *testing.T) {
	r, advance := newTestRegistry(time.Minute)

	lease := r.Issue("owner", "abc", 0)
	if _, ok := r.Get(lease.ID); !ok {
		t.Fatal("lease must be valid right after issuing")
	}

	advance(59 * time.Second)
	if _, ok := r.Get(lease.ID); !ok {
		t.Error("lease must still be valid before the deadline")
	}

	advance(2 * time.Second)
	if _, ok := r.Get(lease.ID); ok {
		t.Error("expired lease must not be returned")
	}
	if r.Len() != 0 {
		t.Errorf("expired lease is still stored: %d", r.Len())
	}
}

func TestRelease(t *testing.T) {
	r, _ := newTestRegistry(time.Minute)

	lease := r.Issue("owner", "abc", 0)
	r.Release(lease.ID)

	if _, ok := r.Get(lease.ID); ok {
		t.Error("released lease must not be returned")
	}
	if r.Len() != 0 {
		t.Errorf("released lease is still stored: %d", r.Len())
	}
}

func TestCleanup(t *testing.T) {
	r, advance := newTestRegistry(time.Minute)

	var ids []string
	for i := 0; i < 3; i++ {
		ids = append(ids, r.Issue("owner", "abc", i).ID)
	}

	advance(30 * time.Second)
	fresh := r.Issue("owner", "def", 0)

	advance(31 * time.Second)

	removed := r.Cleanup()
	if removed != 3 {
		t.Errorf("cleanup removed %d leases, want 3", removed)
	}
	if r.Len() != 1 {
		t.Errorf("leases left = %d, want 1", r.Len())
	}
	if _, ok := r.Get(fresh.ID); !ok {
		t.Error("fresh lease was removed by cleanup")
	}
	for _, id := range ids {
		if _, ok := r.Get(id); ok {
			t.Errorf("expired lease %s survived cleanup", id)
		}
	}
}

func TestRandomIDIsUniqueAndLong(t *testing.T) {
	r := NewRegistry(time.Minute)

	seen := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		id := r.Issue("owner", "abc", 0).ID
		if len(id) != IDBytes*2 {
			t.Fatalf("id %q has length %d, want %d", id, len(id), IDBytes*2)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestDefaultTTL(t *testing.T) {
	r := NewRegistry(0)
	if r.ttl != DefaultTTL {
		t.Errorf("ttl = %v, want %v", r.ttl, DefaultTTL)
	}

	r = NewRegistry(-time.Second)
	if r.ttl != DefaultTTL {
		t.Errorf("ttl = %v, want %v", r.ttl, DefaultTTL)
	}
}

func TestLeaseExpired(t *testing.T) {
	now := time.Now()

	cases := []struct {
		name    string
		expires time.Time
		want    bool
	}{
		{name: "future", expires: now.Add(time.Minute), want: false},
		{name: "past", expires: now.Add(-time.Minute), want: true},
		{name: "exactly now", expires: now, want: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lease := &Lease{ExpiresAt: c.expires}
			if got := lease.Expired(now); got != c.want {
				t.Errorf("Expired = %v, want %v", got, c.want)
			}
		})
	}
}

func TestTouchExtendsLease(t *testing.T) {
	r, advance := newTestRegistry(time.Minute)

	lease := r.Issue("owner", "abc", 0)

	advance(50 * time.Second)
	if !r.Touch(lease.ID) {
		t.Fatal("Touch returned false for a valid lease")
	}

	// После продления доступ живёт ещё минуту
	advance(50 * time.Second)
	if _, ok := r.Get(lease.ID); !ok {
		t.Error("lease must stay valid after being touched")
	}

	advance(11 * time.Second)
	if _, ok := r.Get(lease.ID); ok {
		t.Error("lease must expire after the extended deadline")
	}
}

func TestTouchUnknown(t *testing.T) {
	r, _ := newTestRegistry(time.Minute)

	if r.Touch("missing") {
		t.Error("Touch must fail for an unknown lease")
	}
	if r.Touch("") {
		t.Error("Touch must fail for an empty lease id")
	}
}

func TestTouchExpired(t *testing.T) {
	r, advance := newTestRegistry(time.Minute)

	lease := r.Issue("owner", "abc", 0)
	advance(2 * time.Minute)

	if r.Touch(lease.ID) {
		t.Error("Touch must fail for an expired lease")
	}
	if r.Len() != 0 {
		t.Errorf("expired lease is still stored: %d", r.Len())
	}
}
