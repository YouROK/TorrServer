//go:build gst

package gstreamer

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSubtitleVTTWaitsForVideoRange(t *testing.T) {
	task, store := subtitleTestTask()
	store.appendVTT("00:00:01.000 --> 00:00:02.000\nhello\n\n", 0, 0)

	value, ready := task.SubtitleVTT(2, 0)
	if ready {
		t.Fatal("subtitle range was ready before the video reader passed it")
	}
	if !strings.Contains(value, "hello") {
		t.Fatalf("partial VTT does not contain parsed cue: %q", value)
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		store.advanceVideoReadTo(6_000_000_000)
	}()

	value, err := task.WaitSubtitleVTT(context.Background(), 2, 0, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(value, "hello") {
		t.Fatalf("ready VTT does not contain parsed cue: %q", value)
	}
}

func TestSubtitleVTTEmptyReadRangeIsReady(t *testing.T) {
	task, store := subtitleTestTask()
	store.advanceVideoReadTo(6_000_000_000)

	value, ready := task.SubtitleVTT(2, 0)
	if !ready {
		t.Fatal("empty subtitle range remained pending after video passed it")
	}
	if strings.Contains(value, "-->") {
		t.Fatalf("empty VTT unexpectedly contains a cue: %q", value)
	}
}

func TestCompletedVideoSegmentMakesSubtitleRangeReady(t *testing.T) {
	task, store := subtitleTestTask()
	runner := &gstRunner{
		task:           task,
		subtitleStores: map[int]*subtitleStore{2: store},
	}
	runner.readySegment.segment = Segment{EndNS: 5_900_000_000}
	runner.readySegment.complete = true

	runner.completeReadySegment(0)

	_, ready := task.SubtitleVTT(2, 0)
	if !ready {
		t.Fatal("completed video segment did not advance the subtitle range watermark")
	}
}

func TestWaitSubtitleVTTReturnsPartialDataAtTimeout(t *testing.T) {
	task, store := subtitleTestTask()
	store.appendVTT("00:00:01.000 --> 00:00:02.000\npartial\n\n", 0, 0)

	value, err := task.WaitSubtitleVTT(context.Background(), 2, 0, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(value, "partial") {
		t.Fatalf("timeout discarded partial subtitle data: %q", value)
	}
}

func TestWaitSubtitleVTTReturnsEmptyVTTAtTimeout(t *testing.T) {
	task, _ := subtitleTestTask()

	value, err := task.WaitSubtitleVTT(context.Background(), 2, 0, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(value, "WEBVTT\n") || strings.Contains(value, "-->") {
		t.Fatalf("timeout did not return an empty VTT document: %q", value)
	}
}

func subtitleTestTask() (*Task, *subtitleStore) {
	task := &Task{
		Config: Config{
			SegmentSeconds: 6,
			Subtitles:      true,
		}.normalized(),
		Probe: ProbeInfo{DurationNS: int64(time.Minute)},
	}
	store := newSubtitleStore()
	task.setSubtitleStores(map[int]*subtitleStore{2: store})
	return task, store
}
