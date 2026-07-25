//go:build gst

package gstreamer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
)

func testTfhd(trackID uint32) canonicalTfhd {
	return buildCanonicalTfhd(trackID, 1)
}

func testTrun(duration uint32, size uint32, flags uint32) []byte {
	const trunFlags = trunDataOffsetPresent |
		trunSampleDurationPresent |
		trunSampleSizePresent |
		trunSampleFlagsPresent
	const boxSize = 32

	box := make([]byte, boxSize)
	binary.BigEndian.PutUint32(box[0:4], boxSize)
	binary.BigEndian.PutUint32(box[4:8], boxTrun)
	binary.BigEndian.PutUint32(box[8:12], trunFlags)
	binary.BigEndian.PutUint32(box[12:16], 1)
	binary.BigEndian.PutUint32(box[16:20], 0)
	binary.BigEndian.PutUint32(box[20:24], duration)
	binary.BigEndian.PutUint32(box[24:28], size)
	binary.BigEndian.PutUint32(box[28:32], flags)
	return box
}

func testFragment(
	trackID uint32,
	timescale uint32,
	decodeTime uint64,
	duration uint32,
	payloadSize uint32,
	sync bool,
	fill byte,
) mp4Fragment {
	flags := uint32(0)
	if !sync {
		flags = 0x00010000
	}

	run := mp4Run{
		samples: []mp4Sample{
			{
				duration: duration,
				size:     payloadSize,
				flags:    flags,
			},
		},
		duration:       uint64(duration),
		dataSize:       uint64(payloadSize),
		payloadOffset:  0,
		startsWithSync: sync,
	}

	return mp4Fragment{
		trackID:        trackID,
		timescale:      timescale,
		decodeTime:     decodeTime,
		duration:       uint64(duration),
		startsWithSync: sync,
		tfhd:           testTfhd(trackID),
		runs:           []mp4Run{run},
		payloadLen:     int(payloadSize),
	}
}

func appendTestFragmentPayload(reader *mp4BoxReader, fragment mp4Fragment, fill byte) mp4Fragment {
	payload := bytes.Repeat([]byte{fill}, fragment.payloadLen)
	fragment.payloadStart = reader.sourcePayload.Len()
	_, _ = reader.sourcePayload.Write(payload)
	return fragment
}

func newTestEOSRunner(segmentSeconds float64) (*gstRunner, *mp4BoxReader, *[]Segment) {
	task := &Task{Config: Config{SegmentSeconds: int(segmentSeconds)}.normalized()}
	runner := &gstRunner{task: task}
	task.runner = runner
	segments := make([]Segment, 0, 4)

	reader := Mp4BoxReader(
		func([]byte) {},
		func(seg Segment) {
			segments = append(segments, seg)
			runner.readySegment.segment = seg
			runner.readySegment.complete = true
		},
		segmentSeconds,
		0,
		false,
	)

	runner.reader = reader
	return runner, reader, &segments
}

func resetTestReady(runner *gstRunner) {
	runner.discardReadySegment()
}

func testU32(values ...uint32) []byte {
	out := make([]byte, len(values)*4)
	for i, value := range values {
		binary.BigEndian.PutUint32(out[i*4:(i+1)*4], value)
	}
	return out
}

func testBox(boxType uint32, payload ...[]byte) []byte {
	size := 8
	for _, part := range payload {
		size += len(part)
	}

	out := make([]byte, size)
	binary.BigEndian.PutUint32(out[0:4], uint32(size))
	binary.BigEndian.PutUint32(out[4:8], boxType)

	position := 8
	for _, part := range payload {
		copy(out[position:], part)
		position += len(part)
	}
	return out
}

func testFullBox(boxType uint32, versionFlags uint32, body []byte) []byte {
	fullBody := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(fullBody[0:4], versionFlags)
	copy(fullBody[4:], body)
	return testBox(boxType, fullBody)
}

func TestLinearGrowthBufferDoesNotDoubleCapacity(t *testing.T) {
	const mib = 1024 * 1024
	for _, test := range []struct {
		required int
		want     int
	}{
		{required: 40 * mib, want: 40 * mib},
		{required: 40*mib + 1, want: 41 * mib},
		{required: 41 * mib, want: 41 * mib},
		{required: 41*mib + 1, want: 42 * mib},
	} {
		if got := roundedBufferCapacity(test.required); got != test.want {
			t.Fatalf("rounded capacity for %d = %d, want %d", test.required, got, test.want)
		}
	}

	var buffer linearGrowthBuffer
	buffer.Grow(65 * 1024)
	if got, want := buffer.Cap(), 128*1024; got != want {
		t.Fatalf("initial capacity = %d, want %d", got, want)
	}
	buffer.Reset()
	buffer.Grow(129 * 1024)
	if got, want := buffer.Cap(), 192*1024; got != want {
		t.Fatalf("grown capacity = %d, want %d", got, want)
	}
}

