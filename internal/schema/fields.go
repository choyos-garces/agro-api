package schema

import "github.com/pocketbase/pocketbase/core"

// GeoJSONField returns a standardized JSON field for storing GeoJSON data.
// In PocketBase, complex nested objects are stored as JSON fields.
func GeoJSONField(name string) *core.JSONField {
	return &core.JSONField{
		Name:     name,
		Required: true,    // Ensures it cannot be empty
		MaxSize:  5000000, // 5MB limit for complex polygons
	}
}
