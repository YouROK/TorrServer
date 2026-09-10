package timeindex

import (
	"encoding/binary"
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The sample is built so that the average bitrate is useless: half a minute of busy picture
// followed by half a minute of black, which puts 99% of the bytes in the first half. Judging
// a byte offset by the file's average would put the second half an entire half-minute out.
func buildSample(t testing.TB, dir string) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	src := filepath.Join(dir, "src.mkv")
	run(t, "ffmpeg", "-y", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc2=size=640x480:rate=25:duration=30",
		"-f", "lavfi", "-i", "color=c=black:size=640x480:rate=25:duration=30",
		"-filter_complex", "[0:v][1:v]concat=n=2:v=1:a=0[v]", "-map", "[v]",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "18", "-g", "50",
		"-pix_fmt", "yuv420p", src)
	run(t, "ffmpeg", "-y", "-loglevel", "error", "-i", src, "-c", "copy", filepath.Join(dir, "src.ts"))
	run(t, "ffmpeg", "-y", "-loglevel", "error", "-i", src, "-c", "copy",
		"-movflags", "+faststart", filepath.Join(dir, "src.mp4"))
	run(t, "ffmpeg", "-y", "-loglevel", "error", "-i", src, "-c", "copy",
		"-movflags", "+frag_keyframe+empty_moov", filepath.Join(dir, "frag.mp4"))
	// The old shapes need an old codec: a program stream carries MPEG-2, FLV its own.
	run(t, "ffmpeg", "-y", "-loglevel", "error", "-i", src, "-c:v", "mpeg2video",
		"-qscale:v", "4", "-g", "15", "-f", "vob", filepath.Join(dir, "src.vob"))
	run(t, "ffmpeg", "-y", "-loglevel", "error", "-i", src, "-c:v", "flv",
		"-qscale:v", "4", "-f", "flv", filepath.Join(dir, "src.flv"))
	run(t, "ffmpeg", "-y", "-loglevel", "error", "-i", src, "-c:v", "mpeg4",
		"-qscale:v", "4", "-g", "250", filepath.Join(dir, "src.avi"))
}

func run(t testing.TB, name string, args ...string) string {
	t.Helper()
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return string(out)
}

// feedChunks streams data into the feeder from the given offset, 64 KB at a time, the way a
// player's reads arrive.
func feedChunks(feeder *Feeder, data []byte, from int) {
	const chunk = 64 << 10
	for at := from; at < len(data); at += chunk {
		feeder.Feed(int64(at), data[at:min(at+chunk, len(data))])
	}
}

type truth struct {
	sec float64
	pos int64
}

// groundTruth asks ffprobe where each video packet sits and what time it plays at.
func groundTruth(t testing.TB, path string) ([]truth, float64) {
	t.Helper()
	out := run(t, "ffprobe", "-v", "error", "-select_streams", "v",
		"-show_entries", "packet=pts_time,pos", "-of", "csv=p=0", path)
	rows, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	if err != nil {
		t.Fatalf("ffprobe output: %v", err)
	}
	var list []truth
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		sec, err1 := strconv.ParseFloat(row[0], 64)
		pos, err2 := strconv.ParseInt(row[1], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		list = append(list, truth{sec: sec, pos: pos})
	}
	start, _ := strconv.ParseFloat(strings.TrimSpace(run(t, "ffprobe", "-v", "error",
		"-show_entries", "format=start_time", "-of", "csv=p=0", path)), 64)
	return list, start
}

