package torrfs

import (
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/anacrolix/torrent"

	"server/settings"
	"server/torr"
	"server/torr/offlinetest"
)

func TestOfflineTree(t *testing.T) {
	offlinetest.Prepare(t)
	settings.BTsets.ShowFSActiveTorr = true
	bt := torr.NewBTS()
	torr.InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)

	show := offlinetest.WriteShow(t)
	seed := offlinetest.Seed(t, show)
	tor, err := torr.AddTorrent(show.Spec, "Show", "", "", "tv")
	if err != nil {
		t.Fatal(err)
	}
	if !tor.GotInfo() {
		t.Fatal("GotInfo")
	}
	tor.AddPeers([]torrent.Peer{seed.Peer()})

	if SanitizeName("a/b\\c") != "a_b_c" || SanitizeName(".") != "" || SanitizeName("  ") != "" {
		t.Fatal("SanitizeName")
	}

	root := NewRootDir()
	if _, err := root.Stat(); err != nil || root.Type() == 0 || !root.IsDir() {
		t.Fatal("root stat")
	}
	ents, err := root.ReadDir(0)
	if err != nil || len(ents) == 0 {
		t.Fatalf("root %v %v", ents, err)
	}
	if n, err := root.ReadDir(1); err != nil || len(n) != 1 {
		t.Fatalf("ReadDir 1 %v %v", n, err)
	}

	mkv, rel, err := findFile(root, "", ".mkv")
	if err != nil {
		t.Fatal(err)
	}
	st, err := mkv.Stat()
	if err != nil || st.IsDir() || st.Size() < 1024 {
		t.Fatalf("stat %+v %v", st, err)
	}
	buf := make([]byte, 16)
	done := make(chan error, 1)
	go func() {
		_, err := io.ReadFull(mkv, buf)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("read timed out")
	}
	_ = mkv.Close()

	if _, err := AsFS(root).Open(strings.TrimPrefix(rel, "/")); err != nil {
		t.Fatal(err)
	}
	if _, err := root.Open(rel); err != nil {
		t.Fatal("reopen", rel, err)
	}
	if _, err := root.Open("../etc/passwd"); err == nil {
		t.Fatal("invalid path accepted")
	}
	if f, err := root.Open("tv/missing"); err == nil {
		_ = f.Close()
		t.Fatal("missing path opened")
	}
}

func findFile(n INode, prefix, suffix string) (fs.File, string, error) {
	ents, err := n.ReadDir(-1)
	if err != nil {
		return nil, "", err
	}
	for _, e := range ents {
		child, ok := e.(INode)
		if !ok {
			continue
		}
		path := prefix + child.Name()
		if !child.IsDir() && strings.HasSuffix(child.Name(), suffix) {
			f, err := child.Open(child.Name())
			return f, path, err
		}
		if child.IsDir() {
			f, p, err := findFile(child, path+"/", suffix)
			if err == nil {
				return f, p, nil
			}
		}
	}
	return nil, "", fs.ErrNotExist
}
