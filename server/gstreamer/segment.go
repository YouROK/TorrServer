package gstreamer

import (
	"bytes"
	"io"
)

type Segment struct {
	Header       []byte
	Payloads     [][]byte
	StartSeconds float64
}

func (s Segment) Len() int {
	total := len(s.Header)
	for _, payload := range s.Payloads {
		total += len(payload)
	}
	return total
}

func (s Segment) Empty() bool {
	return s.Len() == 0
}

func (s Segment) WriteTo(dst io.Writer) error {
	if _, err := dst.Write(s.Header); err != nil {
		return err
	}
	for _, payload := range s.Payloads {
		if len(payload) == 0 {
			continue
		}
		if _, err := dst.Write(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s Segment) WriteRange(dst io.Writer, offset int64, count int64) error {
	if count <= 0 || offset < 0 || offset >= int64(s.Len()) {
		return nil
	}

	remainingOffset := offset
	remainingCount := count
	for _, part := range s.parts() {
		if len(part) == 0 {
			continue
		}

		partLength := int64(len(part))
		if remainingOffset >= partLength {
			remainingOffset -= partLength
			continue
		}

		start := remainingOffset
		length := partLength - start
		if length > remainingCount {
			length = remainingCount
		}

		if _, err := dst.Write(part[start : start+length]); err != nil {
			return err
		}

		remainingCount -= length
		if remainingCount <= 0 {
			return nil
		}
		remainingOffset = 0
	}
	return nil
}

func (s Segment) parts() [][]byte {
	parts := make([][]byte, 0, 1+len(s.Payloads))
	if len(s.Header) > 0 {
		parts = append(parts, s.Header)
	}
	parts = append(parts, s.Payloads...)
	return parts
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
