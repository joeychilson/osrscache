package osrscache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type EnumType struct {
	ID            uint16
	KeyType       byte
	ValueType     byte
	DefaultString string
	DefaultInt    int32
	strings       map[int32]string
	ints          map[int32]int32
}

func NewEnumType(id uint16, data []byte) (*EnumType, error) {
	e := &EnumType{
		ID:            id,
		DefaultString: "null",
	}
	if err := e.Read(data); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *EnumType) Read(data []byte) error {
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
			s, err := ReadString(reader)
			if err != nil {
				return err
			}
			e.DefaultString = s
		case 4:
			if err := binary.Read(reader, binary.BigEndian, &e.DefaultInt); err != nil {
				return err
			}
		case 5:
			var size uint16
			if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
				return err
			}
			e.strings = make(map[int32]string, size)
			for i := uint16(0); i < size; i++ {
				var key int32
				if err := binary.Read(reader, binary.BigEndian, &key); err != nil {
					return err
				}
				value, err := ReadString(reader)
				if err != nil {
					return err
				}
				e.strings[key] = value
			}
		case 6:
			var size uint16
			if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
				return err
			}
			e.ints = make(map[int32]int32, size)
			for i := uint16(0); i < size; i++ {
				var key, value int32
				if err := binary.Read(reader, binary.BigEndian, &key); err != nil {
					return err
				}
				if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
					return err
				}
				e.ints[key] = value
			}
		default:
			return fmt.Errorf("unsupported opcode: %d", opcode)
		}
	}
	return nil
}

func (e *EnumType) GetString(key int32) string {
	if e.strings == nil {
		return e.DefaultString
	}
	if value, ok := e.strings[key]; ok {
		return value
	}
	return e.DefaultString
}

func (e *EnumType) GetInt(key int32) int32 {
	if e.ints == nil {
		return e.DefaultInt
	}
	if value, ok := e.ints[key]; ok {
		return value
	}
	return e.DefaultInt
}
