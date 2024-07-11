package osrscache

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	SectorSize               = 520
	SectorHeaderSize         = 8
	SectorExpandedHeaderSize = 10
	SectorDataSize           = 512
	SectorExpandedDataSize   = 510
)

type SectorHeader struct {
	IndexID   IndexID
	ArchiveID ArchiveID
	Chunk     uint16
	Next      uint32
}

type Sector struct {
	Header    SectorHeader
	DataBlock []byte
}

func NewSector(data []byte, headerSize uint32) (*Sector, error) {
	reader := bytes.NewReader(data)
	header := SectorHeader{}

	if headerSize == SectorExpandedHeaderSize {
		if err := binary.Read(reader, binary.BigEndian, &header.ArchiveID); err != nil {
			return nil, fmt.Errorf("failed to read ArchiveID: %w", err)
		}
	} else {
		var shortArchiveID uint16
		if err := binary.Read(reader, binary.BigEndian, &shortArchiveID); err != nil {
			return nil, fmt.Errorf("failed to read short ArchiveID: %w", err)
		}
		header.ArchiveID = ArchiveID(shortArchiveID)
	}

	if err := binary.Read(reader, binary.BigEndian, &header.Chunk); err != nil {
		return nil, fmt.Errorf("failed to read Chunk: %w", err)
	}

	var nextUpper uint16
	if err := binary.Read(reader, binary.BigEndian, &nextUpper); err != nil {
		return nil, fmt.Errorf("failed to read Next upper bits: %w", err)
	}

	var nextLower uint8
	if err := binary.Read(reader, binary.BigEndian, &nextLower); err != nil {
		return nil, fmt.Errorf("failed to read Next lower bits: %w", err)
	}

	header.Next = uint32(nextUpper)<<8 | uint32(nextLower)

	if err := binary.Read(reader, binary.BigEndian, &header.IndexID); err != nil {
		return nil, fmt.Errorf("failed to read IndexID: %w", err)
	}

	return &Sector{Header: header, DataBlock: data[headerSize:]}, nil
}

func (h *SectorHeader) Validate(indexID IndexID, archiveID ArchiveID, chunk uint32) error {
	if h.ArchiveID != archiveID {
		return fmt.Errorf("sector archive mismatch: want %d, got %d", archiveID, h.ArchiveID)
	}
	if h.Chunk != uint16(chunk) {
		return fmt.Errorf("sector chunk mismatch: want %d, got %d", chunk, h.Chunk)
	}
	if h.IndexID != indexID {
		return fmt.Errorf("sector index mismatch: want %d, got %d", indexID, h.IndexID)
	}
	return nil
}
