package migrations

import (
	"fmt"

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
		setCollectionAuthRulesBackfill(tractsCol, apiRule)
		if err := app.Save(tractsCol); err != nil {
			return err
		}

		plotsCol, err := app.FindCollectionByNameOrId("plots")
		if err != nil {
			return err
		}
		setCollectionAuthRulesBackfill(plotsCol, apiRule)
		if err := app.Save(plotsCol); err != nil {
			return err
		}

		workAreasCol, err := app.FindCollectionByNameOrId("work_areas")
		if err != nil {
			return err
		}
		setCollectionAuthRulesBackfill(workAreasCol, apiRule)
		if err := app.Save(workAreasCol); err != nil {
			return err
		}

		infrastructuresCol, err := findCollectionByAnyNameBackfill(app, "infrastructures", "infrastructure")
		if err != nil {
			return err
		}
		setCollectionAuthRulesBackfill(infrastructuresCol, apiRule)
		if err := app.Save(infrastructuresCol); err != nil {
			return err
		}

		spatialFeaturesCol, err := app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return err
		}
		setCollectionAuthRulesBackfill(spatialFeaturesCol, apiRule)
		if err := app.Save(spatialFeaturesCol); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}

func setCollectionAuthRulesBackfill(col *core.Collection, rule string) {
	col.ListRule = &rule
	col.ViewRule = &rule
	col.CreateRule = &rule
	col.UpdateRule = &rule
	col.DeleteRule = &rule
}

func findCollectionByAnyNameBackfill(app core.App, names ...string) (*core.Collection, error) {
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