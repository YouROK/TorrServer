package torrshash

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestPackRoundTrip(t *testing.T) {
	hash := strings.Repeat("ab", 20)
	th := New(hash)
	th.AddField(TagTitle, " Show ")
	th.AddField(TagPoster, "http://img")
	th.AddField(TagCategory, "movie")
	th.AddField(TagTracker, "http://t1")
	th.AddField(TagTracker, "udp://t2")
	th.AddField(TagSize, "42")
	th.AddField(TagTitle, "   ")
	th.AddField(TagSize, "0")

	token, err := Pack(th)
	if err != nil || token == "" || !IsBase62(token) {
		t.Fatalf("pack %q %v", token, err)
	}
	got, err := Unpack("  " + token + "\n")
	if err != nil {
		t.Fatal(err)
	}
	if got.Hash != hash || got.Title() != "Show" || got.Poster() != "http://img" || got.Category() != "movie" {
		t.Fatalf("%s", got)
	}
	if tr := got.Trackers(); len(tr) != 2 || tr[0] != "http://t1" || tr[1] != "udp://t2" {
		t.Fatalf("trackers %#v", tr)
	}
	for _, f := range got.Fields {
		if f.Tag == TagSize && f.Value != "42" {
			t.Fatalf("size field %#v", f)
		}
	}

	raw, err := PackBytes(th)
	if err != nil {
		t.Fatal(err)
	}
	again, err := UnpackBytes(raw)
	if err != nil || again.Title() != "Show" {
		t.Fatalf("bytes unpack %v %#v", err, again)
	}
	if enc := Encode62(raw); Decode62(enc) == nil || !bytes.Equal(Decode62(enc), raw) {
		t.Fatal("base62 mismatch")
	}
}

func TestUnpackRejectsGarbage(t *testing.T) {
	if _, err := Unpack("!!!"); err == nil {
		t.Fatal("expected error for non-zlib token")
	}
	if _, err := UnpackBytes([]byte("not zlib")); err == nil {
		t.Fatal("expected error")
	}
	if _, err := UnpackBytes(nil); err == nil {
		t.Fatal("expected error for nil")
	}
	empty := New("zz")
	if _, err := Pack(empty); err != nil {
		t.Fatal(err)
	}
	if _, err := UnpackBytes(mustPack(t, empty)); err == nil {
		t.Fatal("short hash should fail unpack")
	}
}

func mustPack(t *testing.T, th *TorrsHash) []byte {
	t.Helper()
	b, err := PackBytes(th)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBase62AndTags(t *testing.T) {
	if IsBase62("") || IsBase62("abc!") || !IsBase62(" aZ09 ") {
		t.Fatal("IsBase62")
	}
	if Encode62(nil) != "" || len(Decode62("")) != 0 {
		t.Fatalf("empty encode %q decode %#v", Encode62(nil), Decode62(""))
	}
	if Tag(99).String() != "Unknown" || Tag(TagTitle).String() != "Title" || Tag(TagSize).String() != "Size" {
		t.Fatal(Tag(99).String())
	}
	if err := (&Field{Tag: TagSize, Value: "nope"}).write(&bytes.Buffer{}); err == nil {
		t.Fatal("bad size should fail")
	}
	if err := (&Field{Tag: TagTitle, Value: ""}).write(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
}

func TestPackConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			th := New(strings.Repeat("cd", 20))
			th.AddField(TagTitle, "n")
			th.AddField(TagSize, "7")
			token, err := Pack(th)
			if err != nil {
				t.Error(err)
				return
			}
			got, err := Unpack(token)
			if err != nil || got.Title() != "n" {
				t.Errorf("unpack %v %#v", err, got)
			}
		}()
	}
	wg.Wait()
}
