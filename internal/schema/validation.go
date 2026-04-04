package schema

import (
	"encoding/json"
	"errors"
)

// GeoJSON represents the strict shape we expect from the frontend.
// It acts just like a TypeScript interface.
type GeoJSON struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"` // interface{} means "any type" (nested arrays)
}

// ValidateGeoJSON checks if a string is a valid GeoJSON object.
func ValidateGeoJSON(rawJson string) error {
	// 1. PocketBase might pass an empty string if the field isn't required
	if rawJson == "" {
		return nil
	}

	// 2. Try to parse the string into our struct (like JSON.parse in JS)
	var geo GeoJSON
	if err := json.Unmarshal([]byte(rawJson), &geo); err != nil {
		return errors.New("data is not a valid JSON object")
	}

	// 3. Manually enforce your specific business rules
	if geo.Type != "Polygon" && geo.Type != "MultiPolygon" {
		return errors.New("geojson 'type' must be Polygon or MultiPolygon")
	}

	if geo.Coordinates == nil {
		return errors.New("geojson is missing 'coordinates'")
	}

	return nil
}
