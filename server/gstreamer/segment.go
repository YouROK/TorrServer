//go:build gst

package gstreamer

import "io"

type Segment struct {
	Header       []byte
	Payloads     [][]byte
	StartNS      uint64
	EndNS        uint64
	StartSeconds float64
	EndSeconds   float64
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

func (s Segment) WriteTo(dst io.Writer) (int64, error) {
	written := int64(0)
	count, err := writeBytes(dst, s.Header)
	written += int64(count)
	if err != nil {
		return written, err
	}
	for _, payload := range s.Payloads {
		if len(payload) == 0 {
			continue
		}
		count, err = writeBytes(dst, payload)
		written += int64(count)
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (s Segment) WriteRange(dst io.Writer, offset int64, count int64) error {
	if count <= 0 || offset < 0 || offset >= int64(s.Len()) {
		return nil
	}

	if err := writeRangePart(dst, s.Header, &offset, &count); err != nil || count <= 0 {
		return err
	}
	for _, payload := range s.Payloads {
		if err := writeRangePart(dst, payload, &offset, &count); err != nil || count <= 0 {
			return err
		}
	}
	return nil
}

func writeRangePart(dst io.Writer, part []byte, offset *int64, count *int64) error {
	if len(part) == 0 || *count <= 0 {
		return nil
	}

	partLength := int64(len(part))
	if *offset >= partLength {
		*offset -= partLength
		return nil
	}

	start := *offset
	length := partLength - start
	if length > *count {
		length = *count
	}

	if _, err := writeBytes(dst, part[start:start+length]); err != nil {
		return err
	}

	*count -= length
	*offset = 0
	return nil
}

func writeBytes(dst io.Writer, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	written, err := dst.Write(data)
	if err != nil {
		return written, err
	}
	if written != len(data) {
		return written, io.ErrShortWrite
	}
	return written, nil
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
