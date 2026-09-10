package torrstor

import (
	"testing"
	"time"
)

// client is a player as one actually behaves: it holds up to a fixed amount of film, pulls
// more whenever it has room and the network allows, and empties what it holds only while the
// picture is running. Scripting the delivery by hand instead produces sessions no client
// could have — a buffer sitting empty through steady playback, or a full one still topping
// up — and the tracker is then judged against something that never happens.
type client struct {
	capacity float64 // film it can hold, in seconds
	buffer   float64 // film it is holding
	head     float64 // film time delivered so far
	screen   float64 // film time shown so far
}

// tick advances one second: the network offers up to supply seconds of film, and the picture
// takes one second if it is running. It reports which side was the limit — whether the client
// had nowhere to put what was on offer, or the supply ran out first — because that is what the
// server sees from where it sits, and the reckoning leans on it.
func (c *client) tick(supply float64, playing bool) (clientLimited bool) {
	room := c.capacity - c.buffer
	if playing {
		room++ // what is about to be watched leaves room for more
	}
	got := supply
	if got > room {
		got = room
		clientLimited = true
	}
	if got < 0 {
		got = 0
	}
	c.buffer += got
	c.head += got
	if playing {
		watched := 1.0
		if c.buffer < watched {
			watched = c.buffer // starved: the picture only moves as far as it can
		}
		c.buffer -= watched
		c.screen += watched
	}
	return
}

// steady is a file at one bitrate: the least a test needs to turn bytes into film time and
// back. The size is kept in bytes precisely because real files are not steady, and the test
// that cares about that supplies its own.
type steady float64

func (s steady) TimeAt(off int64) (float64, bool)   { return float64(off) / float64(s), true }
func (s steady) OffsetAt(sec float64) (int64, bool) { return int64(sec * float64(s)), true }

// bytesPerSec is an arbitrary rate: nothing in the reckoning depends on its value.
const bytesPerSec steady = 1 << 20

// waits turns "which side was the limit" into the pair of durations the reader measures.
func waits(clientLimited bool) (client, supply time.Duration) {
	if clientLimited {
		return 800 * time.Millisecond, 200 * time.Millisecond
	}
	return 200 * time.Millisecond, 800 * time.Millisecond
}

type phase struct {
	seconds int
	supply  float64 // film seconds offered per second
	playing bool
}

func run(t *testing.T, capacity float64, phases []phase) (screen, truth, held float64) {
	t.Helper()
	c := &client{capacity: capacity}
	var p progress
	now := time.Unix(1000000, 0)
	p.start(0, 0, now, 0, 0, bytesPerSec)

	for _, ph := range phases {
		for i := 0; i < ph.seconds; i++ {
			held := c.tick(ph.supply, ph.playing)
			now = now.Add(time.Second)
			p.waited(waits(held))
			p.step(c.head, int64(c.head*float64(bytesPerSec)), bytesPerSec, now)
		}
	}
	got, ok := p.screen()
	if !ok {
		t.Fatal("no position")
	}
	return got, c.screen, p.held()
}

