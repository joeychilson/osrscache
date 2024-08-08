package osrscache

import (
	"errors"
	"fmt"
	"io"
)

type Enum struct {
	ID           uint16
	KeyType      byte
	ValueType    byte
	DefaultValue any
	Values       map[int32]any
}

func NewEnum(id uint16) *Enum {
	return &Enum{
		ID:           id,
		DefaultValue: nil,
		Values:       make(map[int32]any),
	}
}

func (e *Enum) Read(data []byte) error {
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
		case 1:
			e.KeyType, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading key type: %w", err)
			}
		case 2:
			e.ValueType, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading value type: %w", err)
			}
		case 3:
			defaultStr, err := reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading default string value: %w", err)
			}
			e.DefaultValue = defaultStr
		case 4:
			defaultInt, err := reader.ReadInt32()
			if err != nil {
				return fmt.Errorf("reading default int value: %w", err)
			}
			e.DefaultValue = defaultInt
		case 5, 6:
			size, err := reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading size: %w", err)
			}
			for i := uint16(0); i < size; i++ {
				key, err := reader.ReadInt32()
				if err != nil {
					return fmt.Errorf("reading key: %w", err)
				}
				if opcode == 5 {
					strValue, err := reader.ReadString()
					if err != nil {
						return fmt.Errorf("reading string value: %w", err)
					}
					e.Values[key] = strValue
				} else {
					intValue, err := reader.ReadInt32()
					if err != nil {
						return fmt.Errorf("reading int value: %w", err)
					}
					e.Values[key] = intValue
				}
			}
		default:
			return fmt.Errorf("unsupported opcode: %d", opcode)
		}
	}
	return nil
}
