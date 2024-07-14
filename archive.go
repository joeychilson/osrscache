package osrscache

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const ArchiveRefLen = 6

type ArchiveID uint32

type ArchiveRef struct {
	ID      ArchiveID
	IndexID IndexID
	Length  uint32
	Sector  uint32
}

func NewArchiveRef(indexID IndexID, id ArchiveID, data []byte) (*ArchiveRef, error) {
	if len(data) != ArchiveRefLen {
		return nil, fmt.Errorf("invalid archive ref length: got %d, want %d", len(data), ArchiveRefLen)
	}

	length := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	sector := uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5])

	return &ArchiveRef{
		ID:      id,
		IndexID: indexID,
		Length:  length,
		Sector:  sector,
	}, nil
}

func (a *ArchiveRef) HeaderSize() uint32 {
	if a.IsExtended() {
		return SectorExpandedHeaderSize
	}
	return SectorHeaderSize
}

func (a *ArchiveRef) DataBlocks() []uint32 {
	headerLen, dataLen := uint32(SectorHeaderSize), uint32(SectorDataSize)
	if a.IsExtended() {
		headerLen, dataLen = uint32(SectorExpandedHeaderSize), uint32(SectorExpandedDataSize)
	}

	n := (a.Length + dataLen - 1) / dataLen
	blocks := make([]uint32, n)
	for i := uint32(0); i < n; i++ {
		if i == n-1 {
			blocks[i] = headerLen + (a.Length % dataLen)
			if blocks[i] == headerLen {
				blocks[i] = headerLen + dataLen
			}
		} else {
			blocks[i] = headerLen + dataLen
		}
	}
	return blocks
}

func (a *ArchiveRef) TotalSize() uint32 {
	return uint32(len(a.DataBlocks())) * SectorSize
}

func (a *ArchiveRef) IsExtended() bool {
	return a.ID > ArchiveID(^uint16(0))
}

type ArchiveMetadata struct {
	ID                 ArchiveID
	NameHash           int32
	CRC                uint32
	CompressedChecksum uint32
	Digests            [64]byte
	CompressedSize     uint32
	UncompressedSize   uint32
	Version            uint32
	EntryCount         int
	ValidIDs           []uint32
}

type ArchiveGroup struct {
	Files []*ArchiveFile
}

type ArchiveFile struct {
	ID   ArchiveID
	Data []byte
}

func NewArchiveGroup(data []byte, entryCount int) (*ArchiveGroup, error) {
	if entryCount == 1 {
		return &ArchiveGroup{Files: []*ArchiveFile{{ID: 0, Data: data}}}, nil
	}

	dataSize := int64(len(data))
	reader := io.NewSectionReader(bytes.NewReader(data), 0, dataSize)

	_, err := reader.Seek(-1, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to end: %w", err)
	}

	var chunkCount uint8
	err = binary.Read(reader, binary.BigEndian, &chunkCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk count: %w", err)
	}

	chunkInfoSize := int64(chunkCount) * int64(entryCount) * 4
	chunkInfoStart := dataSize - 1 - chunkInfoSize
	_, err = reader.Seek(chunkInfoStart, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to chunk size data: %w", err)
	}

	chunkInfo := make([]int32, chunkInfoSize/4)
	err = binary.Read(reader, binary.BigEndian, chunkInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk info: %w", err)
	}

	files := make([]*ArchiveFile, 0, int(chunkCount))
	var offset int64

	for i := 0; i < len(chunkInfo); i += entryCount {
		totalChunkSize := int64(0)
		for j := 0; j < entryCount; j++ {
			totalChunkSize += int64(chunkInfo[i+j])
		}

		buf := make([]byte, totalChunkSize)
		_, err := reader.ReadAt(buf, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk data: %w", err)
		}

		files = append(files, &ArchiveFile{
			ID:   ArchiveID(i / entryCount),
			Data: buf,
		})

		offset += totalChunkSize
	}

	return &ArchiveGroup{Files: files}, nil
}
