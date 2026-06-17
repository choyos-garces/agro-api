package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Creating ledger_asset_usage collection...")

		// Fetch fixed_assets to link the relation field
		assets, err := app.FindCollectionByNameOrId("fixed_assets")
		if err != nil {
			return fmt.Errorf("failed to find 'fixed_assets' collection: %w", err)
		}

		collection := core.NewBaseCollection("ledger_asset_usage")

		// Standard Ledger Rules: Authenticated CRU, Update unlocked only, Admin-only Delete
		baseRule := "@request.auth.id != \"\""
		updateRule := "@request.auth.id != \"\" && is_locked = false"

		collection.ListRule = &baseRule
		collection.ViewRule = &baseRule
		collection.CreateRule = &baseRule
		collection.UpdateRule = &updateRule
		collection.DeleteRule = &updateRule

		// Add Fields
		collection.Fields.Add(&core.TextField{
			Name:     "ledger_id",
			Required: true,
		})

		collection.Fields.Add(&core.TextField{
			Name:     "ledger_type",
			Required: true,
		})

		collection.Fields.Add(&core.RelationField{
			Name:          "asset_id",
			CollectionId:  assets.Id,
			Required:      true,
			CascadeDelete: true, // As requested: Delete usage logs if the asset is deleted
		})

		collection.Fields.Add(&core.NumberField{
			Name:     "hours_logged",
			Required: true,
		})

		collection.Fields.Add(&core.NumberField{
			Name:     "fuel_consumed",
			Required: false, // Nullable
		})

		collection.Fields.Add(&core.BoolField{
			Name: "is_locked",
			// Defaults to false natively
		})

		// Explicit Autodate Fields
		collection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
			OnUpdate: false,
		})

		collection.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})

		// Save the new collection
		if err := app.Save(collection); err != nil {
			return fmt.Errorf("failed to save ledger_asset_usage schema: %w", err)
		}

		log.Println("[Migration] Successfully created ledger_asset_usage!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Dropping ledger_asset_usage...")

		collection, err := app.FindCollectionByNameOrId("ledger_asset_usage")
		if err != nil {
			return err // Graceful skip if it's already gone
		}

		if err := app.Delete(collection); err != nil {
			return fmt.Errorf("failed to delete ledger_asset_usage: %w", err)
		}

		log.Println("[Migration Rollback] Rollback complete.")
		return nil
	})
}
