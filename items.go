package osrscache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type ItemDefinitions map[uint16]*ItemDefinition

func (d ItemDefinitions) Get(id uint16) (*ItemDefinition, error) {
	def, ok := d[id]
	if !ok {
		return nil, fmt.Errorf("item definition not found")
	}
	return def, nil
}

type ItemDefinition struct {
	ID                       uint16             `json:"id"`
	Category                 uint16             `json:"category"`
	Name                     string             `json:"name"`
	Examine                  string             `json:"examine"`
	MembersOnly              bool               `json:"members_only"`
	Stackable                bool               `json:"stackable"`
	Tradeable                bool               `json:"tradeable"`
	Exchangeable             bool               `json:"exchangeable"`
	Value                    int32              `json:"value"`
	Weight                   int16              `json:"weight"`
	ActionsGround            [5]string          `json:"actions_ground"`
	ActionsInventory         [5]string          `json:"actions_inventory"`
	NotedItemID              uint16             `json:"noted_item_id"`
	NotedTemplate            uint16             `json:"noted_template"`
	StackItemIDs             [10]uint16         `json:"stack_item_ids"`
	StackQuantities          [10]uint16         `json:"stack_quantities"`
	Team                     int8               `json:"team"`
	BoughtLinkID             uint16             `json:"bought_link_id"`
	BoughtTemplate           uint16             `json:"bought_template"`
	PlaceholderItemID        uint16             `json:"placeholder_item_id"`
	PlaceholderTemplate      uint16             `json:"placeholder_template"`
	ShiftClickDropIndex      uint8              `json:"shift_click_drop_index"`
	Params                   map[uint32]any     `json:"params"`
	InventoryModelData       InventoryModelData `json:"inventory_model_data"`
	CharacterModelDataMale   CharacterModelData `json:"character_model_data_male"`
	CharacterModelDataFemale CharacterModelData `json:"character_model_data_female"`
	WearPositionPrimary      uint8              `json:"wear_position_primary"`
	WearPositionSecondary    uint8              `json:"wear_position_secondary"`
	WearPositionTertiary     uint8              `json:"wear_position_tertiary"`
}

type InventoryModelData struct {
	ID            uint16   `json:"id"`
	Zoom          uint16   `json:"zoom"`
	RotationX     uint16   `json:"rotation_x"`
	RotationY     uint16   `json:"rotation_y"`
	RotationZ     uint16   `json:"rotation_z"`
	OffsetX       uint16   `json:"offset_x"`
	OffsetY       uint16   `json:"offset_y"`
	ScaleX        uint16   `json:"scale_x"`
	ScaleY        uint16   `json:"scale_y"`
	ScaleZ        uint16   `json:"scale_z"`
	RecolorFrom   []uint16 `json:"recolor_from"`
	RecolorTo     []uint16 `json:"recolor_to"`
	RetextureFrom []uint16 `json:"retexture_from"`
	RetextureTo   []uint16 `json:"retexture_to"`
	Ambient       int8     `json:"ambient"`
	Contrast      int8     `json:"contrast"`
}

type CharacterModelData struct {
	ModelPrimary           uint16 `json:"model_primary"`
	ModelSecondary         uint16 `json:"model_secondary"`
	ModelTertiary          uint16 `json:"model_tertiary"`
	Offset                 uint8  `json:"offset"`
	ChatHeadModelPrimary   uint16 `json:"chat_head_model_primary"`
	ChatHeadModelSecondary uint16 `json:"chat_head_model_secondary"`
}

func NewItemDefinition(id uint16, data []byte) (*ItemDefinition, error) {
	def := &ItemDefinition{
		ID:               id,
		Name:             "null",
		ActionsGround:    [5]string{"", "", "Take", "", ""},
		ActionsInventory: [5]string{"", "", "", "", "Drop"},
		InventoryModelData: InventoryModelData{
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
			def.InventoryModelData.ID, err = reader.ReadUint16()
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
			def.InventoryModelData.Zoom, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model zoom2d: %w", err)
			}
		case 5:
			def.InventoryModelData.RotationX, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model xan2d: %w", err)
			}
		case 6:
			def.InventoryModelData.RotationY, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model yan2d: %w", err)
			}
		case 7:
			def.InventoryModelData.OffsetX, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model xoffset2d: %w", err)
			}
		case 8:
			def.InventoryModelData.OffsetY, err = reader.ReadUint16()
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
			def.CharacterModelDataMale.ModelPrimary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model primary: %w", err)
			}
			def.CharacterModelDataMale.Offset, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading male character model offset: %w", err)
			}
		case 24:
			def.CharacterModelDataMale.ModelSecondary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model secondary: %w", err)
			}
		case 25:
			def.CharacterModelDataFemale.ModelPrimary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model primary: %w", err)
			}
			def.CharacterModelDataFemale.Offset, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading female character model offset: %w", err)
			}
		case 26:
			def.CharacterModelDataFemale.ModelSecondary, err = reader.ReadUint16()
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
			def.InventoryModelData.RecolorFrom = make([]uint16, length)
			def.InventoryModelData.RecolorTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				def.InventoryModelData.RecolorFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				def.InventoryModelData.RecolorTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 41:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			def.InventoryModelData.RetextureFrom = make([]uint16, length)
			def.InventoryModelData.RetextureTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				def.InventoryModelData.RetextureFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				def.InventoryModelData.RetextureTo[i], err = reader.ReadUint16()
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
			def.CharacterModelDataMale.ModelTertiary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model chat head model secondary: %w", err)
			}
		case 79:
			def.CharacterModelDataFemale.ModelTertiary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model chat head model secondary: %w", err)
			}
		case 90:
			def.CharacterModelDataMale.ChatHeadModelPrimary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model chat head model primary: %w", err)
			}
		case 91:
			def.CharacterModelDataFemale.ChatHeadModelPrimary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model chat head model primary: %w", err)
			}
		case 92:
			def.CharacterModelDataMale.ChatHeadModelSecondary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading male character model chat head model secondary: %w", err)
			}
		case 93:
			def.CharacterModelDataFemale.ChatHeadModelSecondary, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading female character model chat head model secondary: %w", err)
			}
		case 94:
			def.Category, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 95:
			def.InventoryModelData.RotationZ, err = reader.ReadUint16()
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
			def.InventoryModelData.ScaleX, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model scale x: %w", err)
			}
		case 111:
			def.InventoryModelData.ScaleY, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model scale y: %w", err)
			}
		case 112:
			def.InventoryModelData.ScaleZ, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading inventory model scale z: %w", err)
			}
		case 113:
			def.InventoryModelData.Ambient, err = reader.ReadInt8()
			if err != nil {
				return fmt.Errorf("reading inventory model ambient: %w", err)
			}
		case 114:
			def.InventoryModelData.Contrast, err = reader.ReadInt8()
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
			def.Params = make(map[uint32]any, length)
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