func TestProgress(t *testing.T) {
	const cap = 30 // the client holds half a minute of film

	for _, tc := range []struct {
		name   string
		phases []phase
		lag    float64 // how far behind the picture the answer may be
		ahead  float64 // and how far in front
	}{
		{
			name: "steady playback",
			phases: []phase{
				{40, 4, true},  // fills while playing
				{180, 1, true}, // then takes only what it watches
			},
			lag: 2, ahead: 1,
		},
		{
			name: "paused with a full buffer",
			phases: []phase{
				{40, 4, true}, {120, 1, true},
				{120, 4, false}, // paused: it is already full, so nothing arrives
				{60, 1, true},
			},
			lag: 3, ahead: 1,
		},
		{
			// Played a little, then paused before the fill was over. The stop is read as
			// "nothing has been shown yet", so the seconds that were shown are counted as
			// still in hand and the answer sits that far back. Bounded by how long the fill
			// lasts, and on the side that is allowed: it used to be able to land three
			// seconds in front instead.
			name: "paused while still filling",
			phases: []phase{
				{10, 4, true},   // barely started, buffer part full
				{120, 4, false}, // paused: tops up, then goes quiet
				{60, 1, true},
			},
			lag: 12, ahead: 1,
		},
		{
			// Paused before a single frame has been shown, to let the torrent warm up. The
			// head races away while the picture stands still, and nothing in the stream says
			// so at the time — what says so is the head stopping, which means the client is
			// full, which is an absolute reading of where the picture must be.
			name: "paused from the very start, warming up",
			phases: []phase{
				{180, 4, false}, // fills, then goes quiet, with nothing shown at all
				{120, 1, true},  // and only then does it start playing
			},
			lag: 6, ahead: 1,
		},
		{
			name: "supply dies under a running picture",
			phases: []phase{
				{40, 4, true}, {120, 1, true},
				{60, 0, true}, // buffer drains, picture keeps going
				{90, 4, true}, // then catches up
			},
			lag: 35, ahead: 1,
		},
		{
			name: "supply below what is watched",
			phases: []phase{
				{40, 4, true}, {120, 1, true},
				{120, 0.6, true}, // slower than the picture: eats into the buffer
				{60, 4, true},
			},
			lag: 35, ahead: 1,
		},
		{
			name: "supply arrives in bursts",
			phases: func() []phase {
				out := []phase{{40, 4, true}, {60, 1, true}}
				for i := 0; i < 40; i++ {
					if i%4 == 0 {
						out = append(out, phase{1, 4, true})
					} else {
						out = append(out, phase{1, 0, true})
					}
				}
				return out
			}(),
			lag: 6, ahead: 1,
		},
		{
			name: "paused twice",
			phases: []phase{
				{40, 4, true}, {90, 1, true},
				{90, 4, false}, {30, 1, true},
				{90, 4, false}, {30, 1, true},
			},
			lag: 3, ahead: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			screen, truth, held := run(t, cap, tc.phases)
			t.Logf("показано %.0fс · счётчик %.0fс (%+.0f) · держит %.0fс",
				truth, screen, screen-truth, held)
			if screen > truth+tc.ahead+0.001 {
				t.Fatalf("картинка на %.0fс, счётчик говорит %.0fс — впереди показанного", truth, screen)
			}
			if truth-screen > tc.lag {
				t.Fatalf("картинка на %.0fс, счётчик говорит %.0fс — отстаёт больше допустимого %.0fс",
					truth, screen, tc.lag)
			}
		})
	}
}

// A player that pauses long enough loses its connection and opens another from the same byte,
// still holding everything it had. The new connection is told what it inherited; without that
// it starts from the only assumption available to it — that nothing is buffered — and reads
// the buffer as whatever it manages to add afterwards. Reported from a real session: 350MB
// before the pause, 60MB after.
func TestInheritsBufferAcrossAReconnect(t *testing.T) {
	const capacity = 30

	c := &client{capacity: capacity}
	now := time.Unix(1000000, 0)

	// A first connection, playing steadily with a full buffer.
	var first progress
	first.start(0, 0, now, 0, 0, bytesPerSec)
	for i := 0; i < 200; i++ {
		held := c.tick(4, true)
		now = now.Add(time.Second)
		first.waited(waits(held))
		first.step(c.head, int64(c.head*float64(bytesPerSec)), bytesPerSec, now)
	}
	held := first.handOn()
	if want := int64((capacity - 2) * float64(bytesPerSec)); held < want {
		t.Fatalf("the first connection should have found the buffer full, it says %dМБ", held>>20)
	}

	// It drops while the picture is paused, and the client sits on what it has.
	now = now.Add(2 * time.Minute)

	// The next connection carries on from the same byte, and is told what came with it.
	var second progress
	second.start(c.head, int64(c.head*float64(bytesPerSec)), now, held, held, bytesPerSec)
	for i := 0; i < 120; i++ {
		held := c.tick(1, true)
		now = now.Add(time.Second)
		second.waited(waits(held))
		second.step(c.head, int64(c.head*float64(bytesPerSec)), bytesPerSec, now)
	}

	screen, _ := second.screen()
	if screen > c.screen+1 {
		t.Fatalf("картинка на %.0fс, счётчик говорит %.0fс — впереди показанного", c.screen, screen)
	}
	if c.screen-screen > 3 {
		t.Fatalf("картинка на %.0fс, счётчик говорит %.0fс — отстаёт слишком сильно", c.screen, screen)
	}
	t.Logf("после переподключения: показано %.0fс, счётчик %.0fс, унаследовано %.0fс",
		c.screen, screen, float64(held)/float64(bytesPerSec))
}

