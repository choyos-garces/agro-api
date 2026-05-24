package migrations

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		trucks, err := app.FindCollectionByNameOrId("trucks")
		if err != nil {
			return fmt.Errorf("failed to find 'trucks' collection: %w", err)
		}

		trucks.Indexes = append(trucks.Indexes, "CREATE UNIQUE INDEX `idx_trucks_license_plate` ON `trucks` (`license_plate`)")

		if err := app.Save(trucks); err != nil {
			return fmt.Errorf("failed saving trucks index: %w", err)
		}

		if plateField, ok := trucks.Fields.GetByName("license_plate").(*core.TextField); ok {
			plateField.Pattern = `^[A-Z]{3}-\d{4}$`
		} else {
			return fmt.Errorf("could not cast license_plate to TextField")
		}

		if err := app.Save(trucks); err != nil {
			return fmt.Errorf("failed saving trucks regex: %w", err)
		}		

		return nil
	}, func(app core.App) error {
		trucks, err := app.FindCollectionByNameOrId("trucks")
		if err != nil {
			return err
		}

		// FIX: Iterate through the slice and keep everything EXCEPT the index we want to remove
		var newIndexes []string
		for _, idx := range trucks.Indexes {
			if !strings.Contains(idx, "`idx_trucks_license_plate`") {
				newIndexes = append(newIndexes, idx)
			}
		}
		
		// Overwrite the indexes slice with our filtered one
		trucks.Indexes = newIndexes

		if err := app.Save(trucks); err != nil {
			return err
		}

		if plateField, ok := trucks.Fields.GetByName("license_plate").(*core.TextField); ok {
			plateField.Pattern = ""
		}

		if err := app.Save(trucks); err != nil {
			return err
		}

		return nil
	})
}