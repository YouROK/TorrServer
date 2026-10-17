package utils

import (
	"strings"
	"testing"

	"server/torrshash"
)

func TestParseTorrsHashBareInfoHash(t *testing.T) {
	const hash = "e3e497e1efd8db8df93822cad554e13740f51997"
	for _, in := range []string{
		"torrs://" + hash,
		"web+torrs://" + hash,
		strings.ToUpper(hash),
	} {
		spec, th, err := ParseTorrsHash(in)
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if th.Hash != hash {
			t.Fatalf("%q: hash want %s, got %s", in, hash, th.Hash)
		}
		if spec.InfoHash.HexString() != hash {
			t.Fatalf("%q: infohash want %s, got %s", in, hash, spec.InfoHash.HexString())
		}
	}
}

func TestParseTorrsHashPackedRoundTrip(t *testing.T) {
	th := torrshash.New("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	th.AddField(torrshash.TagTitle, "Show A")
	token, err := torrshash.Pack(th)
	if err != nil {
		t.Fatal(err)
	}
	spec, parsed, err := ParseTorrsHash("torrs://" + token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Title() != "Show A" {
		t.Fatalf("title: got %q", parsed.Title())
	}
	if spec.DisplayName != "Show A" {
		t.Fatalf("display name: got %q", spec.DisplayName)
	}
}

func TestParseTorrsHashRejectsGarbage(t *testing.T) {
	if _, _, err := ParseTorrsHash("torrs://not-a-hash-or-token!!!"); err == nil {
		t.Fatal("expected error")
	}
}
