package hooks

import (
	"github.com/choyos-garces/agro-api/internal/schema"
	"github.com/pocketbase/pocketbase/core"
)

func Register(app core.App) {

	geoCollections := []string{"tracts"} // Add more collection names here as needed

	// Bind directly to the app instance!
	app.OnRecordCreateRequest(geoCollections...).BindFunc(validateGeometryHook)
	app.OnRecordUpdateRequest(geoCollections...).BindFunc(validateGeometryHook)
}

func validateGeometryHook(e *core.RecordRequestEvent) error {
	geometryJson := e.Record.GetString("geometry")

	if geometryJson != "" {
		if err := schema.ValidateGeoJSON(geometryJson); err != nil {
			return e.BadRequestError("Invalid geometry data: "+err.Error(), nil)
		}
	}

	return e.Next()
}
