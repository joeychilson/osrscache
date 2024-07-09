package osrscache

import (
	"errors"
	"fmt"
	"io"
	"math"
)

type NPCDefinitions map[uint16]*NPCDefinition

func (d NPCDefinitions) Get(id uint16) (*NPCDefinition, error) {
	def, ok := d[id]
	if !ok {
		return nil, fmt.Errorf("npc definition not found")
	}
	return def, nil
}

type NPCDefinition struct {
	ID               uint16           `json:"id"`
	Category         uint16           `json:"category"`
	Name             string           `json:"name"`
	Examine          string           `json:"examine"`
	Size             uint8            `json:"size"`
	Height           uint16           `json:"height"`
	Hitpoints        uint16           `json:"hitpoints"`
	Attack           uint16           `json:"attack"`
	Strength         uint16           `json:"strength"`
	Defense          uint16           `json:"defense"`
	Ranged           uint16           `json:"ranged"`
	Magic            uint16           `json:"magic"`
	CombatLevel      uint16           `json:"combat_level"`
	Actions          [5]string        `json:"actions"`
	Interactable     bool             `json:"interactable"`
	Follower         bool             `json:"follower"`
	LowPriority      bool             `json:"low_priority"`
	Visible          bool             `json:"visible"`
	VisibleOnMinimap bool             `json:"visible_on_minimap"`
	Configs          []uint16         `json:"configs"`
	VarbitID         uint16           `json:"varbit_id"`
	VarpIndex        uint16           `json:"varp_index"`
	OobChild         uint16           `json:"oob_child"`
	Params           map[uint32]any   `json:"params"`
	ModelData        NPCModelData     `json:"model_data"`
	AnimationData    NPCAnimationData `json:"animation_data"`
}

type NPCModelData struct {
	Models              []uint16 `json:"models"`
	ChatHeadModels      []uint16 `json:"chat_head_models"`
	RecolorFrom         []uint16 `json:"recolor_from"`
	RecolorTo           []uint16 `json:"recolor_to"`
	RetextureFrom       []uint16 `json:"retexture_from"`
	RetextureTo         []uint16 `json:"retexture_to"`
	ScaleHeight         uint16   `json:"scale_height"`
	ScaleWidth          uint16   `json:"scale_width"`
	RenderPriority      bool     `json:"render_priority"`
	Ambient             uint8    `json:"ambient"`
	Contrast            uint8    `json:"contrast"`
	HeadIcon            uint16   `json:"head_icon"`
	HeadIconArchive     []int16  `json:"head_icon_archive"`
	HeadIconSpriteIndex []int16  `json:"head_icon_sprite_index"`
	RotateSpeed         uint16   `json:"rotate_speed"`
	RotateFlag          bool     `json:"rotate_flag"`
}

type NPCAnimationData struct {
	Idle                uint16 `json:"idle"`
	IdleRotateLeft      uint16 `json:"idle_rotate_left"`
	IdleRotateRight     uint16 `json:"idle_rotate_right"`
	Walking             uint16 `json:"walking"`
	WalkingRotateLeft   uint16 `json:"walking_rotate_left"`
	WalkingRotateRight  uint16 `json:"walking_rotate_right"`
	WalkingRotate180    uint16 `json:"walking_rotate_180"`
	Running             uint16 `json:"running"`
	RunningRotateLeft   uint16 `json:"running_rotate_left"`
	RunningRotateRight  uint16 `json:"running_rotate_right"`
	RunningRotate180    uint16 `json:"running_rotate_180"`
	Crawling            uint16 `json:"crawling"`
	CrawlingRotateLeft  uint16 `json:"crawling_rotate_left"`
	CrawlingRotateRight uint16 `json:"crawling_rotate_right"`
	CrawlingRotate180   uint16 `json:"crawling_rotate_180"`
}

func NewNPCDefinition(id uint16, data []byte) (*NPCDefinition, error) {
	def := &NPCDefinition{
		ID:               id,
		Name:             "null",
		Interactable:     true,
		VisibleOnMinimap: true,
		ModelData: NPCModelData{
			ScaleHeight: 128,
			ScaleWidth:  128,
			RotateSpeed: 32,
			RotateFlag:  true,
		},
	}
	if err := def.Read(data); err != nil {
		return nil, fmt.Errorf("reading npc definition: %w", err)
	}
	return def, nil
}

