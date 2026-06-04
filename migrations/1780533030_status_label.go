package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Starting work_areas and spatial_features update...")
		db := app.DB()

		// =====================================================================
		// 1. DATA BACKFILL: Set all work_areas status to ACTIVE
		// =====================================================================
		log.Println("[Migration] 1/2: Backfilling work_areas status to 'ACTIVE'...")

		_, err := db.NewQuery("UPDATE work_areas SET status = 'ACTIVE'").Execute()
		if err != nil {
			return fmt.Errorf("failed to backfill work_areas status: %w", err)
		}

		// =====================================================================
		// 2. SCHEMA UPDATE: Remove labels from spatial_features
		// =====================================================================
		log.Println("[Migration] 2/2: Removing 'labels' from spatial_features...")

		sf, err := app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return fmt.Errorf("failed to find 'spatial_features': %w", err)
		}

		sf.Fields.RemoveByName("label")

		if err := app.Save(sf); err != nil {
			return fmt.Errorf("failed saving spatial_features schema: %w", err)
		}

		log.Println("[Migration] Updates applied successfully!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Rolling back schema changes...")

		sf, err := app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return err
		}

		// Re-add the labels field
		// Note: Assuming JSONField as a standard fallback for arrays/objects of labels.
		// If it was a TextField originally, this rollback will safely recreate it as JSON to prevent data structure crashes.
		sf.Fields.Add(&core.JSONField{Name: "label"})

		if err := app.Save(sf); err != nil {
			return err
		}

		// Data Rollback Note:
		// We intentionally DO NOT rollback the work_areas status data back to NULL/Empty,
		// because we cannot distinguish between records that were already 'ACTIVE'
		// prior to this migration and those that were backfilled.

		log.Println("[Migration Rollback] Rollback complete.")
		return nil
	})
}
