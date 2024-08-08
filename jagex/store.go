package jagex

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	DataFileName            = "main_file_cache.dat2"
	IndexFilePrefix         = "main_file_cache.idx"
	MaxIndexFiles           = 256
	IndexEntrySize          = 6
	BlockHeaderSize         = 8
	ExtendedBlockHeaderSize = 10
	BlockDataSize           = 512
	ExtendedBlockDataSize   = 510
)

type JagexStore struct {
	path       string
	dataFile   *os.File
	indexFiles []*os.File
}

func Open(path string) (*JagexStore, error) {
	dataFile, err := os.Open(filepath.Join(path, "main_file_cache.dat2"))
	if err != nil {
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}

	indexFiles := make([]*os.File, MaxIndexFiles)
	for i := 0; i < MaxIndexFiles; i++ {
		indexFile, err := os.Open(filepath.Join(path, fmt.Sprintf("main_file_cache.idx%d", i)))
		if err != nil {
			if os.IsNotExist(err) && i != 255 {
				continue
			}
			return nil, fmt.Errorf("failed to open index file %d: %w", i, err)
		}
		indexFiles[i] = indexFile
	}
	return &JagexStore{path: path, dataFile: dataFile, indexFiles: indexFiles}, nil
}

func (s *JagexStore) Close() error {
	if err := s.dataFile.Close(); err != nil {
		return fmt.Errorf("failed to close data file: %w", err)
	}

	for i, indexFile := range s.indexFiles {
		if indexFile == nil {
			continue
		}
		if err := indexFile.Close(); err != nil {
			return fmt.Errorf("failed to close index file %d: %w", i, err)
		}
	}
	return nil
}

func (s *JagexStore) ArchiveList() ([]uint8, error) {
	if s.indexFiles == nil {
		return nil, fmt.Errorf("no index files loaded")
	}

	var archives []uint8
	for id, file := range s.indexFiles {
		if file == nil {
			continue
		}
		archives = append(archives, uint8(id))
	}

	if len(archives) == 0 {
		return nil, fmt.Errorf("archive does not exist")
	}
	return archives, nil
}

func (s *JagexStore) ArchiveExists(archiveID uint8) bool {
	return s.indexFiles[archiveID] != nil
}

func (s *JagexStore) GroupList(archiveID uint8) ([]uint32, error) {
	indexFile := s.indexFiles[archiveID]
	if indexFile == nil {
		return nil, fmt.Errorf("archive %d does not exist", archiveID)
	}

	stat, err := indexFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat index file: %w", err)
	}

	fileSize := stat.Size()
	if fileSize%IndexEntrySize != 0 {
		return nil, fmt.Errorf("invalid index file size: %d", fileSize)
	}

	groups := make([]uint32, 0, fileSize/IndexEntrySize)
	buffer := make([]byte, IndexEntrySize*1024)

	for position := int64(0); position < fileSize; position += int64(len(buffer)) {
		n, err := indexFile.ReadAt(buffer, position)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read index file at position %d: %w", position, err)
		}

		for i := 0; i < n; i += IndexEntrySize {
			group := int(position)/IndexEntrySize + i/IndexEntrySize
			block := int(buffer[i+3])<<16 | int(buffer[i+4])<<8 | int(buffer[i+5])
			if block != 0 {
				groups = append(groups, uint32(group))
			}
		}

		if err == io.EOF {
			break
		}
	}
	return groups, nil
}

func (s *JagexStore) GroupExists(archiveID uint8, groupID uint32) bool {
	entry, err := s.IndexEntry(archiveID, groupID)
	if err != nil {
		return false
	}
	return entry.Block != 0
}

