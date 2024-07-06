package osrscache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type ItemDefinitions map[uint32]*ItemDefinition

func (d ItemDefinitions) Get(id uint32) (*ItemDefinition, error) {
	def, ok := d[id]
	if !ok {
		return nil, fmt.Errorf("item definition not found")
	}
	return def, nil
}

type ItemDefinition struct {
	ID                    uint16
	Category              uint16
	Name                  string
	Examine               string
	MembersOnly           bool
	Stackable             bool
	Tradeable             bool
	Exchangeable          bool
	Value                 int32
	Weight                int16
	ActionsGround         [5]string
	ActionsInventory      [5]string
	NotedItemID           uint16
	NotedTemplate         uint16
	StackItemIDs          [10]uint16
	StackQuantities       [10]uint16
	Team                  int8
	BoughtLinkID          uint16
	BoughtTemplate        uint16
	PlaceholderItemID     uint16
	PlaceholderTemplate   uint16
	ShiftClickDropIndex   uint8
	Params                map[uint32]any
	InventoryModel        InventoryModel
	CharacterModelMale    CharacterModel
	CharacterModelFemale  CharacterModel
	WearPositionPrimary   uint8
	WearPositionSecondary uint8
	WearPositionTertiary  uint8
}

type InventoryModel struct {
	ID            uint16
	Zoom          uint16
	RotationX     uint16
	RotationY     uint16
	RotationZ     uint16
	OffsetX       uint16
	OffsetY       uint16
	ScaleX        uint16
	ScaleY        uint16
	ScaleZ        uint16
	RecolorFrom   []uint16
	RecolorTo     []uint16
	RetextureFrom []uint16
	RetextureTo   []uint16
	Ambient       int8
	Contrast      int8
}

type CharacterModel struct {
	ModelPrimary           uint16
	ModelSecondary         uint16
	ModelTertiary          uint16
	Offset                 uint8
	ChatHeadModelPrimary   uint16
	ChatHeadModelSecondary uint16
}

func NewItemDefinition(id uint16, data []byte) (*ItemDefinition, error) {
	def := &ItemDefinition{
		ID:               id,
		Name:             "null",
		ActionsGround:    [5]string{"", "", "Take", "", ""},
		ActionsInventory: [5]string{"", "", "", "", "Drop"},
		Params:           make(map[uint32]any),
		InventoryModel: InventoryModel{
			Zoom:   2000,
			ScaleX: 128,
			ScaleY: 128,
			ScaleZ: 128,
		},
	}
	err := def.Read(data)
	if err != nil {
		return nil, fmt.Errorf("reading item definition: %w", err)
	}
	return def, nil
}

