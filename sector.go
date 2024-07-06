package osrscache

import (
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
	if len(data) < int(headerSize) {
		return nil, fmt.Errorf("invalid sector data: got %d bytes, want at least %d", len(data), headerSize)
	}

	header := SectorHeader{}
	if headerSize == SectorExpandedHeaderSize {
		header.ArchiveID = ArchiveID(binary.BigEndian.Uint32(data))
		data = data[4:]
	} else {
		header.ArchiveID = ArchiveID(binary.BigEndian.Uint16(data))
		data = data[2:]
	}

	header.Chunk = binary.BigEndian.Uint16(data)
	header.Next = uint32(binary.BigEndian.Uint16(data[2:])) << 8
	header.Next |= uint32(data[4])
	header.IndexID = IndexID(data[5])

	return &Sector{Header: header, DataBlock: data[6:]}, nil
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
