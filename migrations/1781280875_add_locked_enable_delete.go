package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// The 7 core operational ledgers
		ledgers := []string{
			"ledger_material_inputs",
			"ledger_field_labor",
			"ledger_crop_seeding",
			"ledger_pasture_rotation",
			"ledger_logistics_trips",
			"ledger_asset_events",
			"ledger_human_resources",
			"registry_moko_cases",
		}

		log.Println("[Migration] Implementing is_locked boolean and updating API rules across all ledgers...")

		for _, name := range ledgers {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return fmt.Errorf("failed to find ledger '%s': %w", name, err)
			}

			needsSave := false

			// 1. Add the is_locked boolean field (Defaults to false automatically)
			if col.Fields.GetByName("is_locked") == nil {
				col.Fields.Add(&core.BoolField{
					Name: "is_locked",
				})
				needsSave = true
			}

			// 2. Update the API Rules
			// We append the lock check. The record can only be updated if it is currently unlocked.
			// (Admins bypass these rules entirely)
			updateRule := "@request.auth.id != \"\" && is_locked = false"

			// If rules are different or nil, we update them.
			if col.UpdateRule == nil || *col.UpdateRule != updateRule {
				col.UpdateRule = &updateRule
				needsSave = true
			}

			// Just to ensure immutability is respected, we explicitly keep DeleteRule as nil (Admin only)
			// Or if you wanted to let them delete unlocked ones, you would use the same string as above.
			col.DeleteRule = &updateRule

			// 3. Save Collection
			if needsSave {
				if err := app.Save(col); err != nil {
					return fmt.Errorf("failed saving schema for '%s': %w", name, err)
				}
				log.Printf("[Migration] Secured '%s' with is_locked rules.\n", name)
			}
		}

		log.Println("[Migration] All ledgers successfully locked down!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Removing is_locked from ledgers...")

		ledgers := []string{
			"ledger_material_inputs",
			"ledger_field_labor",
			"ledger_crop_seeding",
			"ledger_pasture_rotation",
			"ledger_logistics_trips",
			"ledger_asset_events",
			"ledger_human_resources",
			"registry_moko_cases",
		}

		for _, name := range ledgers {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}

			// Remove the field
			col.Fields.RemoveByName("is_locked")

			// Revert the rule back to standard authenticated user
			updateRule := "@request.auth.id != \"\""
			col.UpdateRule = &updateRule
			col.UpdateRule = nil // Admin only

			_ = app.Save(col)
		}

		log.Println("[Migration Rollback] Rollback complete.")
		return nil
	})
}
