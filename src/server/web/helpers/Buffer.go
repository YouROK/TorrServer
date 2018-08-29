package helpers

import "bytes"

type SeekingBuffer struct {
	str    string
	buffer *bytes.Buffer
	offset int64
	size   int64
}

func NewSeekingBuffer(str string) *SeekingBuffer {
	return &SeekingBuffer{
		str:    str,
		buffer: bytes.NewBufferString(str),
		offset: 0,
	}
}

func (fb *SeekingBuffer) Read(p []byte) (n int, err error) {
	n, err = fb.buffer.Read(p)
	fb.offset += int64(n)
	return n, err
}

func (fb *SeekingBuffer) Seek(offset int64, whence int) (ret int64, err error) {
	var newoffset int64
	switch whence {
	case 0:
		newoffset = offset
	case 1:
		newoffset = fb.offset + offset
	case 2:
		newoffset = int64(len(fb.str)) - offset
	}
	if newoffset == fb.offset {
		return newoffset, nil
	}
	fb.buffer = bytes.NewBufferString(fb.str[newoffset:])
	fb.offset = newoffset
	return fb.offset, nil
}
