package osrscache

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Reader struct {
	data []byte
	pos  int
}

func NewReader(data []byte) *Reader {
	return &Reader{data: data, pos: 0}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *Reader) ReadByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.data) {
		return nil, io.EOF
	}
	result := make([]byte, n)
	copy(result, r.data[r.pos:r.pos+n])
	r.pos += n
	return result, nil
}

func (r *Reader) ReadInt8() (int8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return int8(b), nil
}

func (r *Reader) ReadInt16() (int16, error) {
	var buf [2]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(buf[:])), nil
}

func (r *Reader) ReadInt32() (int32, error) {
	var buf [4]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf[:])), nil
}

func (r *Reader) ReadUint8() (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return uint8(b), err
}

func (r *Reader) ReadUint16() (uint16, error) {
	var buf [2]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf[:]), nil
}

func (r *Reader) ReadUint24() (uint32, error) {
	var buf [3]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2]), nil
}

func (r *Reader) ReadUint32() (uint32, error) {
	var buf [4]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf[:]), nil
}

func (r *Reader) ReadSmartUint() (uint32, error) {
	firstByte, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	if firstByte&0x80 == 0 {
		secondByte, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint32(binary.BigEndian.Uint16([]byte{firstByte, secondByte})), nil
	}
	var restBytes [3]byte
	_, err = io.ReadFull(r, restBytes[:])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32([]byte{firstByte, restBytes[0], restBytes[1], restBytes[2]}) & 0x7FFFFFFF, nil
}

func (r *Reader) ReadString() (string, error) {
	var result []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && len(result) > 0 {
				break
			}
			return "", err
		}
		if b == 0 {
			break
		}
		result = append(result, b)
	}
	return string(result), nil
}

func (r *Reader) ReadBigSmart2() (int32, error) {
	value, err := r.ReadUint16()
	if err != nil {
		return 0, fmt.Errorf("reading initial uint16 for BigSmart2: %w", err)
	}
	if value == 0 {
		return -1, nil
	}
	if value < 32768 {
		return int32(value - 1), nil
	}
	value2, err := r.ReadUint32()
	if err != nil {
		return 0, fmt.Errorf("reading uint32 for BigSmart2: %w", err)
	}
	return int32(value2 - 0x10000), nil
}

func (r *Reader) ReadUint16SmartMinus1() (uint16, error) {
	value, err := r.ReadUint16()
	if err != nil {
		return 0, fmt.Errorf("reading uint16 for SmartMinus1: %w", err)
	}
	if value == 32767 {
		return 0, nil
	}
	return value + 1, nil
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = int64(r.pos) + offset
	case io.SeekEnd:
		abs = int64(len(r.data)) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}
	if abs < 0 {
		return 0, fmt.Errorf("negative position: %d", abs)
	}
	if abs > int64(len(r.data)) {
		r.pos = len(r.data)
	} else {
		r.pos = int(abs)
	}
	return int64(r.pos), nil
}

func (r *Reader) Len() int {
	return len(r.data) - r.pos
}

func (r *Reader) Size() int64 {
	return int64(len(r.data))
}

func (r *Reader) Reset(b []byte) {
	r.data = b
	r.pos = 0
}
