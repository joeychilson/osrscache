package osrscache

import (
	"fmt"
	"path/filepath"
)

const ReferenceTableID = 255

type Cache struct {
	Data    *DataFile
	Indices *Indices
}

func NewCache(path string) (*Cache, error) {
	data, err := NewDataFile(filepath.Join(path, "main_file_cache.dat2"))
	if err != nil {
		return nil, fmt.Errorf("opening dat2 file: %w", err)
	}

	indices, err := NewIndices(path)
	if err != nil {
		data.Close()
		return nil, fmt.Errorf("creating indices: %w", err)
	}

	return &Cache{Data: data, Indices: indices}, nil
}

func (c *Cache) ArchiveData(indexID IndexID, archiveID ArchiveID) ([]byte, error) {
	index, ok := c.Indices.Get(indexID)
	if !ok {
		return nil, fmt.Errorf("index not found: %d", indexID)
	}

	archiveRef, ok := index.ArchiveRef(archiveID)
	if !ok {
		return nil, fmt.Errorf("archive not found: index %d, archive %d", indexID, archiveID)
	}

	data, err := c.Data.Read(archiveRef)
	if err != nil {
		return nil, fmt.Errorf("reading archive data: %w", err)
	}

	return data, nil
}

func (c *Cache) ArchiveGroup(indexID IndexID, archiveID ArchiveID, entryCount int) (*ArchiveGroup, error) {
	data, err := c.ArchiveData(indexID, archiveID)
	if err != nil {
		return nil, fmt.Errorf("reading archive: %w", err)
	}

	archiveData, err := DecompressArchiveData(data)
	if err != nil {
		return nil, fmt.Errorf("decompressing archive: %w", err)
	}

	group, err := NewArchiveGroup(archiveData, entryCount)
	if err != nil {
		return nil, fmt.Errorf("creating archive group: %w", err)
	}
	return group, nil
}

func (c *Cache) ReferenceTable(indexID IndexID) (*IndexMetadata, error) {
	archive, err := c.ArchiveData(ReferenceTableID, ArchiveID(indexID))
	if err != nil {
		return nil, fmt.Errorf("reading reference table: %w", err)
	}

	archiveData, err := DecompressArchiveData(archive)
	if err != nil {
		return nil, fmt.Errorf("decompressing reference table: %w", err)
	}

	meta, err := NewIndexMetadata(archiveData)
	if err != nil {
		return nil, fmt.Errorf("creating reference table metadata: %w", err)
	}

	return meta, nil
}

func (c *Cache) EntityCount(indexID IndexID, archiveID ArchiveID) (int, error) {
	meta, err := c.ReferenceTable(indexID)
	if err != nil {
		return 0, fmt.Errorf("reading reference table: %w", err)
	}
	return meta.Archives[archiveID-1].EntryCount, nil
}

func (c *Cache) ItemDefinitions() (ItemDefinitions, error) {
	entryCount, err := c.EntityCount(2, 10)
	if err != nil {
		return nil, fmt.Errorf("getting entity count: %w", err)
	}

	group, err := c.ArchiveGroup(2, 10, entryCount)
	if err != nil {
		return nil, fmt.Errorf("getting item definitions: %w", err)
	}

	definitions := make(ItemDefinitions, len(group.Files))
	for _, file := range group.Files {
		def, err := NewItemDefinition(uint16(file.ID), file.Data)
		if err != nil {
			return nil, fmt.Errorf("creating item definition: %w", err)
		}
		definitions[uint16(file.ID)] = def
	}

	return definitions, nil
}

func (c *Cache) NPCDefinitions() (NPCDefinitions, error) {
	entryCount, err := c.EntityCount(2, 9)
	if err != nil {
		return nil, fmt.Errorf("getting entity count: %w", err)
	}

	group, err := c.ArchiveGroup(2, 9, entryCount)
	if err != nil {
		return nil, fmt.Errorf("getting npc definitions: %w", err)
	}

	definitions := make(NPCDefinitions, len(group.Files))
	for _, file := range group.Files {
		def, err := NewNPCDefinition(uint16(file.ID), file.Data)
		if err != nil {
			return nil, fmt.Errorf("creating npc definition: %w", err)
		}
		definitions[uint16(file.ID)] = def
	}

	return definitions, nil
}

func (c *Cache) Close() error {
	return c.Data.Close()
}
