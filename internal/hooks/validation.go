package hooks

import (
	"github.com/choyos-garces/agro-api/internal/schema"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase/core"
)

func Register(app core.App) {

	geoCollections := []string{"tracts", "plots", "infrastructure"} // Add more collection names here as needed

	// Register the Data Validation
	app.OnRecordCreateRequest(geoCollections...).BindFunc(validateGeometryHook)
	app.OnRecordUpdateRequest(geoCollections...).BindFunc(validateGeometryHook)
}

func validateGeometryHook(e *core.RecordRequestEvent) error {
	geometryJson := e.Record.GetString("geometry")

	if geometryJson != "" {
		if err := schema.ValidateGeoJSON(geometryJson); err != nil {
			fieldErrors := validation.Errors{
				// NewError takes your custom code string, and the error message!
				"geometry": validation.NewError("invalid_geojson", err.Error()),
			}

			return e.BadRequestError("Failed to create record.", fieldErrors)
		}
	}

	return e.Next()
}
