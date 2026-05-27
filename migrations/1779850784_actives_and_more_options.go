package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Starting operational schema updates...")

		// ---------------------------------------------------------
		// 1. ADD 'active' BOOLEAN TO CORE ENTITIES
		// ---------------------------------------------------------
		collectionsWithActiveFlag := []string{
			"logistic_routes", 
			"logistic_locations", 
			"trucks", 
			"drivers",
		}

		for _, name := range collectionsWithActiveFlag {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return fmt.Errorf("failed to find '%s': %w", name, err)
			}
			col.Fields.Add(&core.BoolField{Name: "active"})
			if err := app.Save(col); err != nil {
				return fmt.Errorf("failed saving '%s' active flag: %w", name, err)
			}
		}

		// ---------------------------------------------------------
		// 2. EXPAND SELECT OPTIONS (infrastructure & work_areas)
		// ---------------------------------------------------------
		
		// Infrastructure -> type
		infra, err := app.FindCollectionByNameOrId("infrastructure")
		if err == nil {
			if typeField, ok := infra.Fields.GetByName("type").(*core.SelectField); ok {
				typeField.Values = append(typeField.Values, "packing_plant")
				if err := app.Save(infra); err != nil {
					return fmt.Errorf("failed saving infrastructure: %w", err)
				}
			}
		}

		// Work Areas -> category
		workAreas, err := app.FindCollectionByNameOrId("work_areas")
		if err == nil {
			if catField, ok := workAreas.Fields.GetByName("category").(*core.SelectField); ok {
				catField.Values = append(catField.Values, "replanted", "reseeded")
				if err := app.Save(workAreas); err != nil {
					return fmt.Errorf("failed saving work_areas: %w", err)
				}
			}
		}

		// ---------------------------------------------------------
		// 3. SPATIAL FEATURES DOMAIN REFACOR (Pluralization)
		// ---------------------------------------------------------
		spatial, err := app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return fmt.Errorf("failed to find 'spatial_features': %w", err)
		}

		// Step A: Raw SQL Data update
		// We explicitly target the exact singular strings to append 's'. 
		// 'infrastructure' is excluded naturally by not being in the IN clause.
		log.Println("[Migration] Pluralizing legacy 'domain_type' data...")
		query := `
			UPDATE spatial_features 
			SET domain_type = domain_type || 's' 
			WHERE domain_type IN ('tract', 'plot', 'pasture', 'logistic_route', 'logistic_location', 'work_area')
		`
		if _, err := app.DB().NewQuery(query).Execute(); err != nil {
			return fmt.Errorf("failed pluralizing spatial_features data: %w", err)
		}

		// Step B: Update Schema Constraints
		log.Println("[Migration] Updating spatial_features 'domain_type' options...")
		if domainField, ok := spatial.Fields.GetByName("domain_type").(*core.SelectField); ok {
			domainField.Values = []string{
				"tracts", "plots", "pastures", "logistic_routes", 
				"logistic_locations", "work_areas", "infrastructure",
			}
			if err := app.Save(spatial); err != nil {
				return fmt.Errorf("failed saving spatial_features schema: %w", err)
			}
		}

		log.Println("[Migration] Updates applied successfully!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Starting rollback...")

		// 1. Remove 'active' boolean
		collectionsWithActiveFlag := []string{
			"logistic_routes", "logistic_locations", "trucks", "drivers",
		}
		for _, name := range collectionsWithActiveFlag {
			col, err := app.FindCollectionByNameOrId(name)
			if err == nil {
				col.Fields.RemoveByName("active")
				_ = app.Save(col)
			}
		}

		// 2. Revert Infrastructure options
		infra, err := app.FindCollectionByNameOrId("infrastructure")
		if err == nil {
			if typeField, ok := infra.Fields.GetByName("type").(*core.SelectField); ok {
				var reverted []string
				for _, val := range typeField.Values {
					if val != "packing_plant" {
						reverted = append(reverted, val)
					}
				}
				typeField.Values = reverted
				_ = app.Save(infra)
			}
		}

		// 3. Revert Work Areas options
		workAreas, err := app.FindCollectionByNameOrId("work_areas")
		if err == nil {
			if catField, ok := workAreas.Fields.GetByName("category").(*core.SelectField); ok {
				var reverted []string
				for _, val := range catField.Values {
					if val != "replanted" && val != "reseeded" {
						reverted = append(reverted, val)
					}
				}
				catField.Values = reverted
				_ = app.Save(workAreas)
			}
		}

		// 4. Revert Spatial Features
		spatial, err := app.FindCollectionByNameOrId("spatial_features")
		if err == nil {
			// Revert Data: Remove the trailing 's'
			query := `
				UPDATE spatial_features 
				SET domain_type = SUBSTR(domain_type, 1, LENGTH(domain_type) - 1) 
				WHERE domain_type IN ('tracts', 'plots', 'pastures', 'logistic_routes', 'logistic_locations', 'work_areas')
			`
			_, _ = app.DB().NewQuery(query).Execute()

			// Revert Schema Options
			if domainField, ok := spatial.Fields.GetByName("domain_type").(*core.SelectField); ok {
				domainField.Values = []string{
					"tract", "plot", "pasture", "logistic_route", 
					"logistic_location", "work_area", "infrastructure",
				}
				_ = app.Save(spatial)
			}
		}

		log.Println("[Migration Rollback] Rollback complete.")
		return nil
	})
}