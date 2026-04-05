package schema

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GeoJSON represents the strict shape we expect from the frontend.
// It acts just like a TypeScript interface.
type GeoJSON struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"` // interface{} means "any type" (nested arrays)
}

// ValidateGeoJSON checks if a string is a valid GeoJSON object.
func ValidateGeoJSON(rawJson string) error {
	if rawJson == "" {
		return nil
	}

	var geo GeoJSON
	if err := json.Unmarshal([]byte(rawJson), &geo); err != nil {
		return errors.New("data is not a valid JSON object")
	}

	// 1. Validate the 'type' (The Go equivalent of a TS Union Type)
	switch geo.Type {
	case "Point", "LineString", "Polygon", "MultiPoint", "MultiLineString", "MultiPolygon":
		// It's valid, do nothing and continue!
	default:
		// fmt.Errorf allows us to inject variables into strings, just like template literals (`...`) in JS
		return fmt.Errorf("invalid type '%s'. Must be Point, LineString, Polygon, MultiPoint, MultiLineString, or MultiPolygon", geo.Type)
	}

	// 2. Validate 'coordinates' exists
	if geo.Coordinates == nil {
		return errors.New("missing 'coordinates' property")
	}

	// 3. Validate 'coordinates' is actually an array (any[])
	// This is a "Type Assertion" in Go. It checks if the dynamic interface{}
	// is actually a JSON array under the hood.
	if _, isArray := geo.Coordinates.([]interface{}); !isArray {
		return errors.New("'coordinates' must be an array")
	}

	return nil
}
