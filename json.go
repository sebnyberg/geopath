package geopath

import (
	"encoding/json"
	"fmt"
	"io"
)

// ParsePaths parses lines found within the provided GeoJSON. Any
// non line-strings are ignored. An error is returned if the file does not
// contain valid GeoJSON.
func ParsePaths(file io.Reader) ([]path, error) {
	var data geoJSON
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}
	if data.Typ != geoJSONTypeFeatureCollection {
		return nil, fmt.Errorf(
			"invalid root geojson type %v, expected %v",
			data.Typ, geoJSONTypeFeatureCollection,
		)
	}
	res := make([]path, 0, len(data.Features))
	for i, f := range data.Features {
		if f.Typ != geoJSONFeatureTypeFeature {
			return nil, fmt.Errorf(
				"invalid geojson feature type %v, expected %v",
				f.Typ, geoJSONFeatureTypeFeature,
			)
		}
		if f.Geom.Typ != geoJSONGeometryTypeLineString {
			continue
		}
		if len(f.Geom.Coordinates) != 2 {
			return nil, fmt.Errorf(
				"LineString geom for feature #%v did not have two coordinates", i)
		}
		p := path{
			f.Geom.Coordinates[0],
			f.Geom.Coordinates[1],
		}
		res = append(res, p)
	}
	return res, nil
}

const (
	geoJSONTypeFeatureCollection  = "FeatureCollection"
	geoJSONFeatureTypeFeature     = "Feature"
	geoJSONGeometryTypeLineString = "LineString"
)

type geoJSON struct {
	Typ      string           `json:"type"`
	Features []geoJSONFeature `json:"features"`
}

type geoJSONFeature struct {
	Typ  string          `json:"type"`
	Geom geoJSONGeometry `json:"geometry"`
}

type geoJSONGeometry struct {
	Coordinates [][2]float64 `json:"coordinates"`
	Typ         string       `json:"type"`
}
