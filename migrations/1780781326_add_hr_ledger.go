package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Creating ledger_human_resources collection...")

		// Fetch personnel to link the relation field
		personnel, err := app.FindCollectionByNameOrId("personnel")
		if err != nil {
			return fmt.Errorf("failed to find 'personnel' collection: %w", err)
		}

		collection := core.NewBaseCollection("ledger_human_resources")

		// Standard Ledger Rules: Authenticated CRU, Admin-only Delete
		rule := "@request.auth.id != \"\""
		collection.ListRule = &rule
		collection.ViewRule = &rule
		collection.CreateRule = &rule
		collection.UpdateRule = &rule
		collection.DeleteRule = nil // Immutable log

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
			Name:          "personnel_id",
			CollectionId:  personnel.Id,
			Required:      true,
			CascadeDelete: false, // Do not delete payroll logs if personnel is deleted
		})

		collection.Fields.Add(&core.SelectField{
			Name:     "role_played",
			Values:   []string{"EXECUTOR", "AUDITOR", "SUPERVISOR"},
			Required: true,
		})

		collection.Fields.Add(&core.NumberField{
			Name: "hours_logged",
			// Nullable natively by omitting Required: true
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
			return fmt.Errorf("failed to save ledger_human_resources schema: %w", err)
		}

		log.Println("[Migration] Successfully created ledger_human_resources!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Dropping ledger_human_resources...")

		collection, err := app.FindCollectionByNameOrId("ledger_human_resources")
		if err != nil {
			return err // Graceful skip if it's already gone
		}

		if err := app.Delete(collection); err != nil {
			return fmt.Errorf("failed to delete ledger_human_resources: %w", err)
		}

		log.Println("[Migration Rollback] Rollback complete.")
		return nil
	})
}
