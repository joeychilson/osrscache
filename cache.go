package osrscache

import "fmt"

type Cache struct {
	Store Store
}

func NewCache(store Store) *Cache {
	return &Cache{Store: store}
}

func (c *Cache) Index(groupID int) (*Index, error) {
	groupData, err := c.Store.Read(255, groupID)
	if err != nil {
		return nil, fmt.Errorf("reading reference table: %w", err)
	}

	decompressedGroupData, err := DecompressData(groupData)
	if err != nil {
		return nil, fmt.Errorf("decompressing reference table: %w", err)
	}

	index, err := NewIndex(decompressedGroupData)
	if err != nil {
		return nil, fmt.Errorf("creating reference table index: %w", err)
	}
	return index, nil
}

func (c *Cache) Files(archiveID int, groupID int) (map[int][]byte, error) {
	groupData, err := c.Store.Read(archiveID, groupID)
	if err != nil {
		return nil, fmt.Errorf("reading group data: %w", err)
	}

	decompresGroupData, err := DecompressData(groupData)
	if err != nil {
		return nil, fmt.Errorf("decompressing group data: %w", err)
	}

	index, err := c.Index(archiveID)
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	group, err := index.Group(groupID)
	if err != nil {
		return nil, fmt.Errorf("getting group: %w", err)
	}

	files, err := group.Unpack(decompresGroupData)
	if err != nil {
		return nil, fmt.Errorf("unpacking group: %w", err)
	}
	return files, nil
}

func (c *Cache) ItemDefinitions() (map[int]*ItemDefinition, error) {
	files, err := c.Files(2, 10)
	if err != nil {
		return nil, fmt.Errorf("getting item definition files: %w", err)
	}

	definitions := make(map[int]*ItemDefinition, len(files))
	for id, data := range files {
		def := NewItemDefinition(id)
		if err := def.Read(data); err != nil {
			return nil, fmt.Errorf("reading item definition: %w", err)
		}
		definitions[id] = def
	}
	return definitions, nil
}

func (c *Cache) ExportItemDefinitions(outputDir string, mode JSONExportMode) error {
	items, err := c.ItemDefinitions()
	if err != nil {
		return fmt.Errorf("getting item definitions: %w", err)
	}
	return NewJSONExporter(items, outputDir).ExportToJSON(mode, "item")
}

func (c *Cache) NPCDefinitions() (map[int]*NPCDefinition, error) {
	files, err := c.Files(2, 9)
	if err != nil {
		return nil, fmt.Errorf("getting npc definition files: %w", err)
	}

	definitions := make(map[int]*NPCDefinition, len(files))
	for id, data := range files {
		def := NewNPCDefinition(id)
		if err := def.Read(data); err != nil {
			return nil, fmt.Errorf("reading npc definition: %w", err)
		}
		definitions[id] = def
	}
	return definitions, nil
}

func (c *Cache) ExportNPCDefinitions(outputDir string, mode JSONExportMode) error {
	npcs, err := c.NPCDefinitions()
	if err != nil {
		return fmt.Errorf("getting npc definitions: %w", err)
	}
	return NewJSONExporter(npcs, outputDir).ExportToJSON(mode, "npc")
}

func (c *Cache) ObjectDefinitions() (map[int]*ObjectDefinition, error) {
	files, err := c.Files(2, 6)
	if err != nil {
		return nil, fmt.Errorf("getting object definition files: %w", err)
	}

	definitions := make(map[int]*ObjectDefinition, len(files))
	for id, data := range files {
		def := NewObjectDefinition(id)
		if err := def.Read(data); err != nil {
			return nil, fmt.Errorf("reading object definition: %w", err)
		}
		definitions[id] = def
	}
	return definitions, nil
}

func (c *Cache) ExportObjectDefinitions(outputDir string, mode JSONExportMode) error {
	npcs, err := c.ObjectDefinitions()
	if err != nil {
		return fmt.Errorf("getting object definitions: %w", err)
	}
	return NewJSONExporter(npcs, outputDir).ExportToJSON(mode, "object")
}

func (c *Cache) Enums() (map[int]*Enum, error) {
	files, err := c.Files(2, 8)
	if err != nil {
		return nil, fmt.Errorf("getting enums files: %w", err)
	}

	enums := make(map[int]*Enum, len(files))
	for id, data := range files {
		enum := NewEnum(id)
		if err := enum.Read(data); err != nil {
			return nil, fmt.Errorf("reading enum: %w", err)
		}
		enums[id] = enum
	}
	return enums, nil
}

func (c *Cache) ExportEnums(outputDir string, mode JSONExportMode) error {
	enums, err := c.Enums()
	if err != nil {
		return fmt.Errorf("getting enums: %w", err)
	}
	return NewJSONExporter(enums, outputDir).ExportToJSON(mode, "enum")
}

func (c *Cache) Structs() (map[int]*Struct, error) {
	files, err := c.Files(2, 34)
	if err != nil {
		return nil, fmt.Errorf("getting struct types files: %w", err)
	}

	structs := make(map[int]*Struct, len(files))
	for id, data := range files {
		def := NewStruct(id)
		if err := def.Read(data); err != nil {
			return nil, fmt.Errorf("reading struct type: %w", err)
		}
		structs[id] = def
	}
	return structs, nil
}

func (c *Cache) ExportStructs(outputDir string, mode JSONExportMode) error {
	structs, err := c.Structs()
	if err != nil {
		return fmt.Errorf("getting structs: %w", err)
	}
	return NewJSONExporter(structs, outputDir).ExportToJSON(mode, "struct")
}

func (c *Cache) Sprites() (map[uint32]*Sprite, error) {
	groups, err := c.Store.GroupList(8)
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	sprites := make(map[uint32]*Sprite, len(groups))
	for group := range groups {
		archiveData, err := c.Store.Read(8, group)
		if err != nil {
			return nil, fmt.Errorf("reading sprite archive: %w", err)
		}

		decompressedData, err := DecompressData(archiveData)
		if err != nil {
			return nil, fmt.Errorf("decompressing sprite archive: %w", err)
		}

		sprite := NewSprite(group)
		if err := sprite.Read(decompressedData); err != nil {
			return nil, fmt.Errorf("reading sprite: %w", err)
		}
		sprites[uint32(group)] = sprite
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
