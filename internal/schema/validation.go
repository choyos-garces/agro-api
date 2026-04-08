package schema

import (
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb/geojson"
)

type geoJSONTypeProbe struct {
	Type string `json:"type"`
}

func ValidateGeoJSON(rawJSON string) error {
	if rawJSON == "" {
		return nil
	}

	rawData := []byte(rawJSON)

	var probe geoJSONTypeProbe
	if err := json.Unmarshal(rawData, &probe); err != nil {
		return fmt.Errorf("invalid GeoJSON: data is not a valid JSON object")
	}

	if probe.Type == "" {
		return fmt.Errorf("invalid GeoJSON: missing 'type' property")
	}

	switch probe.Type {
	case "Feature":
		if _, err := geojson.UnmarshalFeature(rawData); err != nil {
			return normalizeGeoJSONError(err)
		}
	case "FeatureCollection":
		if _, err := geojson.UnmarshalFeatureCollection(rawData); err != nil {
			return normalizeGeoJSONError(err)
		}
	default:
		if _, err := geojson.UnmarshalGeometry(rawData); err != nil {
			return normalizeGeoJSONError(err)
		}
	}

	return nil
}

func normalizeGeoJSONError(err error) error {
	return fmt.Errorf("invalid GeoJSON: %v", err)
}
