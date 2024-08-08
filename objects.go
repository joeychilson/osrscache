package osrscache

import (
	"errors"
	"fmt"
	"io"
	"math"
)

type Object struct {
	ID                         uint16          `json:"id"`
	Category                   uint16          `json:"category"`
	Name                       string          `json:"name"`
	ConfigID                   uint16          `json:"config_id"`
	MapAreaID                  uint16          `json:"map_area_id"`
	MapSceneID                 uint16          `json:"map_scene_id"`
	AnimationID                uint16          `json:"animation_id"`
	Solid                      bool            `json:"solid"`
	Shadow                     bool            `json:"shadow"`
	ObstructGround             bool            `json:"obstruct_ground"`
	SupportsItems              uint8           `json:"supports_items"`
	Actions                    [5]string       `json:"actions"`
	InteractType               uint8           `json:"interact_type"`
	Rotated                    bool            `json:"rotated"`
	AmbientSoundID             uint16          `json:"ambient_sound_id"`
	AmbientSoundIDs            []uint16        `json:"ambient_sound_ids"`
	AmbientSoundDistance       uint8           `json:"ambient_sound_distance"`
	AmbientSoundRetain         uint8           `json:"ambient_sound_retain"`
	AmbientSoundChangeTicksMin uint16          `json:"ambient_sound_change_ticks_min"`
	AmbientSoundChangeTicksMax uint16          `json:"ambient_sound_change_ticks_max"`
	BlocksProjectile           bool            `json:"blocks_projectile"`
	WallOrDoor                 uint8           `json:"wall_or_door"`
	ContouredGround            uint8           `json:"contoured_ground"`
	ConfigChangeDest           []uint16        `json:"config_change_dest"`
	Params                     map[uint32]any  `json:"params"`
	ModelData                  ObjectModelData `json:"model_data"`
}

type ObjectModelData struct {
	Models             []uint16 `json:"models"`
	Types              []uint8  `json:"types"`
	RecolorFrom        []uint16 `json:"recolor_from"`
	RecolorTo          []uint16 `json:"recolor_to"`
	RetextureFrom      []uint16 `json:"retexture_from"`
	RetextureTo        []uint16 `json:"retexture_to"`
	SizeX              uint8    `json:"size_x"`
	SizeY              uint8    `json:"size_y"`
	OffsetX            uint16   `json:"offset_x"`
	OffsetY            uint16   `json:"offset_y"`
	OffsetZ            uint16   `json:"offset_z"`
	ModelSizeX         uint16   `json:"model_size_x"`
	ModelSizeY         uint16   `json:"model_size_y"`
	ModelSizeZ         uint16   `json:"model_size_z"`
	VarpID             uint16   `json:"varp_id"`
	Ambient            uint8    `json:"ambient"`
	Contrast           uint8    `json:"contrast"`
	DecordDisplacement uint8    `json:"decord_displacement"`
	MergeNormals       bool     `json:"merge_normals"`
	BlockingMask       uint8    `json:"blocking_mask"`
}

func NewObject(id uint16) *Object {
	return &Object{
		ID:               id,
		InteractType:     3,
		BlocksProjectile: true,
		Solid:            true,
		ModelData: ObjectModelData{
			DecordDisplacement: 16,
			SizeX:              1,
			SizeY:              1,
			ModelSizeX:         128,
			ModelSizeY:         128,
			ModelSizeZ:         128,
		},
	}
}

