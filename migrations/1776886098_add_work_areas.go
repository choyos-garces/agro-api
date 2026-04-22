package migrations

import (
	"github.com/choyos-garces/agro-api/internal/schema"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		apiRule := "@request.auth.id != ''"

		tractsCol, err := app.FindCollectionByNameOrId("tracts")
		if err != nil {
			return err
		}

		// Create the "work_areas" collection
		workAreaCol := core.NewBaseCollection("work_areas")

		workAreaCol.Fields.Add(&core.RelationField{
			Name:         "tract_id",
			Required:     true,
			CollectionId: tractsCol.Id,
			MaxSelect:    1,
		})

		workAreaCol.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
		})

		workAreaCol.Fields.Add(&core.SelectField{
			Name:      "category",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"healthy", "needs_attention", "diseased", "treatment_active", "fallow"},
		})

		workAreaCol.Fields.Add(schema.GeoJSONField("geometry"))

		workAreaCol.Fields.Add(&core.BoolField{
			Name: "active",
		})

		workAreaCol.ListRule = types.Pointer(apiRule)
		workAreaCol.ViewRule = types.Pointer(apiRule)
		workAreaCol.CreateRule = types.Pointer(apiRule)
		workAreaCol.UpdateRule = types.Pointer(apiRule)
		workAreaCol.DeleteRule = types.Pointer(apiRule)

		if err := app.Save(workAreaCol); err != nil {
			return err
		}

		// Edit the "plots" collection
		plotsCol, err := app.FindCollectionByNameOrId("plots")
		if err != nil {
			return err
		}

		plotsCol.Fields.RemoveByName("name")
		plotsCol.Fields.RemoveByName("category")
		plotsCol.Fields.RemoveByName("status")
		plotsCol.Fields.RemoveByName("condition")
		plotsCol.Fields.Add(&core.BoolField{
			Name: "active",
		})

		return app.Save(plotsCol)
	}, func(app core.App) error {
		workAreaCol, err := app.FindCollectionByNameOrId("work_areas")
		if err == nil {
			if err := app.Delete(workAreaCol); err != nil {
				return err
			}
		}

		plotsCol, err := app.FindCollectionByNameOrId("plots")
		if err != nil {
			return err
		}

		plotsCol.Fields.RemoveByName("active")

		plotsCol.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
		})

		plotsCol.Fields.Add(&core.SelectField{
			Name:      "category",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"permanent", "temporary"},
		})

		plotsCol.Fields.Add(&core.SelectField{
			Name:      "status",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"active", "archived"},
		})

		plotsCol.Fields.Add(&core.SelectField{
			Name:      "condition",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"healthy", "needs_attention", "diseased", "treatment_active", "fallow"},
		})

		return app.Save(plotsCol)
	})
}