func TestIndexReadsTimeFromTheStream(t *testing.T) {
	dir := t.TempDir()
	buildSample(t, dir)

	for _, tc := range []struct {
		file      string
		source    string
		tolerance float64 // how far behind the real time a lookup may be, in seconds
	}{
		{"src.mkv", "matroska", 5.5}, // clusters run up to five seconds apart
		{"src.ts", "mpegts", 4},
		{"src.mp4", "mp4", 1},
		{"frag.mp4", "mp4", 4},
		{"src.vob", "mpegps", 1},
		{"src.flv", "flv", 1},
		{"src.avi", "avi", 1},
	} {
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join(dir, tc.file)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			packets, start := groundTruth(t, path)
			if len(packets) == 0 {
				t.Fatal("no packets from ffprobe")
			}

			ix := New(path)
			if ix == nil {
				t.Fatalf("no parser for %s", tc.file)
			}
			if got := ix.Source(); got != tc.source {
				t.Fatalf("source = %q, want %q", got, tc.source)
			}
			ix.SetOrigin(start)
			feeder := ix.Feeder()

			// Fed the way a player reads it: in chunks, from the beginning.
			feedChunks(feeder, data, 0)

			duration := packets[len(packets)-1].sec
			var worst, worstLinear float64
			checked := 0
			for _, p := range packets {
				got, ok := ix.TimeAt(p.pos)
				if !ok {
					continue
				}
				checked++
				if got > p.sec+0.001 {
					t.Fatalf("offset %d: index says %.3fs, real time is %.3fs — ahead of the picture",
						p.pos, got, p.sec)
				}
				if behind := p.sec - got; behind > worst {
					worst = behind
				}
				linear := float64(p.pos) / float64(len(data)) * duration
				if off := linear - p.sec; off > worstLinear {
					worstLinear = off
				}
			}
			if checked < len(packets)/2 {
				t.Fatalf("only %d of %d packets resolved", checked, len(packets))
			}
			if worst > tc.tolerance {
				t.Fatalf("worst lag %.2fs exceeds %.2fs", worst, tc.tolerance)
			}
			t.Logf("%s: %d packets, worst lag %.2fs; the average-bitrate guess would run %.2fs ahead",
				tc.file, checked, worst, worstLinear)
		})
	}
}

