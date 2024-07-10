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
	index, err := c.Indices.Get(indexID)
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	archiveRef, err := index.ArchiveRef(archiveID)
	if err != nil {
		return nil, fmt.Errorf("getting archive reference: %w", err)
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
		return nil, fmt.Errorf("getting item entity count: %w", err)
	}

	group, err := c.ArchiveGroup(2, 10, entryCount)
	if err != nil {
		return nil, fmt.Errorf("getting items archive group: %w", err)
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

func (c *Cache) ExportItemDefinitions(outputDir string, mode JSONExportMode, filename string) error {
	items, err := c.ItemDefinitions()
	if err != nil {
		return fmt.Errorf("getting item definitions: %w", err)
	}
	return NewJSONExporter(items, outputDir).ExportToJSON(mode, filename)
}

func (c *Cache) NPCDefinitions() (NPCDefinitions, error) {
	entryCount, err := c.EntityCount(2, 9)
	if err != nil {
		return nil, fmt.Errorf("getting npc entity count: %w", err)
	}

	group, err := c.ArchiveGroup(2, 9, entryCount)
	if err != nil {
		return nil, fmt.Errorf("getting npcs archive group: %w", err)
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

func (c *Cache) ExportNPCDefinitions(outputDir string, mode JSONExportMode, filename string) error {
	npcs, err := c.NPCDefinitions()
	if err != nil {
		return fmt.Errorf("getting npc definitions: %w", err)
	}
	return NewJSONExporter(npcs, outputDir).ExportToJSON(mode, filename)
}

func (c *Cache) ObjectDefinitions() (ObjectDefinitions, error) {
	entryCount, err := c.EntityCount(2, 6)
	if err != nil {
		return nil, fmt.Errorf("getting object entity count: %w", err)
	}

	group, err := c.ArchiveGroup(2, 6, entryCount)
	if err != nil {
		return nil, fmt.Errorf("getting objects archive group: %w", err)
	}

	definitions := make(ObjectDefinitions, len(group.Files))
	for _, file := range group.Files {
		def, err := NewObjectDefinition(uint16(file.ID), file.Data)
		if err != nil {
			return nil, fmt.Errorf("creating object definition: %w", err)
		}
		definitions[uint16(file.ID)] = def
	}
	return definitions, nil
}

func (c *Cache) ExportObjectDefinitions(outputDir string, mode JSONExportMode, filename string) error {
	npcs, err := c.ObjectDefinitions()
	if err != nil {
		return fmt.Errorf("getting object definitions: %w", err)
	}
	return NewJSONExporter(npcs, outputDir).ExportToJSON(mode, filename)
}

func (c *Cache) Sprites() (Sprites, error) {
	index, err := c.Indices.Get(8)
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	sprites := make(Sprites, len(index.ArchiveIDs()))
	for _, id := range index.ArchiveIDs() {
		archiveData, err := c.ArchiveData(8, id)
		if err != nil {
			return nil, fmt.Errorf("reading sprite archive: %w", err)
		}

		decompressedData, err := DecompressArchiveData(archiveData)
		if err != nil {
			return nil, fmt.Errorf("decompressing sprite archive: %w", err)
		}

		sprite, err := NewSprite(uint32(id), decompressedData)
		if err != nil {
			return nil, fmt.Errorf("creating sprite: %w", err)
		}
		sprites[uint32(id)] = sprite
	}
	return sprites, nil
}

func (c *Cache) ExportSprites(outputDir string) error {
	sprites, err := c.Sprites()
	if err != nil {
		return fmt.Errorf("getting sprites: %w", err)
	}
	return NewImageExporter(sprites, outputDir).ExportToImage("sprite")
}

func (c *Cache) Close() error {
	return c.Data.Close()
}
