package torrshash

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Field struct {
	Tag   Tag    `json:"tag"`
	Value string `json:"value"`
}

type Tag int

const (
	TagTitle    = iota // string
	TagPoster          // string
	TagTracker         // string
	TagCategory        // string
	TagSize            // binary
)

func (tag Tag) String() string {
	switch tag {
	case TagTitle:
		return "Title"
	case TagPoster:
		return "Poster"
	case TagTracker:
		return "Tracker"
	case TagCategory:
		return "Category"
	case TagSize:
		return "Size"
	default:
		return "Unknown"
	}
}

func NewField(tag Tag, value string) *Field {
	return &Field{Tag: tag, Value: value}
}

func ReadField(reader io.Reader) (*Field, error) {
	tagb := make([]byte, 1)
	if _, err := reader.Read(tagb); err == io.EOF {
		return nil, nil
	}
	tag := Tag(tagb[0])

	if isBinary(tag) {
		var value int64
		err := binary.Read(reader, binary.LittleEndian, &value)
		if err != nil {
			return nil, err
		}
		return NewField(tag, strconv.FormatInt(value, 10)), nil
	}

	var length uint16
	err := binary.Read(reader, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}

	valBytes := make([]byte, length)
	n, err := io.ReadFull(reader, valBytes)

	if err != nil {
		return nil, err
	}

	if n != int(length) {
		return nil, fmt.Errorf("expected to read %v bytes, read %v", length, n)
	}

	return NewField(tag, string(valBytes)), nil
}

func (f *Field) write(writer io.Writer) error {
	value := strings.TrimSpace(f.Value)
	if len(value) == 0 {
		return nil
	}
	if isBinary(f.Tag) && value == "0" {
		return nil
	}

	_, err := writer.Write([]byte{byte(f.Tag)})
	if err != nil {
		return err
	}

	if isBinary(f.Tag) {
		ii, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		if ii == 0 {
			return nil
		}
		return binary.Write(writer, binary.LittleEndian, ii)
	}

	strBytes := []byte(value)
	err = binary.Write(writer, binary.LittleEndian, uint16(len(strBytes)))
	if err != nil {
		return err
	}
	_, err = writer.Write(strBytes)
	return err
}

func isBinary(t Tag) bool {
	switch t {
	case TagTitle:
	case TagPoster:
	case TagTracker:
	case TagCategory:
		return false
	case TagSize:
		return true
	default:
		return false
	}
	return false
}