func (obj *Object) Read(data []byte) error {
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
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			obj.ModelData.Models = make([]uint16, length)
			obj.ModelData.Types = make([]uint8, length)
			for i := 0; i < int(length); i++ {
				obj.ModelData.Models[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading model: %w", err)
				}
				obj.ModelData.Types[i], err = reader.ReadUint8()
				if err != nil {
					return fmt.Errorf("reading model type: %w", err)
				}
			}
		case 2:
			obj.Name, err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading name: %w", err)
			}
		case 5:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			clear(obj.ModelData.Types)
			obj.ModelData.Models = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				obj.ModelData.Models[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading model: %w", err)
				}
			}
		case 14:
			obj.ModelData.SizeX, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading size x: %w", err)
			}
		case 15:
			obj.ModelData.SizeY, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading size y: %w", err)
			}
		case 17:
			obj.InteractType = 0
			obj.BlocksProjectile = false
		case 18:
			obj.BlocksProjectile = false
		case 19:
			obj.WallOrDoor, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading wall or door: %w", err)
			}
		case 21:
			obj.ContouredGround = 0
		case 22:
			obj.ModelData.MergeNormals = true
		case 24:
			obj.AnimationID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading animation id: %w", err)
			}
		case 27:
			obj.InteractType = 1
		case 28:
			obj.ModelData.DecordDisplacement, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading decord displacement: %w", err)
			}
		case 29:
			obj.ModelData.Ambient, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading ambient: %w", err)
			}
		case 30, 31, 32, 33, 34:
			obj.Actions[opcode-30], err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading action: %w", err)
			}
		case 39:
			obj.ModelData.Contrast, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading contrast: %w", err)
			}
		case 40:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			obj.ModelData.RecolorFrom = make([]uint16, length)
			obj.ModelData.RecolorTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				obj.ModelData.RecolorFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				obj.ModelData.RecolorTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 41:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			obj.ModelData.RetextureFrom = make([]uint16, length)
			obj.ModelData.RetextureTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				obj.ModelData.RetextureFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading retexture from: %w", err)
				}
				obj.ModelData.RetextureTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading retexture to: %w", err)
				}
			}
		case 61:
			obj.Category, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 62:
			obj.Rotated = true
		case 64:
			obj.Shadow = false
		case 65:
			obj.ModelData.ModelSizeX, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading model size x: %w", err)
			}
		case 66:
			obj.ModelData.ModelSizeZ, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading model size height: %w", err)
			}
		case 67:
			obj.ModelData.ModelSizeY, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading model size y: %w", err)
			}
		case 68:
			obj.MapSceneID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading map scene id: %w", err)
			}
		case 69:
			obj.ModelData.BlockingMask, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading blocking mask: %w", err)
			}
		case 70:
			obj.ModelData.OffsetX, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading offset x: %w", err)
			}
		case 71:
			obj.ModelData.OffsetZ, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading offset z: %w", err)
			}
		case 72:
			obj.ModelData.OffsetY, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading offset y: %w", err)
			}
		case 73:
			obj.ObstructGround = true
		case 74:
			obj.Solid = false
		case 75:
			obj.SupportsItems, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading supports items: %w", err)
			}
		case 77:
			obj.ModelData.VarpID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varp id: %w", err)
			}
			if obj.ModelData.VarpID == math.MaxUint16 {
				obj.ModelData.VarpID = 0
			}
			obj.ConfigID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading config id: %w", err)
			}
			if obj.ConfigID == math.MaxUint16 {
				obj.ConfigID = 0
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading config length: %w", err)
			}
			obj.ConfigChangeDest = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				obj.ConfigChangeDest[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading config change dest: %w", err)
				}
				if obj.ConfigChangeDest[i] == math.MaxUint16 {
					obj.ConfigChangeDest[i] = 0
				}
			}
			obj.ConfigChangeDest[length+1] = 0
		case 78:
			obj.AmbientSoundID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading ambient sound id: %w", err)
			}
			obj.AmbientSoundDistance, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading ambient sound distance: %w", err)
			}
			obj.AmbientSoundRetain, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading ambient sound retain: %w", err)
			}
		case 79:
			obj.AmbientSoundChangeTicksMin, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading ambient sound change ticks min: %w", err)
			}
			obj.AmbientSoundChangeTicksMax, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading ambient sound change ticks max: %w", err)
			}
			obj.AmbientSoundDistance, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading ambient sound distance: %w", err)
			}
			obj.AmbientSoundRetain, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading ambient sound retain: %w", err)
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading ambient sound length: %w", err)
			}
			obj.AmbientSoundIDs = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				obj.AmbientSoundIDs[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading ambient sound ids: %w", err)
				}
			}
		case 81:
			obj.ContouredGround, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading contoured ground: %w", err)
			}
		case 82:
			obj.MapAreaID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading map area id: %w", err)
			}
		case 92:
			obj.ModelData.VarpID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varp id: %w", err)
			}
			obj.ConfigID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading config id: %w", err)
			}
			varValue, err := reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading var value: %w", err)
			}
			if varValue == math.MaxUint16 {
				varValue = 0
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading config length: %w", err)
			}
			obj.ConfigChangeDest = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				obj.ConfigChangeDest[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading config change dest: %w", err)
				}
			}
			obj.ConfigChangeDest[length+1] = varValue
		case 249:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			obj.Params = make(map[uint32]any, length)
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
				obj.Params[key] = value
			}
		}
	}
	return nil
}
