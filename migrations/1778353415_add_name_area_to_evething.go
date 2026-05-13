package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		plotsCol, err := app.FindCollectionByNameOrId("plots")
		if err != nil {
			return err
		}

		// Add "name" field to "plots" collection
		plotsCol.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Max:      255,
		})

		// Add "area" field to "plots" collection
		minArea := 0.1
		plotsCol.Fields.Add(&core.NumberField{
			Name:     "area",
			Required: true,
			Min:      &minArea,
		})

		// Add "active" field to "plots" collection
		plotsCol.Fields.Add(&core.BoolField{
			Name: "active",
		})

		infrastructureCol, err := app.FindCollectionByNameOrId("infrastructure")
		if err != nil {
			return err
		}

		// Add "area" field to "infrastructure" collection
		infrastructureCol.Fields.Add(&core.NumberField{
			Name:     "area",
			Required: true,
			Min:      &minArea,
		})

		// Add "active" field to "infrastructure" collection
		infrastructureCol.Fields.Add(&core.BoolField{
			Name: "active",
		})

		workAreaCol, err := app.FindCollectionByNameOrId("work_areas")
		if err != nil {
			return err
		}

		// Add "area" field to "work_areas" collection
		workAreaCol.Fields.Add(&core.NumberField{
			Name:     "area",
			Required: true,
			Min:      &minArea,
		})

		app.Save(infrastructureCol)
		app.Save(plotsCol)
		app.Save(workAreaCol)

		return nil
	}, func(app core.App) error {
		plotsCol, err := app.FindCollectionByNameOrId("plots")
		if err != nil {
			return err		
		}
		plotsCol.Fields.RemoveByName("name")
		plotsCol.Fields.RemoveByName("area")
		plotsCol.Fields.RemoveByName("active")

		infrastructureCol, err := app.FindCollectionByNameOrId("infrastructure")
		if err != nil {
			return err
		}
		infrastructureCol.Fields.RemoveByName("area")
		infrastructureCol.Fields.RemoveByName("active")

		workAreaCol, err := app.FindCollectionByNameOrId("work_areas")
		if err != nil {
			return err
		}
		workAreaCol.Fields.RemoveByName("area")

		app.Save(infrastructureCol)
		app.Save(plotsCol)
		app.Save(workAreaCol)
		return nil
	})
}
