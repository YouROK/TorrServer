package torr

import (
	"strings"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
)

func TestBytesHuman(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1 << 20, "1.0 MiB"},
		{1 << 30, "1.0 GiB"},
		{3 << 30, "3.0 GiB"},
	}
	for _, c := range cases {
		if got := bytesHuman(c.n); got != c.want {
			t.Errorf("bytesHuman(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestCheckOfflineSpaceSkipsWhenNothingNeeded(t *testing.T) {
	if err := checkOfflineSpace(t.TempDir(), 0); err != nil {
		t.Fatalf("checkOfflineSpace with remaining 0: %v", err)
	}
}

func TestCheckOfflineSpaceRefusesImpossibleSize(t *testing.T) {
	// 1 PiB does not fit on any test filesystem: the error must mention the
	// amounts, so the user sees what was requested and what was available.
	err := checkOfflineSpace(t.TempDir(), 1<<50)
	if err == nil {
		t.Fatal("checkOfflineSpace must refuse a 1 PiB download")
	}
	if !strings.Contains(err.Error(), "not enough disk space") {
		t.Fatalf("unexpected error text: %v", err)
	}
}

func TestCheckOfflineSpaceToleratesUnknownFreeSpace(t *testing.T) {
	// A missing directory cannot be statfs'ed: the check degrades to "not
	// blocked" instead of failing the download.
	if err := checkOfflineSpace(t.TempDir()+"/does-not-exist", 1<<50); err != nil {
		t.Fatalf("checkOfflineSpace must not block when free space is unknown: %v", err)
	}
}

func TestOfflineTotalLengthAndRemainingOfEmptySpec(t *testing.T) {
	if n := offlineTotalLength(nil); n != 0 {
		t.Fatalf("offlineTotalLength(nil) = %d, want 0", n)
	}
	if n := offlineRemainingBytes(nil, metainfo.Hash{}); n != 0 {
		t.Fatalf("offlineRemainingBytes(nil) = %d, want 0", n)
	}
}
