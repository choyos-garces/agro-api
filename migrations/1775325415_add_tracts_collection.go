package migrations

import (
	"github.com/choyos-garces/agro-api/internal/schema"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// 1. Create a new "Base" collection named "tracts"
		collection := core.NewBaseCollection("tracts")

		// 2. Set your API Rules! (Tying back to your RBAC array logic)
		// Note: In Go, these require a pointer to a string (*string) so PocketBase
		// can tell the difference between "public" (empty string) and "admin only" (nil).
		apiRule := "@request.auth.id != ''"
		collection.ListRule = &apiRule
		collection.ViewRule = &apiRule
		collection.CreateRule = &apiRule
		collection.UpdateRule = &apiRule
		collection.DeleteRule = &apiRule

		// 3. Add your fields
		collection.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Max:      255,
		})

		collection.Fields.Add(&core.TextField{
			Name: "region",
		})

		collection.Fields.Add(&core.NumberField{
			Name: "area",
		})

		collection.Fields.Add(&core.BoolField{
			Name: "active",
		})

		collection.Fields.Add(schema.GeoJSONField("geometry"))

		// 4. Save the collection to the database
		return app.Save(collection)

	}, func(app core.App) error {
		// --- THE DOWN FUNCTION (Undo) ---

		// 1. Find the collection we just made
		collection, err := app.FindCollectionByNameOrId("tracts")
		if err != nil {
			return err // If it doesn't exist, just return the error
		}

		// 2. Delete it
		return app.Delete(collection)
	})
}
