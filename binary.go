package osrscache

import (
	"encoding/binary"
	"io"
)

type BinaryReader struct {
	reader io.Reader
}

func NewBinaryReader(r io.Reader) *BinaryReader {
	return &BinaryReader{reader: r}
}

func (br *BinaryReader) ReadUint8() (uint8, error) {
	var value uint8
	err := binary.Read(br.reader, binary.BigEndian, &value)
	return value, err
}

func (br *BinaryReader) ReadInt8() (int8, error) {
	var value int8
	err := binary.Read(br.reader, binary.BigEndian, &value)
	return value, err
}

func (br *BinaryReader) ReadUint16() (uint16, error) {
	var value uint16
	err := binary.Read(br.reader, binary.BigEndian, &value)
	return value, err
}

func (br *BinaryReader) ReadInt16() (int16, error) {
	var value int16
	err := binary.Read(br.reader, binary.BigEndian, &value)
	return value, err
}

func (br *BinaryReader) ReadUint32() (uint32, error) {
	var value uint32
	err := binary.Read(br.reader, binary.BigEndian, &value)
	return value, err
}

func (br *BinaryReader) ReadInt32() (int32, error) {
	var value int32
	err := binary.Read(br.reader, binary.BigEndian, &value)
	return value, err
}

func (br *BinaryReader) ReadUint24() (uint32, error) {
	var buf [3]byte
	if _, err := io.ReadFull(br.reader, buf[:]); err != nil {
		return 0, err
	}
	return uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2]), nil
}

func (br *BinaryReader) ReadString() (string, error) {
	var result []byte
	for {
		b, err := br.ReadUint8()
		if err != nil {
			return "", err
		}
		if b == 0 {
			break
		}
		result = append(result, b)
	}
	return string(result), nil
}
