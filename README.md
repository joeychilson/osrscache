# osrscache

A Go library for reading the Old School Runescape cache.

Note: This library is still under development.

## Usage

```go
package main

import (
	"log"
	"time"

	"github.com/joeychilson/osrscache"
	"github.com/joeychilson/osrscache/openrs2"
)

func main() {
	startTime := time.Now()

	store, err := openrs2.Open("./cache")
	if err != nil {
		log.Fatalf("opening store: %v", err)
	}

	cache := osrscache.NewCache(store)

	items, err := cache.Items()
	if err != nil {
		log.Fatalf("getting items: %v", err)
	}
	log.Printf("loaded %d items", len(items))

	npcs, err := cache.NPCs()
	if err != nil {
		log.Fatalf("getting npcs: %v", err)
	}
	log.Printf("loaded %d npcs", len(npcs))

	objs, err := cache.Objects()
	if err != nil {
		log.Fatalf("getting objects: %v", err)
	}
	log.Printf("loaded %d objects", len(objs))

	structs, err := cache.Structs()
	if err != nil {
		log.Fatalf("getting structs: %v", err)
	}
	log.Printf("loaded %d structs", len(structs))

	enums, err := cache.Enums()
	if err != nil {
		log.Fatalf("getting enums: %v", err)
	}
	log.Printf("loaded %d enums", len(enums))

	sprites, err := cache.Sprites()
	if err != nil {
		log.Fatalf("getting sprites: %v", err)
	}
	log.Printf("loaded %d sprites", len(sprites))

	textures, err := cache.Textures()
	if err != nil {
		log.Fatalf("getting textures: %v", err)
	}
	log.Printf("loaded %d textures", len(textures))

	log.Printf("took: %v", time.Since(startTime))
}
```

## Acknowledgements

- [runelite](https://github.com/runelite/runelite)
- [osrs-wiki](https://github.com/osrs-wiki/cache-mediawiki)
- [openrs2](https://github.com/openrs2/openrs2)
