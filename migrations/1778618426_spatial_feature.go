package migrations

import (
	"fmt"

	"github.com/choyos-garces/agro-api/internal/schema"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		apiRule := "@request.auth.id != ''"

		tractsCol, err := app.FindCollectionByNameOrId("tracts")
		if err != nil {
			return err
		}
		setCollectionAuthRules(tractsCol, apiRule)

		spatialFeaturesCol := core.NewBaseCollection("spatial_features")

		spatialFeaturesCol.Fields.Add(&core.RelationField{
			Name:          "tract_id",
			Required:      true,
			CollectionId:  tractsCol.Id,
			MaxSelect:     1,
			CascadeDelete: true,
		})

		spatialFeaturesCol.Fields.Add(&core.SelectField{
			Name:      "feature_type",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"tract_boundary", "plot", "work_area", "infrastructure"},
		})

		spatialFeaturesCol.Fields.Add(&core.TextField{
			Name: "label",
		})

		spatialFeaturesCol.Fields.Add(&core.JSONField{
			Name:     "geometry",
			Required: true,
		})

		spatialFeaturesCol.Fields.Add(&core.NumberField{
			Name: "area",
		})

		spatialFeaturesCol.Fields.Add(&core.BoolField{
			Name: "archived",
		})

		spatialFeaturesCol.Fields.Add(&core.JSONField{
			Name: "import_meta",
		})
		setCollectionAuthRules(spatialFeaturesCol, apiRule)

		if err := app.Save(spatialFeaturesCol); err != nil {
			return err
		}

		tractsCol.Fields.RemoveByName("geometry")
		tractsCol.Fields.RemoveByName("area")
		tractsCol.Fields.Add(&core.RelationField{
			Name:         "spatial_feature_id",
			CollectionId: spatialFeaturesCol.Id,
			MaxSelect:    1,
		})
		if err := app.Save(tractsCol); err != nil {
			return err
		}

		plotsCol, err := app.FindCollectionByNameOrId("plots")
		if err != nil {
			return err
		}
		setCollectionAuthRules(plotsCol, apiRule)
		plotsCol.Fields.RemoveByName("geometry")
		plotsCol.Fields.RemoveByName("area")
		plotsCol.Fields.Add(&core.RelationField{
			Name:         "spatial_feature_id",
			CollectionId: spatialFeaturesCol.Id,
			MaxSelect:    1,
		})
		if err := app.Save(plotsCol); err != nil {
			return err
		}

		workAreasCol, err := app.FindCollectionByNameOrId("work_areas")
		if err != nil {
			return err
		}
		setCollectionAuthRules(workAreasCol, apiRule)
		workAreasCol.Fields.RemoveByName("geometry")
		workAreasCol.Fields.RemoveByName("area")
		workAreasCol.Fields.Add(&core.RelationField{
			Name:         "spatial_feature_id",
			CollectionId: spatialFeaturesCol.Id,
			MaxSelect:    1,
		})
		if err := app.Save(workAreasCol); err != nil {
			return err
		}

		infrastructuresCol, err := findCollectionByAnyName(app, "infrastructures", "infrastructure")
		if err != nil {
			return err
		}
		setCollectionAuthRules(infrastructuresCol, apiRule)
		infrastructuresCol.Fields.RemoveByName("geometry")
		infrastructuresCol.Fields.RemoveByName("area")
		infrastructuresCol.Fields.Add(&core.RelationField{
			Name:         "spatial_feature_id",
			CollectionId: spatialFeaturesCol.Id,
			MaxSelect:    1,
		})
		if err := app.Save(infrastructuresCol); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		tractsCol, err := app.FindCollectionByNameOrId("tracts")
		if err != nil {
			return err
		}
		tractsCol.Fields.RemoveByName("spatial_feature_id")
		minTractArea := 1.0
		tractsCol.Fields.Add(&core.NumberField{
			Name:     "area",
			Required: true,
			Min:      &minTractArea,
		})
		tractsCol.Fields.Add(schema.GeoJSONField("geometry"))
		if err := app.Save(tractsCol); err != nil {
			return err
		}

		plotsCol, err := app.FindCollectionByNameOrId("plots")
		if err != nil {
			return err
		}
		plotsCol.Fields.RemoveByName("spatial_feature_id")
		minSubArea := 0.1
		plotsCol.Fields.Add(&core.NumberField{
			Name:     "area",
			Required: true,
			Min:      &minSubArea,
		})
		plotsCol.Fields.Add(schema.GeoJSONField("geometry"))
		if err := app.Save(plotsCol); err != nil {
			return err
		}

		workAreasCol, err := app.FindCollectionByNameOrId("work_areas")
		if err != nil {
			return err
		}
		workAreasCol.Fields.RemoveByName("spatial_feature_id")
		workAreasCol.Fields.Add(&core.NumberField{
			Name:     "area",
			Required: true,
			Min:      &minSubArea,
		})
		workAreasCol.Fields.Add(schema.GeoJSONField("geometry"))
		if err := app.Save(workAreasCol); err != nil {
			return err
		}

		infrastructuresCol, err := findCollectionByAnyName(app, "infrastructures", "infrastructure")
		if err != nil {
			return err
		}
		infrastructuresCol.Fields.RemoveByName("spatial_feature_id")
		infrastructuresCol.Fields.Add(&core.NumberField{
			Name:     "area",
			Required: true,
			Min:      &minSubArea,
		})
		infrastructuresCol.Fields.Add(schema.GeoJSONField("geometry"))
		if err := app.Save(infrastructuresCol); err != nil {
			return err
		}

		spatialFeaturesCol, err := app.FindCollectionByNameOrId("spatial_features")
		if err == nil {
			if err := app.Delete(spatialFeaturesCol); err != nil {
				return err
			}
		}

		return nil
	})
}

func findCollectionByAnyName(app core.App, names ...string) (*core.Collection, error) {
	var lastErr error

	for _, name := range names {
		col, err := app.FindCollectionByNameOrId(name)
		if err == nil {
			return col, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to find any collection in %v: %w", names, lastErr)
	}

	return nil, fmt.Errorf("no collection names provided")
}

func setCollectionAuthRules(col *core.Collection, rule string) {
	col.ListRule = &rule
	col.ViewRule = &rule
	col.CreateRule = &rule
	col.UpdateRule = &rule
	col.DeleteRule = &rule
}
