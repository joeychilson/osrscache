package osrscache

import (
	"bytes"
	"encoding/binary"
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
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.ID); err != nil {
				return fmt.Errorf("reading inventory model id: %w", err)
			}
		case 2:
			name, err := ReadString(reader)
			if err != nil {
				return fmt.Errorf("reading name: %w", err)
			}
			def.Name = name
		case 3:
			examine, err := ReadString(reader)
			if err != nil {
				return fmt.Errorf("reading examine: %w", err)
			}
			def.Examine = examine
		case 4:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.Zoom); err != nil {
				return fmt.Errorf("reading inventory model zoom2d: %w", err)
			}
		case 5:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.RotationX); err != nil {
				return fmt.Errorf("reading inventory model xan2d: %w", err)
			}
		case 6:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.RotationY); err != nil {
				return fmt.Errorf("reading inventory model yan2d: %w", err)
			}
		case 7:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.OffsetX); err != nil {
				return fmt.Errorf("reading inventory model xoffset2d: %w", err)
			}
		case 8:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.OffsetY); err != nil {
				return fmt.Errorf("reading inventory model yoffset2d: %w", err)
			}
		case 11:
			def.Stackable = true
		case 12:
			if err := binary.Read(reader, binary.BigEndian, &def.Value); err != nil {
				return fmt.Errorf("reading value: %w", err)
			}
		case 13:
			if err := binary.Read(reader, binary.BigEndian, &def.WearPositionPrimary); err != nil {
				return fmt.Errorf("reading wear position primary: %w", err)
			}
		case 14:
			if err := binary.Read(reader, binary.BigEndian, &def.WearPositionSecondary); err != nil {
				return fmt.Errorf("reading wear position secondary: %w", err)
			}
		case 16:
			def.MembersOnly = true
		case 23:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataMale.ModelPrimary); err != nil {
				return fmt.Errorf("reading male character model primary: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataMale.Offset); err != nil {
				return fmt.Errorf("reading male character model offset: %w", err)
			}
		case 24:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataMale.ModelSecondary); err != nil {
				return fmt.Errorf("reading male character model secondary: %w", err)
			}
		case 25:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataFemale.ModelPrimary); err != nil {
				return fmt.Errorf("reading female character model primary: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataFemale.Offset); err != nil {
				return fmt.Errorf("reading female character model offset: %w", err)
			}
		case 26:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataFemale.ModelSecondary); err != nil {
				return fmt.Errorf("reading female character model secondary: %w", err)
			}
		case 27:
			if err := binary.Read(reader, binary.BigEndian, &def.WearPositionTertiary); err != nil {
				return fmt.Errorf("reading wear position tertiary: %w", err)
			}
		case 30, 31, 32, 33, 34:
			action, err := ReadString(reader)
			if err != nil {
				return fmt.Errorf("reading ground action 4: %w", err)
			}
			def.ActionsGround[opcode-30] = action
		case 35, 36, 37, 38, 39:
			action, err := ReadString(reader)
			if err != nil {
				return fmt.Errorf("reading inventory action 4: %w", err)
			}
			def.ActionsInventory[opcode-35] = action
		case 40:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			def.InventoryModelData.RecolorFrom = make([]uint16, length)
			def.InventoryModelData.RecolorTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.RecolorFrom[i]); err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.RecolorTo[i]); err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 41:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			def.InventoryModelData.RetextureFrom = make([]uint16, length)
			def.InventoryModelData.RetextureTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.RetextureFrom[i]); err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.RetextureTo[i]); err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 42:
			if err := binary.Read(reader, binary.BigEndian, &def.ShiftClickDropIndex); err != nil {
				return fmt.Errorf("reading shift click drop index: %w", err)
			}
		case 65:
			def.Exchangeable = true
		case 75:
			if err := binary.Read(reader, binary.BigEndian, &def.Weight); err != nil {
				return fmt.Errorf("reading weight: %w", err)
			}
		case 78:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataMale.ModelTertiary); err != nil {
				return fmt.Errorf("reading male character model chat head model secondary: %w", err)
			}
		case 79:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataFemale.ModelTertiary); err != nil {
				return fmt.Errorf("reading female character model chat head model secondary: %w", err)
			}
		case 90:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataMale.ChatHeadModelPrimary); err != nil {
				return fmt.Errorf("reading male character model chat head model primary: %w", err)
			}
		case 91:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataFemale.ChatHeadModelPrimary); err != nil {
				return fmt.Errorf("reading female character model chat head model primary: %w", err)
			}
		case 92:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataMale.ChatHeadModelSecondary); err != nil {
				return fmt.Errorf("reading male character model chat head model secondary: %w", err)
			}
		case 93:
			if err := binary.Read(reader, binary.BigEndian, &def.CharacterModelDataFemale.ChatHeadModelSecondary); err != nil {
				return fmt.Errorf("reading female character model chat head model secondary: %w", err)
			}
		case 94:
			if err := binary.Read(reader, binary.BigEndian, &def.Category); err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 95:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.RotationZ); err != nil {
				return fmt.Errorf("reading inventory model rotation z: %w", err)
			}
		case 97:
			if err := binary.Read(reader, binary.BigEndian, &def.NotedItemID); err != nil {
				return fmt.Errorf("reading noted item id: %w", err)
			}
		case 98:
			if err := binary.Read(reader, binary.BigEndian, &def.NotedTemplate); err != nil {
				return fmt.Errorf("reading noted template: %w", err)
			}
		case 100, 101, 102, 103, 104, 105, 106, 107, 108, 109:
			if err := binary.Read(reader, binary.BigEndian, &def.StackItemIDs[opcode-100]); err != nil {
				return fmt.Errorf("reading stack item ids: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.StackQuantities[opcode-100]); err != nil {
				return fmt.Errorf("reading stack quantities: %w", err)
			}
		case 110:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.ScaleX); err != nil {
				return fmt.Errorf("reading inventory model scale x: %w", err)
			}
		case 111:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.ScaleY); err != nil {
				return fmt.Errorf("reading inventory model scale y: %w", err)
			}
		case 112:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.ScaleZ); err != nil {
				return fmt.Errorf("reading inventory model scale z: %w", err)
			}
		case 113:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.Ambient); err != nil {
				return fmt.Errorf("reading inventory model ambient: %w", err)
			}
		case 114:
			if err := binary.Read(reader, binary.BigEndian, &def.InventoryModelData.Contrast); err != nil {
				return fmt.Errorf("reading inventory model contrast: %w", err)
			}
		case 115:
			if err := binary.Read(reader, binary.BigEndian, &def.Team); err != nil {
				return fmt.Errorf("reading team: %w", err)
			}
		case 139:
			if err := binary.Read(reader, binary.BigEndian, &def.BoughtLinkID); err != nil {
				return fmt.Errorf("reading bought link id: %w", err)
			}
		case 140:
			if err := binary.Read(reader, binary.BigEndian, &def.BoughtTemplate); err != nil {
				return fmt.Errorf("reading bought template: %w", err)
			}
		case 148:
			if err := binary.Read(reader, binary.BigEndian, &def.PlaceholderItemID); err != nil {
				return fmt.Errorf("reading placeholder item id: %w", err)
			}
		case 149:
			if err := binary.Read(reader, binary.BigEndian, &def.PlaceholderTemplate); err != nil {
				return fmt.Errorf("reading placeholder template: %w", err)
			}
		case 249:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			def.Params = make(map[uint32]any, length)
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
				def.Params[key] = value
			}
		default:
			return fmt.Errorf("unknown opcode: %d", opcode)
		}
	}
	return nil
}
