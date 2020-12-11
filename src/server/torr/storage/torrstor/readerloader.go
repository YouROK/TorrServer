package torrstor

import (
	"io"

	"github.com/dustin/go-humanize"
	"server/log"
)

func (r *Reader) getUsedPiecesRange() (int, int) {
	startOff := r.getStartLoaderOffset()
	endOff := r.getEndLoaderOffset()
	return r.getPieceNum(startOff), r.getPieceNum(endOff)
}

func (r *Reader) getReaderPieces() (int, int) {
	readerOff := r.offset
	loaderOff := r.loaderOffset
	return r.getPieceNum(readerOff), r.getPieceNum(loaderOff)
}

//////////////////////////////////////////////////////////////
/// Прелоадер начинает загрузку от старта плеера+RAH и имеет свой RAH
/// Прелоадер грузит до конца-RAH
func (r *Reader) preload() {
	r.muPreload.Lock()
	// загрузка уже идет или конец меньше RAH, тогда старается основной ридер
	if r.isPreload || r.getEndLoaderOffset()-r.loaderOffset < r.readahead {
		r.muPreload.Unlock()
		return
	}
	r.isPreload = true
	r.muPreload.Unlock()

	log.TLogln("Start buffering from", humanize.IBytes(uint64(r.loaderOffset)))
	go func() {
		// получаем ридер
		buffReader := r.file.NewReader()
		defer func() {
			r.isPreload = false
			buffReader.Close()
			log.TLogln("End buffering")
		}()
		// ищем не прочитанный кусок
		r.loaderOffset = r.findPreloadedStart()
		// выходим если ничего подгружать не нужно
		if r.loaderOffset >= r.getEndLoaderOffset() {
			return
		}
		// двигаем лоадер
		buffReader.Seek(r.loaderOffset, io.SeekStart)
		buff := make([]byte, 1024)
		// isReadahead чтобы меньше переключать RAH
		isReadahead := false
		buffReader.SetReadahead(0)
		// читаем пока позиция лоадера меньше конца и не закрыт ридер
		for r.loaderOffset < r.getEndLoaderOffset()-1024 && !r.isClosed {
			off, err := buffReader.Read(buff)
			if err != nil {
				log.TLogln("Error read e head buffer", err)
				return
			}
			r.loaderOffset += int64(off)
			// если лоадер не успевает загрузить данные и вошел на границу загрузки основного ридера, двигаем его
			if r.loaderOffset < r.offset+r.readahead {
				// подвигаем за границу основного ридера+1 кусок
				r.loaderOffset = r.offset + r.readahead + r.cache.pieceLength
				buffReader.Seek(r.loaderOffset, io.SeekStart)
			}
			// если ридер подобрался к концу-RAH
			if r.loaderOffset > r.getEndLoaderOffset()-r.readahead-1024 && isReadahead {
				// читаем конец без RAH
				log.TLogln("disable buffering RAH")
				buffReader.SetReadahead(0)
				isReadahead = false
			} else if r.loaderOffset < r.getEndLoaderOffset()-r.readahead-1024 && !isReadahead {
				// Конец удалился и можно включить RAH
				log.TLogln("enable buffering RAH")
				buffReader.SetReadahead(r.readahead)
				isReadahead = true
			}
			//log.TLogln(humanize.IBytes(uint64(r.offset)), humanize.IBytes(uint64(r.loaderOffset)), humanize.IBytes(uint64(r.endLoaderOffset)))
		}
	}()
}

func (r *Reader) getStartLoaderOffset() int64 {
	off := r.offset - r.cache.capacity/2
	if off < 0 {
		off = 0
	}
	return off
}

func (r *Reader) getEndLoaderOffset() int64 {
	off := r.offset + r.cache.capacity/2
	if off > r.file.Length() {
		off = r.file.Length()
	}
	return off
}

func (r *Reader) findPreloadedStart() int64 {
	found := false
	pstart := r.getPieceNum(r.offset + r.readahead)
	pend := r.getPieceNum(r.getEndLoaderOffset())
	for i := pstart; i < pend; i++ {
		if r.cache.pieces[i].Size < r.cache.pieces[i].Length {
			pstart = i
			found = true
			break
		}
	}
	if !found {
		return r.getEndLoaderOffset()
	}
	return int64(pstart) * r.cache.pieceLength
}

func (r *Reader) getPieceNum(offset int64) int {
	return int((offset + r.file.Offset()) / r.cache.pieceLength)
}
