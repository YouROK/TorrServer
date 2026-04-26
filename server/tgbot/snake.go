package tgbot

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	"server/log"
	"server/torr"
	cacheSt "server/torr/storage/state"
)

var (
	snakeStopChans     = make(map[int]chan struct{})
	snakeStopChansMu   sync.Mutex
	snakeWindowStart   = make(map[string]int)
	snakeWindowStartMu sync.Mutex
)

const (
	snakeBlockFilled    = "🟩"
	snakeBlockEmpty     = "⬜"
	snakeBlockReader    = "🔵"
	snakeBlockInRange   = "🟦"
	snakeTitleMaxLen    = 55
	snakeHashDisplayLen = 8
)

func cmdSnake(c tele.Context) error {
	args := c.Args()
	hash := ""
	cols, rows := 20, 3

	if len(args) > 0 {
		hash = resolveHash(c, args[0])
	}
	if len(args) > 1 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 && n <= 50 {
			cols = n
		}
	}
	if len(args) > 2 {
		if n, err := strconv.Atoi(args[2]); err == nil && n > 0 && n <= 15 {
			rows = n
		}
	}

	if hash == "" {
		return c.Send(tr(c.Sender().ID, "snake_usage"))
	}

	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Send(tr(c.Sender().ID, "torrent_not_found") + ":\n<code>" + hash + "</code>")
	}

	st := t.CacheState()
	if st == nil {
		return c.Send(fmt.Sprintf(tr(c.Sender().ID, "cache_unavailable"), hash))
	}

	uid := c.Sender().ID
	txt := formatSnake(uid, st, hash, cols, rows)
	kbd := snakeKeyboard(uid, hash, cols, rows, true)
	msg, err := c.Bot().Send(c.Sender(), txt, kbd)
	if err != nil {
		return err
	}
	log.TLogln("tg snake sent", logUserID(uid), logSafeStr(st.Torrent.Title, 40), hash)
	go snakeRefreshLoop(c.Bot(), msg, hash, uid, cols, rows)
	return nil
}

