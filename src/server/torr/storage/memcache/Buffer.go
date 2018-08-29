package memcache

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
	buffs map[int]*buffer
	frees int
	size  int64
	mu    sync.Mutex
}

func NewBufferPool(bufferLength int64, capacity int64) *BufferPool {
	bp := new(BufferPool)
	buffsSize := int(capacity/bufferLength) + 3
	bp.frees = buffsSize
	bp.size = bufferLength
	return bp
}

func (b *BufferPool) mkBuffs() {
	if b.buffs != nil {
		return
	}
	b.buffs = make(map[int]*buffer, b.frees)
	fmt.Println("Create", b.frees, "buffers")
	for i := 0; i < b.frees; i++ {
		buf := buffer{
			-1,
			make([]byte, b.size),
			false,
		}
		b.buffs[i] = &buf
	}
}

func (b *BufferPool) GetBuffer(p *Piece) (buff []byte, index int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mkBuffs()
	for id, buf := range b.buffs {
		if !buf.used {
			buf.used = true
			buf.pieceId = p.Id
			buff = buf.buf
			index = id
			b.frees--
			//fmt.Printf("Get buffer: %v %v %v %p\n", id, p.Id, b.frees, buff)
			return
		}
	}
	fmt.Println("Create slow buffer")
	return make([]byte, b.size), -1
}

func (b *BufferPool) ReleaseBuffer(index int) {
	if index == -1 {
		utils.FreeOSMem()
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mkBuffs()
	if buff, ok := b.buffs[index]; ok {
		buff.used = false
		buff.pieceId = -1
		b.frees++
		//fmt.Println("Release buffer:", index, b.frees)
	} else {
		utils.FreeOSMem()
	}
}

func (b *BufferPool) Len() int {
	return b.frees
}
