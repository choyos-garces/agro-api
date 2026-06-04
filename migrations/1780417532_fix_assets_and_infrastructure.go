package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Starting Blueprint Patch (Fixed Assets & Infrastructures)...")

		db := app.DB()

		// =====================================================================
		// 1. DATA NORMALIZATION (Raw SQL)
		// =====================================================================
		log.Println("[Migration] 1/3: Normalizing data for new enum constraints...")

		// Map legacy fixed_assets enums to new values
		db.NewQuery("UPDATE fixed_assets SET asset_category = 'HEAVY_MACHINERY' WHERE asset_category = 'MACHINERY'").Execute()
		db.NewQuery("UPDATE fixed_assets SET status = 'IN_REPAIR' WHERE status = 'IN_SHOP'").Execute()
		db.NewQuery("UPDATE fixed_assets SET status = 'RETIRED' WHERE status = 'OUT_OF_SERVICE'").Execute()

		// Map spatial_features domain_type to plural
		db.NewQuery("UPDATE spatial_features SET domain_type = 'infrastructures' WHERE domain_type = 'infrastructure'").Execute()

		// =====================================================================
		// 2. FIXED ASSETS OVERHAUL
		// =====================================================================
		log.Println("[Migration] 2/3: Applying fixed_assets schema updates...")
		if assets, err := app.FindCollectionByNameOrId("fixed_assets"); err == nil {

			// Update asset_category options
			if f, ok := assets.Fields.GetByName("asset_category").(*core.SelectField); ok {
				f.Values = []string{"VEHICLE", "HEAVY_MACHINERY", "IMPLEMENT", "POWER_TOOL", "FACILITY_PUMP"}
			}

			// Update status options
			if f, ok := assets.Fields.GetByName("status").(*core.SelectField); ok {
				f.Values = []string{"OPERATIONAL", "IN_REPAIR", "RETIRED"}
			}

			// Add new fields
			assets.Fields.Add(&core.NumberField{Name: "capacity_value"})
			assets.Fields.Add(&core.TextField{Name: "capacity_unit"})
			assets.Fields.Add(&core.SelectField{Name: "fuel_type", Values: []string{"DIESEL", "GASOLINE", "ELECTRIC", "NONE"}})
			assets.Fields.Add(&core.NumberField{Name: "model_year", Required: true})
			assets.Fields.Add(&core.NumberField{Name: "purchase_year", Required: true})
			// Soft link to spatial_features
			assets.Fields.Add(&core.TextField{Name: "home_base_id", Pattern: `^[a-zA-Z0-9]{15}$`})

			if err := app.Save(assets); err != nil {
				return fmt.Errorf("failed saving fixed_assets updates: %w", err)
			}
		}

		// =====================================================================
		// 3. INFRASTRUCTURE(S) RENAME & SPATIAL MAPPING
		// =====================================================================
		log.Println("[Migration] 3/3: Renaming infrastructure collection...")

		// Rename Collection
		if infra, err := app.FindCollectionByNameOrId("infrastructure"); err == nil {
			infra.Name = "infrastructures"
			if err := app.Save(infra); err != nil {
				return fmt.Errorf("failed renaming infrastructures: %w", err)
			}
		}

		// Update spatial_features domain_type options
		if sf, err := app.FindCollectionByNameOrId("spatial_features"); err == nil {
			if f, ok := sf.Fields.GetByName("domain_type").(*core.SelectField); ok {
				f.Values = []string{
					"tracts", "plots", "pastures", "logistic_routes",
					"logistic_locations", "work_areas", "infrastructures",
				}
			}
			if err := app.Save(sf); err != nil {
				return fmt.Errorf("failed updating spatial_features: %w", err)
			}
		}

		log.Println("[Migration] Patch applied successfully!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Rolling back blueprint patch...")
		db := app.DB()

		// 1. Revert infrastructures rename
		if infra, err := app.FindCollectionByNameOrId("infrastructures"); err == nil {
			infra.Name = "infrastructure"
			_ = app.Save(infra)
		}

		// 2. Revert spatial_features
		if sf, err := app.FindCollectionByNameOrId("spatial_features"); err == nil {
			if f, ok := sf.Fields.GetByName("domain_type").(*core.SelectField); ok {
				f.Values = []string{
					"tracts", "plots", "pastures", "logistic_routes",
					"logistic_locations", "work_areas", "infrastructure",
				}
			}
			_ = app.Save(sf)
			// Revert data mapping
			db.NewQuery("UPDATE spatial_features SET domain_type = 'infrastructure' WHERE domain_type = 'infrastructures'").Execute()
		}

		// 3. Revert fixed_assets fields and options
		if assets, err := app.FindCollectionByNameOrId("fixed_assets"); err == nil {
			assets.Fields.RemoveByName("capacity_value")
			assets.Fields.RemoveByName("capacity_unit")
			assets.Fields.RemoveByName("fuel_type")
			assets.Fields.RemoveByName("model_year")
			assets.Fields.RemoveByName("purchase_year")
			assets.Fields.RemoveByName("home_base_id")

			if f, ok := assets.Fields.GetByName("asset_category").(*core.SelectField); ok {
				f.Values = []string{"VEHICLE", "MACHINERY", "INFRASTRUCTURE"}
			}
			if f, ok := assets.Fields.GetByName("status").(*core.SelectField); ok {
				f.Values = []string{"OPERATIONAL", "IN_SHOP", "OUT_OF_SERVICE"}
			}
			_ = app.Save(assets)

			// Revert data mapping
			db.NewQuery("UPDATE fixed_assets SET asset_category = 'MACHINERY' WHERE asset_category = 'HEAVY_MACHINERY'").Execute()
			db.NewQuery("UPDATE fixed_assets SET status = 'IN_SHOP' WHERE status = 'IN_REPAIR'").Execute()
			db.NewQuery("UPDATE fixed_assets SET status = 'OUT_OF_SERVICE' WHERE status = 'RETIRED'").Execute()
		}

		return nil
	})
}
