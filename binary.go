package osrscache

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type BinaryReader struct {
	reader io.ReadSeeker
}

func NewBinaryReader(data []byte) *BinaryReader {
	return &BinaryReader{
		reader: bytes.NewReader(data),
	}
}

func (br *BinaryReader) ReadByte() (byte, error) {
	var value byte
	err := binary.Read(br.reader, binary.BigEndian, &value)
	return value, err
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

func (br *BinaryReader) ReadUint24() (uint32, error) {
	var buf [3]byte
	_, err := io.ReadFull(br.reader, buf[:])
	if err != nil {
		return 0, err
	}
	return uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2]), nil
}

func (br *BinaryReader) ReadInt32() (int32, error) {
	var value int32
	err := binary.Read(br.reader, binary.BigEndian, &value)
	return value, err
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

func (br *BinaryReader) Seek(offset int64, whence int) (int64, error) {
	return br.reader.Seek(offset, whence)
}

func ReadUint24(r io.Reader) (uint32, error) {
	var buf [3]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2]), nil
}

func ReadBigSmart2(r io.Reader) (int32, error) {
	var value uint16
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, fmt.Errorf("reading initial uint16 for BigSmart2: %w", err)
	}
	if value == 0 {
		return -1, nil
	}
	if value < 32768 {
		return int32(value - 1), nil
	}
	var value2 uint32
	if err := binary.Read(r, binary.BigEndian, &value2); err != nil {
		return 0, fmt.Errorf("reading uint32 for BigSmart2: %w", err)
	}
	return int32(value2 - 0x10000), nil
}

func ReadUint16SmartMinus1(r io.Reader) (uint16, error) {
	var value uint16
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, fmt.Errorf("reading uint16 for SmartMinus1: %w", err)
	}
	if value == 32767 {
		return 0, nil
	}
	return value + 1, nil
}

func ReadString(r io.Reader) (string, error) {
	var result []byte
	for {
		var b [1]byte
		_, err := r.Read(b[:])
		if err != nil {
			if err == io.EOF && len(result) > 0 {
				break
			}
			return "", err
		}
		if b[0] == 0 {
			break
		}
		result = append(result, b[0])
	}
	return string(result), nil
}
