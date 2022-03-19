# geopath

Dumb shortest-path implementation using GeoJSON linestrings to navigate.

Just plain ol Dijkstra's.

Usage:

```go
package main

import (
	"fmt"
	"log"
	"os"

	geopath "github.com/sebnyberg/geopath"
)

func main() {
	f, err := os.OpenFile("path/to/geo.json", os.O_RDONLY, 0444)
	if err != nil {
		log.Fatalln(err)
	}
	paths, err := geopath.ParsePaths(f)
	if err != nil {
		log.Fatalln(err)
	}
	start := [2]float64{-84.397252, 33.792997}
	end := [2]float64{-84.395111, 33.791666}
	precision := 0.00001
	path, distance, err := geopath.FindShortestPath(paths, start, end, precision)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Shortest distance: %.05f\n", distance)
	fmt.Printf("Path:\n")
	for _, p := range path {
		fmt.Printf("(%.06f, %.06f)\n", p[0], p[1])
	}
}
```
