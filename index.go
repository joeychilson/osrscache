package osrscache

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const IndexFilePrefix = "main_file_cache.idx"

type IndexID uint8

type Indices struct {
	indices map[IndexID]*Index
	mu      sync.RWMutex
}

func NewIndices(path string) (*Indices, error) {
	indices := &Indices{
		indices: make(map[IndexID]*Index),
	}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(info.Name(), IndexFilePrefix) {
			idStr := strings.TrimPrefix(info.Name(), IndexFilePrefix)
			id, err := strconv.ParseUint(idStr, 10, 8)
			if err != nil {
				return fmt.Errorf("parsing index ID: %w", err)
			}
			index, err := NewIndex(IndexID(id), filePath)
			if err != nil {
				return fmt.Errorf("creating index: %w", err)
			}
			indices.indices[IndexID(id)] = index
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking cache directory: %w", err)
	}

	return indices, nil
}

func (i *Indices) Get(id IndexID) (*Index, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	index, ok := i.indices[IndexID(id)]
	return index, ok
}

func (i *Indices) Count() int {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return len(i.indices)
}

func (i *Indices) IndexIDs() []uint8 {
	i.mu.RLock()
	defer i.mu.RUnlock()

	ids := make([]uint8, 0, len(i.indices))
	for id := range i.indices {
		ids = append(ids, uint8(id))
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

type Index struct {
	ID          IndexID
	ArchiveRefs map[ArchiveID]*ArchiveRef
	Metadata    *IndexMetadata
	mu          sync.RWMutex
}

func NewIndex(id IndexID, file string) (*Index, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading index file: %w", err)
	}

	index := &Index{
		ID:          id,
		ArchiveRefs: make(map[ArchiveID]*ArchiveRef),
	}

	for i := 0; i < len(data); i += ArchiveRefLen {
		if i+ArchiveRefLen > len(data) {
			break
		}

		archiveID := ArchiveID(i / ArchiveRefLen)
		archiveRef, err := NewArchiveRef(id, archiveID, data[i:i+ArchiveRefLen])
		if err != nil {
			return nil, fmt.Errorf("parsing archive ref: %w", err)
		}

		index.ArchiveRefs[archiveID] = archiveRef
	}

	return index, nil
}

func (i *Index) ArchiveRef(id ArchiveID) (*ArchiveRef, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	ref, ok := i.ArchiveRefs[id]
	return ref, ok
}

func (i *Index) ArchiveIDs() []ArchiveID {
	i.mu.RLock()
	defer i.mu.RUnlock()

	ids := make([]ArchiveID, 0, len(i.ArchiveRefs))
	for id := range i.ArchiveRefs {
		ids = append(ids, id)
	}
	return ids
}

type IndexMetadata struct {
	Archives []ArchiveMetadata
}

func NewIndexMetadata(data []byte) (*IndexMetadata, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty metadata buffer")
	}

	reader := bytes.NewReader(data)

	var protocol uint8
	if err := binary.Read(reader, binary.BigEndian, &protocol); err != nil {
		return nil, fmt.Errorf("reading protocol: %w", err)
	}

	if protocol >= 6 {
		// TODO: Revision?
		_, err := reader.Seek(4, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("skipping unknown uint32: %w", err)
		}
	}

	var flags uint8
	if err := binary.Read(reader, binary.BigEndian, &flags); err != nil {
		return nil, fmt.Errorf("reading flags: %w", err)
	}

	hasNames := (flags & 0x1) != 0
	hasDigests := (flags & 0x2) != 0
	hasLengths := (flags & 0x4) != 0
	hasCompressedChecksum := (flags & 0x8) != 0

	var archiveCount uint16
	if err := binary.Read(reader, binary.BigEndian, &archiveCount); err != nil {
		return nil, fmt.Errorf("reading archive count: %w", err)
	}

	im := &IndexMetadata{
		Archives: make([]ArchiveMetadata, archiveCount),
	}

	var prevArchiveId uint32
	for i := range im.Archives {
		var delta uint16
		if err := binary.Read(reader, binary.BigEndian, &delta); err != nil {
			return nil, fmt.Errorf("reading archive ID delta: %w", err)
		}
		im.Archives[i].ID = ArchiveID(prevArchiveId + uint32(delta))
		prevArchiveId = uint32(im.Archives[i].ID)
	}

	if hasNames {
		for i := range im.Archives {
			if err := binary.Read(reader, binary.BigEndian, &im.Archives[i].NameHash); err != nil {
				return nil, fmt.Errorf("reading name hash: %w", err)
			}
		}
	}

	for i := range im.Archives {
		if err := binary.Read(reader, binary.BigEndian, &im.Archives[i].CRC); err != nil {
			return nil, fmt.Errorf("reading CRC: %w", err)
		}
	}

	if hasCompressedChecksum {
		for i := range im.Archives {
			if err := binary.Read(reader, binary.BigEndian, &im.Archives[i].CompressedChecksum); err != nil {
				return nil, fmt.Errorf("reading compressed checksum: %w", err)
			}
		}
	}

	if hasDigests {
		for i := range im.Archives {
			if err := binary.Read(reader, binary.BigEndian, &im.Archives[i].Digests); err != nil {
				return nil, fmt.Errorf("reading digests: %w", err)
			}
		}
	}

	if hasLengths {
		for i := range im.Archives {
			if err := binary.Read(reader, binary.BigEndian, &im.Archives[i].CompressedSize); err != nil {
				return nil, fmt.Errorf("reading compressed size: %w", err)
			}
			if err := binary.Read(reader, binary.BigEndian, &im.Archives[i].UncompressedSize); err != nil {
				return nil, fmt.Errorf("reading uncompressed size: %w", err)
			}
		}
	}

	for i := range im.Archives {
		if err := binary.Read(reader, binary.BigEndian, &im.Archives[i].Version); err != nil {
			return nil, fmt.Errorf("reading version: %w", err)
		}
	}

	for i := range im.Archives {
		var entryCount uint16
		if err := binary.Read(reader, binary.BigEndian, &entryCount); err != nil {
			return nil, fmt.Errorf("reading entry count: %w", err)
		}
		im.Archives[i].EntryCount = int(entryCount)
	}

	for i := range im.Archives {
		if im.Archives[i].EntryCount > 1 {
			im.Archives[i].ValidIDs = make([]uint32, im.Archives[i].EntryCount)
			var prevEntryId uint32
			for j := range im.Archives[i].ValidIDs {
				var delta uint16
				if err := binary.Read(reader, binary.BigEndian, &delta); err != nil {
					return nil, fmt.Errorf("reading entry ID delta: %w", err)
				}
				im.Archives[i].ValidIDs[j] = prevEntryId + uint32(delta)
				prevEntryId = im.Archives[i].ValidIDs[j]
			}
		}
	}
	return im, nil
}
