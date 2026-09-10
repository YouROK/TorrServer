package torrstor

import "time"

// progress follows where the picture is.
//
// It rests on one quantity, measured once: how much of the file the client is holding that it
// has not shown yet. That is a property of the device — a box of a fixed number of megabytes,
// set in the player — so it is kept in bytes and in nothing else. How much film fits in it is
// not fixed at all: the same three hundred megabytes ran from forty-two seconds down to
// twenty-eight across eight minutes of one film as the bitrate climbed. Time comes in only at
// the very end, by asking the file what plays at the byte that far behind the read head.
//
//	picture = the film at (read head − buffer)
//
// Nothing accumulates. An earlier version added up film against the clock second by second,
// and every one of the failures that took to find came out of that: readings that drifted,
// a largest-ever value that ratcheted and never came down, and a figure that grew by being
// handed from one connection to the next. A size measured once and then simply subtracted has
// nowhere to drift to.
//
// The measurement itself is the fill at the start of a connection. The client takes film
// faster than it can watch it until it has no room left, and what it took beyond what it could
// have watched in that time is the size of the box. Which of two things ends the fill is the
// whole of the reasoning:
//
//   - it settles to taking only what it watches — the client is playing and full, and what it
//     has watched by then is the clock since it started;
//   - it stops dead — the client was never watching. This is a player paused to let the
//     torrent warm up, and everything delivered is in its hands.
//
// Both look identical while the filling is going on, which is why the answer is only taken
// when it ends.
type progress struct {
	ref    float64 // film time this session started at
	refOff int64   // and the byte it started at
	begun  time.Time

	head float64   // film time at the read head
	off  int64     // and the byte it is at
	at   time.Time // when that was last read
	set  bool

	buffer int64 // film the client is holding, in bytes
	fixed  bool  // and whether that has been settled

	screenSec float64 // where the picture is, worked out at the last step
	screenOff int64

	gap     float64   // how far apart this file's timestamps arrive, as seen so far
	arrived time.Time // when film last actually arrived

	// Whether the client has stopped accumulating — which is the whole of the question, and is
	// not a question about speed. Deciding it by how fast film arrives was tried at length and
	// does not work: a fill on a poorly seeded release runs at 1.2 times what is being watched,
	// slow enough to pass for a client taking only what it consumes while its buffer is still
	// filling up. Measured on such a release the size settled at 227MB while the player went on
	// to hold 300, and the position sat 26 seconds in front of the picture.
	//
	// What separates them is not the rate but the amount in hand: whatever the head is doing,
	// the film between it and the picture stands still for a client that is full and watching,
	// and keeps climbing for one that is still filling — by the difference between the two
	// speeds, however small that difference is.
	played bool

	// When what the client holds was last seen to grow by more than a hair, and whether the
	// head has ever stopped at all.
	grew       time.Time
	everSilent bool

	// The recent fill rate, over a stretch rather than step to step: film arrives in chunks,
	// and a single step reads 0.46 or 1.69 in the middle of perfectly ordinary playback.
	winArr    float64
	winPass   float64
	winClient float64
	winSupply float64
	win       []span

	// The two waits for the step being taken now.
	stepClient float64
	stepSupply float64

	// Time during which film was actually arriving. The picture is placed at this much past
	// where the session started, and the distinction from the wall clock is not academic: the
	// moment a client pauses, the clock goes on and the picture does not, so a picture placed
	// by the clock eats a second of the buffer for every second of the pause. Measured, a
	// buffer of 378MB was chewed down to 268 before the stop was even recognised, and that
	// shrunken figure was what got frozen and handed to the next connection — which then put
	// the picture ten seconds in front of itself.
	watched float64

	// What was handed over by the connection before, which is the most this session may hand
	// on in turn. Without that cap a session inherits a buffer, adds its own reading to it and
	// passes on the sum; the next inherits that, and the figure grows with every reconnection.
	inherited int64
}

