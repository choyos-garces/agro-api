package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collectionsToUpdate := []string{
			"field_observations",
			"fixed_assets",
			"ledger_crop_seeding",
			"personnel",
			"personnel_qualifications",
		}

		log.Println("[Migration] Starting missing autodate fields patch...")

		for _, name := range collectionsToUpdate {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return fmt.Errorf("failed to find collection '%s': %w", name, err)
			}

			needsSave := false

			// 1. Check and add 'created' field
			if col.Fields.GetByName("created") == nil {
				col.Fields.Add(&core.AutodateField{
					Name:     "created",
					OnCreate: true,
					OnUpdate: false,
				})
				needsSave = true
			}

			// 2. Check and add 'updated' field
			if col.Fields.GetByName("updated") == nil {
				col.Fields.Add(&core.AutodateField{
					Name:     "updated",
					OnCreate: true,
					OnUpdate: true,
				})
				needsSave = true
			}

			// 3. Save if we made modifications
			if needsSave {
				if err := app.Save(col); err != nil {
					return fmt.Errorf("failed saving autodate fields for '%s': %w", name, err)
				}
				log.Printf("[Migration] Successfully restored autodate fields to '%s'\n", name)
			} else {
				log.Printf("[Migration] '%s' already has autodate fields. Skipping...\n", name)
			}
		}

		log.Println("[Migration] Autodate patch applied successfully!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Rolling back autodate fields patch...")

		collectionsToUpdate := []string{
			"field_observations",
			"fixed_assets",
			"ledger_crop_seeding",
			"personnel",
			"personnel_qualifications",
		}

		for _, name := range collectionsToUpdate {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue // Skip gracefully if it doesn't exist
			}

			// Remove the autodate fields during rollback
			col.Fields.RemoveByName("created")
			col.Fields.RemoveByName("updated")

			_ = app.Save(col)
		}

		log.Println("[Migration Rollback] Rollback complete.")
		return nil
	})
}
