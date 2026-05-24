package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collections := []string{
			"drivers",
			"ledger_asset_events",
			"ledger_crop_seeding",
			"ledger_field_labor",
			"ledger_logistics_trips",
			"ledger_material_inputs",
			"ledger_pasture_rotation",
			"logistic_locations",
			"logistic_routes",
			"pastures",
			"registry_moko_cases",
			"trucks",
		}

		for _, name := range collections {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}

			// Add 'created' field
			collection.Fields.Add(&core.AutodateField{
				Name:     "created",
				OnCreate: true,
				OnUpdate: false,
			})

			// Add 'updated' field
			collection.Fields.Add(&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			})

			if err := app.Save(collection); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		collections := []string{
			"drivers",
			"ledger_asset_events",
			"ledger_crop_seeding",
			"ledger_field_labor",
			"ledger_logistics_trips",
			"ledger_material_inputs",
			"ledger_pasture_rotation",
			"logistic_locations",
			"logistic_routes",
			"pastures",
			"registry_moko_cases",
			"trucks",
		}

		for _, name := range collections {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}

			collection.Fields.RemoveByName("created")
			collection.Fields.RemoveByName("updated")

			if err := app.Save(collection); err != nil {
				return err
			}
		}
		return nil
	})
}
