package geopath_test

import (
	"os"
	"testing"

	"github.com/sebnyberg/geopath"
)

func TestFindPath(t *testing.T) {
	t.Run("small dataset", func(t *testing.T) {
		f, err := os.Open("testdata/sample_large.json")
		if err != nil {
			t.Errorf("failed to open file, %v", err)
		}
		paths, err := geopath.ParsePaths(f)
		if err != nil {
			t.Errorf("failed to parse paths, %v", err)
		}
		start := [2]float64{-84.396863, 33.792908}
		end := [2]float64{-84.396535, 33.792578}
		_, distance, err := geopath.FindShortestPath(paths, start, end, 0.00001)
		if err != nil {
			t.Errorf("returned an error %s", err)
		}
		if distance < 66 || distance > 67 {
			t.Errorf("Distance wasn't correct, had %f meters, should be ~66.265", distance)
		}
	})
	t.Run("Test FindPath Large Dataset", func(t *testing.T) {
		f, err := os.Open("testdata/sample_large.json")
		if err != nil {
			t.Errorf("failed to open file, %v", err)
		}
		paths, err := geopath.ParsePaths(f)
		if err != nil {
			t.Errorf("failed to parse paths, %v", err)
		}
		start := [2]float64{-84.397252, 33.792997}
		end := [2]float64{-84.395111, 33.791666}
		precision := 0.00001
		_, distance, err := geopath.FindShortestPath(paths, start, end, precision)
		if err != nil {
			t.Errorf("returned an error %s", err)
		}
		if distance < 365 || distance > 366 {
			t.Errorf("Distance wasn't correct, had %f meters, should be ~66.265", distance)
		}
	})
}

func BenchmarkFindPath(b *testing.B) {
	f, err := os.OpenFile("testdata/sample_large.json", os.O_RDONLY, 0444)
	if err != nil {
		b.Fail()
	}
	paths, err := geopath.ParsePaths(f)
	if err != nil {
		b.Fail()
	}
	var path [][2]float64
	var distance float64
	start := [2]float64{-84.397252, 33.792997}
	end := [2]float64{-84.395111, 33.791666}
	precision := 0.00001
	for i := 0; i < b.N; i++ {
		path, distance, _ = geopath.FindShortestPath(paths, start, end, precision)
	}
	_ = path
	_ = distance
}