func formatSnake(uid int64, st *cacheSt.CacheState, hash string, cols, rows int) string {
	totalBlocks := cols * rows
	if totalBlocks <= 0 {
		return tr(uid, "snake_no_data")
	}
	if st.PiecesCount <= 0 {
		title := ""
		if st.Torrent != nil {
			title = escapeHtml(st.Torrent.Title)
		}
		txt := "📊 <b>" + title + "</b>\n"
		txt += fmt.Sprintf("%s: %s / %s\n", tr(uid, "snake_cache"),
			humanize.IBytes(uint64(st.Filled)), humanize.IBytes(uint64(st.Capacity)))
		dispHash := st.Hash
		if len(dispHash) > snakeHashDisplayLen {
			dispHash = dispHash[:snakeHashDisplayLen]
		}
		txt += tr(uid, "snake_no_data") + " <code>" + dispHash + "</code>"
		return txt
	}

	pieceFilled := make(map[int]bool)
	for id, p := range st.Pieces {
		if id >= 0 && id < st.PiecesCount && p.Size > 0 {
			pieceFilled[id] = true
		}
	}

	readerPositions := make(map[int]bool)
	readerRanges := make(map[int]bool)
	for _, r := range st.Readers {
		readerPositions[r.Reader] = true
		for p := r.Start; p < r.End && p < st.PiecesCount; p++ {
			readerRanges[p] = true
		}
	}

	cacheWindowPieces := int64(totalBlocks) * 2
	if st.PiecesLength > 0 {
		cacheWindowPieces = st.Capacity / st.PiecesLength
	}
	if cacheWindowPieces < int64(totalBlocks) {
		cacheWindowPieces = int64(totalBlocks)
	}

	startPiece, endPiece := 0, st.PiecesCount
	if len(st.Readers) > 0 {
		minReader, maxReader := st.PiecesCount, 0
		for _, r := range st.Readers {
			if r.Reader < minReader {
				minReader = r.Reader
			}
			if r.Reader > maxReader {
				maxReader = r.Reader
			}
		}
		windowSize := int(cacheWindowPieces)
		snakeWindowStartMu.Lock()
		lastStart := snakeWindowStart[hash]
		scrollThreshold := windowSize * 3 / 4
		if lastStart == 0 || minReader < lastStart {
			lastStart = minReader
		} else if minReader >= lastStart+scrollThreshold {
			lastStart = minReader - windowSize/5
		}
		if lastStart < 0 {
			lastStart = 0
		}
		snakeWindowStart[hash] = lastStart
		snakeWindowStartMu.Unlock()
		startPiece = lastStart
		endPiece = startPiece + windowSize
		if endPiece > st.PiecesCount {
			endPiece = st.PiecesCount
			startPiece = endPiece - windowSize
			if startPiece < 0 {
				startPiece = 0
			}
		}
	} else if len(pieceFilled) > 0 {
		minP, maxP := st.PiecesCount, 0
		for id := range pieceFilled {
			if id < minP {
				minP = id
			}
			if id > maxP {
				maxP = id
			}
		}
		window := maxP - minP + 1
		if window > int(cacheWindowPieces) {
			window = int(cacheWindowPieces)
		}
		startPiece = minP
		endPiece = minP + window
		if endPiece > st.PiecesCount {
			endPiece = st.PiecesCount
		}
	}

	windowSize := endPiece - startPiece
	if windowSize <= 0 {
		windowSize = 1
	}

	blocks := make([]string, totalBlocks)
	piecesPerBlock := (windowSize + totalBlocks - 1) / totalBlocks
	if piecesPerBlock < 1 {
		piecesPerBlock = 1
	}

	for i := 0; i < totalBlocks; i++ {
		start := startPiece + i*piecesPerBlock
		end := start + piecesPerBlock
		if end > endPiece {
			end = endPiece
		}
		if start >= end {
			blocks[i] = snakeBlockEmpty
			continue
		}

		blockFilled := false
		blockHasReader := false
		blockInRange := false
		for p := start; p < end; p++ {
			if pieceFilled[p] {
				blockFilled = true
			}
			if readerPositions[p] {
				blockHasReader = true
			}
			if readerRanges[p] {
				blockInRange = true
			}
		}

		switch {
		case blockHasReader:
			blocks[i] = snakeBlockReader
		case blockFilled:
			blocks[i] = snakeBlockFilled
		case blockInRange:
			blocks[i] = snakeBlockInRange
		default:
			blocks[i] = snakeBlockEmpty
		}
	}

	var sb strings.Builder
	title := ""
	if st.Torrent != nil {
		title = st.Torrent.Title
	}
	if len([]rune(title)) > snakeTitleMaxLen {
		title = string([]rune(title)[:snakeTitleMaxLen]) + "…"
	}
	title = escapeHtml(title)
	sb.WriteString("📊 <b>")
	sb.WriteString(title)
	sb.WriteString("</b>\n")
	fmt.Fprintf(&sb, "%s: %s / %s",
		tr(uid, "snake_cache"),
		humanize.IBytes(uint64(st.Filled)),
		humanize.IBytes(uint64(st.Capacity)))
	if len(st.Readers) > 1 {
		fmt.Fprintf(&sb, " · %d %s", len(st.Readers), tr(uid, "status_streams"))
	}
	if endPiece-startPiece < st.PiecesCount {
		fmt.Fprintf(&sb, " · %s %d-%d", tr(uid, "snake_pieces"), startPiece+1, endPiece)
	}
	sb.WriteString("\n")

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			var idx int
			if r%2 == 0 {
				idx = r*cols + c
			} else {
				idx = r*cols + (cols - 1 - c)
			}
			if idx < len(blocks) {
				sb.WriteString(blocks[idx])
			}
		}
		sb.WriteString("\n")
	}
	dispHash := st.Hash
	if len(dispHash) > snakeHashDisplayLen {
		dispHash = dispHash[:snakeHashDisplayLen]
	}
	sb.WriteString(tr(uid, "snake_legend"))
	sb.WriteString(" <code>")
	sb.WriteString(dispHash)
	sb.WriteString("</code>")
	return sb.String()
}

func snakeData(hash string, cols, rows int) string {
	return fmt.Sprintf("%s|%d|%d", hash, cols, rows)
}

func parseSnakeData(data string) (hash string, cols, rows int) {
	cols, rows = 20, 3
	parts := strings.Split(data, "|")
	if len(parts) > 0 {
		hash = parts[0]
	}
	if len(parts) > 1 {
		if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
			cols = n
		}
	}
	if len(parts) > 2 {
		if n, err := strconv.Atoi(parts[2]); err == nil && n > 0 {
			rows = n
		}
	}
	return hash, cols, rows
}

