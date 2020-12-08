package torrstor

import (
	"io"

	"github.com/dustin/go-humanize"
	"server/log"
)

func (r *Reader) getUsedPieces() (int, int) {
	startOff, endOff := r.offset, r.endOffsetPreload
	if startOff < endOff {
		endOff = startOff + r.readahead
	}
	return r.getRangePieces(r.offset, r.endOffsetPreload)
}

func (r *Reader) preload() {
	r.endOffsetPreload = r.offset + r.cache.capacity - 1024

	if r.endOffsetPreload > r.file.Length() {
		r.endOffsetPreload = r.file.Length()
	}

	if r.isPreload || r.endOffsetPreload < r.readahead {
		return
	}

	r.isPreload = true
	log.TLogln("Start buffering from", humanize.IBytes(uint64(r.currOffsetPreload)))
	go func() {
		buffReader := r.file.NewReader()
		defer func() {
			r.isPreload = false
			buffReader.Close()
		}()
		buffReader.SetReadahead(0)
		r.currOffsetPreload = r.offset + r.readahead
		buffReader.Seek(r.currOffsetPreload, io.SeekStart)
		buff := make([]byte, 1024)
		for r.currOffsetPreload < r.endOffsetPreload && !r.isClosed {
			off, err := buffReader.Read(buff)
			if err != nil {
				log.TLogln("Error read e head buffer", err)
				return
			}
			r.currOffsetPreload += int64(off)
			r.endOffsetPreload = r.offset + r.cache.capacity - 1024
			if r.currOffsetPreload < r.offset {
				r.currOffsetPreload = r.offset + r.readahead
				buffReader.Seek(r.currOffsetPreload, io.SeekStart)
			}
			//log.TLogln(humanize.IBytes(uint64(r.offset)), humanize.IBytes(uint64(r.currOffsetPreload)), humanize.IBytes(uint64(r.endOffsetPreload)))
		}
		log.TLogln("End buffering")
	}()
}

func (r *Reader) getRangePieces(offCurr, offEnd int64) (int, int) {
	currPiece := r.getPieceNum(offCurr)
	endPiece := r.getPieceNum(offEnd)
	return currPiece, endPiece
}

func (r *Reader) getPieceNum(offset int64) int {
	return int((offset + r.file.Offset()) / r.cache.pieceLength)
}
