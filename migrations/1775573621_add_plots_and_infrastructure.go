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

		plotsCol := core.NewBaseCollection("plots")

		plotsCol.Fields.Add(&core.RelationField{
			Name:         "tract_id",
			Required:     true,
			CollectionId: tractsCol.Id,
			MaxSelect:    1,
		})

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

		plotsCol.Fields.Add(&core.TextField{
			Name: "crop_type",
		})

		plotsCol.Fields.Add(schema.GeoJSONField("geometry"))

		infrastructureCol := core.NewBaseCollection("infrastructure")

		infrastructureCol.Fields.Add(&core.RelationField{
			Name:         "tract_id",
			Required:     true,
			CollectionId: tractsCol.Id,
			MaxSelect:    1,
		})

		infrastructureCol.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
		})

		infrastructureCol.Fields.Add(&core.SelectField{
			Name:      "type",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"building", "water_body", "road", "facility", "parking", "eating_area"},
		})

		infrastructureCol.Fields.Add(schema.GeoJSONField("geometry"))

		plotsCol.ListRule = types.Pointer(apiRule)
		plotsCol.ViewRule = types.Pointer(apiRule)
		plotsCol.CreateRule = types.Pointer(apiRule)
		plotsCol.UpdateRule = types.Pointer(apiRule)
		plotsCol.DeleteRule = types.Pointer(apiRule)

		infrastructureCol.ListRule = types.Pointer(apiRule)
		infrastructureCol.ViewRule = types.Pointer(apiRule)
		infrastructureCol.CreateRule = types.Pointer(apiRule)
		infrastructureCol.UpdateRule = types.Pointer(apiRule)
		infrastructureCol.DeleteRule = types.Pointer(apiRule)

		if err := app.Save(plotsCol); err != nil {
			return err
		}

		if err := app.Save(infrastructureCol); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		plotsCol, err := app.FindCollectionByNameOrId("plots")
		if err == nil {
			if err := app.Delete(plotsCol); err != nil {
				return err
			}
		}

		infrastructureCol, err := app.FindCollectionByNameOrId("infrastructure")
		if err == nil {
			if err := app.Delete(infrastructureCol); err != nil {
				return err
			}
		}

		return nil
	})
}