func snakeKeyboard(uid int64, hash string, cols, rows int, active bool) *tele.ReplyMarkup {
	data := snakeData(hash, cols, rows)
	if active {
		return &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
			{
				{Text: "🔄", Unique: "fsnakerefresh", Data: data},
				{Text: tr(uid, "status_stop_btn"), Unique: "fsnakestop", Data: data},
			},
		}}
	}
	return &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
		{{Text: tr(uid, "status_refresh_btn"), Unique: "fsnakerefresh", Data: data}},
	}}
}

func snakeRefreshLoop(api tele.API, msg *tele.Message, hash string, uid int64, cols, rows int) {
	const interval = 2 * time.Second
	const duration = 2 * time.Minute
	stopCh := make(chan struct{})
	snakeStopChansMu.Lock()
	snakeStopChans[msg.ID] = stopCh
	snakeStopChansMu.Unlock()
	defer func() {
		snakeStopChansMu.Lock()
		delete(snakeStopChans, msg.ID)
		snakeStopChansMu.Unlock()
	}()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	deadline := time.Now().Add(duration)
	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if time.Now().After(deadline) {
				t := torr.GetTorrent(hash)
				if t != nil {
					if st := t.CacheState(); st != nil {
						txt := formatSnake(uid, st, hash, cols, rows) + "\n" + tr(uid, "status_auto_ended")
						_, _ = api.Edit(msg, txt, snakeKeyboard(uid, hash, cols, rows, false), tele.ModeHTML)
					}
				}
				return
			}
			t := torr.GetTorrent(hash)
			if t == nil {
				return
			}
			st := t.CacheState()
			if st == nil {
				return
			}
			txt := formatSnake(uid, st, hash, cols, rows)
			if _, err := api.Edit(msg, txt, snakeKeyboard(uid, hash, cols, rows, true), tele.ModeHTML); err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "message is not modified") {
					continue
				}
				if strings.Contains(errStr, "message to edit not found") {
					return
				}
				log.TLogln("tg snake refresh err", err)
				return
			}
		}
	}
}

func stopSnakeRefresh(msgID int) {
	snakeStopChansMu.Lock()
	ch := snakeStopChans[msgID]
	delete(snakeStopChans, msgID)
	snakeStopChansMu.Unlock()
	if ch != nil {
		close(ch)
	}
}

func callbackSnakeRefresh(c tele.Context, data string) error {
	hash, cols, rows := parseSnakeData(data)
	if hash == "" {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "callback_unknown")})
	}
	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "torrent_not_found")})
	}
	st := t.CacheState()
	if st == nil {
		return c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf(tr(c.Sender().ID, "cache_unavailable"), hash)})
	}
	if c.Callback().Message != nil {
		stopSnakeRefresh(c.Callback().Message.ID)
		_ = c.Bot().Delete(c.Callback().Message)
	}
	_ = c.Respond(&tele.CallbackResponse{})
	uid := c.Sender().ID
	txt := formatSnake(uid, st, hash, cols, rows)
	kbd := snakeKeyboard(uid, hash, cols, rows, true)
	msg, err := c.Bot().Send(c.Sender(), txt, kbd)
	if err != nil {
		return err
	}
	go snakeRefreshLoop(c.Bot(), msg, hash, uid, cols, rows)
	return nil
}

func callbackSnakeStop(c tele.Context, data string) error {
	uid := c.Sender().ID
	hash, cols, rows := parseSnakeData(data)
	if hash != "" {
		if t := torr.GetTorrent(hash); t != nil {
			log.TLogln("tg snake stop", logUserID(uid), logSafeStr(t.Title, 40), hash)
		}
	}
	if c.Callback().Message != nil {
		stopSnakeRefresh(c.Callback().Message.ID)
		if hash != "" {
			msg := c.Callback().Message
			t := torr.GetTorrent(hash)
			txt := ""
			if t != nil {
				if st := t.CacheState(); st != nil {
					txt = formatSnake(uid, st, hash, cols, rows)
				}
			}
			if txt == "" {
				txt = "<code>" + hash + "</code>"
			}
			txt += "\n" + tr(uid, "status_stopped")
			_, _ = c.Bot().Edit(msg, txt, snakeKeyboard(uid, hash, cols, rows, false), tele.ModeHTML)
		}
	}
	return c.Respond(&tele.CallbackResponse{Text: "🛑"})
}
