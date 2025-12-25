package torrshash

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"io"
	"strings"
)

func Pack(hash *TorrsHash) (string, error) {
	data, err := pack(hash)
	if err != nil {
		return "", err
	}
	return Encode62(data), nil
}

func PackBytes(hash *TorrsHash) ([]byte, error) {
	return pack(hash)
}

func Unpack(token string) (*TorrsHash, error) {
	data := Decode62(strings.TrimSpace(token))
	return UnpackBytes(data)
}

func UnpackBytes(data []byte) (*TorrsHash, error) {
	return unpack(data)
}

func pack(t *TorrsHash) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw, err := zlib.NewWriterLevel(buf, zlib.BestCompression)
	if err != nil {
		return nil, err
	}

	hashBytes, _ := hex.DecodeString(strings.TrimSpace(t.Hash))
	_, err = zw.Write(hashBytes)
	if err != nil {
		return nil, err
	}

	for _, f := range t.Fields {
		err = f.write(zw)
		if err != nil {
			return nil, err
		}
	}

	err = zw.Close()
	return buf.Bytes(), err
}

func unpack(data []byte) (*TorrsHash, error) {
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	hashBuf := make([]byte, 20)
	if _, err = io.ReadFull(zr, hashBuf); err != nil {
		return nil, err
	}

	th := &TorrsHash{}
	th.Hash = hex.EncodeToString(hashBuf)

	for {
		field, err := ReadField(zr)
		if err == nil && field == nil {
			//End on read
			return th, nil
		}
		if err != nil {
			return th, err
		}

		th.Fields = append(th.Fields, field)
	}
}