func (s *JagexStore) Read(archiveID uint8, groupID uint32) ([]byte, error) {
	entry, err := s.IndexEntry(archiveID, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to read index entry: %w", err)
	}

	if entry.Block == 0 {
		return nil, fmt.Errorf("group %d does not exist in archive %d", groupID, archiveID)
	}

	extended := groupID >= 65536

	var (
		blockHeaderSize int
		blockDataSize   int
	)
	if extended {
		blockHeaderSize = ExtendedBlockHeaderSize
		blockDataSize = ExtendedBlockDataSize
	} else {
		blockHeaderSize = BlockHeaderSize
		blockDataSize = BlockDataSize
	}

	dataFileStat, err := s.dataFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat data file: %w", err)
	}

	entryBuffer := make([]byte, entry.Size)
	blockBuffer := make([]byte, blockHeaderSize)

	currentBlock := entry.Block
	blockNum := 0

	var bytesRead int
	for bytesRead < int(entry.Size) {
		if currentBlock == 0 {
			return nil, fmt.Errorf("group shorter than expected")
		}

		pos := int64(currentBlock) * int64(blockHeaderSize+blockDataSize)

		if pos+int64(blockHeaderSize) > dataFileStat.Size() {
			return nil, fmt.Errorf("next block is outside the data file")
		}

		_, err := s.dataFile.ReadAt(blockBuffer, pos)
		if err != nil {
			return nil, fmt.Errorf("failed to read block header at position %d: %w", pos, err)
		}

		var (
			actualGroup   int
			actualNum     int
			nextBlock     int
			actualArchive int
		)

		if extended {
			actualGroup = int(blockBuffer[0])<<24 | int(blockBuffer[1])<<16 | int(blockBuffer[2])<<8 | int(blockBuffer[3])
			actualNum = int(blockBuffer[4])<<8 | int(blockBuffer[5])
			nextBlock = int(blockBuffer[6])<<16 | int(blockBuffer[7])<<8 | int(blockBuffer[8])
			actualArchive = int(blockBuffer[9])
		} else {
			actualGroup = int(blockBuffer[0])<<8 | int(blockBuffer[1])
			actualNum = int(blockBuffer[2])<<8 | int(blockBuffer[3])
			nextBlock = int(blockBuffer[4])<<16 | int(blockBuffer[5])<<8 | int(blockBuffer[6])
			actualArchive = int(blockBuffer[7])
		}

		if actualGroup != int(groupID) {
			return nil, fmt.Errorf("expected group %d, but got %d", groupID, actualGroup)
		}

		if actualNum != blockNum {
			return nil, fmt.Errorf("expected block number %d, but got %d", blockNum, actualNum)
		}

		if actualArchive != int(archiveID) {
			return nil, fmt.Errorf("expected archive %d, but got %d", archiveID, actualArchive)
		}

		dataSize := int(entry.Size) - bytesRead
		if dataSize > blockDataSize {
			dataSize = blockDataSize
		}

		_, err = s.dataFile.ReadAt(entryBuffer[bytesRead:bytesRead+dataSize], pos+int64(blockHeaderSize))
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read block data at position %d: %w", pos+int64(blockHeaderSize), err)
		}

		bytesRead += dataSize
		currentBlock = uint32(nextBlock)
		blockNum++
	}
	return entryBuffer, nil
}

type IndexEntry struct {
	Size  uint32
	Block uint32
}

func (s *JagexStore) IndexEntry(archiveID uint8, groupID uint32) (*IndexEntry, error) {
	indexFile := s.indexFiles[int(archiveID)]
	if indexFile == nil {
		return nil, fmt.Errorf("archive %d does not exist", archiveID)
	}

	stat, err := indexFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat index file: %w", err)
	}

	fileSize := stat.Size()
	if fileSize%IndexEntrySize != 0 {
		return nil, fmt.Errorf("invalid index file size: %d", fileSize)
	}

	buffer := make([]byte, IndexEntrySize)
	position := int64(groupID) * IndexEntrySize

	n, err := indexFile.ReadAt(buffer, position)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file at position %d: %w", position, err)
	}

	if n != IndexEntrySize {
		return nil, fmt.Errorf("invalid index entry size: %d", n)
	}

	return &IndexEntry{
		Size:  uint32(buffer[0])<<16 | uint32(buffer[1])<<8 | uint32(buffer[2]),
		Block: uint32(buffer[3])<<16 | uint32(buffer[4])<<8 | uint32(buffer[5]),
	}, nil
}
