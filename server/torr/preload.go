package torr

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/anacrolix/torrent"

	"server/log"
	"server/settings"
	"server/torr/state"
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
				t.AddExpiredTime(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
				time.Sleep(time.Second)
			}
		}()

		mb5 := int64(5 * 1024 * 1024)

		readerStart := file.NewReader()
		defer readerStart.Close()
		readerStart.SetResponsive()
		readerStart.SetReadahead(0)
		readerStartEnd := size - mb5

		if readerStartEnd < 0 {
			// Если конец начального ридера оказался за началом
			readerStartEnd = size
		}
		if readerStartEnd > file.Length() {
			// Если конец начального ридера оказался после конца файла
			readerStartEnd = file.Length()
		}

		readerEndStart := file.Length() - mb5
		readerEndEnd := file.Length()

		var wa sync.WaitGroup
		go func() {
			offset := int64(0)
			if readerEndStart > readerStartEnd {
				// Если конечный ридер не входит в диапозон начального
				wa.Add(1)
				defer wa.Done()
				readerEnd := file.NewReader()
				readerEnd.SetResponsive()
				readerEnd.SetReadahead(0)
				readerEnd.Seek(readerEndStart, io.SeekStart)
				offset = readerEndStart
				tmp := make([]byte, 32768, 32768)
				for offset+int64(len(tmp)) < readerEndEnd {
					n, err := readerEnd.Read(tmp)
					if err != nil {
						break
					}
					offset += int64(n)
				}
				readerEnd.Close()
			}
		}()

		pieceLength := t.Info().PieceLength
		readahead := pieceLength * 4
		if readerStartEnd < readahead {
			readahead = 0
		}
		readerStart.SetReadahead(readahead)
		offset := int64(0)
		tmp := make([]byte, 32768, 32768)
		for offset+int64(len(tmp)) < readerStartEnd {
			n, err := readerStart.Read(tmp)
			if err != nil {
				log.TLogln("Error preload:", err)
				return
			}
			offset += int64(n)
			if readahead > 0 && readerStartEnd-(offset+int64(len(tmp))) < readahead {
				readahead = 0
				readerStart.SetReadahead(0)
			}
		}

		wa.Wait()
	}
	log.TLogln("End preload:", file.Torrent().InfoHash().HexString(), "Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
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