func TestEquivalentTfhdTrunRepresentationsCanMerge(t *testing.T) {
	videoTrack := trackInfo{
		id:        1,
		timescale: 1000,
		trex: trexInfo{
			descriptionIndex: 1,
		},
	}

	tfhdA := testFullBox(
		boxTfhd,
		tfhdDefaultSampleDurationPresent|
			tfhdDefaultSampleSizePresent|
			tfhdDefaultSampleFlagsPresent,
		testU32(
			1,
			40,
			100,
			0,
		),
	)
	tfdtA := testFullBox(boxTfdt, 0, testU32(0))
	trunA := testFullBox(
		boxTrun,
		trunDataOffsetPresent,
		testU32(
			2,
			0,
		),
	)

	tfhdB := testFullBox(
		boxTfhd,
		tfhdSampleDescriptionIndexPresent,
		testU32(
			1,
			1,
		),
	)
	tfdtB := testFullBox(boxTfdt, 0, testU32(80))
	trunB := testFullBox(
		boxTrun,
		trunDataOffsetPresent|
			trunSampleDurationPresent|
			trunSampleSizePresent|
			trunSampleFlagsPresent,
		testU32(
			2,
			0,
			40, 100, 0,
			40, 100, 0,
		),
	)

	var fragmentA mp4Fragment
	err := parseTraf(
		testBox(boxTraf, tfhdA, tfdtA, trunA),
		8,
		videoTrack,
		trackInfo{},
		0,
		0,
		&fragmentA,
	)
	if err != nil {
		t.Fatal(err)
	}

	var fragmentB mp4Fragment
	err = parseTraf(
		testBox(boxTraf, tfhdB, tfdtB, trunB),
		8,
		videoTrack,
		trackInfo{},
		0,
		0,
		&fragmentB,
	)
	if err != nil {
		t.Fatal(err)
	}

	if fragmentA.tfhd != fragmentB.tfhd {
		t.Fatal("canonical tfhd differs")
	}
	if fragmentA.runs[0].version != fragmentB.runs[0].version ||
		fragmentA.runs[0].hasCompositionTimeOffsets != fragmentB.runs[0].hasCompositionTimeOffsets ||
		!slices.Equal(fragmentA.runs[0].samples, fragmentB.runs[0].samples) {
		t.Fatal("canonical trun differs")
	}

	if err := validateTrack([]mp4Fragment{fragmentA, fragmentB}, 2); err != nil {
		t.Fatalf("semantically equivalent fragments do not merge: %v", err)
	}
}

func TestWriteCanonicalTrunMatchesCanonicalLayout(t *testing.T) {
	run := mp4Run{
		version: 0,
		samples: []mp4Sample{
			{duration: 1000, size: 3, flags: 0x01020304},
		},
	}

	var output bytes.Buffer
	if err := writeCanonicalTrun(&output, &run, 1234); err != nil {
		t.Fatal(err)
	}

	want := testTrun(1000, 3, 0x01020304)
	binary.BigEndian.PutUint32(want[16:20], 1234)
	if !bytes.Equal(output.Bytes(), want) {
		t.Fatalf("canonical trun = %x, want %x", output.Bytes(), want)
	}
}