func (def *ItemDefinition) Read(data []byte) error {
	reader := NewBinaryReader(bytes.NewReader(data))
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
			def.InventoryModel.ID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model id: %w", err)
			}
		case 2:
			def.Name, err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading name: %w", err)
			}
		case 3:
			def.Examine, err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading examine: %w", err)
			}
		case 4:
			def.InventoryModel.Zoom, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model zoom2d: %w", err)
			}
		case 5:
			def.InventoryModel.RotationX, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model xan2d: %w", err)
			}
		case 6:
			def.InventoryModel.RotationY, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model yan2d: %w", err)
			}
		case 7:
			def.InventoryModel.OffsetX, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model xoffset2d: %w", err)
			}
		case 8:
			def.InventoryModel.OffsetY, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model yoffset2d: %w", err)
			}
		case 11:
			def.Stackable = true
		case 12:
			def.Value, err = reader.ReadInt32()
			if err != nil {
				return fmt.Errorf("reading value: %w", err)
			}
		case 13:
			def.WearPositionPrimary, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading wear position primary: %w", err)
			}
		case 14:
			def.WearPositionSecondary, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading wear position secondary: %w", err)
			}
		case 16:
			def.MembersOnly = true
		case 23:
			def.CharacterModelMale.ModelPrimary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model primary: %w", err)
			}
			def.CharacterModelMale.Offset, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading male character model offset: %w", err)
			}
		case 24:
			def.CharacterModelMale.ModelSecondary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model secondary: %w", err)
			}
		case 25:
			def.CharacterModelFemale.ModelPrimary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model primary: %w", err)
			}
			def.CharacterModelFemale.Offset, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading female character model offset: %w", err)
			}
		case 26:
			def.CharacterModelFemale.ModelSecondary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model secondary: %w", err)
			}
		case 27:
			def.WearPositionTertiary, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading wear position tertiary: %w", err)
			}
		case 30, 31, 32, 33, 34:
			def.ActionsGround[opcode-30], err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading ground action 4: %w", err)
			}
		case 35, 36, 37, 38, 39:
			def.ActionsInventory[opcode-35], err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading inventory action 4: %w", err)
			}
		case 40:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			def.InventoryModel.RecolorFrom = make([]uint16, length)
			def.InventoryModel.RecolorTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				def.InventoryModel.RecolorFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				def.InventoryModel.RecolorTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 41:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			def.InventoryModel.RetextureFrom = make([]uint16, length)
			def.InventoryModel.RetextureTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				def.InventoryModel.RetextureFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				def.InventoryModel.RetextureTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 42:
			def.ShiftClickDropIndex, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading shift click drop index: %w", err)
			}
		case 65:
			def.Exchangeable = true
		case 75:
			def.Weight, err = reader.ReadInt16()
			if err != nil {
				return fmt.Errorf("reading weight: %w", err)
			}
		case 78:
			def.CharacterModelMale.ModelTertiary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model chat head model secondary: %w", err)
			}
		case 79:
			def.CharacterModelFemale.ModelTertiary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model chat head model secondary: %w", err)
			}
		case 90:
			def.CharacterModelMale.ChatHeadModelPrimary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model chat head model primary: %w", err)
			}
		case 91:
			def.CharacterModelFemale.ChatHeadModelPrimary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model chat head model primary: %w", err)
			}
		case 92:
			def.CharacterModelMale.ChatHeadModelSecondary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model chat head model secondary: %w", err)
			}
		case 93:
			def.CharacterModelFemale.ChatHeadModelSecondary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model chat head model secondary: %w", err)
			}
		case 94:
			def.Category, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 95:
			def.InventoryModel.RotationZ, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model rotation z: %w", err)
			}
		case 97:
			def.NotedItemID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading noted item id: %w", err)
			}
		case 98:
			def.NotedTemplate, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading noted template: %w", err)
			}
		case 100, 101, 102, 103, 104, 105, 106, 107, 108, 109:
			def.StackItemIDs[opcode-100], err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading stack item ids: %w", err)
			}
			def.StackQuantities[opcode-100], err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading stack quantities: %w", err)
			}
		case 110:
			def.InventoryModel.ScaleX, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model scale x: %w", err)
			}
		case 111:
			def.InventoryModel.ScaleY, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model scale y: %w", err)
			}
		case 112:
			def.InventoryModel.ScaleZ, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model scale z: %w", err)
			}
		case 113:
			def.InventoryModel.Ambient, err = reader.ReadInt8()
			if err != nil {
				return fmt.Errorf("reading inventory model ambient: %w", err)
			}
		case 114:
			def.InventoryModel.Contrast, err = reader.ReadInt8()
			if err != nil {
				return fmt.Errorf("reading inventory model contrast: %w", err)
			}
		case 115:
			def.Team, err = reader.ReadInt8()
			if err != nil {
				return fmt.Errorf("reading team: %w", err)
			}
		case 139:
			def.BoughtLinkID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading bought link id: %w", err)
			}
		case 140:
			def.BoughtTemplate, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading bought template: %w", err)
			}
		case 148:
			def.PlaceholderItemID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading placeholder item id: %w", err)
			}
		case 149:
			def.PlaceholderTemplate, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading placeholder template: %w", err)
			}
		case 249:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			for i := 0; i < int(length); i++ {
				isString, err := reader.ReadUint8()
				if err != nil {
					return err
				}
				key, err := reader.ReadUint24()
				if err != nil {
					return err
				}
				var value interface{}
				if isString == 1 {
					value, err = reader.ReadString()
				} else {
					value, err = reader.ReadInt32()
				}
				if err != nil {
					return err
				}
				def.Params[uint32(key)] = value
			}
		default:
			return fmt.Errorf("unknown opcode: %d", opcode)
		}
	}
	return nil
}
