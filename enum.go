package osrscache

import (
	"bytes"
	"encoding/binary"
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

func NewEnum(id uint16, data []byte) (*Enum, error) {
	e := &Enum{
		ID:           id,
		DefaultValue: nil,
		Values:       make(map[int32]any),
	}
	if err := e.Read(data); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Enum) Read(data []byte) error {
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
		case 1:
			if err := binary.Read(reader, binary.BigEndian, &e.KeyType); err != nil {
				return err
			}
		case 2:
			if err := binary.Read(reader, binary.BigEndian, &e.ValueType); err != nil {
				return err
			}
		case 3:
			defaultStr, err := ReadString(reader)
			if err != nil {
				return err
			}
			e.DefaultValue = defaultStr
		case 4:
			var defaultInt int32
			if err := binary.Read(reader, binary.BigEndian, &defaultInt); err != nil {
				return err
			}
			e.DefaultValue = defaultInt
		case 5, 6:
			var size uint16
			if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
				return err
			}
			for i := uint16(0); i < size; i++ {
				var key int32
				if err := binary.Read(reader, binary.BigEndian, &key); err != nil {
					return err
				}
				if opcode == 5 {
					strValue, err := ReadString(reader)
					if err != nil {
						return err
					}
					e.Values[key] = strValue
				} else {
					var intValue int32
					if err := binary.Read(reader, binary.BigEndian, &intValue); err != nil {
						return err
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