func (def *NPCDefinition) Read(data []byte) error {
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
				return fmt.Errorf("reading length for models: %w", err)
			}
			def.ModelData.Models = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				model, err := reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading model at index %d: %w", i, err)
				}
				def.ModelData.Models[i] = model
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
		case 12:
			def.Size, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading size: %w", err)
			}
		case 13:
			def.AnimationData.Idle, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading idle animation: %w", err)
			}
		case 14:
			def.AnimationData.Walking, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
		case 15:
			def.AnimationData.IdleRotateLeft, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading idle rotate left: %w", err)
			}
		case 16:
			def.AnimationData.IdleRotateRight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading idle rotate right: %w", err)
			}
		case 17:
			def.AnimationData.Walking, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			def.AnimationData.WalkingRotate180, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			def.AnimationData.WalkingRotateLeft, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			def.AnimationData.WalkingRotateRight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
		case 18:
			def.Category, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 30, 31, 32, 33, 34:
			def.Actions[opcode-30], err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading action: %w", err)
			}
		case 40:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading recolor length: %w", err)
			}
			def.ModelData.RecolorFrom = make([]uint16, length)
			def.ModelData.RecolorTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				def.ModelData.RecolorFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				def.ModelData.RecolorTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 41:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading retexture length: %w", err)
			}
			def.ModelData.RetextureFrom = make([]uint16, length)
			def.ModelData.RetextureTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				def.ModelData.RetextureFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading retexture from: %w", err)
				}
				def.ModelData.RetextureTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading retexture to: %w", err)
				}
			}
		case 60:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading chat head model length: %w", err)
			}
			def.ModelData.ChatHeadModels = make([]uint16, length)
			for i := range int(length) {
				def.ModelData.ChatHeadModels[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading chat head model: %w", err)
				}
			}
		case 74:
			def.Attack, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading attack: %w", err)
			}
		case 75:
			def.Defense, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading defense: %w", err)
			}
		case 76:
			def.Strength, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading strength: %w", err)
			}
		case 77:
			def.Hitpoints, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading range: %w", err)
			}
		case 78:
			def.Ranged, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading ranged: %w", err)
			}
		case 79:
			def.Magic, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading magic: %w", err)
			}
		case 93:
			def.VisibleOnMinimap = false
		case 95:
			def.CombatLevel, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading combat level: %w", err)
			}
		case 97:
			def.ModelData.ScaleWidth, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading scale width: %w", err)
			}
		case 98:
			def.ModelData.ScaleHeight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading scale height: %w", err)
			}
		case 99:
			def.Visible = true
		case 100:
			def.ModelData.Ambient, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading ambient: %w", err)
			}
		case 101:
			def.ModelData.Contrast, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading contrast: %w", err)
			}
		case 102:
			bitfield, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading head icon bitfield: %w", err)
			}
			def.ModelData.HeadIconArchive = []int16{}
			def.ModelData.HeadIconSpriteIndex = []int16{}
			for bits := bitfield; bits != 0; bits >>= 1 {
				if bits&1 == 0 {
					def.ModelData.HeadIconArchive = append(def.ModelData.HeadIconArchive, -1)
					def.ModelData.HeadIconSpriteIndex = append(def.ModelData.HeadIconSpriteIndex, -1)
				} else {
					archive, err := reader.ReadBigSmart2()
					if err != nil {
						return fmt.Errorf("reading head icon archive: %w", err)
					}
					spriteIndex, err := reader.ReadUint16SmartMinus1()
					if err != nil {
						return fmt.Errorf("reading head icon sprite index: %w", err)
					}
					def.ModelData.HeadIconArchive = append(def.ModelData.HeadIconArchive, int16(archive))
					def.ModelData.HeadIconSpriteIndex = append(def.ModelData.HeadIconSpriteIndex, int16(spriteIndex))
				}
			}
		case 103:
			def.ModelData.RotateSpeed, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading rotate speed: %w", err)
			}
		case 106:
			def.VarbitID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varbit id (106): %w", err)
			}
			if def.VarbitID == math.MaxUint16 {
				def.VarbitID = 0
			}
			def.VarpIndex, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varp index (106): %w", err)
			}
			if def.VarpIndex == math.MaxUint16 {
				def.VarpIndex = 0
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading config length (106): %w", err)
			}
			def.Configs = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				def.Configs[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading config (106): %w", err)
				}
				if def.Configs[i] == math.MaxUint16 {
					def.Configs[i] = 0
				}
			}
			def.Configs[length+1] = 0
		case 107:
			def.Interactable = false
		case 109:
			def.ModelData.RotateFlag = true
		case 111:
			def.Follower = true
			def.LowPriority = true
		case 114:
			def.AnimationData.Running, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
		case 115:
			def.AnimationData.Running, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			def.AnimationData.RunningRotate180, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			def.AnimationData.RunningRotateLeft, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			def.AnimationData.RunningRotateRight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
		case 116:
			def.AnimationData.Crawling, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
		case 117:
			def.AnimationData.Crawling, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			def.AnimationData.CrawlingRotate180, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			def.AnimationData.CrawlingRotateLeft, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			def.AnimationData.CrawlingRotateRight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
		case 118:
			def.VarbitID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varbit id (118): %w", err)
			}
			if def.VarbitID == math.MaxUint16 {
				def.VarbitID = 0
			}
			def.VarpIndex, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varp index (118): %w", err)
			}
			if def.VarpIndex == math.MaxUint16 {
				def.VarpIndex = 0
			}
			var varValue uint16
			varValue, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading var value (118): %w", err)
			}
			if varValue == math.MaxUint16 {
				varValue = 0
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading config length (118): %w", err)
			}
			def.Configs = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				def.Configs[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading config at index %d (118): %w", i, err)
				}
			}
			def.Configs[length+1] = varValue
		case 122:
			def.Follower = true
		case 123:
			def.LowPriority = true
		case 124:
			def.Height, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading height: %w", err)
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
		}
	}
	return nil
}
