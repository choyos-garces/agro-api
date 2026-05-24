package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracts, err := app.FindCollectionByNameOrId("tracts")
		if err != nil {
			return err
		}

		pastures, err := app.FindCollectionByNameOrId("pastures")
		if err != nil {
			return err
		}

		locs, err := app.FindCollectionByNameOrId("logistic_locations")
		if err != nil {
			return err
		}	

		drivers, err := app.FindCollectionByNameOrId("drivers")
		if err != nil {
			return err
		}

		routes, err := app.FindCollectionByNameOrId("logistic_routes")
		if err != nil {
			return err
		}

		// add active boolean to the pasture collection
		pastures.Fields.Add(&core.BoolField{Name: "active"})
		// add tract relation to the pasture collection
		pastures.Fields.Add(&core.RelationField{Name: "tract_id", CollectionId: tracts.Id})
		// remove total_area_hectares from the pasture collection
		pastures.Fields.RemoveByName("total_area_hectares")
		// rename pasture_name to name in the pasture collection
		pastures.Fields.RemoveByName("pasture_name")
		pastures.Fields.Add(&core.TextField{ Name: "name", Required: true, Max: 255,})
		// rename pasture_status to status in the pasture collection
		pastures.Fields.RemoveByName("pasture_status")
		pastures.Fields.Add(&core.SelectField{
			Name:      "status",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"open", "closed", "resting", "maintenance", "seeding"},
		})
		// add pasture_type to the pasture collection
		pastures.Fields.Add(&core.TextField{ Name: "pasture_type", Required: true, Max: 255,})
		// remove user_id from the driver collection
		drivers.Fields.RemoveByName("user_id")
		// rename destiny to destination in the logistics_routes collection	
		routes.Fields.RemoveByName("destiny_location")
		routes.Fields.Add(&core.RelationField{Name: "destination_location_id", CollectionId: locs.Id})
		
		if err := app.Save(pastures); err != nil {

			return fmt.Errorf("failed to save pastures collection: %w", err)
		}

		if err := app.Save(drivers); err != nil {
			return fmt.Errorf("failed to save drivers collection: %w", err)
		}

		if err := app.Save(routes); err != nil {
			return fmt.Errorf("failed to save logistics_routes collection: %w", err)
		}

		return nil
	}, func(app core.App) error {
		pastures, err := app.FindCollectionByNameOrId("pastures")
		if err != nil {
			return err
		}

		locs, err := app.FindCollectionByNameOrId("logistic_locations")
		if err != nil {
			return err
		}

		drivers, err := app.FindCollectionByNameOrId("drivers")
		if err != nil {
			return err
		}

		routes, err := app.FindCollectionByNameOrId("logistic_routes")
		if err != nil {
			return err
		}

		users, _ := app.FindCollectionByNameOrId("users")

		// rollback pasture changes
		pastures.Fields.RemoveByName("active")
		pastures.Fields.RemoveByName("name")
		pastures.Fields.RemoveByName("status")
		pastures.Fields.RemoveByName("pasture_type")
		pastures.Fields.Add(&core.TextField{ Name: "pasture_name", Required: true, Max: 255,})
		pastures.Fields.Add(&core.TextField{ Name: "pasture_status", Required: true, Max: 255,})
		pastures.Fields.Add(&core.NumberField{Name: "total_area_hectares"})
		pastures.Fields.RemoveByName("tract_id")


		// rollback driver changes
		if users != nil {
			drivers.Fields.Add(&core.RelationField{Name: "user_id", CollectionId: users.Id, Required: false})
		}

		// rollback logistics_routes destination rename
		routes.Fields.RemoveByName("destination_location_id")
		routes.Fields.Add(&core.RelationField{Name: "destiny_location", CollectionId: locs.Id})

		if err := app.Save(pastures); err != nil {
			return err
		}

		if err := app.Save(drivers); err != nil {
			return err
		}

		if err := app.Save(routes); err != nil {
			return err
		}

		return nil
	})
}
