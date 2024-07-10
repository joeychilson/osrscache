package osrscache

import (
	"bytes"
	"encoding/binary"
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
				return fmt.Errorf("reading length for models: %w", err)
			}
			def.ModelData.Models = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.Models[i]); err != nil {
					return fmt.Errorf("reading model at index %d: %w", i, err)
				}
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
		case 12:
			if err := binary.Read(reader, binary.BigEndian, &def.Size); err != nil {
				return fmt.Errorf("reading size: %w", err)
			}
		case 13:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.Idle); err != nil {
				return fmt.Errorf("reading idle animation: %w", err)
			}
		case 14:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.Walking); err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
		case 15:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.IdleRotateLeft); err != nil {
				return fmt.Errorf("reading idle rotate left: %w", err)
			}
		case 16:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.IdleRotateRight); err != nil {
				return fmt.Errorf("reading idle rotate right: %w", err)
			}
		case 17:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.Walking); err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.WalkingRotate180); err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.WalkingRotateLeft); err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.WalkingRotateRight); err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
		case 18:
			if err := binary.Read(reader, binary.BigEndian, &def.Category); err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 30, 31, 32, 33, 34:
			action, err := ReadString(reader)
			if err != nil {
				return fmt.Errorf("reading action: %w", err)
			}
			def.Actions[opcode-30] = action
		case 40:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading recolor length: %w", err)
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
				return fmt.Errorf("reading retexture length: %w", err)
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
		case 60:
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading model length: %w", err)
			}
			def.ModelData.ChatHeadModels = make([]uint16, length)
			for i := range int(length) {
				if err := binary.Read(reader, binary.BigEndian, &def.ModelData.ChatHeadModels[i]); err != nil {
					return fmt.Errorf("reading chat head model: %w", err)
				}
			}
		case 74:
			if err := binary.Read(reader, binary.BigEndian, &def.Attack); err != nil {
				return fmt.Errorf("reading attack: %w", err)
			}
		case 75:
			if err := binary.Read(reader, binary.BigEndian, &def.Defense); err != nil {
				return fmt.Errorf("reading defense: %w", err)
			}
		case 76:
			if err := binary.Read(reader, binary.BigEndian, &def.Strength); err != nil {
				return fmt.Errorf("reading strength: %w", err)
			}
		case 77:
			if err := binary.Read(reader, binary.BigEndian, &def.Hitpoints); err != nil {
				return fmt.Errorf("reading range: %w", err)
			}
		case 78:
			if err := binary.Read(reader, binary.BigEndian, &def.Ranged); err != nil {
				return fmt.Errorf("reading ranged: %w", err)
			}
		case 79:
			if err := binary.Read(reader, binary.BigEndian, &def.Magic); err != nil {
				return fmt.Errorf("reading magic: %w", err)
			}
		case 93:
			def.VisibleOnMinimap = false
		case 95:
			if err := binary.Read(reader, binary.BigEndian, &def.CombatLevel); err != nil {
				return fmt.Errorf("reading combat level: %w", err)
			}
		case 97:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.ScaleWidth); err != nil {
				return fmt.Errorf("reading scale width: %w", err)
			}
		case 98:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.ScaleHeight); err != nil {
				return fmt.Errorf("reading scale height: %w", err)
			}
		case 99:
			def.Visible = true
		case 100:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.Ambient); err != nil {
				return fmt.Errorf("reading ambient: %w", err)
			}
		case 101:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.Contrast); err != nil {
				return fmt.Errorf("reading contrast: %w", err)
			}
		case 102:
			var bitfield uint8
			if err := binary.Read(reader, binary.BigEndian, &bitfield); err != nil {
				return fmt.Errorf("reading head icon bitfield: %w", err)
			}
			def.ModelData.HeadIconArchive = []int16{}
			def.ModelData.HeadIconSpriteIndex = []int16{}
			for bits := bitfield; bits != 0; bits >>= 1 {
				if bits&1 == 0 {
					def.ModelData.HeadIconArchive = append(def.ModelData.HeadIconArchive, -1)
					def.ModelData.HeadIconSpriteIndex = append(def.ModelData.HeadIconSpriteIndex, -1)
				} else {
					archive, err := ReadBigSmart2(reader)
					if err != nil {
						return fmt.Errorf("reading head icon archive: %w", err)
					}
					spriteIndex, err := ReadUint16SmartMinus1(reader)
					if err != nil {
						return fmt.Errorf("reading head icon sprite index: %w", err)
					}
					def.ModelData.HeadIconArchive = append(def.ModelData.HeadIconArchive, int16(archive))
					def.ModelData.HeadIconSpriteIndex = append(def.ModelData.HeadIconSpriteIndex, int16(spriteIndex))
				}
			}
		case 103:
			if err := binary.Read(reader, binary.BigEndian, &def.ModelData.RotateSpeed); err != nil {
				return fmt.Errorf("reading rotate speed: %w", err)
			}
		case 106:
			if err := binary.Read(reader, binary.BigEndian, &def.VarbitID); err != nil {
				return fmt.Errorf("reading varbit id (106): %w", err)
			}
			if def.VarbitID == math.MaxUint16 {
				def.VarbitID = 0
			}
			if err := binary.Read(reader, binary.BigEndian, &def.VarpIndex); err != nil {
				return fmt.Errorf("reading varp index (106): %w", err)
			}
			if def.VarpIndex == math.MaxUint16 {
				def.VarpIndex = 0
			}
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading config length (106): %w", err)
			}
			def.Configs = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.Configs[i]); err != nil {
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
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.Running); err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
		case 115:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.Running); err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.RunningRotate180); err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.RunningRotateLeft); err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.RunningRotateRight); err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
		case 116:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.Crawling); err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
		case 117:
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.Crawling); err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.CrawlingRotate180); err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.CrawlingRotateLeft); err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &def.AnimationData.CrawlingRotateRight); err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
		case 118:
			if err := binary.Read(reader, binary.BigEndian, &def.VarbitID); err != nil {
				return fmt.Errorf("reading varbit id (118): %w", err)
			}
			if def.VarbitID == math.MaxUint16 {
				def.VarbitID = 0
			}
			if err := binary.Read(reader, binary.BigEndian, &def.VarpIndex); err != nil {
				return fmt.Errorf("reading varp index (118): %w", err)
			}
			if def.VarpIndex == math.MaxUint16 {
				def.VarpIndex = 0
			}
			var varValue uint16
			if err := binary.Read(reader, binary.BigEndian, &varValue); err != nil {
				return fmt.Errorf("reading var value (118): %w", err)
			}
			if varValue == math.MaxUint16 {
				varValue = 0
			}
			var length uint8
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return fmt.Errorf("reading config length (118): %w", err)
			}
			def.Configs = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				if err := binary.Read(reader, binary.BigEndian, &def.Configs[i]); err != nil {
					return fmt.Errorf("reading config at index %d (118): %w", i, err)
				}
			}
			def.Configs[length+1] = varValue
		case 122:
			def.Follower = true
		case 123:
			def.LowPriority = true
		case 124:
			if err := binary.Read(reader, binary.BigEndian, &def.Height); err != nil {
				return fmt.Errorf("reading height: %w", err)
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
		}
	}
	return nil
}
