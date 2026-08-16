//go:build gst

package gstreamer

import (
	"bytes"
	"testing"
)

func testAudioSplitExplicitRun(
	t *testing.T,
	sampleCount int,
	duration uint32,
	size uint32,
) mp4Run {
	t.Helper()

	samples := make([]mp4Sample, sampleCount)
	for i := range samples {
		samples[i] = mp4Sample{
			duration: duration,
			size:     size,
			flags:    0,
		}
	}

	run, err := buildExplicitRun(
		mp4Run{version: 0},
		samples,
	)
	if err != nil {
		t.Fatalf("buildExplicitRun: %v", err)
	}
	return run
}

func testAudioSplitFragment(
	t *testing.T,
	decodeTime uint64,
	timescale uint32,
	runs ...mp4Run,
) mp4Fragment {
	t.Helper()

	var duration uint64
	var payloadLength int
	var payloadOffset int64

	for i := range runs {
		runs[i].payloadOffset = payloadOffset
		payloadOffset += int64(runs[i].dataSize)
		duration += runs[i].duration
		payloadLength += int(runs[i].dataSize)
	}

	payload := make([]byte, payloadLength)
	for i := range payload {
		payload[i] = byte(i)
	}

	return mp4Fragment{
		trackID:        2,
		timescale:      timescale,
		decodeTime:     decodeTime,
		duration:       duration,
		startsWithSync: true,
		tfhd:           buildCanonicalTfhd(2, 1),
		runs:           runs,
		payloadStart:   0,
		payloadLen:     len(payload),
	}
}

func TestSplitFragmentAtSamplePreservesTimelineAndPayload(t *testing.T) {
	source := testAudioSplitFragment(
		t,
		10_000,
		48_000,
		testAudioSplitExplicitRun(t, 4, 1024, 10),
		testAudioSplitExplicitRun(t, 3, 1024, 20),
	)
	sourcePayload := make([]byte, source.payloadLen)
	for i := range sourcePayload {
		sourcePayload[i] = byte(i)
	}

	left, right, err := splitFragmentAtSample(source, 5)
	if err != nil {
		t.Fatalf("splitFragmentAtSample: %v", err)
	}

	if got, want := left.duration, uint64(5*1024); got != want {
		t.Fatalf("left duration = %d, want %d", got, want)
	}
	if got, want := right.duration, uint64(2*1024); got != want {
		t.Fatalf("right duration = %d, want %d", got, want)
	}
	if got, want := right.decodeTime, source.decodeTime+left.duration; got != want {
		t.Fatalf("right decode time = %d, want %d", got, want)
	}
	if got, want := left.payloadLen, 4*10+1*20; got != want {
		t.Fatalf("left payload length = %d, want %d", got, want)
	}
	if got, want := right.payloadLen, 2*20; got != want {
		t.Fatalf("right payload length = %d, want %d", got, want)
	}

	joined := append(
		append([]byte(nil), sourcePayload[left.payloadStart:left.payloadStart+left.payloadLen]...),
		sourcePayload[right.payloadStart:right.payloadStart+right.payloadLen]...,
	)
	if !bytes.Equal(joined, sourcePayload) {
		t.Fatal("left+right payload does not reproduce source payload")
	}

	rightSamples, err := fragmentSampleCount(right)
	if err != nil {
		t.Fatal(err)
	}
	leftSamples, err := fragmentSampleCount(left)
	if err != nil {
		t.Fatal(err)
	}
	if leftSamples != 5 {
		t.Fatalf("left samples = %d, want 5", leftSamples)
	}
	if rightSamples != 2 {
		t.Fatalf("right samples = %d, want 2", rightSamples)
	}
}

func TestSelectAudioCountSplitsAtFirstBoundaryNotBeforeVideoEnd(t *testing.T) {
	reader := Mp4BoxReader(
		func([]byte) {},
		func(Segment) {},
		6,
		0,
		false,
	)
	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.audioTrack = trackInfo{id: 2, timescale: 48_000}

	reader.audio = append(
		reader.audio,
		testAudioSplitFragment(
			t,
			0,
			48_000,
			testAudioSplitExplicitRun(t, 10, 1024, 10),
		),
	)

	count, err := reader.selectAudioCount(100)
	if err != nil {
		t.Fatalf("selectAudioCount: %v", err)
	}
	if count != 1 {
		t.Fatalf("audio count = %d, want 1", count)
	}
	if len(reader.audio) != 2 {
		t.Fatalf("audio fragments = %d, want 2 after split", len(reader.audio))
	}

	left := reader.audio[0]
	right := reader.audio[1]

	if got, want := left.duration, uint64(5*1024); got != want {
		t.Fatalf("left duration = %d, want %d", got, want)
	}
	if got, want := right.decodeTime, left.endTime(); got != want {
		t.Fatalf("right decode time = %d, want %d", got, want)
	}
	if left.payloadLen+right.payloadLen != 100 {
		t.Fatalf(
			"payload partition = %d+%d, want 100",
			left.payloadLen,
			right.payloadLen,
		)
	}

	if !scaledGreaterOrEqual(
		left.endTime(),
		reader.videoTrack.timescale,
		100,
		reader.audioTrack.timescale,
	) {
		t.Fatal("selected audio ends before video")
	}

	previousBoundary := left.endTime() - 1024
	if scaledGreaterOrEqual(
		previousBoundary,
		reader.videoTrack.timescale,
		100,
		reader.audioTrack.timescale,
	) {
		t.Fatal("split was not made at the first audio boundary at/after video end")
	}
}

func TestReclaimPayloadsKeepsSplitAudioTail(t *testing.T) {
	reader := Mp4BoxReader(
		func([]byte) {},
		func(Segment) {},
		6,
		0,
		false,
	)
	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.audioTrack = trackInfo{id: 2, timescale: 48_000}

	source := testAudioSplitFragment(
		t,
		0,
		48_000,
		testAudioSplitExplicitRun(t, 10, 1024, 10),
	)
	sourcePayload := make([]byte, source.payloadLen)
	for i := range sourcePayload {
		sourcePayload[i] = byte(i)
	}
	_, _ = reader.sourcePayload.Write(sourcePayload)
	reader.audio = append(reader.audio, source)

	count, err := reader.selectAudioCount(100)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || len(reader.audio) != 2 {
		t.Fatalf("split result count=%d fragments=%d, want 1/2", count, len(reader.audio))
	}

	rightBefore := reader.audio[1]
	rightPayload := append(
		[]byte(nil),
		reader.sourcePayload.Bytes()[rightBefore.payloadStart:rightBefore.payloadStart+rightBefore.payloadLen]...,
	)

	reader.audio = removeFragments(reader.audio, 1)
	reader.ReclaimPayloads()

	if len(reader.audio) != 1 {
		t.Fatalf("audio fragments=%d, want 1", len(reader.audio))
	}
	rightAfter := reader.audio[0]
	if rightAfter.payloadStart != 0 {
		t.Fatalf("right payload start=%d, want 0", rightAfter.payloadStart)
	}
	if rightAfter.payloadLen != len(rightPayload) {
		t.Fatalf("right payload len=%d, want %d", rightAfter.payloadLen, len(rightPayload))
	}
	if !bytes.Equal(reader.sourcePayload.Bytes()[:rightAfter.payloadLen], rightPayload) {
		t.Fatal("right payload bytes changed after reclaim")
	}
}
