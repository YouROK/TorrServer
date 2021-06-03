package torr

import (
	"fmt"
	"io"
	"time"

	"github.com/anacrolix/torrent"

	"server/log"
	"server/torr/state"
	"server/torr/storage/torrstor"
	utils2 "server/utils"
)

func (t *Torrent) Preload(index int, size int64) {
	if size <= 0 {
		return
	}
	t.PreloadSize = size

	if t.Stat == state.TorrentGettingInfo {
		if !t.WaitInfo() {
			return
		}
		// wait change status
		time.Sleep(100 * time.Millisecond)
	}

	t.muTorrent.Lock()
	if t.Stat != state.TorrentWorking {
		t.muTorrent.Unlock()
		return
	}

	t.Stat = state.TorrentPreload
	t.muTorrent.Unlock()

	defer func() {
		if t.Stat == state.TorrentPreload {
			t.Stat = state.TorrentWorking
		}
	}()

	file := t.findFileIndex(index)
	if file == nil {
		file = t.Files()[0]
	}

	if size > file.Length() {
		size = file.Length()
	}

	if t.Info() != nil {
		// Запуск лога в отдельном потоке
		go func() {
			for t.Stat == state.TorrentPreload {
				stat := fmt.Sprint(file.Torrent().InfoHash().HexString(), " ", utils2.Format(float64(t.PreloadedBytes)), "/", utils2.Format(float64(t.PreloadSize)), " Speed:", utils2.Format(t.DownloadSpeed), " Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
				log.TLogln("Preload:", stat)
				time.Sleep(time.Second)
			}
		}()

		mb5 := int64(5 * 1024 * 1024)

		readerStart := file.NewReader()
		defer readerStart.Close()
		readerStart.SetResponsive()
		readerStart.SetReadahead(0)
		readerStartEnd := file.Offset() + size - mb5

		if readerStartEnd < file.Offset() {
			// Если конец начального ридера оказался за началом
			readerStartEnd = file.Offset() + size
		}
		if readerStartEnd > file.Offset()+file.Length() {
			// Если конец начального ридера оказался после конца файла
			readerStartEnd = file.Offset() + file.Length()
		}

		readerEndStart := file.Offset() + file.Length() - mb5
		readerEndEnd := file.Offset() + file.Length()

		tmp := make([]byte, 32768, 32768)
		offset := int64(0)
		if readerEndStart > readerStartEnd {
			// Если конечный ридер не входит в диапозон начального
			readerEnd := file.NewReader()
			readerEnd.SetResponsive()
			readerEnd.SetReadahead(0)
			readerEnd.Seek(readerEndStart, io.SeekStart)
			offset = readerEndStart
			for offset+int64(len(tmp)) < readerEndEnd {
				n, err := readerEnd.Read(tmp)
				if err != nil {
					log.TLogln("Error preload:", err)
					readerEnd.Close()
					return
				}
				offset += int64(n)
			}
			readerEnd.Close()
		}

		offset = 0
		for offset+int64(len(tmp)) < readerStartEnd {
			n, err := readerStart.Read(tmp)
			if err != nil {
				log.TLogln("Error preload:", err)
				return
			}
			offset += int64(n)
		}

		/*pieceLength := t.Info().PieceLength
		mb5 := int64(5 * 1024 * 1024)

		pieceFileStart := int(file.Offset() / pieceLength)
		pieceFileStartEnd := int((file.Offset()+size-mb5)/pieceLength) - 1
		if pieceFileStartEnd < pieceFileStart {
			pieceFileStartEnd = pieceFileStart
		}

		pieceFileEnd := int((file.Offset() + file.Length() - mb5) / pieceLength)
		pieceFileEndEnd := int((file.Offset() + file.Length()) / pieceLength)
		if file.Length() < mb5 {
			pieceFileStartEnd = pieceFileEndEnd
			pieceFileEnd = -1
			pieceFileEndEnd = -1
		}

		lastStat := time.Now().Add(-time.Second)

		for true {
			t.muTorrent.Lock()
			if t.Torrent == nil {
				return
			}

			t.PreloadedBytes = t.cache.GetState().Filled
			t.muTorrent.Unlock()

			stat := fmt.Sprint(file.Torrent().InfoHash().HexString(), " ", utils2.Format(float64(t.PreloadedBytes)), "/", utils2.Format(float64(t.PreloadSize)), " Speed:", utils2.Format(t.DownloadSpeed), " Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
			if time.Since(lastStat) > time.Second {
				log.TLogln("Preload:", stat)
				lastStat = time.Now()
			}

			beginLoadingPieces := t.piecesLoading(pieceFileStart, pieceFileStartEnd)
			endLoadingPieces := t.piecesLoading(pieceFileEnd, pieceFileEndEnd)

			if beginLoadingPieces == 0 && endLoadingPieces == 0 {
				break
			}

			t.AddExpiredTime(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
			time.Sleep(time.Second)
		}*/
	}
	log.TLogln("End preload:", file.Torrent().InfoHash().HexString(), "Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
}

func (t *Torrent) piecesLoading(start, end int) int {
	count := 0
	if start < 0 || end < 0 {
		return 0
	}
	limitLoading := 5
	for i := start; i <= end; i++ {
		if !t.Piece(i).Storage().PieceImpl.(*torrstor.Piece).Complete {
			count++
			if limitLoading > 0 && t.PieceState(i).Priority == torrent.PiecePriorityNone {
				t.Piece(i).SetPriority(torrent.PiecePriorityNormal)
			}
			limitLoading--
		}
	}
	return count
}

func (t *Torrent) findFileIndex(index int) *torrent.File {
	st := t.Status()
	var stFile *state.TorrentFileStat
	for _, f := range st.FileStats {
		if index == f.Id {
			stFile = f
			break
		}
	}
	if stFile == nil {
		return nil
	}
	for _, file := range t.Files() {
		if file.Path() == stFile.Path {
			return file
		}
	}
	return nil
}
