package osrscache

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type BinaryReader struct {
	reader io.ReadSeeker
	offset int64
}

func NewBinaryReader(data []byte) *BinaryReader {
	return &BinaryReader{
		reader: bytes.NewReader(data),
		offset: 0,
	}
}

func (br *BinaryReader) Seek(offset int64, whence int) (int64, error) {
	newOffset, err := br.reader.Seek(offset, whence)
	if err != nil {
		return 0, err
	}
	br.offset = newOffset
	return newOffset, nil
}

func (br *BinaryReader) GetOffset() int64 {
	return br.offset
}

func (br *BinaryReader) ReadUint8() (uint8, error) {
	var value uint8
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.offset += 1
	return value, nil
}

func (br *BinaryReader) ReadInt8() (int8, error) {
	var value int8
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.offset += 1
	return value, nil
}

func (br *BinaryReader) ReadUint16() (uint16, error) {
	var value uint16
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.offset += 2
	return value, nil
}

func (br *BinaryReader) ReadInt16() (int16, error) {
	var value int16
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.offset += 2
	return value, nil
}

func (br *BinaryReader) ReadUint32() (uint32, error) {
	var value uint32
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.offset += 4
	return value, nil
}

func (br *BinaryReader) ReadUint24() (uint32, error) {
	var buf [3]byte
	_, err := io.ReadFull(br.reader, buf[:])
	if err != nil {
		return 0, err
	}
	br.offset += 3
	return uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2]), nil
}

func (br *BinaryReader) ReadInt32() (int32, error) {
	var value int32
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.offset += 4
	return value, nil
}

func (br *BinaryReader) ReadBigSmart2() (int32, error) {
	value, err := br.ReadUint16()
	if err != nil {
		return 0, fmt.Errorf("reading initial uint16 for BigSmart2: %w", err)
	}
	if value == 0 {
		return -1, nil
	}
	if value < 32768 {
		return int32(value - 1), nil
	}
	value2, err := br.ReadUint32()
	if err != nil {
		return 0, fmt.Errorf("reading uint32 for BigSmart2: %w", err)
	}
	return int32(value2 - 0x10000), nil
}

func (br *BinaryReader) ReadUint16SmartMinus1() (uint16, error) {
	value, err := br.ReadUint16()
	if err != nil {
		return 0, fmt.Errorf("reading uint16 for SmartMinus1: %w", err)
	}
	if value == 32767 {
		return 0, nil
	}
	return value + 1, nil
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
