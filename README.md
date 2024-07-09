# osrscache

A Go library for reading the Old School Runescape cache.

Note: This library is still under development.

## Usage

```go
package main

import (
	"fmt"

	"github.com/joeychilson/osrscache"
)

func main() {
	cache, err := osrscache.NewCache("path/to/osrs/cache")
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	items, err := cache.ItemDefinitions()
	if err != nil {
		panic(err)
	}

	item, err := items.Get(4151)
	if err != nil {
		panic(err)
	}

	fmt.Println(item)
}
```

## Acknowledgements

- [rune-fs](https://github.com/jimvdl/rune-fs)
- [runelite](https://github.com/runelite/runelite)
- [osrs-wiki](https://github.com/osrs-wiki/cache-mediawiki)
- [openrs2](https://github.com/openrs2/openrs2)
