package osrscache

import (
	"errors"
	"fmt"
	"io"
	"math"
)

type NPC struct {
	ID               int              `json:"id"`
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

func NewNPC(id int) *NPC {
	return &NPC{
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
}

func (npc *NPC) Read(data []byte) error {
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
				return fmt.Errorf("reading length for models: %w", err)
			}
			npc.ModelData.Models = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				npc.ModelData.Models[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading model at index %d: %w", i, err)
				}
			}
		case 2:
			npc.Name, err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading name: %w", err)
			}
		case 3:
			npc.Examine, err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading examine: %w", err)
			}
		case 12:
			npc.Size, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading size: %w", err)
			}
		case 13:
			npc.AnimationData.Idle, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading idle animation: %w", err)
			}
		case 14:
			npc.AnimationData.Walking, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
		case 15:
			npc.AnimationData.IdleRotateLeft, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading idle rotate left: %w", err)
			}
		case 16:
			npc.AnimationData.IdleRotateRight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading idle rotate right: %w", err)
			}
		case 17:
			npc.AnimationData.Walking, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			npc.AnimationData.WalkingRotate180, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			npc.AnimationData.WalkingRotateLeft, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
			npc.AnimationData.WalkingRotateRight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading walking animation: %w", err)
			}
		case 18:
			npc.Category, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading category: %w", err)
			}
		case 30, 31, 32, 33, 34:
			npc.Actions[opcode-30], err = reader.ReadString()
			if err != nil {
				return fmt.Errorf("reading action: %w", err)
			}
		case 40:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading recolor length: %w", err)
			}
			npc.ModelData.RecolorFrom = make([]uint16, length)
			npc.ModelData.RecolorTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				npc.ModelData.RecolorFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor from: %w", err)
				}
				npc.ModelData.RecolorTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading recolor to: %w", err)
				}
			}
		case 41:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading retexture length: %w", err)
			}
			npc.ModelData.RetextureFrom = make([]uint16, length)
			npc.ModelData.RetextureTo = make([]uint16, length)
			for i := 0; i < int(length); i++ {
				npc.ModelData.RetextureFrom[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading retexture from: %w", err)
				}
				npc.ModelData.RetextureTo[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading retexture to: %w", err)
				}
			}
		case 60:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading model length: %w", err)
			}
			npc.ModelData.ChatHeadModels = make([]uint16, length)
			for i := range npc.ModelData.ChatHeadModels {
				npc.ModelData.ChatHeadModels[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading chat head model: %w", err)
				}
			}
		case 74:
			npc.Attack, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading attack: %w", err)
			}
		case 75:
			npc.Defense, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading defense: %w", err)
			}
		case 76:
			npc.Strength, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading strength: %w", err)
			}
		case 77:
			npc.Hitpoints, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading range: %w", err)
			}
		case 78:
			npc.Ranged, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading ranged: %w", err)
			}
		case 79:
			npc.Magic, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading magic: %w", err)
			}
		case 93:
			npc.VisibleOnMinimap = false
		case 95:
			npc.CombatLevel, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading combat level: %w", err)
			}
		case 97:
			npc.ModelData.ScaleWidth, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading scale width: %w", err)
			}
		case 98:
			npc.ModelData.ScaleHeight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading scale height: %w", err)
			}
		case 99:
			npc.Visible = true
		case 100:
			npc.ModelData.Ambient, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading ambient: %w", err)
			}
		case 101:
			npc.ModelData.Contrast, err = reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading contrast: %w", err)
			}
		case 102:
			bitfield, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading head icon bitfield: %w", err)
			}
			npc.ModelData.HeadIconArchive = []int16{}
			npc.ModelData.HeadIconSpriteIndex = []int16{}
			for bits := bitfield; bits != 0; bits >>= 1 {
				if bits&1 == 0 {
					npc.ModelData.HeadIconArchive = append(npc.ModelData.HeadIconArchive, -1)
					npc.ModelData.HeadIconSpriteIndex = append(npc.ModelData.HeadIconSpriteIndex, -1)
				} else {
					archive, err := reader.ReadBigSmart2()
					if err != nil {
						return fmt.Errorf("reading head icon archive: %w", err)
					}
					spriteIndex, err := reader.ReadUint16SmartMinus1()
					if err != nil {
						return fmt.Errorf("reading head icon sprite index: %w", err)
					}
					npc.ModelData.HeadIconArchive = append(npc.ModelData.HeadIconArchive, int16(archive))
					npc.ModelData.HeadIconSpriteIndex = append(npc.ModelData.HeadIconSpriteIndex, int16(spriteIndex))
				}
			}
		case 103:
			npc.ModelData.RotateSpeed, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading rotate speed: %w", err)
			}
		case 106:
			npc.VarbitID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varbit id (106): %w", err)
			}
			if npc.VarbitID == math.MaxUint16 {
				npc.VarbitID = 0
			}
			npc.VarpIndex, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varp index (106): %w", err)
			}
			if npc.VarpIndex == math.MaxUint16 {
				npc.VarpIndex = 0
			}
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading config length (106): %w", err)
			}
			npc.Configs = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				npc.Configs[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading config (106): %w", err)
				}
				if npc.Configs[i] == math.MaxUint16 {
					npc.Configs[i] = 0
				}
			}
			npc.Configs[length+1] = 0
		case 107:
			npc.Interactable = false
		case 109:
			npc.ModelData.RotateFlag = true
		case 111:
			npc.Follower = true
			npc.LowPriority = true
		case 114:
			npc.AnimationData.Running, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
		case 115:
			npc.AnimationData.Running, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			npc.AnimationData.RunningRotate180, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			npc.AnimationData.RunningRotateLeft, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
			npc.AnimationData.RunningRotateRight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading running animation: %w", err)
			}
		case 116:
			npc.AnimationData.Crawling, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
		case 117:
			npc.AnimationData.Crawling, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			npc.AnimationData.CrawlingRotate180, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			npc.AnimationData.CrawlingRotateLeft, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
			npc.AnimationData.CrawlingRotateRight, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading crawling animation: %w", err)
			}
		case 118:
			npc.VarbitID, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varbit id (118): %w", err)
			}
			if npc.VarbitID == math.MaxUint16 {
				npc.VarbitID = 0
			}
			npc.VarpIndex, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading varp index (118): %w", err)
			}
			if npc.VarpIndex == math.MaxUint16 {
				npc.VarpIndex = 0
			}
			varValue, err := reader.ReadUint16()
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
			npc.Configs = make([]uint16, int(length)+2)
			for i := 0; i <= int(length); i++ {
				npc.Configs[i], err = reader.ReadUint16()
				if err != nil {
					return fmt.Errorf("reading config at index %d (118): %w", i, err)
				}
			}
			npc.Configs[length+1] = varValue
		case 122:
			npc.Follower = true
		case 123:
			npc.LowPriority = true
		case 124:
			npc.Height, err = reader.ReadUint16()
			if err != nil {
				return fmt.Errorf("reading height: %w", err)
			}
		case 249:
			length, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading param length: %w", err)
			}
			npc.Params = make(map[uint32]any, length)
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
				npc.Params[key] = value
			}
		}
	}
	return nil
}