// A player that resumes reads the header first and then jumps, on a separate connection.
// The index is shared per file, so it sees both, and the jump must not confuse the parser.
func TestIndexSurvivesASeek(t *testing.T) {
	dir := t.TempDir()
	buildSample(t, dir)
	path := filepath.Join(dir, "src.mkv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	packets, _ := groundTruth(t, path)

	ix := New(path)
	feeder := ix.Feeder()
	feeder.Feed(0, data[:256<<10]) // the header read

	seek := len(data) / 2
	feedChunks(feeder, data, seek)

	found := 0
	for _, p := range packets {
		if p.pos < int64(seek) {
			continue
		}
		got, ok := ix.TimeAt(p.pos)
		if !ok {
			continue
		}
		found++
		if got > p.sec+0.001 {
			t.Fatalf("offset %d after a seek: index says %.3fs, real time is %.3fs", p.pos, got, p.sec)
		}
	}
	if found == 0 {
		t.Fatal("nothing resolved after the seek")
	}
	t.Logf("resolved %d packets past the seek", found)
}

// Over tens of gigabytes the four bytes that mark a Matroska cluster turn up inside frame
// data by chance, and one that is followed by something shaped like a timestamp used to be
// filed as if it were real — which is how a film two hours in reported its fourth minute.
// Here one is planted deliberately.
func TestStrayHeaderInFrameData(t *testing.T) {
	dir := t.TempDir()
	buildSample(t, dir)
	path := filepath.Join(dir, "src.mkv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	packets, _ := groundTruth(t, path)

	// A cluster header claiming the fourth second, planted deep inside the picture data of a
	// file that is by then half a minute in. It parses: an id, a length, then a timestamp.
	plant := []byte{0x1F, 0x43, 0xB6, 0x75, 0xA0, 0xE7, 0x81, 0x04}
	at := len(data) / 2
	copy(data[at:], plant)

	ix := New(path)
	feeder := ix.Feeder()
	feedChunks(feeder, data, 0)

	for _, p := range packets {
		if p.pos < int64(at) {
			continue
		}
		got, ok := ix.TimeAt(p.pos)
		if !ok {
			continue
		}
		if got > p.sec+0.001 {
			t.Fatalf("offset %d: index says %.3fs, real time is %.3fs", p.pos, got, p.sec)
		}
		// The planted header would drag everything after it back to four seconds.
		if p.sec > 20 && got < 4.5 {
			t.Fatalf("offset %d plays at %.1fs but the index fell back to %.1fs — the stray header was believed",
				p.pos, p.sec, got)
		}
	}
}

// A player opens more than one connection to the same file: the header on one, the picture
// on another, and a probe alongside. Fed through a single parser those interleave into what
// looks like a stream jumping about, and Matroska — whose clusters have to be followed in
// order — stops indexing altogether. That is what froze the read head's time in place while
// the head itself ran on, and made the buffer appear to grow without end.
func TestTwoStreamsOnOneFile(t *testing.T) {
	dir := t.TempDir()
	buildSample(t, dir)
	path := filepath.Join(dir, "src.mkv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	packets, _ := groundTruth(t, path)

	ix := New(path)
	picture, probe := ix.Feeder(), ix.Feeder()

	const chunk = 64 << 10
	for at := 0; at < len(data); at += chunk {
		end := at + chunk
		if end > len(data) {
			end = len(data)
		}
		picture.Feed(int64(at), data[at:end])
		// The other connection keeps rereading the front of the file, the way a probe does.
		probe.Feed(int64((at/chunk%4)*chunk), data[(at/chunk%4)*chunk:(at/chunk%4+1)*chunk])
	}

	// Coverage is what suffers: with one parser between them most clusters are dropped, and
	// the gaps left behind are what freeze a lookup in place while the read head moves on.
	var worst float64
	for _, p := range packets {
		got, ok := ix.TimeAt(p.pos)
		if !ok {
			continue
		}
		if got > p.sec+0.001 {
			t.Fatalf("offset %d: index says %.3fs, real time is %.3fs", p.pos, got, p.sec)
		}
		if behind := p.sec - got; behind > worst {
			worst = behind
		}
	}
	count := len(ix.samples)
	if worst > 5.5 {
		t.Fatalf("worst lag %.1fs across the file, from %d timestamps — clusters are being dropped", worst, count)
	}
	t.Logf("%d timestamps indexed, worst lag %.1fs", count, worst)
}

// The three bytes that mark TimestampScale turn up in picture data like any other pattern,
// and the number behind them then scales every timestamp in the file. One planted deep in the
// stream used to turn a two-hour film into a sixty-eight-hour one — and because the inflated
// times still rose in order, nothing further along noticed.
func TestStrayScaleInFrameData(t *testing.T) {
	dir := t.TempDir()
	buildSample(t, dir)
	path := filepath.Join(dir, "src.mkv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	packets, _ := groundTruth(t, path)

	// TimestampScale, four bytes long, claiming 188 million nanoseconds per tick.
	plant := []byte{0x2A, 0xD7, 0xB1, 0x84, 0x0B, 0x35, 0x0D, 0xC0}
	copy(data[len(data)/3:], plant)

	ix := New(path)
	ix.SetDuration(packets[len(packets)-1].sec)
	feeder := ix.Feeder()
	feedChunks(feeder, data, 0)

	for _, p := range packets {
		got, ok := ix.TimeAt(p.pos)
		if !ok {
			continue
		}
		if got > p.sec+0.001 {
			t.Fatalf("offset %d: index says %.1fs, real time is %.1fs — the planted scale was believed",
				p.pos, got, p.sec)
		}
	}
}

// The names of MP4 boxes are found by searching, so a stray match inside picture data is
// followed by arbitrary bytes. One shape of those — a box claiming its length is stated as
// 64 bits, with fewer than sixteen bytes behind it — used to leave the walk resliced from
// nothing, spinning for ever under the index lock inside a read.
func TestMalformedBoxesDoNotSpin(t *testing.T) {
	ix := New("x.mp4")
	feeder := ix.Feeder()

	frame := make([]byte, 4<<10)
	for i := range frame {
		frame[i] = byte(i)
	}
	// A length of one says "the real length follows as 64 bits", and then it does not.
	plant := []byte{0x00, 0x00, 0x00, 0x01, 'm', 'o', 'o', 'v', 0x00, 0x00, 0x00, 0x01}
	copy(frame[1024:], plant)

	done := make(chan struct{})
	go func() {
		feeder.Feed(0, frame)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("разбор не вернулся: обход боксов не двигается вперёд")
	}
}

// An FLV header states where its first tag begins. Counted as a 32-bit signed number that
// turns negative past two gigabytes, and a negative start reaches off the front of the slice.
func TestFLVHeaderOffsetOutOfRange(t *testing.T) {
	for _, offset := range []uint32{0xFFFFFFFF, 0x80000000, 0, 3} {
		buf := make([]byte, 4<<10)
		copy(buf, "FLV")
		buf[3], buf[4] = 1, 5
		binary.BigEndian.PutUint32(buf[5:9], offset)
		feeder := New("x.flv").Feeder()
		feeder.Feed(0, buf) // паника здесь и есть отказ
	}
}
