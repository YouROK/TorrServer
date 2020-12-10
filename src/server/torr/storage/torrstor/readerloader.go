package torrstor

import (
	"io"

	"github.com/dustin/go-humanize"
	"server/log"
)

func (r *Reader) getUsedPieces() (int, int, int) {
	startOff, endOff := r.offset, r.endOffsetPreload
	if startOff < endOff {
		endOff = startOff + r.readahead
	}
	return r.getRangePieces(r.offset, r.currOffsetPreload, r.endOffsetPreload)
}

//////////////////////////////////////////////////////////////
/// Прелоадер начинает загрузку от старта плеера+RAH и имеет свой RAH
/// Прелоадер грузит до конца-RAH
func (r *Reader) preload() {
	// определяем конец загрузки
	r.endOffsetPreload = r.offset + r.cache.capacity - 1024

	// конец за пределами конца файла, тримим
	if r.endOffsetPreload > r.file.Length() {
		r.endOffsetPreload = r.file.Length()
	}
	r.muPreload.Lock()
	// загрузка уже идет или конец меньше RAH, тогда старается основной ридер
	if r.isPreload || r.endOffsetPreload < r.readahead {
		r.muPreload.Unlock()
		return
	}
	r.isPreload = true
	r.muPreload.Unlock()

	log.TLogln("Start buffering from", humanize.IBytes(uint64(r.currOffsetPreload)))
	go func() {
		// получаем ридер
		buffReader := r.file.NewReader()
		defer func() {
			r.isPreload = false
			buffReader.Close()
		}()
		// ищем не прочитанный кусок
		r.currOffsetPreload = r.findPreloadedStart()
		// выходим если ничего подгружать не нужно
		if r.currOffsetPreload >= r.endOffsetPreload {
			return
		}
		// двигаем лоадер
		buffReader.Seek(r.currOffsetPreload, io.SeekStart)
		buff := make([]byte, 1024)
		// isReadahead чтобы меньше переключать RAH
		isReadahead := false
		buffReader.SetReadahead(0)
		// читаем пока позиция лоадера меньше конца и не закрыт ридер
		for r.currOffsetPreload < r.endOffsetPreload-1024 && !r.isClosed {
			off, err := buffReader.Read(buff)
			if err != nil {
				log.TLogln("Error read e head buffer", err)
				return
			}
			r.currOffsetPreload += int64(off)
			// пересчитываем конец загрузки
			r.endOffsetPreload = r.offset + r.cache.capacity
			// если лоадер не успевает загрузить данные и вошел на границу загрузки основного ридера, двигаем его
			if r.currOffsetPreload < r.offset+r.readahead {
				// подвигаем за границу основного ридера+1 кусок
				r.currOffsetPreload = r.offset + r.readahead + r.cache.pieceLength
				buffReader.Seek(r.currOffsetPreload, io.SeekStart)
			}
			// если ридер подобрался к концу-RAH
			if r.currOffsetPreload > r.endOffsetPreload-r.readahead-1024 && isReadahead {
				// читаем конец без RAH
				log.TLogln("disable buffering RAH")
				buffReader.SetReadahead(0)
				isReadahead = false
			} else if r.currOffsetPreload < r.endOffsetPreload-r.readahead-1024 && !isReadahead {
				// Конец удалился и можно включить RAH
				log.TLogln("enable buffering RAH")
				buffReader.SetReadahead(r.readahead)
				isReadahead = true
			}
			//log.TLogln(humanize.IBytes(uint64(r.offset)), humanize.IBytes(uint64(r.currOffsetPreload)), humanize.IBytes(uint64(r.endOffsetPreload)))
		}
		log.TLogln("End buffering")
	}()
}

func (r *Reader) findPreloadedStart() int64 {
	found := false
	pstart := r.getPieceNum(r.offset + r.readahead)
	pend := r.getPieceNum(r.endOffsetPreload)
	for i := pstart; i < pend; i++ {
		if r.cache.pieces[i].Size < r.cache.pieces[i].Length {
			pstart = i
			found = true
			break
		}
	}
	if !found {
		return r.endOffsetPreload
	}
	return int64(pstart) * r.cache.pieceLength
}

func (r *Reader) getRangePieces(offCurr, offReader, offEnd int64) (int, int, int) {
	currPiece := r.getPieceNum(offCurr)
	readerPiece := r.getPieceNum(offReader)
	endPiece := r.getPieceNum(offEnd)
	return currPiece, readerPiece, endPiece
}

func (r *Reader) getPieceNum(offset int64) int {
	return int((offset + r.file.Offset()) / r.cache.pieceLength)
}
