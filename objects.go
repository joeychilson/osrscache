package osrscache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

type ObjectDefinitions map[uint16]*ObjectDefinition

func (d ObjectDefinitions) Get(id uint16) (*ObjectDefinition, error) {
	def, ok := d[id]
	if !ok {
		return nil, fmt.Errorf("object definition not found")
	}
	return def, nil
}

type ObjectDefinition struct {
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

func NewObjectDefinition(id uint16, data []byte) (*ObjectDefinition, error) {
	def := &ObjectDefinition{
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
	if err := def.Read(data); err != nil {
		return nil, fmt.Errorf("reading object definition: %w", err)
	}
	return def, nil
}

func (def *ObjectDefinition) Read(data []byte) error {
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
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			def.ModelData.Models = make([]uint16, length)
			def.ModelData.Types = make([]uint8, length)
			for i := 0; i < int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.Models[i]); err != nil {
					return fmt.Errorf("reading model: %w", err)
				}
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.Types[i]); err != nil {
					return fmt.Errorf("reading model type: %w", err)
				}
			}
		case 2:
			name, err := ReadString(reader)
			if err != nil {
				return fmt.Errorf("reading name: %w", err)
			}
			def.Name = name
		case 5:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			clear(def.ModelData.Types)
			def.ModelData.Models = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.Models[i]); err != nil {
					return fmt.Errorf("reading model: %w", err)
				}
			}
		case 14:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.SizeX); err != nil {
				return fmt.Errorf("reading size x: %w", err)
			}
		case 15:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.SizeY); err != nil {
				return fmt.Errorf("reading size y: %w", err)
			}
		case 17:
			def.InteractType = 0
			def.BlocksProjectile = false
		case 18:
			def.BlocksProjectile = false
		case 19:
			if err := binary.Read(reader, binary.BigEndian, &def.WallOrDoor); err != nil {
				return fmt.Errorf("reading wall or door: %w", err)
			}
		case 21:
			def.ContouredGround = 0
		case 22:
			def.ModelData.MergeNormals = true
		case 24:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationID); err != nil {
				return fmt.Errorf("reading animation id: %w", err)
			}
		case 27:
			def.InteractType = 1
		case 28:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.DecordDisplacement); err != nil {
				return fmt.Errorf("reading decord displacement: %w", err)
			}
		case 29:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.Ambient); err != nil {
				return fmt.Errorf("reading ambient: %w", err)
			}
		case 30, 31, 32, 33, 34:
			action, err := ReadString(reader)
			if err != nil {
				return fmt.Errorf("reading action: %w", err)
			}
			def.Actions[opcode-30] = action
		case 39:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.Contrast); err != nil {
				return fmt.Errorf("reading contrast: %w", err)
			}
		case 40:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			def.ModelData.RecolorFrom = make([]uint16, length)
			def.ModelData.RecolorTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.RecolorFrom[i]); err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.RecolorTo[i]); err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 41:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			def.ModelData.RetextureFrom = make([]uint16, length)
			def.ModelData.RetextureTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.RetextureFrom[i]); err != nil {
					return fmt.Errorf("reading retexture from: %w", err)
				}
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.RetextureTo[i]); err != nil {
					return fmt.Errorf("reading retexture to: %w", err)
				}
			}
		case 61:
			if err := binary.Read(reader, binary.BigEndian, &def.Category); err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 62:
			def.Rotated = true
		case 64:
			def.Shadow = false
		case 65:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.ModelSizeX); err != nil {
				return fmt.Errorf("reading model size x: %w", err)
			}
		case 66:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.ModelSizeZ); err != nil {
				return fmt.Errorf("reading model size height: %w", err)
			}
		case 67:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.ModelSizeY); err != nil {
				return fmt.Errorf("reading model size y: %w", err)
			}
		case 68:
			if err := binary.Read(reader, binary.BigEndian, &def.MapSceneID); err != nil {
				return fmt.Errorf("reading map scene id: %w", err)
			}
		case 69:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.BlockingMask); err != nil {
				return fmt.Errorf("reading blocking mask: %w", err)
			}
		case 70:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.OffsetX); err != nil {
				return fmt.Errorf("reading offset x: %w", err)
			}
		case 71:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.OffsetZ); err != nil {
				return fmt.Errorf("reading offset z: %w", err)
			}
		case 72:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.OffsetY); err != nil {
				return fmt.Errorf("reading offset y: %w", err)
			}
		case 73:
			def.ObstructGround = true
		case 74:
			def.Solid = false
		case 75:
			if err := binary.Read(reader, binary.BigEndian, &def.SupportsItems); err != nil {
				return fmt.Errorf("reading supports items: %w", err)
			}
		case 77:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.VarpID); err != nil {
				return fmt.Errorf("reading varp id: %w", err)
			}
			if def.ModelData.VarpID == math.MaxUint16 {
				def.ModelData.VarpID = 0
			}
			if err := binary.Read(reader, binary.BigEndian, &def.ConfigID); err != nil {
				return fmt.Errorf("reading config id: %w", err)
			}
			if def.ConfigID == math.MaxUint16 {
				def.ConfigID = 0
			}
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading config length: %w", err)
			}
			def.ConfigChangeDest = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.ConfigChangeDest[i]); err != nil {
					return fmt.Errorf("reading config change dest: %w", err)
				}
				if def.ConfigChangeDest[i] == math.MaxUint16 {
					def.ConfigChangeDest[i] = 0
				}
			}
			def.ConfigChangeDest[length+1] = 0
		case 78:
			if err := binary.Read(reader, binary.BigEndian, &def.AmbientSoundID); err != nil {
				return fmt.Errorf("reading ambient sound id: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AmbientSoundDistance); err != nil {
				return fmt.Errorf("reading ambient sound distance: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AmbientSoundRetain); err != nil {
				return fmt.Errorf("reading ambient sound retain: %w", err)
			}
		case 79:
			if err := binary.Read(reader, binary.BigEndian, &def.AmbientSoundChangeTicksMin); err != nil {
				return fmt.Errorf("reading ambient sound change ticks min: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AmbientSoundChangeTicksMax); err != nil {
				return fmt.Errorf("reading ambient sound change ticks max: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AmbientSoundDistance); err != nil {
				return fmt.Errorf("reading ambient sound distance: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AmbientSoundRetain); err != nil {
				return fmt.Errorf("reading ambient sound retain: %w", err)
			}
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading ambient sound length: %w", err)
			}
			def.AmbientSoundIDs = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.AmbientSoundIDs[i]); err != nil {
					return fmt.Errorf("reading ambient sound ids: %w", err)
				}
			}
		case 81:
			if err := binary.Read(reader, binary.BigEndian, &def.ContouredGround); err != nil {
				return fmt.Errorf("reading contoured ground: %w", err)
			}
		case 82:
			if err := binary.Read(reader, binary.BigEndian, &def.MapAreaID); err != nil {
				return fmt.Errorf("reading map area id: %w", err)
			}
		case 92:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.VarpID); err != nil {
				return fmt.Errorf("reading varp id: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.ConfigID); err != nil {
				return fmt.Errorf("reading config id: %w", err)
			}
			var varValue uint16
			if err := binary.Read(reader, binary.BigEndian, &varValue); err != nil {
				return fmt.Errorf("reading var value: %w", err)
			}
			if varValue == math.MaxUint16 {
				varValue = 0
			}
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading config length: %w", err)
			}
			def.ConfigChangeDest = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.ConfigChangeDest[i]); err != nil {
					return fmt.Errorf("reading config change dest: %w", err)
				}
			}
			def.ConfigChangeDest[length+1] = varValue
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
		}
	}
	return nil
}
