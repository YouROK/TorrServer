package torrstor

import (
	"github.com/anacrolix/torrent"
	"server/settings"
)

func (r *Reader) getPiecesRange() (int, int) {
	startOff, endOff := r.getReaderRange()
	return r.getPieceNum(startOff), r.getPieceNum(endOff)
}

func (r *Reader) getReaderPiece() int {
	readerOff := r.offset
	return r.getPieceNum(readerOff)
}

func (r *Reader) getPieceNum(offset int64) int {
	return int((offset + r.file.Offset()) / r.cache.pieceLength)
}

func (r *Reader) getReaderRange() (int64, int64) {
	prc := int64(settings.BTsets.ReaderPreload)
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
	begin, end := r.getPiecesRange()
	rahPiece := int(r.readahead / torr.Info().PieceLength)
	readerPiece := r.getReaderPiece()

	for i := r.lastRangeBegin; i < r.lastRangeEnd; i++ {
		if i >= readerPiece && i <= readerPiece+rahPiece { // reader pieces
			continue
		}
		piece := torr.Piece(i)
		piece.SetPriority(torrent.PiecePriorityNone)
	}

	for i := begin; i < end; i++ {
		if i <= readerPiece+rahPiece { // reader pieces
			continue
		}
		torr.Piece(i).SetPriority(torrent.PiecePriorityNormal)
	}
	r.lastRangeBegin, r.lastRangeEnd = begin, end
}
