package osrscache

import (
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
	reader := NewBinaryReader(data)
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
			def.ModelData.Models = make([]uint16, length)
			def.ModelData.Types = make([]uint8, length)
			for i := 0; i < int(length); i++ {
				if def.ModelData.Models[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading model: %w", err)
				}
				if def.ModelData.Types[i], err = reader.ReadUint8(); err != nil {
					return fmt.Errorf("reading model type: %w", err)
				}
			}
		case 2:
			if def.Name, err = reader.ReadString(); err != nil {
				return fmt.Errorf("reading name: %w", err)
			}
		case 5:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			clear(def.ModelData.Types)
			def.ModelData.Models = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if def.ModelData.Models[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading model: %w", err)
				}
			}
		case 14:
			if def.ModelData.SizeX, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading size x: %w", err)
			}
		case 15:
			if def.ModelData.SizeY, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading size y: %w", err)
			}
		case 17:
			def.InteractType = 0
			def.BlocksProjectile = false
		case 18:
			def.BlocksProjectile = false
		case 19:
			if def.WallOrDoor, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading wall or door: %w", err)
			}
		case 21:
			def.ContouredGround = 0
		case 22:
			def.ModelData.MergeNormals = true
		case 24:
			if def.AnimationID, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading animation id: %w", err)
			}
		case 27:
			def.InteractType = 1
		case 28:
			if def.ModelData.DecordDisplacement, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading decord displacement: %w", err)
			}
		case 29:
			if def.ModelData.Ambient, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading ambient (29): %w", err)
			}
		case 30, 31, 32, 33, 34:
			if def.Actions[opcode-30], err = reader.ReadString(); err != nil {
				return fmt.Errorf("reading action: %w", err)
			}
		case 39:
			if def.ModelData.Contrast, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading contrast: %w", err)
			}
		case 40:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			def.ModelData.RecolorFrom = make([]uint16, length)
			def.ModelData.RecolorTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if def.ModelData.RecolorFrom[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				if def.ModelData.RecolorTo[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 41:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			def.ModelData.RetextureFrom = make([]uint16, length)
			def.ModelData.RetextureTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if def.ModelData.RetextureFrom[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading retexture from: %w", err)
				}
				if def.ModelData.RetextureTo[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading retexture to: %w", err)
				}
			}
		case 61:
			if def.Category, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 62:
			def.Rotated = true
		case 64:
			def.Shadow = false
		case 65:
			if def.ModelData.ModelSizeX, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading model size x: %w", err)
			}
		case 66:
			if def.ModelData.ModelSizeZ, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading model size height: %w", err)
			}
		case 67:
			if def.ModelData.ModelSizeY, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading model size y: %w", err)
			}
		case 68:
			if def.MapSceneID, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading map scene id: %w", err)
			}
		case 69:
			if def.ModelData.BlockingMask, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading blocking mask: %w", err)
			}
		case 70:
			if def.ModelData.OffsetX, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading offset x: %w", err)
			}
		case 71:
			if def.ModelData.OffsetZ, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading offset z: %w", err)
			}
		case 72:
			if def.ModelData.OffsetY, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading offset y: %w", err)
			}
		case 73:
			def.ObstructGround = true
		case 74:
			def.Solid = false
		case 75:
			if def.SupportsItems, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading supports items: %w", err)
			}
		case 77:
			if def.ModelData.VarpID, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading varp id (77): %w", err)
			}
			if def.ModelData.VarpID == math.MaxUint16 {
				def.ModelData.VarpID = 0
			}
			if def.ConfigID, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading config id (77): %w", err)
			}
			if def.ConfigID == math.MaxUint16 {
				def.ConfigID = 0
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading config length (77): %w", err)
			}
			def.ConfigChangeDest = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				if def.ConfigChangeDest[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading config change dest (77): %w", err)
				}
				if def.ConfigChangeDest[i] == math.MaxUint16 {
					def.ConfigChangeDest[i] = 0
				}
			}
			def.ConfigChangeDest[length+1] = 0
		case 78:
			if def.AmbientSoundID, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading ambient sound id (78): %w", err)
			}
			if def.AmbientSoundDistance, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading ambient sound distance (78): %w", err)
			}
			if def.AmbientSoundRetain, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading ambient sound retain: %w", err)
			}
		case 79:
			if def.AmbientSoundChangeTicksMin, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading ambient sound change ticks min: %w", err)
			}
			if def.AmbientSoundChangeTicksMax, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading ambient sound change ticks max: %w", err)
			}
			if def.AmbientSoundDistance, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading ambient sound distance (79): %w", err)
			}
			if def.AmbientSoundRetain, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading ambient sound retain: %w", err)
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length: %w", err)
			}
			def.AmbientSoundIDs = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if def.AmbientSoundIDs[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading ambient sound ids: %w", err)
				}
			}
		case 81:
			if def.ContouredGround, err = reader.ReadUint8(); err != nil {
				return fmt.Errorf("reading contoured ground: %w", err)
			}
		case 82:
			if def.MapAreaID, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading map area id: %w", err)
			}
		case 92:
			if def.ModelData.VarpID, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading varp id (92): %w", err)
			}
			if def.ConfigID, err = reader.ReadUint16(); err != nil {
				return fmt.Errorf("reading config id (92): %w", err)
			}
			var varValue uint16
			varValue, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading var value (92): %w", err)
			}
			if varValue == math.MaxUint16 {
				varValue = 0
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading config length (92): %w", err)
			}
			def.ConfigChangeDest = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				if def.ConfigChangeDest[i], err = reader.ReadUint16(); err != nil {
					return fmt.Errorf("reading config change dest at index %d (92): %w", i, err)
				}
			}
			def.ConfigChangeDest[length+1] = varValue
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
		}
	}
	return nil
}
