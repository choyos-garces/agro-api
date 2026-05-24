package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		spatialFeaturesCol, err := app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return err
		}

		featureTypeField, ok := spatialFeaturesCol.Fields.GetByName("feature_type").(*core.SelectField)
		if !ok {
			return nil // if the field doesn't exist, we can skip this migration
		}

		featureTypeField.Values = []string{"tract_boundary", "plot", "work_area", "infrastructure", "pasture", "logistic_location"} 
		return app.Save(spatialFeaturesCol);
	}, func(app core.App) error {
		return nil
	})
}
