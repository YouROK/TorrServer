package torrstor

import (
	"github.com/anacrolix/torrent"
	"server/settings"
)

type Range struct {
	Start, End int
}

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
	beginOffset := r.offset - r.cache.capacity*(100-prc)/100
	endOffset := r.offset + r.cache.capacity*prc/100

	if beginOffset < 0 {
		beginOffset = 0
	}

	if endOffset > r.file.Length() {
		endOffset = r.file.Length()
	}
	return beginOffset, endOffset
}

func (r *Reader) preload() {
	torr := r.file.Torrent()
	rrange := r.getPiecesRange()
	rahPiece := int(r.readahead / torr.Info().PieceLength)
	readerPiece := r.getReaderPiece()

	// from reader readahead to end of range
	for i := readerPiece + rahPiece; i < rrange.End; i++ {
		if torr.Piece(i).State().Priority == torrent.PiecePriorityNone {
			torr.Piece(i).SetPriority(torrent.PiecePriorityHigh)
		}
	}
}
