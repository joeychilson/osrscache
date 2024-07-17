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
	Files map[ArchiveID]*ArchiveFile
}

type ArchiveFile struct {
	ID   ArchiveID
	Data []byte
}

func NewArchiveGroup(metadata *ArchiveMetadata, data []byte) (*ArchiveGroup, error) {
	if metadata.EntryCount == 1 {
		return &ArchiveGroup{
			Files: map[ArchiveID]*ArchiveFile{metadata.ID: {ID: metadata.ID, Data: data}},
		}, nil
	}

	reader := bytes.NewReader(data)

	_, err := reader.Seek(-1, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to end: %w", err)
	}

	var chunkCount uint8
	if err := binary.Read(reader, binary.BigEndian, &chunkCount); err != nil {
		return nil, fmt.Errorf("failed to read chunk count: %w", err)
	}

	readPtr := reader.Size() - 1 - int64(chunkCount)*int64(metadata.EntryCount)*4

	_, err = reader.Seek(readPtr, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to chunk size data: %w", err)
	}

	type cachedChunk struct {
		entryID   ArchiveID
		chunkSize int
	}

	cachedChunks := make([]cachedChunk, 0, int(chunkCount)*metadata.EntryCount)

	for i := 0; i < int(chunkCount); i++ {
		totalChunkSize := 0
		for j := 0; j < metadata.EntryCount; j++ {
			var delta int32
			if err := binary.Read(reader, binary.BigEndian, &delta); err != nil {
				return nil, fmt.Errorf("failed to read chunk delta: %w", err)
			}

			totalChunkSize += int(delta)
			readPtr += 4

			cachedChunks = append(cachedChunks, cachedChunk{
				entryID:   ArchiveID(metadata.ValidIDs[j]),
				chunkSize: totalChunkSize,
			})
		}
	}

	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to start: %w", err)
	}

	files := make(map[ArchiveID]*ArchiveFile, metadata.EntryCount)

	for _, chunk := range cachedChunks {
		buf := make([]byte, chunk.chunkSize)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk data: %w", err)
		}

		if existingFile, ok := files[chunk.entryID]; ok {
			existingFile.Data = append(existingFile.Data, buf...)
		} else {
			files[chunk.entryID] = &ArchiveFile{
				ID:   chunk.entryID,
				Data: buf,
			}
		}
	}

	return &ArchiveGroup{Files: files}, nil
}