func BenchmarkWriteCanonicalTrun(b *testing.B) {
	samples := make([]mp4Sample, 300)
	for i := range samples {
		samples[i] = mp4Sample{duration: 1024, size: 512, flags: 0}
	}
	run := mp4Run{samples: samples}

	var output bytes.Buffer
	output.Grow(int(canonicalTrunSize(run)))
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		output.Reset()
		if err := writeCanonicalTrun(&output, &run, 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildSegmentSteadyState(b *testing.B) {
	const sampleCount = 150
	const sampleSize = 512

	samples := make([]mp4Sample, sampleCount)
	for i := range samples {
		samples[i] = mp4Sample{duration: 1000, size: sampleSize, flags: 0}
	}
	payload := make([]byte, sampleCount*sampleSize)
	fragment := mp4Fragment{
		trackID:                1,
		timescale:              1000,
		decodeTime:             0,
		duration:               sampleCount * 1000,
		startsWithSync:         true,
		sampleDescriptionIndex: 1,
		tfhd:                   buildCanonicalTfhd(1, 1),
		runs: []mp4Run{
			{
				samples:        samples,
				duration:       sampleCount * 1000,
				dataSize:       sampleCount * sampleSize,
				startsWithSync: true,
			},
		},
		payloadLen: len(payload),
	}

	reader := Mp4BoxReader(func([]byte) {}, func(Segment) {}, 6, 0, false)
	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.video = make([]mp4Fragment, 0, 1)
	reader.sourcePayload.Grow(len(payload))

	build := func() {
		fragment.decodeTime = uint64(reader.sequence-1) * fragment.duration
		fragment.payloadStart = reader.sourcePayload.Len()
		_, _ = reader.sourcePayload.Write(payload)
		reader.video = append(reader.video, fragment)
		if err := reader.buildSegment(1, 0, false); err != nil {
			b.Fatal(err)
		}
		reader.ReleaseSegment()
	}

	build()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		build()
	}
}

func BenchmarkParseTraf(b *testing.B) {
	const sampleCount = 300
	const trunFlags = trunDataOffsetPresent |
		trunSampleDurationPresent |
		trunSampleSizePresent |
		trunSampleFlagsPresent

	body := make([]byte, 8+sampleCount*12)
	binary.BigEndian.PutUint32(body[0:4], sampleCount)
	position := 8
	for range sampleCount {
		binary.BigEndian.PutUint32(body[position:position+4], 1024)
		binary.BigEndian.PutUint32(body[position+4:position+8], 512)
		binary.BigEndian.PutUint32(body[position+8:position+12], 0)
		position += 12
	}

	traf := testBox(
		boxTraf,
		testFullBox(boxTfhd, tfhdSampleDescriptionIndexPresent, testU32(1, 1)),
		testFullBox(boxTfdt, 0, testU32(0)),
		testFullBox(boxTrun, trunFlags, body),
	)
	videoTrack := trackInfo{id: 1, timescale: 48_000}
	var fragment mp4Fragment

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := parseTraf(traf, 8, videoTrack, trackInfo{}, 0, 0, &fragment); err != nil {
			b.Fatal(err)
		}
	}
}

func TestBuildSegmentAllowsVideoOnly(t *testing.T) {
	samples := []mp4Sample{
		{
			duration: 1000,
			size:     3,
			flags:    0,
		},
	}
	var got Segment
	reader := Mp4BoxReader(
		func([]byte) {},
		func(segment Segment) {
			got = segment
		},
		1,
		0,
		false,
	)

	reader.videoTrack = trackInfo{
		id:        1,
		timescale: 1000,
		trex: trexInfo{
			descriptionIndex: 1,
		},
	}
	reader.video = []mp4Fragment{
		appendTestFragmentPayload(reader, mp4Fragment{
			trackID:                1,
			timescale:              1000,
			decodeTime:             0,
			duration:               1000,
			startsWithSync:         true,
			sampleDescriptionIndex: 1,
			tfhd:                   buildCanonicalTfhd(1, 1),
			runs: []mp4Run{
				{
					samples:        samples,
					duration:       1000,
					dataSize:       3,
					startsWithSync: true,
				},
			},
			payloadLen: 3,
		}, 0x01),
	}

	if err := reader.buildSegment(1, 0, false); err != nil {
		t.Fatal(err)
	}
	if got.Empty() {
		t.Fatal("video-only segment is empty")
	}
	if len(got.Payloads) != 1 {
		t.Fatalf("Payloads=%d, want 1", len(got.Payloads))
	}
}

func TestReleaseSegmentClearsPayloadReferencesAndKeepsSliceCapacity(t *testing.T) {
	reader := Mp4BoxReader(func([]byte) {}, func(Segment) {}, 6, 0, false)
	_, _ = reader.sourcePayload.Write([]byte("payload"))
	reader.segmentPayloads = append(reader.segmentPayloads, reader.sourcePayload.Bytes())
	capacity := cap(reader.segmentPayloads)

	reader.ReleaseSegment()

	if len(reader.segmentPayloads) != 0 {
		t.Fatalf("payload ranges length = %d, want 0", len(reader.segmentPayloads))
	}
	if cap(reader.segmentPayloads) != capacity {
		t.Fatalf("payload ranges capacity = %d, want %d", cap(reader.segmentPayloads), capacity)
	}
	for i, payload := range reader.segmentPayloads[:capacity] {
		if payload != nil {
			t.Fatalf("payload range %d still retains storage", i)
		}
	}
}

