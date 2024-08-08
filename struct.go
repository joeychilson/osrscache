package osrscache

import (
	"errors"
	"fmt"
	"io"
)

type Struct struct {
	ID     uint16
	Params map[uint32]any
}

func NewStruct(id uint16) *Struct {
	return &Struct{ID: id, Params: make(map[uint32]any)}
}

func (s *Struct) Read(data []byte) error {
	reader := NewReader(data)
	for {
		opcode, err := reader.ReadUint8()
		if err != nil {
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
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			s.Params = make(map[uint32]any, length)
			for i := 0; i < int(length); i++ {
				isString, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf("reading is string: %w", err)
				}

				key, err := reader.ReadUint24()
				if err != nil {
					return fmt.Errorf("reading key: %w", err)
				}

				var value interface{}
				if isString == 1 {
					strValue, err := reader.ReadString()
					if err != nil {
						return fmt.Errorf("reading string value: %w", err)
					}
					value = strValue
				} else {
					intValue, err := reader.ReadUint32()
					if err != nil {
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