// A buffer handed over by an earlier connection is decided on before this point: whether it
// belongs to whoever turned up is settled by where they turned up, in takeOver. What is left
// to check here is the cost of getting that wrong, because it is the whole reason the rule is
// allowed to be generous. Handed a buffer a client is not holding, the reckoning must still
// never place the picture past where it really is.
func TestAWrongCarryOnlyEverLandsBehind(t *testing.T) {
	const capacity = 30.0
	const carried = int64(capacity * float64(bytesPerSec)) // the same half-minute, in bytes

	t.Run("the same client, still holding it", func(t *testing.T) {
		c := &client{capacity: capacity, head: 500, screen: 500}
		for i := 0; i < 40; i++ { // fills up
			c.tick(4, true)
		}
		var p progress
		now := time.Unix(1000000, 0)
		p.start(c.head, int64(c.head*float64(bytesPerSec)), now, carried, carried, bytesPerSec) // reconnects with the buffer still in hand

		for i := 0; i < 120; i++ { // full, so it takes only what it watches
			held := c.tick(1, true)
			now = now.Add(time.Second)
			p.waited(waits(held))
			p.step(c.head, int64(c.head*float64(bytesPerSec)), bytesPerSec, now)
		}
		screen, _ := p.screen()
		if screen > c.screen+1 || c.screen-screen > 3 {
			t.Fatalf("картинка на %.0fс, счётчик говорит %.0fс — перенос не удержан", c.screen, screen)
		}
		t.Logf("перенос удержан: показано %.0fс, счётчик %.0fс", c.screen, screen)
	})

	t.Run("someone else, starting empty", func(t *testing.T) {
		c := &client{capacity: capacity, head: 500, screen: 500} // nothing buffered
		var p progress
		now := time.Unix(1000000, 0)
		p.start(c.head, int64(c.head*float64(bytesPerSec)), now, carried, carried, bytesPerSec) // handed a buffer that is not its own

		for i := 0; i < 40; i++ { // fills from empty, far faster than it watches
			held := c.tick(4, true)
			now = now.Add(time.Second)
			p.waited(waits(held))
			p.step(c.head, int64(c.head*float64(bytesPerSec)), bytesPerSec, now)
		}
		for i := 0; i < 90; i++ {
			held := c.tick(1, true)
			now = now.Add(time.Second)
			p.waited(waits(held))
			p.step(c.head, int64(c.head*float64(bytesPerSec)), bytesPerSec, now)
		}
		screen, _ := p.screen()
		if screen > c.screen+1 {
			t.Fatalf("картинка на %.0fс, счётчик говорит %.0fс — впереди показанного", c.screen, screen)
		}
		t.Logf("чужой буфер стоил отставания: показано %.0fс, счётчик %.0fс (на %.0fс позади)",
			c.screen, screen, c.screen-screen)
	})
}

// Where a connection begins is what says whether the client is the one that was watching. A
// player that lost its connection carries on from inside what it already had; a seek lands
// minutes away and must inherit nothing, or the position it reports is a whole buffer behind
// somewhere it was never playing.
func TestHandoverGoesOnlyToTheSamePlayback(t *testing.T) {
	c := &Cache{handovers: map[string]*handover{}}
	const path = "film.mkv"
	c.noteRead(path, 1, 30<<20, 30<<20, 600<<20, 600) // picture at 600s, holding 30s, so served 600..630

	for _, tc := range []struct {
		name  string
		start int64
		want  int64 // what the client still has in hand there
	}{
		{"resumes at the old read head, so it kept the lot", 630 << 20, 30 << 20},
		{"resumes halfway, so half of it died with the line", 615 << 20, 15 << 20},
		{"resumes at its own picture, holding nothing", 600 << 20, 0},
		{"seeks forward", 2400 << 20, 0},
		{"seeks back", 120 << 20, 0},
	} {
		got, _, ok := c.takeOver(path, tc.start)
		if ok != (tc.want > 0) || got != tc.want {
			t.Fatalf("%s: старт на %dМБ — перенесено %dМБ (%v), ждали %dМБ",
				tc.name, tc.start>>20, got>>20, ok, tc.want>>20)
		}
	}
}