func TestSegmentCarriesEndSecondsForResume(t *testing.T) {
	var got Segment
	reader := Mp4BoxReader(
		func([]byte) {},
		func(seg Segment) {
			got = seg
		},
		6,
		0,
		false,
	)
	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.video = []mp4Fragment{
		appendTestFragmentPayload(reader, testFragment(1, 1000, 0, 1000, 10, true, 0x11), 0x11),
		appendTestFragmentPayload(reader, testFragment(1, 1000, 1000, 1000, 11, true, 0x22), 0x22),
	}

	if err := reader.buildSegment(2, 0, false); err != nil {
		t.Fatal(err)
	}
	if got.StartSeconds != 0 {
		t.Fatalf("start=%v, want 0", got.StartSeconds)
	}
	if got.EndSeconds != 2 {
		t.Fatalf("end=%v, want 2", got.EndSeconds)
	}
}

func TestDrainEndOfStreamOrdersRegularRemainderThenTruncation(t *testing.T) {
	runner, reader, segments := newTestEOSRunner(2)

	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.audioTrack = trackInfo{}

	reader.video = append(
		reader.video,
		appendTestFragmentPayload(reader, testFragment(1, 1000, 0, 1000, 10, true, 0x11), 0x11),
		appendTestFragmentPayload(reader, testFragment(1, 1000, 1000, 1000, 11, true, 0x22), 0x22),
		appendTestFragmentPayload(reader, testFragment(1, 1000, 2000, 500, 12, true, 0x33), 0x33),
	)

	reader.pendingStorage = mp4Fragment{
		trackID:    1,
		timescale:  1000,
		decodeTime: 2500,
	}
	reader.pending = &reader.pendingStorage
	_, _ = reader.sourceMoof.Write(make([]byte, 16))
	_, _ = reader.sourcePayload.Write([]byte{1, 2, 3})
	reader.currentBoxType = boxMdat
	reader.currentBoxRemaining = 7

	regular, err := runner.drainEndOfStream(10)
	if err != nil {
		t.Fatalf("regular EOS drain failed: %v", err)
	}
	if regular.StartSeconds != 0 {
		t.Fatalf("regular StartSeconds=%v, want 0", regular.StartSeconds)
	}
	if len(reader.video) != 1 {
		t.Fatalf("video fragments after regular=%d, want 1", len(reader.video))
	}

	resetTestReady(runner)

	remainder, err := runner.drainEndOfStream(11)
	if err != nil {
		t.Fatalf("EOS remainder failed: %v", err)
	}
	if remainder.StartSeconds != 2 {
		t.Fatalf("remainder StartSeconds=%v, want 2", remainder.StartSeconds)
	}
	if len(reader.video) != 0 {
		t.Fatalf("video fragments after remainder=%d, want 0", len(reader.video))
	}

	resetTestReady(runner)

	_, err = runner.drainEndOfStream(12)
	if !errors.Is(err, ErrTruncatedMP4Fragment) {
		t.Fatalf("final error=%v, want ErrTruncatedMP4Fragment", err)
	}

	if len(*segments) != 2 {
		t.Fatalf("emitted segments=%d, want 2", len(*segments))
	}
}

func TestTryBuildEndOfStreamRemainderAllowsAudioOnly(t *testing.T) {
	_, reader, segments := newTestEOSRunner(6)

	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.audioTrack = trackInfo{id: 2, timescale: 48000}
	reader.audio = append(
		reader.audio,
		appendTestFragmentPayload(reader, testFragment(2, 48000, 144000, 48000, 32, true, 0x44), 0x44),
	)

	completed, err := reader.TryBuildEndOfStreamRemainder()
	if err != nil {
		t.Fatal(err)
	}
	if !completed {
		t.Fatal("audio-only remainder was not completed")
	}
	if len(reader.audio) != 0 {
		t.Fatalf("remaining audio fragments=%d, want 0", len(reader.audio))
	}
	if len(*segments) != 1 {
		t.Fatalf("emitted segments=%d, want 1", len(*segments))
	}
	if (*segments)[0].StartSeconds != 3 {
		t.Fatalf("StartSeconds=%v, want 3", (*segments)[0].StartSeconds)
	}
}

