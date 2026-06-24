package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Updating spatial_target_type constraints...")
		db := app.DB()

		// =====================================================================
		// 1. DATA NORMALIZATION (ledger_material_inputs)
		// =====================================================================
		// We must sanitize existing string data before enforcing the SelectField enum,
		// otherwise PocketBase will throw a validation error on app.Save().
		log.Println("[Migration] 1/3: Normalizing legacy string data...")

		normalizeQuery := `
			UPDATE ledger_material_inputs
			SET spatial_target_type = CASE
				WHEN spatial_target_type IN ('plot', 'PLOT', 'plots') THEN 'plots'
				WHEN spatial_target_type IN ('work_area', 'WORK_AREA', 'work_areas') THEN 'work_areas'
				WHEN spatial_target_type IN ('pasture', 'PASTURE', 'pastures') THEN 'pastures'
				ELSE ''
			END
			WHERE spatial_target_type != '' AND spatial_target_type IS NOT NULL;
		`
		if _, err := db.NewQuery(normalizeQuery).Execute(); err != nil {
			return fmt.Errorf("failed to normalize spatial_target_type data: %w", err)
		}

		// =====================================================================
		// 2. UPDATE ledger_material_inputs
		// =====================================================================
		log.Println("[Migration] 2/3: Converting spatial_target_type to SelectField...")

		matInputs, err := app.FindCollectionByNameOrId("ledger_material_inputs")
		if err != nil {
			return fmt.Errorf("failed to find 'ledger_material_inputs': %w", err)
		}

		// Safest way to change field type in PB: Remove old field and add new one with the same name.
		// SQLite will retain the underlying column data.
		matInputs.Fields.RemoveByName("spatial_target_type")
		matInputs.Fields.Add(&core.SelectField{
			Name:   "spatial_target_type",
			Values: []string{"plots", "work_areas", "pastures"},
		})

		if err := app.Save(matInputs); err != nil {
			return fmt.Errorf("failed saving ledger_material_inputs: %w", err)
		}

		// =====================================================================
		// 3. UPDATE ledger_field_labor
		// =====================================================================
		log.Println("[Migration] 3/3: Adding spatial_target_type to ledger_field_labor...")

		fieldLabor, err := app.FindCollectionByNameOrId("ledger_field_labor")
		if err != nil {
			return fmt.Errorf("failed to find 'ledger_field_labor': %w", err)
		}

		if fieldLabor.Fields.GetByName("spatial_target_type") == nil {
			fieldLabor.Fields.Add(&core.SelectField{
				Name:   "spatial_target_type",
				Values: []string{"plots", "work_areas", "pastures"},
			})

			if err := app.Save(fieldLabor); err != nil {
				return fmt.Errorf("failed saving ledger_field_labor: %w", err)
			}
		}

		log.Println("[Migration] Updates applied successfully!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Rolling back spatial_target_type changes...")

		// 1. Revert ledger_material_inputs back to TextField
		matInputs, err := app.FindCollectionByNameOrId("ledger_material_inputs")
		if err == nil {
			matInputs.Fields.RemoveByName("spatial_target_type")
			matInputs.Fields.Add(&core.TextField{
				Name: "spatial_target_type",
			})
			_ = app.Save(matInputs)
		}

		// 2. Remove field from ledger_field_labor
		fieldLabor, err := app.FindCollectionByNameOrId("ledger_field_labor")
		if err == nil {
			fieldLabor.Fields.RemoveByName("spatial_target_type")
			_ = app.Save(fieldLabor)
		}

		log.Println("[Migration Rollback] Rollback complete.")
		return nil
	})
}
