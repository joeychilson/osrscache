package osrscache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type StructType struct {
	ID     uint16
	Params map[uint32]any
}

func NewStructType(id uint16, data []byte) (*StructType, error) {
	s := &StructType{
		ID:     id,
		Params: make(map[uint32]any),
	}
	if err := s.Read(data); err != nil {
		return nil, fmt.Errorf("reading struct type: %w", err)
	}
	return s, nil
}

func (s *StructType) Read(data []byte) error {
	reader := bytes.NewReader(data)
	for {
		var opcode uint8
		if err := binary.Read(reader, binary.BigEndian, &opcode); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("reading opcode: %w", err)
		}
		if opcode == 0 {
			break
		}
		switch opcode {
		case 249:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			s.Params = make(map[uint32]any, length)
			for i := 0; i < int(length); i++ {
				var isString uint8
				if err := binary.Read(reader, binary.BigEndian, &isString); err != nil {
					return fmt.Errorf("reading is string: %w", err)
				}

				key, err := ReadUint24(reader)
				if err != nil {
					return fmt.Errorf("reading key: %w", err)
				}

				var value interface{}
				if isString == 1 {
					strValue, err := ReadString(reader)
					if err != nil {
						return fmt.Errorf("reading string value: %w", err)
					}
					value = strValue
				} else {
					var intValue uint32
					if err := binary.Read(reader, binary.BigEndian, &intValue); err != nil {
						return fmt.Errorf("reading uint32 value: %w", err)
					}
					value = intValue
				}
				s.Params[key] = value
			}
		default:
			return fmt.Errorf("unsupported opcode: %d", opcode)

		}
	}
	return nil
}
