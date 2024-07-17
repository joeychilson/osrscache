package osrscache

import (
	"fmt"
	"log"
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

func (c *Cache) ArchiveMetadata(indexID IndexID, archiveID ArchiveID) (*ArchiveMetadata, error) {
	meta, err := c.ReferenceTable(indexID)
	if err != nil {
		return nil, fmt.Errorf("reading reference table: %w", err)
	}
	archive, err := meta.ArchiveByID(archiveID)
	if err != nil {
		return nil, fmt.Errorf("getting archive: %w", err)
	}
	return archive, nil
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

func (c *Cache) ArchiveGroup(indexID IndexID, archiveID ArchiveID) (*ArchiveGroup, error) {
	data, err := c.ArchiveData(indexID, archiveID)
	if err != nil {
		return nil, fmt.Errorf("reading archive: %w", err)
	}

	archiveData, err := DecompressArchiveData(data)
	if err != nil {
		return nil, fmt.Errorf("decompressing archive: %w", err)
	}

	archiveMetadata, err := c.ArchiveMetadata(indexID, archiveID)
	if err != nil {
		return nil, fmt.Errorf("getting archive metadata: %w", err)
	}

	group, err := NewArchiveGroup(archiveMetadata, archiveData)
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

func (c *Cache) ItemDefinitions() (map[uint16]*ItemDefinition, error) {
	group, err := c.ArchiveGroup(2, 10)
	if err != nil {
		return nil, fmt.Errorf("getting items archive group: %w", err)
	}

	definitions := make(map[uint16]*ItemDefinition, len(group.Files))
	for _, file := range group.Files {
		def, err := NewItemDefinition(uint16(file.ID), file.Data)
		if err != nil {
			return nil, fmt.Errorf("creating item definition: %w", err)
		}
		definitions[uint16(file.ID)] = def
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

func (c *Cache) NPCDefinitions() (map[uint16]*NPCDefinition, error) {
	group, err := c.ArchiveGroup(2, 9)
	if err != nil {
		return nil, fmt.Errorf("getting npcs archive group: %w", err)
	}

	definitions := make(map[uint16]*NPCDefinition, len(group.Files))
	for _, file := range group.Files {
		def, err := NewNPCDefinition(uint16(file.ID), file.Data)
		if err != nil {
			return nil, fmt.Errorf("creating npc definition: %w", err)
		}
		definitions[uint16(file.ID)] = def
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

func (c *Cache) ObjectDefinitions() (map[uint16]*ObjectDefinition, error) {
	group, err := c.ArchiveGroup(2, 6)
	if err != nil {
		return nil, fmt.Errorf("getting objects archive group: %w", err)
	}

	definitions := make(map[uint16]*ObjectDefinition, len(group.Files))
	for _, file := range group.Files {
		def, err := NewObjectDefinition(uint16(file.ID), file.Data)
		if err != nil {
			return nil, fmt.Errorf("creating object definition: %w", err)
		}
		definitions[uint16(file.ID)] = def
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

func (c *Cache) Sprites() (map[uint32]*Sprite, error) {
	index, err := c.Indices.Get(8)
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	sprites := make(map[uint32]*Sprite, len(index.ArchiveIDs()))
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

func (c *Cache) Enums() (map[uint16]*Enum, error) {
	group, err := c.ArchiveGroup(2, 8)
	if err != nil {
		return nil, fmt.Errorf("getting enums archive group: %w", err)
	}

	enums := make(map[uint16]*Enum, len(group.Files))
	for _, file := range group.Files {
		enum, err := NewEnum(uint16(file.ID), file.Data)
		if err != nil {
			return nil, fmt.Errorf("creating enum type: %w", err)
		}
		enums[uint16(file.ID)] = enum
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

func (c *Cache) Structs() (map[uint16]*Struct, error) {
	group, err := c.ArchiveGroup(2, 34)
	if err != nil {
		return nil, fmt.Errorf("getting struct types archive group: %w", err)
	}

	structs := make(map[uint16]*Struct, len(group.Files))
	for _, file := range group.Files {
		def, err := NewStruct(uint16(file.ID), file.Data)
		if err != nil {
			return nil, fmt.Errorf("creating struct type: %w", err)
		}
		structs[uint16(file.ID)] = def
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

func (c *Cache) CombatAchievements() (map[uint32]*CombatAchievement, error) {
	enums, err := c.Enums()
	if err != nil {
		return nil, fmt.Errorf("getting enums: %w", err)
	}

	structs, err := c.Structs()
	if err != nil {
		return nil, fmt.Errorf("getting structs: %w", err)
	}

	tierEnum, ok := enums[TierEnumID]
	if !ok {
		return nil, fmt.Errorf("tier enum not found")
	}

	typeEnum, ok := enums[TypeEnumID]
	if !ok {
		return nil, fmt.Errorf("type enum not found")
	}

	monsterEnum, ok := enums[MonsterEnumID]
	if !ok {
		return nil, fmt.Errorf("monster enum not found")
	}

	var (
		combatAchievementTaskIDs = []uint16{3981, 3982, 3983, 3984, 3985, 3986}
		combatAchievements       = make(map[uint32]*CombatAchievement)
	)
	for _, taskID := range combatAchievementTaskIDs {
		taskEnum, ok := enums[taskID]
		if !ok {
			return nil, fmt.Errorf("task enum not found")
		}
		for _, taskID := range taskEnum.Values {
			taskStruct, ok := structs[uint16(taskID.(int32))]
			if !ok {
				log.Printf("task struct not found: %d", taskID)
				continue
			}

			id, ok := taskStruct.Params[CAIDParamID].(uint32)
			if !ok {
				return nil, fmt.Errorf("invalid or missing ID")
			}

			title, ok := taskStruct.Params[TitleParamID].(string)
			if !ok {
				return nil, fmt.Errorf("invalid or missing title")
			}

			description, ok := taskStruct.Params[DescriptionParamID].(string)
			if !ok {
				return nil, fmt.Errorf("invalid or missing description")
			}

			monsterKey, ok := taskStruct.Params[MonsterParamID].(uint32)
			if !ok {
				return nil, fmt.Errorf("invalid or missing monster key")
			}

			tierKey, ok := taskStruct.Params[TierParamID].(uint32)
			if !ok {
				return nil, fmt.Errorf("invalid or missing tier key")
			}

			typeKey, ok := taskStruct.Params[TypeParamID].(uint32)
			if !ok {
				return nil, fmt.Errorf("invalid or missing type key")
			}

			monster, ok := monsterEnum.Values[int32(monsterKey)].(string)
			if !ok {
				log.Printf("invalid monster mapping: %d", monsterKey)
				continue
			}

			tier, ok := tierEnum.Values[int32(tierKey)].(string)
			if !ok {
				return nil, fmt.Errorf("invalid tier mapping")
			}

			achievementType, ok := typeEnum.Values[int32(typeKey)].(string)
			if !ok {
				return nil, fmt.Errorf("invalid type mapping")
			}

			combatAchievements[id] = &CombatAchievement{
				ID:          id,
				Title:       title,
				Description: description,
				Tier:        CombatAchievementTier(tier),
				Type:        CombatAchievementType(achievementType),
				Monster:     monster,
			}
		}
	}
	return combatAchievements, nil
}

func (c *Cache) ExportCombatAchievements(outputDir string, mode JSONExportMode) error {
	achievements, err := c.CombatAchievements()
	if err != nil {
		return fmt.Errorf("getting item definitions: %w", err)
	}
	return NewJSONExporter(achievements, outputDir).ExportToJSON(mode, "combat_achievement")
}

func (c *Cache) Close() error {
	return c.Data.Close()
}