func TestTryBuildEndOfStreamRemainderDoesNotEmitNonSyncVideo(t *testing.T) {
	_, reader, segments := newTestEOSRunner(6)

	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.video = append(
		reader.video,
		appendTestFragmentPayload(reader, testFragment(1, 1000, 1000, 1000, 16, false, 0x55), 0x55),
	)

	completed, err := reader.TryBuildEndOfStreamRemainder()
	if completed {
		t.Fatal("non-sync video remainder must not be emitted")
	}
	if err != nil {
		t.Fatalf("error=%v, want nil for a non-sync EOS tail", err)
	}
	if len(*segments) != 0 {
		t.Fatalf("emitted segments=%d, want 0", len(*segments))
	}
	if len(reader.video) != 1 {
		t.Fatalf("video fragments=%d, want retained 1", len(reader.video))
	}
}

func TestNormalizeTrunRejectsExcessiveSampleCountBeforeAllocation(t *testing.T) {
	box := make([]byte, 16)
	binary.BigEndian.PutUint32(box[0:4], uint32(len(box)))
	binary.BigEndian.PutUint32(box[4:8], boxTrun)
	binary.BigEndian.PutUint32(box[12:16], maxTrunSamples+1)

	_, err := normalizeTrun(box, 8, 1, 1, 0, false)
	if err == nil || !strings.Contains(err.Error(), "sample_count exceeds safety limit") {
		t.Fatalf("error=%v, want sample_count safety error", err)
	}
}

func TestDrainEndOfStreamReturnsUndecodableRemainder(t *testing.T) {
	runner, reader, _ := newTestEOSRunner(6)
	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.video = append(reader.video, appendTestFragmentPayload(reader, testFragment(1, 1000, 1000, 1000, 16, false, 0x55), 0x55))

	_, err := runner.drainEndOfStream(100)
	if !errors.Is(err, ErrUndecodableEOSRemainder) {
		t.Fatalf("error=%v, want ErrUndecodableEOSRemainder", err)
	}
}

func TestDrainEndOfStreamReturnsCleanExhaustion(t *testing.T) {
	runner, reader, _ := newTestEOSRunner(6)
	reader.videoTrack = trackInfo{id: 1, timescale: 1000}

	_, err := runner.drainEndOfStream(100)
	if !errors.Is(err, ErrEndOfStreamExhausted) {
		t.Fatalf("error=%v, want ErrEndOfStreamExhausted", err)
	}
}

func TestFinishEndOfStreamFreezesOnReaderError(t *testing.T) {
	runner, reader, _ := newTestEOSRunner(6)
	reader.videoTrack = trackInfo{id: 1, timescale: 1000}
	reader.video = append(reader.video, appendTestFragmentPayload(reader, testFragment(1, 1000, 1000, 1000, 16, false, 0x55), 0x55))
	runner.task.setInitMP4([]byte{1, 2, 3})

	_, err := runner.finishEndOfStream(100)
	if !errors.Is(err, ErrUndecodableEOSRemainder) {
		t.Fatalf("error=%v, want ErrUndecodableEOSRemainder", err)
	}
	if !runner.IsFrozen() || runner.reader != nil {
		t.Fatal("reader error did not freeze and release the runner")
	}
	if runner.task.hasInitMP4() {
		t.Fatal("reader error retained init mp4")
	}
}

func TestFinishEndOfStreamKeepsCleanExhaustion(t *testing.T) {
	runner, reader, _ := newTestEOSRunner(6)
	reader.videoTrack = trackInfo{id: 1, timescale: 1000}

	_, err := runner.finishEndOfStream(100)
	if !errors.Is(err, ErrEndOfStreamExhausted) {
		t.Fatalf("error=%v, want ErrEndOfStreamExhausted", err)
	}
	if runner.IsFrozen() || runner.reader != reader {
		t.Fatal("clean end of stream unexpectedly froze the runner")
	}
}

func TestAddTfdtOffsetNoOffset(t *testing.T) {
	got, err := addTfdtOffset(100, 1000, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 100 {
		t.Fatalf("got %d, want 100", got)
	}
}

func TestAddTfdtOffsetValid(t *testing.T) {
	got, err := addTfdtOffset(100, 1000, 1.5)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1600 {
		t.Fatalf("got %d, want 1600", got)
	}
}

func TestAddTfdtOffsetInvalidSeconds(t *testing.T) {
	if _, err := addTfdtOffset(100, 1000, math.Inf(1)); err == nil {
		t.Fatal("expected error")
	}
}

func TestAddTfdtOffsetOverflow(t *testing.T) {
	if _, err := addTfdtOffset(math.MaxUint64-10, 1000, 1); err == nil {
		t.Fatal("expected overflow error")
	}
}
