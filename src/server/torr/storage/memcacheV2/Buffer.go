package memcacheV2

import (
	"fmt"
	"sync"

	"server/utils"
)

type buffer struct {
	pieceId int
	buf     []byte
	used    bool
}

type BufferPool struct {
	buffs        map[int]*buffer
	bufferLength int64
	bufferCount  int
	mu           sync.Mutex
}

func NewBufferPool(bufferLength int64) *BufferPool {
	bp := new(BufferPool)
	bp.bufferLength = bufferLength
	bp.buffs = make(map[int]*buffer)
	return bp
}

func (b *BufferPool) GetBuffer(p *Piece) (buff []byte, index int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, buf := range b.buffs {
		if !buf.used {
			fmt.Println("Get buffer:", id)
			buf.used = true
			buf.pieceId = p.Id
			buff = buf.buf
			index = id
			return
		}
	}

	fmt.Println("Create buffer:", b.bufferCount)
	buf := new(buffer)
	buf.buf = make([]byte, b.bufferLength)
	buf.used = true
	buf.pieceId = p.Id
	b.buffs[b.bufferCount] = buf
	index = b.bufferCount
	buff = buf.buf
	b.bufferCount++
	return
}

func (b *BufferPool) ReleaseBuffer(index int) {
	if index == -1 {
		utils.FreeOSMem()
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if buff, ok := b.buffs[index]; ok {
		fmt.Println("Release buffer:", index)
		buff.used = false
		buff.pieceId = -1
	} else {
		utils.FreeOSMem()
	}
}

func (b *BufferPool) Used() map[int]struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()
	used := make(map[int]struct{})
	for _, b := range b.buffs {
		if b.used {
			used[b.pieceId] = struct{}{}
		}
	}
	return used
}

func (b *BufferPool) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	count := 0
	for _, b := range b.buffs {
		if b.used {
			count++
		}
	}
	return count
}
