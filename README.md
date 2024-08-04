# osrscache

A Go library for reading the Old School Runescape cache.

Note: This library is still under development.

## Usage

```go
package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joeychilson/osrscache"
	"github.com/joeychilson/osrscache/jagex"
)

func main() {
	startTime := time.Now()

	homeFolder, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("getting home directory: %v", err)
	}

	cachePath := filepath.Join(homeFolder, "Jagex", "Old School Runescape", "data")

	store, err := jagex.Open(cachePath)
	if err != nil {
		log.Fatalf("opening store: %v", err)
	}
	defer store.Close()

	cache := osrscache.NewCache(store)

	err = cache.ExportItemDefinitions("items", osrscache.JsonExportModeIndividual)
	if err != nil {
		log.Fatalf("exporting item definitions: %v", err)
	}

	err = cache.ExportNPCDefinitions("npcs", osrscache.JsonExportModeIndividual)
	if err != nil {
		log.Fatalf("exporting npc definitions: %v", err)
	}

	err = cache.ExportObjectDefinitions("objects", osrscache.JsonExportModeIndividual)
	if err != nil {
		log.Fatalf("exporting object definitions: %v", err)
	}

	err = cache.ExportEnums("enums", osrscache.JsonExportModeIndividual)
	if err != nil {
		log.Fatalf("exporting enums: %v", err)
	}

	err = cache.ExportStructs("structs", osrscache.JsonExportModeIndividual)
	if err != nil {
		log.Fatalf("exporting structs: %v", err)
	}

	err = cache.ExportSprites("sprites")
	if err != nil {
		log.Fatalf("exporting sprites: %v", err)
	}

	log.Printf("took %v", time.Since(startTime))
}
```

## Acknowledgements

- [rune-fs](https://github.com/jimvdl/rune-fs)
- [runelite](https://github.com/runelite/runelite)
- [osrs-wiki](https://github.com/osrs-wiki/cache-mediawiki)
- [openrs2](https://github.com/openrs2/openrs2)
