package osrscache

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
)

const (
	FlagColumnMajor = 0x1
	FlagAlpha       = 0x2
)

type Sprites map[uint32]*Sprite

func (s Sprites) Get(id uint32) (*Sprite, error) {
	sprite, ok := s[id]
	if !ok {
		return nil, fmt.Errorf("sprite not found")
	}
	return sprite, nil
}

type Sprite struct {
	ID      uint32   `json:"id"`
	Width   uint16   `json:"width"`
	Height  uint16   `json:"height"`
	Palette []uint32 `json:"palette"`
	Frames  []*Frame `json:"frames"`
}

func NewSprite(id uint32, data []byte) (*Sprite, error) {
	sprite := &Sprite{
		ID: id,
	}
	err := sprite.Read(data)
	if err != nil {
		return nil, fmt.Errorf("reading sprite: %w", err)
	}
	return sprite, nil
}

func (s *Sprite) Read(data []byte) error {
	reader := bytes.NewReader(data)

	if _, err := reader.Seek(reader.Size()-2, io.SeekStart); err != nil {
		return fmt.Errorf("seeking to frame length: %w", err)
	}

	var frameLength uint16
	if err := binary.Read(reader, binary.BigEndian, &frameLength); err != nil {
		return fmt.Errorf("reading frame length: %w", err)
	}

	trailerLen := int64(frameLength*8 + 7)

	if _, err := reader.Seek(reader.Size()-trailerLen, io.SeekStart); err != nil {
		return fmt.Errorf("seeking to trailer start: %w", err)
	}

	if err := binary.Read(reader, binary.BigEndian, &s.Width); err != nil {
		return fmt.Errorf("reading width: %w", err)
	}

	if err := binary.Read(reader, binary.BigEndian, &s.Height); err != nil {
		return fmt.Errorf("reading height: %w", err)
	}

	var paletteLength uint8
	if err := binary.Read(reader, binary.BigEndian, &paletteLength); err != nil {
		return fmt.Errorf("reading palette length: %w", err)
	}

	trailerLen += int64(paletteLength) * 3

	xOffsets := make([]uint16, frameLength)
	for i := range xOffsets {
		if err := binary.Read(reader, binary.BigEndian, &xOffsets[i]); err != nil {
			return fmt.Errorf("reading x offset: %w", err)
		}
	}

	yOffsets := make([]uint16, frameLength)
	for i := range yOffsets {
		if err := binary.Read(reader, binary.BigEndian, &yOffsets[i]); err != nil {
			return fmt.Errorf("reading y offset: %w", err)
		}
	}

	maxWidths := make([]uint16, frameLength)
	for i := range maxWidths {
		if err := binary.Read(reader, binary.BigEndian, &maxWidths[i]); err != nil {
			return fmt.Errorf("reading max width: %w", err)
		}
	}

	maxHeights := make([]uint16, frameLength)
	for i := range maxHeights {
		if err := binary.Read(reader, binary.BigEndian, &maxHeights[i]); err != nil {
			return fmt.Errorf("reading max height: %w", err)
		}
	}

	if _, err := reader.Seek(reader.Size()-trailerLen, io.SeekStart); err != nil {
		return fmt.Errorf("seeking to palette start: %w", err)
	}

	s.Palette = make([]uint32, paletteLength)
	for i := range s.Palette {
		var color [3]byte
		if _, err := reader.Read(color[:]); err != nil {
			return fmt.Errorf("reading palette: %w", err)
		}
		s.Palette[i] = uint32(color[0])<<16 | uint32(color[1])<<8 | uint32(color[2])
	}

	s.Frames = make([]*Frame, frameLength)
	for i := range s.Frames {
		frame, err := NewFrame(i, xOffsets[i], yOffsets[i], maxWidths[i], maxHeights[i], data)
		if err != nil {
			return fmt.Errorf("creating frame: %w", err)
		}
		s.Frames[i] = frame
	}
	return nil
}

func (s *Sprite) Image() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, int(s.Width), int(s.Height)))

	for _, frame := range s.Frames {
		index := 0
		for y := 0; y < int(frame.MaxHeight); y++ {
			for x := 0; x < int(frame.MaxWidth); x++ {
				paletteIndex := int(frame.Pixels[index]) & 0xFF

				var c color.RGBA
				if paletteIndex == 0 {
					c = color.RGBA{0, 0, 0, 0}
				} else {
					paletteColor := s.Palette[paletteIndex-1]
					c = color.RGBA{
						R: uint8((paletteColor >> 16) & 0xFF),
						G: uint8((paletteColor >> 8) & 0xFF),
						B: uint8(paletteColor & 0xFF),
						A: 255,
					}
				}
				if frame.Alpha != nil {
					c.A = frame.Alpha[index]
				}
				imgX, imgY := int(frame.OffsetX)+x, int(frame.OffsetY)+y
				if imgX >= 0 && imgX < int(s.Width) && imgY >= 0 && imgY < int(s.Height) {
					img.Set(imgX, imgY, c)
				}
				index++
			}
		}
	}
	return img
}

type Frame struct {
	ID        int    `json:"id"`
	OffsetX   uint16 `json:"offset_x"`
	OffsetY   uint16 `json:"offset_y"`
	MaxWidth  uint16 `json:"max_width"`
	MaxHeight uint16 `json:"max_height"`
	Pixels    []byte `json:"pixels"`
	Alpha     []byte `json:"alpha"`
}

func NewFrame(id int, offsetX uint16, offsetY uint16, maxWidth uint16, maxHeight uint16, data []byte) (*Frame, error) {
	frame := &Frame{
		ID:        id,
		OffsetX:   offsetX,
		OffsetY:   offsetY,
		MaxWidth:  maxWidth,
		MaxHeight: maxHeight,
		Pixels:    make([]byte, int(maxWidth)*int(maxHeight)),
	}
	err := frame.Read(data)
	if err != nil {
		return nil, fmt.Errorf("reading frame: %w", err)
	}
	return frame, nil
}

func (f *Frame) Read(data []byte) error {
	reader := bytes.NewReader(data)

	var flags uint8
	if err := binary.Read(reader, binary.BigEndian, &flags); err != nil {
		return fmt.Errorf("reading flags: %w", err)
	}

	if flags&FlagAlpha != 0 {
		f.Alpha = make([]byte, f.MaxWidth*f.MaxHeight)
	}

	if flags&FlagColumnMajor != 0 {
		for x := 0; x < int(f.MaxWidth); x++ {
			for y := 0; y < int(f.MaxHeight); y++ {
				pixel, err := reader.ReadByte()
				if err != nil {
					return fmt.Errorf("reading pixel: %w", err)
				}
				f.Pixels[y*int(f.MaxWidth)+x] = pixel
			}
		}
		if f.Alpha != nil {
			for x := 0; x < int(f.MaxWidth); x++ {
				for y := 0; y < int(f.MaxHeight); y++ {
					alpha, err := reader.ReadByte()
					if err != nil {
						return fmt.Errorf("reading alpha: %w", err)
					}
					f.Alpha[y*int(f.MaxWidth)+x] = alpha
				}
			}
		}
	} else {
		for i := 0; i < len(f.Pixels); i++ {
			pixel, err := reader.ReadByte()
			if err != nil {
				return fmt.Errorf("reading pixel: %w", err)
			}
			f.Pixels[i] = pixel
		}
		if f.Alpha != nil {
			for i := 0; i < len(f.Alpha); i++ {
				alpha, err := reader.ReadByte()
				if err != nil {
					return fmt.Errorf("reading alpha: %w", err)
				}
				f.Alpha[i] = alpha
			}
		}
	}
	return nil
}
