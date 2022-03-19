package geopath_test

import (
	"os"
	"testing"

	"github.com/sebnyberg/geopath"
)

func TestParseLines(t *testing.T) {
	f, err := os.Open("testdata/sample.json")
	if err != nil {
		t.Errorf("failed to open sample GeoJSON, %v", err)
	}
	_, err = geopath.ParsePaths(f)
	if err != nil {
		t.Errorf("unexpected error parsing sample file, %v", err)
	}
}
