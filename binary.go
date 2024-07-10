package osrscache

import (
	"encoding/binary"
	"fmt"
	"io"
)

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
