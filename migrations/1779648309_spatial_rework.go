package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Starting spatial_features update and data backfill...")

		// 1. Fetch the collection
		spatial, err := app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return fmt.Errorf("failed to find 'spatial_features': %w", err)
		}

		// 2. Structural changes: Add domain_id and remove tract_id
		log.Println("[Migration] Adding 'domain_id' and removing 'tract_id'...")
		spatial.Fields.RemoveByName("tract_id")
		spatial.Fields.Add(&core.TextField{
			Name: "domain_id",
			// PocketBase IDs are exactly 15 alphanumeric characters
			Pattern: `^[a-zA-Z0-9]{15}$`, 
		})

		// Save schema first to physically create the 'domain_id' column in the database
		if err := app.Save(spatial); err != nil {
			return fmt.Errorf("failed initial schema save: %w", err)
		}

		// 3. Backfill data via Raw SQL
		log.Println("[Migration] Backfilling 'domain_id' and mapping feature types...")
		
		mappings := []struct {
			OldType    string
			Collection string
			NewType    string
		}{
			{"tract_boundary", "tracts", "tract"},
			{"work_area", "work_areas", "work_area"},
			{"infrastructure", "infrastructure", "infrastructure"},
			{"plot", "plots", "plot"},
			{"pasture", "pastures", "pasture"},
			{"logistic_location", "logistic_locations", "logistic_location"},
		}

		for _, m := range mappings {
			// Subquery grabs the ID from the related collection where spatial_feature_id matches.
			// COALESCE ensures it defaults to an empty string instead of NULL if no relation is found.
			// We check both OldType and NewType (and the legacy 'tract_boundry' spelling) to be safe.
			query := fmt.Sprintf(`
				UPDATE spatial_features 
				SET domain_id = COALESCE((SELECT id FROM %s WHERE spatial_feature_id = spatial_features.id LIMIT 1), ''),
					feature_type = '%s'
				WHERE feature_type = '%s' OR feature_type = '%s' OR feature_type = 'tract_boundry'
			`, m.Collection, m.NewType, m.OldType, m.NewType)
			
			_, err := app.DB().NewQuery(query).Execute()
			if err != nil {
				return fmt.Errorf("failed mapping %s data: %w", m.OldType, err)
			}
		}

		// 4. Update Schema: Rename feature_type to domain_type and enforce Select rules
		// Re-fetch the collection to ensure we have the latest state before saving again
		spatial, err = app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return err
		}

		field := spatial.Fields.GetByName("feature_type")
		if field == nil {
			return fmt.Errorf("could not find 'feature_type' field")
		}

		selectField, ok := field.(*core.SelectField)
		if !ok {
			return fmt.Errorf("field 'feature_type' is not a SelectField")
		}

		log.Println("[Migration] Renaming field to 'domain_type' and applying constraints...")
		selectField.Name = "domain_type"
		selectField.Values = []string{
			"tract", "plot", "pasture", "logistic_route", 
			"logistic_location", "work_area", "infrastructure",
		}

		// 5. Final Schema Save
		if err := app.Save(spatial); err != nil {
			return fmt.Errorf("failed final saving spatial_features schema: %w", err)
		}

		log.Println("[Migration] Successfully updated spatial_features!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Rolling back spatial_features schema...")
		spatial, err := app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return err
		}

		// 1. Revert domain_type back to feature_type
		field := spatial.Fields.GetByName("domain_type")
		if field != nil {
			if selectField, ok := field.(*core.SelectField); ok {
				selectField.Name = "feature_type"
				selectField.Values = []string{
					"tract", "tract_boundary", "plot", "pasture", 
					"logistic_route", "logistic_location", "work_area", "infrastructure",
				}
			}
		}

		// 2. Remove domain_id
		spatial.Fields.RemoveByName("domain_id")

		// 3. Restore tract_id relation
		spatial.Fields.Add(&core.RelationField{
			Name: "tract_id", 
			// Fallback placeholder relation to keep the compiler happy
			CollectionId: spatial.Id, 
		})

		if err := app.Save(spatial); err != nil {
			return err
		}
		
		log.Println("[Migration Rollback] Rollback complete.")
		return nil
	})
}