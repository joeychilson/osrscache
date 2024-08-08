package osrscache

import "fmt"

type Cache struct {
	Store Store
}

func NewCache(store Store) *Cache {
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

func (c *Cache) Files(archiveID uint8, groupID uint32) (map[int][]byte, error) {
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

func (c *Cache) Item(id int) (*Item, error) {
	files, err := c.Files(2, 10)
	if err != nil {
		return nil, fmt.Errorf("getting item files: %w", err)
	}

	data, ok := files[id]
	if !ok {
		return nil, fmt.Errorf("item %d not found", id)
	}

	item := NewItem(id)
	if err := item.Read(data); err != nil {
		return nil, fmt.Errorf("reading item: %w", err)
	}
	return item, nil
}

func (c *Cache) Items() (map[int]*Item, error) {
	files, err := c.Files(2, 10)
	if err != nil {
		return nil, fmt.Errorf("getting item files: %w", err)
	}

	items := make(map[int]*Item, len(files))
	for id, data := range files {
		item := NewItem(id)
		if err := item.Read(data); err != nil {
			return nil, fmt.Errorf("reading item: %w", err)
		}
		items[id] = item
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

func (c *Cache) NPC(id int) (*NPC, error) {
	files, err := c.Files(2, 9)
	if err != nil {
		return nil, fmt.Errorf("getting npc files: %w", err)
	}

	data, ok := files[id]
	if !ok {
		return nil, fmt.Errorf("npc %d not found", id)
	}

	npc := NewNPC(id)
	if err := npc.Read(data); err != nil {
		return nil, fmt.Errorf("reading npc: %w", err)
	}
	return npc, nil
}

func (c *Cache) NPCs() (map[int]*NPC, error) {
	files, err := c.Files(2, 9)
	if err != nil {
		return nil, fmt.Errorf("getting npc files: %w", err)
	}

	npcs := make(map[int]*NPC, len(files))
	for id, data := range files {
		npc := NewNPC(id)
		if err := npc.Read(data); err != nil {
			return nil, fmt.Errorf("reading npc: %w", err)
		}
		npcs[id] = npc
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

func (c *Cache) Object(id int) (*Object, error) {
	files, err := c.Files(2, 6)
	if err != nil {
		return nil, fmt.Errorf("getting object files: %w", err)
	}

	data, ok := files[id]
	if !ok {
		return nil, fmt.Errorf("object %d not found", id)
	}

	obj := NewObject(id)
	if err := obj.Read(data); err != nil {
		return nil, fmt.Errorf("reading object: %w", err)
	}
	return obj, nil
}

func (c *Cache) Objects() (map[int]*Object, error) {
	files, err := c.Files(2, 6)
	if err != nil {
		return nil, fmt.Errorf("getting object files: %w", err)
	}

	objs := make(map[int]*Object, len(files))
	for id, data := range files {
		obj := NewObject(id)
		if err := obj.Read(data); err != nil {
			return nil, fmt.Errorf("reading object: %w", err)
		}
		objs[id] = obj
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

func (c *Cache) Enum(id int) (*Enum, error) {
	files, err := c.Files(2, 8)
	if err != nil {
		return nil, fmt.Errorf("getting enum files: %w", err)
	}

	data, ok := files[id]
	if !ok {
		return nil, fmt.Errorf("enum %d not found", id)
	}

	enum := NewEnum(id)
	if err := enum.Read(data); err != nil {
		return nil, fmt.Errorf("reading enum: %w", err)
	}
	return enum, nil
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

func (c *Cache) Struct(id int) (*Struct, error) {
	files, err := c.Files(2, 34)
	if err != nil {
		return nil, fmt.Errorf("getting struct files: %w", err)
	}

	data, ok := files[id]
	if !ok {
		return nil, fmt.Errorf("struct %d not found", id)
	}

	str := NewStruct(id)
	if err := str.Read(data); err != nil {
		return nil, fmt.Errorf("reading struct: %w", err)
	}
	return str, nil
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

func (c *Cache) Sprite(id int) (*Sprite, error) {
	archiveData, err := c.Store.Read(8, id)
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

func (c *Cache) Sprites() (map[int]*Sprite, error) {
	groups, err := c.Store.GroupList(8)
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	sprites := make(map[int]*Sprite, len(groups))
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
		sprites[group] = sprite
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

func (c *Cache) Texture(id int) (*Texture, error) {
	files, err := c.Files(9, 0)
	if err != nil {
		return nil, fmt.Errorf("getting texture files: %w", err)
	}

	data, ok := files[id]
	if !ok {
		return nil, fmt.Errorf("texture %d not found", id)
	}

	texture := NewTexture(id)
	if err := texture.Read(data); err != nil {
		return nil, fmt.Errorf("reading texture: %w", err)
	}
	return texture, nil
}

func (c *Cache) Textures() (map[int]*Texture, error) {
	files, err := c.Files(9, 0)
	if err != nil {
		return nil, fmt.Errorf("getting object definition files: %w", err)
	}

	textures := make(map[int]*Texture, len(files))
	for id, data := range files {
		texture := NewTexture(id)
		if err := texture.Read(data); err != nil {
			return nil, fmt.Errorf("reading texture: %w", err)
		}
		textures[id] = texture
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
