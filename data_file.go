package osrscache

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type DataFile struct {
	File *os.File
	Size int64
	mu   sync.RWMutex
}

func NewDataFile(path string) (*DataFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	return &DataFile{File: file, Size: info.Size()}, nil
}

func (d *DataFile) Read(ref *ArchiveRef) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	buffer := make([]byte, 0, ref.Length)
	current := ref.Sector
	headerSize := ref.HeaderSize()
	maxSector := uint32(d.Size / SectorSize)

	for chunk, dataLen := range ref.DataBlocks() {
		if current >= maxSector {
			return nil, fmt.Errorf("sector %d out of bounds (max: %d)", current, maxSector)
		}

		offset := int64(current * SectorSize)

		if _, err := d.File.Seek(offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking to sector: %w", err)
		}

		sectorData := make([]byte, dataLen)
		if _, err := io.ReadFull(d.File, sectorData); err != nil {
			return nil, fmt.Errorf("reading sector data: %w", err)
		}

		sector, err := NewSector(sectorData, headerSize)
		if err != nil {
			return nil, fmt.Errorf("creating sector: %w", err)
		}

		if err := sector.Header.Validate(ref.IndexID, ref.ID, uint32(chunk)); err != nil {
			return nil, fmt.Errorf("validating sector header: %w", err)
		}

		buffer = append(buffer, sector.DataBlock...)

		if sector.Header.Next == 0 {
			break
		}
		current = sector.Header.Next
	}

	if len(buffer) != int(ref.Length) {
		return nil, fmt.Errorf("read data length mismatch: got %d, want %d", len(buffer), ref.Length)
	}

	return buffer, nil
}

func (d *DataFile) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.File.Close()
}
