package settings

import (
	"fmt"
	"sync"
	"testing"
)

type memDB struct {
	mu sync.Mutex
	m  map[string][]byte
}

func newMemDB() *memDB { return &memDB{m: map[string][]byte{}} }

func (d *memDB) key(x, n string) string { return x + "\x00" + n }

func (d *memDB) CloseDB() {}
func (d *memDB) Get(x, n string) []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]byte(nil), d.m[d.key(x, n)]...)
}
func (d *memDB) Set(x, n string, v []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.m[d.key(x, n)] = append([]byte(nil), v...)
}
func (d *memDB) List(x string) []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	var names []string
	prefix := x + "\x00"
	for k := range d.m {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			names = append(names, k[len(prefix):])
		}
	}
	return names
}
func (d *memDB) Rem(x, n string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.m, d.key(x, n))
}
func (d *memDB) Clear(x string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	prefix := x + "\x00"
	for k := range d.m {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(d.m, k)
		}
	}
}

func TestXPathRouterRoutesAndRejectsDuplicate(t *testing.T) {
	r := NewXPathDBRouter()
	a, b := newMemDB(), newMemDB()
	if err := r.RegisterRoute(a, "Torrents"); err != nil {
		t.Fatal(err)
	}
	if err := r.RegisterRoute(b, " Viewed "); err != nil {
		t.Fatal(err)
	}
	if err := r.RegisterRoute(a, "torrents"); err == nil {
		t.Fatal("duplicate route")
	}
	r.Set("Torrents", "h", []byte("a"))
	r.Set("Viewed", "h", []byte("b"))
	if string(r.Get("Torrents", "h")) != "a" || string(r.Get("Viewed", "h")) != "b" {
		t.Fatalf("get %q %q", r.Get("Torrents", "h"), r.Get("Viewed", "h"))
	}
	if len(r.List("Torrents")) != 1 {
		t.Fatalf("list %#v", r.List("Torrents"))
	}
	r.Rem("Torrents", "h")
	if r.Get("Torrents", "h") != nil {
		t.Fatal("removed key still present")
	}
	r.Set("Viewed", "x", []byte("z"))
	r.Clear("Viewed")
	if r.Get("Viewed", "h") != nil || r.Get("Viewed", "x") != nil {
		t.Fatal("clear left keys")
	}
	r.CloseDB()
}

func TestXPathRouterConcurrent(t *testing.T) {
	r := NewXPathDBRouter()
	if err := r.RegisterRoute(newMemDB(), "Bucket"); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 40; i++ {
				name := fmt.Sprintf("n%d-%d", g, i)
				r.Set("Bucket", name, []byte(name))
				_ = r.Get("Bucket", name)
				_ = r.List("Bucket")
				r.Rem("Bucket", name)
			}
		}(g)
	}
	wg.Wait()
}

func TestJsonDBConcurrent(t *testing.T) {
	old := Path
	Path = t.TempDir()
	t.Cleanup(func() { Path = old })
	db := NewJsonDB()
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				name := fmt.Sprintf("k%d-%d", g, i)
				db.Set("Bucket", name, []byte(fmt.Sprintf(`{"n":%d}`, i)))
				_ = db.Get("Bucket", name)
				_ = db.List("Bucket")
				db.Rem("Bucket", name)
			}
		}(g)
	}
	wg.Wait()
	db.Clear("Bucket")
	if names := db.List("Bucket"); len(names) != 0 {
		t.Fatalf("left %#v", names)
	}
}

func TestReadIndexes(t *testing.T) {
	got := readIndexes([]byte(`{"1":1.5,"2":0}`))
	if got[1] != 1.5 || got[2] != 0 {
		t.Fatalf("%#v", got)
	}
	old := readIndexes([]byte(`{"3":{},"4":{}}`))
	if _, ok := old[3]; !ok || old[3] != 0 || old[4] != 0 {
		t.Fatalf("legacy %#v", old)
	}
	if len(readIndexes([]byte("not json"))) != 0 {
		t.Fatal("garbage")
	}
	if len(readIndexes(nil)) != 0 {
		t.Fatal("nil")
	}
}