// film turns byte offsets in a file into playback time and back. The index does this for a
// real file; the tests stand in for it with a plain bitrate.
type film interface {
	TimeAt(off int64) (float64, bool)
	OffsetAt(sec float64) (int64, bool)
}

const (
	// Bounds on how long an absence of film has to last before it means anything. It follows
	// the spacing of the file's own timestamps: between two of them nothing arrives however
	// well playback is going, so a shorter gap says nothing at all.
	minQuiet = 6.0
	maxQuiet = 30.0
	// The most film a second of reading can plausibly deliver. A jump past this is a
	// timestamp that does not belong where it was found.
	maxLeap = 300.0
	// How long a stretch is judged over, how much the amount in hand may drift in it without
	// counting as movement, and the band within which a client is taking about what it
	// watches. The band needs both ends: without a floor, "no faster than it watches" is also
	// true of a client that cannot keep up, and so is "no longer growing" — a draining buffer
	// is certainly not growing. Both are true in the middle of a supply dip, which is the
	// worst possible moment to decide how much a client holds. Measured on an undisturbed run,
	// the size settled at 179MB while the client was starving down to 100, and stayed there
	// after it had recovered to 300.
	settleFor  = 15.0
	settleSlop = 4 << 20
	// A client reading without a break for this long cannot be paused: a buffer is finite, so
	// one that is filling reaches the top and stops. Nothing that goes on taking film for five
	// minutes straight has anywhere left to put it except a picture.
	//
	// This is the same reasoning as "the head stopped, so it is paused", from the other end,
	// and it is here to bound the one case that was otherwise unbounded. Until a session
	// concludes something, the picture is held where it began; a session that never concludes
	// anything holds it there for as long as it runs, and the answer drifts a second further
	// behind every second. Five minutes is where that stops.
	watchingAfter = 300.0
	keepUpRate    = 0.9
	// Anything faster than this is filling, not watching. It sits just above one because that
	// is what watching is — a second of film a second — and the margin is for the measurement,
	// not for the client: film arrives in chunks, so a fifteen-second window reads between
	// 0.76 and 1.25 during perfectly ordinary playback.
	//
	// It was 1.5 to keep that noise from reading as a fill, and that was the wrong way round.
	// Set too high, a client filling at 1.2 — which happens on a release with few seeders —
	// passes for one that is watching, its size is settled while it is still filling up, and
	// the position ends up in front of the picture. Set too low, the only cost is that
	// "watching" is recognised a window or two later, and until then the picture is held where
	// the session began, which is behind. One of those is allowed and the other is not.
	watchRate = 1.1
)

// start fixes the point the session began at: the film time and byte of the first timestamp
// this connection found, and the moment it first read.
//
// A connection normally opens with nothing buffered, which is the one moment the buffer is
// known without measuring it. holding is the exception: a player that pauses long enough
// loses its connection and opens another from the same byte, still carrying everything it
// had. Such a client is also known to have been watching — there would have been nothing to
// hand over otherwise — so there is no fill to wait for and nothing left to settle.
func (p *progress) start(ref float64, refOff int64, at time.Time, holding, size int64, ix film) {
	*p = progress{ref: ref, refOff: refOff, begun: at, head: ref, off: refOff,
		at: at, set: true, gap: maxQuiet / 3, screenSec: ref, screenOff: refOff, grew: at}
	if holding > 0 {
		p.buffer = holding
		// The floor is the box, not what is left in it: the refill of everything that died
		// with the line belongs to the client just as much as what survived.
		p.inherited = max(holding, size)
		// Settled, and not measured again. A buffer is a fixed number of megabytes set on the
		// device; a dropped connection does not resize it, and the same client coming back is
		// holding the same box it was holding before. Re-measuring it on the new connection was
		// tried and is gone: it only ever found a different number, and the figure wandered
		// from one reconnection to the next — 439MB, then 496 — with nothing about the player
		// having changed at all.
		//
		// What is carried over as still in hand is smaller than the box, because everything the
		// server had sent and the player had not yet received died with the line. That says
		// where the picture is. The box is what the position is worked out from, so it is the
		// box that gets frozen; until the player has asked again for what it lost, the picture
		// is reported a little further back than it really is, which is the side to be on.
		p.played, p.fixed = true, true
		// The picture is that far behind the byte this connection began on, so the session
		// counts from there rather than from where reading resumes — in film time as well as
		// in bytes. Moving only the byte and leaving the time where reading resumed leaves the
		// two describing different places, and since the reported position may not go below
		// the session's own start, it then cannot go below where the connection opened: a
		// picture forty-six seconds back was reported at the read head for want of this line.
		p.refOff = refOff - holding
		if sec, ok := ix.TimeAt(p.refOff); ok {
			p.ref = sec
		}
		p.screenSec, p.screenOff = p.ref, p.refOff
	}
}

