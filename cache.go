package osrscache

import "fmt"

type Cache struct {
	Store Store
}

func New(store Store) *Cache {
	return &Cache{Store: store}
}

func (c *Cache) Index(archiveID uint8) (*Index, error) {
	groupData, err := c.Store.Read(255, uint32(archiveID))
	if err != nil {
		return nil, fmt.Errorf("reading reference table: %w", err)
	}

	decompressedGroupData, err := DecompressData(groupData)
	if err != nil {
		return nil, fmt.Errorf("decompressing reference table: %w", err)
	}

	index, err := ReadIndex(decompressedGroupData)
	if err != nil {
		return nil, fmt.Errorf("creating reference table index: %w", err)
	}
	return index, nil
}

func (c *Cache) Files(archiveID uint8, groupID uint32) (map[uint32][]byte, error) {
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

func (c *Cache) Item(id uint16) (*Item, error) {
	files, err := c.Files(2, 10)
	if err != nil {
		return nil, fmt.Errorf("getting item files: %w", err)
	}

	data, ok := files[uint32(id)]
	if !ok {
		return nil, fmt.Errorf("item %d not found", id)
	}

	item := NewItem(id)
	if err := item.Read(data); err != nil {
		return nil, fmt.Errorf("reading item: %w", err)
	}
	return item, nil
}

func (c *Cache) Items() (map[uint16]*Item, error) {
	files, err := c.Files(2, 10)
	if err != nil {
		return nil, fmt.Errorf("getting item files: %w", err)
	}

	items := make(map[uint16]*Item, len(files))
	for id, data := range files {
		item := NewItem(uint16(id))
		if err := item.Read(data); err != nil {
			return nil, fmt.Errorf("reading item: %w", err)
		}
		items[uint16(id)] = item
	}
	return items, nil
}

func (c *Cache) ExportItems(outputDir string, mode JSONExportMode) error {
	items, err := c.Items()
	if err != nil {
		return fmt.Errorf("getting items: %w", err)
	}
	return NewJSONExporter(items, outputDir).ExportToJSON(mode, "item")
}

func (c *Cache) NPC(id uint16) (*NPC, error) {
	files, err := c.Files(2, 9)
	if err != nil {
		return nil, fmt.Errorf("getting npc files: %w", err)
	}

	data, ok := files[uint32(id)]
	if !ok {
		return nil, fmt.Errorf("npc %d not found", id)
	}

	npc := NewNPC(uint16(id))
	if err := npc.Read(data); err != nil {
		return nil, fmt.Errorf("reading npc: %w", err)
	}
	return npc, nil
}

func (c *Cache) NPCs() (map[uint16]*NPC, error) {
	files, err := c.Files(2, 9)
	if err != nil {
		return nil, fmt.Errorf("getting npc files: %w", err)
	}

	npcs := make(map[uint16]*NPC, len(files))
	for id, data := range files {
		npc := NewNPC(uint16(id))
		if err := npc.Read(data); err != nil {
			return nil, fmt.Errorf("reading npc: %w", err)
		}
		npcs[uint16(id)] = npc
	}
	return npcs, nil
}

func (c *Cache) ExportNPCs(outputDir string, mode JSONExportMode) error {
	npcs, err := c.NPCs()
	if err != nil {
		return fmt.Errorf("getting npcs: %w", err)
	}
	return NewJSONExporter(npcs, outputDir).ExportToJSON(mode, "npc")
}

func (c *Cache) Object(id uint16) (*Object, error) {
	files, err := c.Files(2, 6)
	if err != nil {
		return nil, fmt.Errorf("getting object files: %w", err)
	}

	data, ok := files[uint32(id)]
	if !ok {
		return nil, fmt.Errorf("object %d not found", id)
	}

	obj := NewObject(uint16(id))
	if err := obj.Read(data); err != nil {
		return nil, fmt.Errorf("reading object: %w", err)
	}
	return obj, nil
}

func (c *Cache) Objects() (map[uint16]*Object, error) {
	files, err := c.Files(2, 6)
	if err != nil {
		return nil, fmt.Errorf("getting object files: %w", err)
	}

	objs := make(map[uint16]*Object, len(files))
	for id, data := range files {
		obj := NewObject(uint16(id))
		if err := obj.Read(data); err != nil {
			return nil, fmt.Errorf("reading object: %w", err)
		}
		objs[uint16(id)] = obj
	}
	return objs, nil
}

func (c *Cache) ExportObjects(outputDir string, mode JSONExportMode) error {
	npcs, err := c.Objects()
	if err != nil {
		return fmt.Errorf("getting objects: %w", err)
	}
	return NewJSONExporter(npcs, outputDir).ExportToJSON(mode, "object")
}

func (c *Cache) Enum(id uint16) (*Enum, error) {
	files, err := c.Files(2, 8)
	if err != nil {
		return nil, fmt.Errorf("getting enum files: %w", err)
	}

	data, ok := files[uint32(id)]
	if !ok {
		return nil, fmt.Errorf("enum %d not found", id)
	}

	enum := NewEnum(id)
	if err := enum.Read(data); err != nil {
		return nil, fmt.Errorf("reading enum: %w", err)
	}
	return enum, nil
}

func (c *Cache) Enums() (map[uint16]*Enum, error) {
	files, err := c.Files(2, 8)
	if err != nil {
		return nil, fmt.Errorf("getting enums files: %w", err)
	}

	enums := make(map[uint16]*Enum, len(files))
	for id, data := range files {
		enum := NewEnum(uint16(id))
		if err := enum.Read(data); err != nil {
			return nil, fmt.Errorf("reading enum: %w", err)
		}
		enums[uint16(id)] = enum
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

func (c *Cache) Struct(id uint16) (*Struct, error) {
	files, err := c.Files(2, 34)
	if err != nil {
		return nil, fmt.Errorf("getting struct files: %w", err)
	}

	data, ok := files[uint32(id)]
	if !ok {
		return nil, fmt.Errorf("struct %d not found", id)
	}

	str := NewStruct(id)
	if err := str.Read(data); err != nil {
		return nil, fmt.Errorf("reading struct: %w", err)
	}
	return str, nil
}

func (c *Cache) Structs() (map[uint16]*Struct, error) {
	files, err := c.Files(2, 34)
	if err != nil {
		return nil, fmt.Errorf("getting struct types files: %w", err)
	}

	structs := make(map[uint16]*Struct, len(files))
	for id, data := range files {
		def := NewStruct(uint16(id))
		if err := def.Read(data); err != nil {
			return nil, fmt.Errorf("reading struct type: %w", err)
		}
		structs[uint16(id)] = def
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

func (c *Cache) Sprite(id uint16) (*Sprite, error) {
	archiveData, err := c.Store.Read(8, uint32(id))
	if err != nil {
		return nil, fmt.Errorf("reading sprite archive: %w", err)
	}

	decompressedData, err := DecompressData(archiveData)
	if err != nil {
		return nil, fmt.Errorf("decompressing sprite archive: %w", err)
	}

	sprite := NewSprite(id)
	if err := sprite.Read(decompressedData); err != nil {
		return nil, fmt.Errorf("reading sprite: %w", err)
	}
	return sprite, nil
}

func (c *Cache) Sprites() (map[uint16]*Sprite, error) {
	groups, err := c.Store.GroupList(8)
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	sprites := make(map[uint16]*Sprite, len(groups))
	for group := range groups {
		archiveData, err := c.Store.Read(8, uint32(group))
		if err != nil {
			return nil, fmt.Errorf("reading sprite archive: %w", err)
		}

		decompressedData, err := DecompressData(archiveData)
		if err != nil {
			return nil, fmt.Errorf("decompressing sprite archive: %w", err)
		}

		sprite := NewSprite(uint16(group))
		if err := sprite.Read(decompressedData); err != nil {
			return nil, fmt.Errorf("reading sprite: %w", err)
		}
		sprites[uint16(group)] = sprite
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

func (c *Cache) Texture(id uint16) (*Texture, error) {
	files, err := c.Files(9, 0)
	if err != nil {
		return nil, fmt.Errorf("getting texture files: %w", err)
	}

	data, ok := files[uint32(id)]
	if !ok {
		return nil, fmt.Errorf("texture %d not found", id)
	}

	texture := NewTexture(id)
	if err := texture.Read(data); err != nil {
		return nil, fmt.Errorf("reading texture: %w", err)
	}
	return texture, nil
}

func (c *Cache) Textures() (map[uint16]*Texture, error) {
	files, err := c.Files(9, 0)
	if err != nil {
		return nil, fmt.Errorf("getting object definition files: %w", err)
	}

	textures := make(map[uint16]*Texture, len(files))
	for id, data := range files {
		texture := NewTexture(uint16(id))
		if err := texture.Read(data); err != nil {
			return nil, fmt.Errorf("reading texture: %w", err)
		}
		textures[uint16(id)] = texture
	}
	return textures, nil
}

func (c *Cache) ExportTextures(outputDir string, mode JSONExportMode) error {
	textures, err := c.Textures()
	if err != nil {
		return fmt.Errorf("getting textures: %w", err)
	}
	return NewJSONExporter(textures, outputDir).ExportToJSON(mode, "texture")
}
