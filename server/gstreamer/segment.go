package gstreamer

import "bytes"

type Segment struct {
	Data         []byte
	StartSeconds float64
}

func (s Segment) Len() int {
	return len(s.Data)
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func takeBuffer(buf *bytes.Buffer) []byte {
	if buf.Len() == 0 {
		return nil
	}

	data := buf.Bytes()
	*buf = bytes.Buffer{}
	return data
}