// step takes in the film time and byte now at the read head.
func (p *progress) step(head float64, off int64, ix film, now time.Time) {
	if !p.set || ix == nil {
		return
	}
	passed := now.Sub(p.at).Seconds()
	if passed <= 0 {
		return
	}
	if head < p.head {
		head, off = p.head, p.off // timestamps never go backwards within a file
	}
	if head-p.head > maxLeap*passed {
		return // not a timestamp from where it was found
	}

	arrived := head - p.head
	if arrived > 0 {
		if !p.arrived.IsZero() {
			// The widest recent spacing, easing down: guessing it too short reads the quiet
			// between timestamps as a stopped picture.
			if seen := now.Sub(p.arrived).Seconds(); seen > p.gap {
				p.gap = seen
			} else {
				p.gap = max(p.gap*0.97, seen)
			}
		}
		p.arrived = now
	}
	quiet := min(max(p.gap*3, minQuiet), maxQuiet)
	silent := !p.arrived.IsZero() && now.Sub(p.arrived).Seconds() >= quiet

	// Only while film is coming, and only while the client is taking no more than it watches.
	// A head that has stopped is a picture that has stopped with it — but so, possibly, is a
	// head that has sped up: a client topping up is doing something the picture takes no part
	// in, and there is no telling from the stream whether the picture is running alongside.
	// This is the run-up to a pause, where the player fills the last of its buffer after the
	// picture has already stopped; counted as watched, it walked the picture 24 seconds past
	// itself and handed that on to the next connection.
	//
	// The cost when the client really is playing while it fills is that the picture is held
	// where it is until the filling stops. That is the allowed direction, and it is over as
	// soon as the buffer is.
	// And not at all while film is arriving more slowly than it plays, so long as nothing has
	// been established yet. Watching cannot be sustained below one times: the player eats
	// through whatever it had and stops to reload. So until a buffer has been established,
	// a slow supply means the client is not consuming — everything arriving is going into
	// store, whether it is paused or waiting for the torrent to pick up.
	//
	// The arithmetic without this is brutal on a low-bitrate file, where a fixed box holds
	// many minutes: three hundred megabytes of a cartoon is nine minutes of film, and filling
	// it at half speed takes eighteen. Credit those eighteen minutes as watched and the count
	// finds nothing in hand at all, putting the position at the read head — nine minutes in
	// front of a picture that has not started.
	//
	// Once it is established that the client is watching and holding a buffer, the same slow
	// supply means the opposite: it is draining what it has, and the picture is running.
	if arrived > 0 && (p.played || p.keepsUp()) {
		p.watched += passed
	}

	// Nothing can have been shown that has not arrived. The picture may be at the read head at
	// the very most, and that is the case of a client holding nothing at all.
	//
	// The cap is what settles a question that no measure of speed can. A client taking film
	// more slowly than it plays cannot have been watching from the start of the session — it
	// would have run dry — however ordinary that speed looks: measured on a warm-up pause over
	// a link slower than the film, the fill averaged 0.94 times real time, which is squarely
	// inside any band that would call it watching. Yet its buffer went from nothing to three
	// hundred megabytes, and a client that is watching cannot be filling up at less than one
	// times. So while the cap is binding there is nothing in hand, and nothing in hand means
	// it has not been watching.
	delivered := head - p.ref
	if p.watched > delivered {
		p.watched = delivered
	}

	grown := p.buffer
	p.window(arrived, passed)
	// The same band, from both sides. Asking only that it be no faster than it watches lets in
	// the trickle after a fill has finished: the head has stopped, silence has not yet been
	// declared, and a step where a hundredth of a second of film arrives reads as a client
	// taking exactly what it consumes. That is how a player paused to warm up came to be taken
	// for one that had been watching all along — and then, when the stop was recognised, it
	// was credited with only what it had measured rather than with everything delivered.
	// Requiring the client to keep this up for a good while before believing it — long enough
	// that the tail of a fill could not fake it — was tried and had to go. It works: playback
	// piles up eighty seconds inside the band where a fill's tail manages thirteen. But while
	// the session has not concluded anything, the picture is held where it started, and the
	// size is then frozen together with however far the picture had really got by then. On the
	// scenarios that means a permanent sixty-nine seconds of lag, bought to save thirty on a
	// warm-up pause. The tail of a fill is still counted, and the note above says what that
	// costs.
	if arrived > 0 && !silent && !p.played && p.keepingUp() {
		p.played = true
	}
	if !p.played && !p.everSilent && now.Sub(p.begun).Seconds() >= watchingAfter {
		p.played = true
	}
	if silent {
		p.everSilent = true
	}

	switch {
	case p.fixed:
		// The size is settled. The picture is that many bytes behind the head and follows it,
		// which is also why a pause needs no handling of its own: the head stops, so it stops.
	case p.played:
		// Full and watching: what it holds is the film between the head and the picture, and
		// the picture has been running since the session began.
		if at, ok := ix.OffsetAt(p.ref + p.watched); ok && off > at {
			p.buffer = off - at
		}
	default:
		// Nothing settled yet, so the picture is taken to be where the session started and
		// everything delivered to be in hand. Reading it the other way round — as if the
		// picture had been running all along — was tried and is worse: over the opening
		// seconds a client takes film at about the speed it plays it, so that reading finds
		// nothing in hand at all and puts the picture at the read head, eighteen seconds in
		// front of itself.
		//
		// The cost is that a session which never settles never leaves its own start. That is
		// paid for by settling reliably, not by reading this differently.
		//
		// Choosing between two readings here by who was holding things up was tried and had to
		// go, because the inference does not hold in the direction that matters. "The torrent
		// is the bottleneck" was read as "the client is taking everything it can get, so it is
		// empty" — but a client paused with a full buffer over a slow torrent looks exactly the
		// same, and reading it as empty puts the position at the read head. Measured on a
		// warm-up pause: forty-seven seconds in front of a picture that had not started, and it
		// stayed in front for the whole session.
		//
		// So the reading here stays the pessimistic one. It costs lag on a client that really
		// is starving — the position sits at the session's start while the picture runs on —
		// and that is the side which is allowed to be wrong.
		p.buffer = off - p.refOff
	}

	// What was carried over is a floor, never a ceiling: a buffer handed to a client that is
	// not in fact holding it must still leave the picture behind where it really is, while
	// upwards it must not bind, or the refill of everything that died with the line is never
	// counted.
	if p.buffer < p.inherited {
		p.buffer = p.inherited
	}
	if p.buffer < 0 {
		p.buffer = 0
	}
	// When what the client holds was last seen to grow. A size still climbing is a box only
	// part filled, and freezing there settles on less than the client will hold.
	if p.buffer > grown+settleSlop {
		p.grew = now
	}

	// What ends the fill decides what the answer was.
	if !p.fixed {
		switch {
		case silent:
			p.fixed = true
		case p.played && p.clientLimited() && now.Sub(p.grew).Seconds() >= settleFor:
			p.fixed = true
		}
	}

	p.screenOff = off - p.buffer
	if p.screenOff < p.refOff {
		p.screenOff = p.refOff
	}
	if sec, ok := ix.TimeAt(p.screenOff); ok {
		p.screenSec = max(p.ref, sec)
	}

	p.head, p.off, p.at = head, off, now
}

