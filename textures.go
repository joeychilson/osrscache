package osrscache

import "fmt"

type Texture struct {
	ID                 int
	SpriteIDs          []uint16
	Opaque             bool    // Not 100% sure this is Opaque
	Colors             []int32 // Not 100% sure this is Colors
	AverageRGB         uint16  // Not 100% sure this is AverageRGB
	BlendModes         []uint8 // Not 100% sure this is BlendModes
	AnimationSpeed     uint8
	AnimationDirection uint8
	AnimationFrames    []uint8 // Not 100% sure this is AnimationFrames
}

func NewTexture(id int) *Texture {
	return &Texture{ID: id}
}

func (t *Texture) Read(data []byte) error {
	var err error

	reader := NewReader(data)

	t.AverageRGB, err = reader.ReadUint16()
	if err != nil {
		return fmt.Errorf("reading average rgb: %w", err)
	}

	opaque, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("reading opaque: %w", err)
	}
	t.Opaque = opaque != 0

	spriteCount, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("reading sprite count: %w", err)
	}

	t.SpriteIDs = make([]uint16, 0, spriteCount)
	for i := 0; i < int(spriteCount); i++ {
		spriteID, err := reader.ReadUint16()
		if err != nil {
			return fmt.Errorf("reading sprite id: %w", err)
		}
		t.SpriteIDs = append(t.SpriteIDs, spriteID)
	}

	if spriteCount > 1 {
		t.BlendModes = make([]uint8, spriteCount-1)
		for i := 0; i < int(spriteCount-1); i++ {
			blendMode, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading blend mode: %w", err)
			}
			t.BlendModes[i] = blendMode
		}

		t.AnimationFrames = make([]uint8, spriteCount-1)
		for i := 0; i < int(spriteCount-1); i++ {
			animationFrame, err := reader.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading animation frame: %w", err)
			}
			t.AnimationFrames[i] = animationFrame
		}
	}

	t.Colors = make([]int32, spriteCount)
	for i := 0; i < int(spriteCount); i++ {
		color, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading color: %w", err)
		}
		t.Colors[i] = color
	}

	t.AnimationDirection, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("reading animation direction: %w", err)
	}

	t.AnimationSpeed, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("reading animation speed: %w", err)
	}
	return nil
}
