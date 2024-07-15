package osrscache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Enum[K int | int32, V string | int | int32] struct {
	ID           uint16
	KeyType      byte
	ValueType    byte
	DefaultValue V
	Values       map[K]V
}

func NewEnum[K int | int32, V string | int | int32](id uint16, data []byte) (*Enum[K, V], error) {
	e := &Enum[K, V]{
		ID: id,
	}
	if err := e.Read(data); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Enum[K, V]) Read(data []byte) error {
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
				return fmt.Errorf("reading key type: %w", err)
			}
		case 2:
			if err := binary.Read(reader, binary.BigEndian, &e.ValueType); err != nil {
				return fmt.Errorf("reading value type: %w", err)
			}
		case 3:
			s, err := ReadString(reader)
			if err != nil {
				return fmt.Errorf("reading default string value: %w", err)
			}
			e.DefaultValue = any(s).(V)
		case 4:
			var defaultInt int32
			if err := binary.Read(reader, binary.BigEndian, &defaultInt); err != nil {
				return fmt.Errorf("reading default int value: %w", err)
			}
			e.DefaultValue = any(defaultInt).(V)
		case 5, 6:
			var size uint16
			if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
				return fmt.Errorf("reading size: %w", err)
			}
			for i := uint16(0); i < size; i++ {
				var key K
				if err := binary.Read(reader, binary.BigEndian, &key); err != nil {
					return fmt.Errorf("reading key: %w", err)
				}
				var value V
				if opcode == 5 {
					s, err := ReadString(reader)
					if err != nil {
						return fmt.Errorf("reading string value: %w", err)
					}
					value = any(s).(V)
				} else {
					var intValue int32
					if err := binary.Read(reader, binary.BigEndian, &intValue); err != nil {
						return fmt.Errorf("reading int value: %w", err)
					}
					value = any(intValue).(V)
				}
				e.Values[key] = value
			}
		default:
			return fmt.Errorf("unsupported opcode: %d", opcode)
		}
	}
	return nil
}