// span is one step's worth of the window: film arrived, the clock it took, and which side the
// reading waited on while it did.
type span struct{ arr, pass, client, supply float64 }

// waited takes in how long the last stretch spent held up by the client and by the torrent.
func (p *progress) waited(client, supply time.Duration) {
	p.stepClient, p.stepSupply = client.Seconds(), supply.Seconds()
}

// clientLimited reports that the hold-up is the client rather than the supply: film is there
// to be sent and it is not taking it. A client only stops taking when it has nowhere to put
// more, so this is the one plain statement that it is full — and being full is the only state
// in which what it holds can be read off at all.
//
// The distinction cannot be made from the outside. A starving client and a full one both leave
// the head crawling, and telling them apart by speed is what settled the size in the middle of
// a supply dip: 179MB recorded while the player was starving down to 100 and went on to hold
// 300, with the position ten seconds in front of the picture for the rest of the film.
func (p *progress) clientLimited() bool {
	return p.winClient+p.winSupply > 0 && p.winClient > p.winSupply
}

// window keeps the recent fill rate.
func (p *progress) window(arrived, passed float64) {
	p.win = append(p.win, span{arrived, passed, p.stepClient, p.stepSupply})
	p.winArr, p.winPass = p.winArr+arrived, p.winPass+passed
	p.winClient, p.winSupply = p.winClient+p.stepClient, p.winSupply+p.stepSupply
	for len(p.win) > 1 && p.winPass-p.win[0].pass >= settleFor {
		p.winArr, p.winPass = p.winArr-p.win[0].arr, p.winPass-p.win[0].pass
		p.winClient, p.winSupply = p.winClient-p.win[0].client, p.winSupply-p.win[0].supply
		p.win = p.win[1:]
	}
}

