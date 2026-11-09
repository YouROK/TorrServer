//go:build gst

package gstreamer

import (
	"testing"
	"time"
)

// Per-viewer tasks ("<hash>.<sid>").

const testHash = "d9c69fc5a6d501c1022cad09df463e6083b3514d"

func TestTaskIDForSession(t *testing.T) {
	if got := taskIDFor(testHash, ""); got != testHash {
		t.Fatalf("no session: %q", got)
	}
	if got := taskIDFor(testHash, "ab12"); got != testHash+".ab12" {
		t.Fatalf("session: %q", got)
	}
	if HashOfTaskID(testHash+".ab12") != testHash || HashOfTaskID(testHash) != testHash {
		t.Fatal("HashOfTaskID")
	}
	for _, bad := range []string{"", "a.b", "a/b", "../x", "x y", string(make([]byte, 65))} {
		if validSession(bad) {
			t.Fatalf("session %q accepted", bad)
		}
	}
	if !validSession("Ab-9_z") {
		t.Fatal("valid session rejected")
	}
}

func TestSessionsAreIndependentTasks(t *testing.T) {
	now := time.Now().UTC()
	a, ra := newTrackedTask(testHash+".a", now)
	b, rb := newTrackedTask(testHash+".b", now)
	other, ro := newTrackedTask("0000000000000000000000000000000000000000.a", now)
	service := &Service{tasks: map[string]*Task{a.ID: a, b.ID: b, other.ID: other}}

	// one viewer closes the player: only its own task goes, the torrent stays for the other viewer
	if !service.TryRemove(a.ID) || !ra.disposed.Load() {
		t.Fatal("own task not removed")
	}
	if service.tasks[b.ID] != b || rb.disposed.Load() {
		t.Fatal("other viewer's task was touched")
	}
	if !service.hasTasksForHash(testHash) {
		t.Fatal("hasTasksForHash: viewer b still streams")
	}
	// clients without sid remove / heartbeat with the bare hash
	if service.TryRemove(testHash) {
		t.Fatal("bare hash removed a viewer's task")
	}
	if service.GetAnyForHash(testHash) != b {
		t.Fatal("GetAnyForHash")
	}
	// the torrent itself is dropped: all its viewers go, other torrents stay
	if !service.RemoveTorrent(testHash) || !rb.disposed.Load() || len(service.tasks) != 1 || ro.disposed.Load() {
		t.Fatal("RemoveTorrent")
	}
}
