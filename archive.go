package osrscache

import (
	"encoding/binary"
	"fmt"
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

func NewArchiveGroup(buffer []byte, entryCount int) (*ArchiveGroup, error) {
	if len(buffer) < 1 {
		return nil, fmt.Errorf("buffer is too short")
	}

	numChunks := int(buffer[len(buffer)-1])
	minBufferSize := 1 + numChunks*entryCount*4
	if len(buffer) < minBufferSize {
		return nil, fmt.Errorf("buffer is too short: expected at least %d bytes, got %d", minBufferSize, len(buffer))
	}

	type CachedChunk struct {
		entryID   uint32
		chunkSize int
	}

	chunkInfo := make([]CachedChunk, 0, numChunks*entryCount)
	data := make([]*ArchiveFile, 0, numChunks*entryCount)

	readPtr := len(buffer) - 1 - numChunks*entryCount*4

	for i := 0; i < numChunks; i++ {
		totalChunkSize := 0
		for entryID := 0; entryID < entryCount; entryID++ {
			if readPtr+4 > len(buffer) {
				return nil, fmt.Errorf("unexpected end of buffer while reading chunk sizes")
			}
			delta := int32(binary.BigEndian.Uint32(buffer[readPtr : readPtr+4]))
			readPtr += 4
			totalChunkSize += int(delta)

			chunkInfo = append(chunkInfo, CachedChunk{
				entryID:   uint32(entryID),
				chunkSize: totalChunkSize,
			})
		}
	}

	readPtr = 0
	for _, chunk := range chunkInfo {
		if readPtr+chunk.chunkSize > len(buffer) {
			return nil, fmt.Errorf("unexpected end of buffer while reading chunk data")
		}
		buf := buffer[readPtr : readPtr+chunk.chunkSize]

		data = append(data, &ArchiveFile{
			ID:   ArchiveID(chunk.entryID),
			Data: buf,
		})
		readPtr += chunk.chunkSize
	}
	return &ArchiveGroup{Files: data}, nil
}