// keepsUp reports whether film is at least arriving as fast as it plays. Below that a client
// cannot be watching for long, whatever else is true.
func (p *progress) keepsUp() bool {
	return p.winPass < settleFor || p.winArr >= p.winPass*keepUpRate
}

// keepingUp reports whether film has lately been arriving at about the speed it is watched —
// neither filling nor falling behind. It is the only state in which what the client holds can
// be read off, because it is the only one in which what it holds is not changing.
func (p *progress) keepingUp() bool {
	return p.winPass >= settleFor &&
		p.winArr >= p.winPass*keepUpRate && p.winArr <= p.winPass*watchRate
}

// screen is where the picture is, in film time.
func (p *progress) screen() (float64, bool) {
	if !p.set {
		return 0, false
	}
	return p.screenSec, true
}

// box is the size of the client's buffer as this session has it — what it measured, or what
// it was told by the session before. A device does not change size, so this is what travels.
func (p *progress) box() int64 {
	if !p.set {
		return 0
	}
	return max(p.buffer, p.inherited)
}

// pictureAt is the byte the picture is at, for the connection that comes next.
func (p *progress) pictureAt() int64 {
	return p.screenOff
}

// held is how much film the client has that has not been shown yet, in seconds — for display
// only, since the reckoning itself never needs it.
func (p *progress) held() float64 {
	if !p.set {
		return 0
	}
	return max(0, p.head-p.screenSec)
}

// handOn is the buffer this session may pass to the next one, in bytes. Never more than it was
// handed itself.
func (p *progress) handOn() int64 {
	if !p.set {
		return 0
	}
	if p.inherited > 0 && p.buffer > p.inherited {
		return p.inherited
	}
	return p.buffer
}
