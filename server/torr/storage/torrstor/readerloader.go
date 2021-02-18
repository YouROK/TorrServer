package torrstor

import (
	"github.com/anacrolix/torrent"
	"server/settings"
)

func (r *Reader) getPiecesRange() Range {
	startOff, endOff := r.getOffsetRange()
	return Range{r.getPieceNum(startOff), r.getPieceNum(endOff)}
}

func (r *Reader) getReaderPiece() int {
	readerOff := r.offset
	return r.getPieceNum(readerOff)
}

func (r *Reader) getPieceNum(offset int64) int {
	return int((offset + r.file.Offset()) / r.cache.pieceLength)
}

func (r *Reader) getOffsetRange() (int64, int64) {
	prc := int64(settings.BTsets.ReaderReadAHead)
	readers := int64(len(r.cache.readers))
	if readers == 0 {
		readers = 1
	}

	beginOffset := r.offset - (r.cache.capacity/readers)*(100-prc)/100
	endOffset := r.offset + (r.cache.capacity/readers)*prc/100

	if beginOffset < 0 {
		beginOffset = 0
	}

	if endOffset > r.file.Length() {
		endOffset = r.file.Length()
	}
	return beginOffset, endOffset
}

func (r *Reader) preload() {
	rrange := r.getPiecesRange()
	if rrange.Start == r.ranges.Start && rrange.End == r.ranges.End {
		return
	}

	torr := r.file.Torrent()
	r.ranges = rrange
	rahPiece := int(r.readahead / torr.Info().PieceLength)
	readerPiece := r.getReaderPiece()

	// from reader readahead to end of range
	for i := readerPiece + rahPiece; i < rrange.End; i++ {
		if torr.Piece(i).State().Priority == torrent.PiecePriorityNone {
			torr.Piece(i).SetPriority(torrent.PiecePriorityNormal)
		}
	}
}
