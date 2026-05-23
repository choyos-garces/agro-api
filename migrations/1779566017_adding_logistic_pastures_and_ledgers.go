package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Helper to apply the CRU rules (Admin-only Delete)
		applyAuthRules := func(c *core.Collection) {
			rule := "@request.auth.id != \"\""
			c.ListRule = &rule
			c.ViewRule = &rule
			c.CreateRule = &rule
			c.UpdateRule = &rule
			c.DeleteRule = nil // Admin only
		}

		// ---------------------------------------------------------
		// 0. DEPENDENCIES
		// ---------------------------------------------------------

		// Ensure spatial_features exists for the relations
		spatial, err := app.FindCollectionByNameOrId("spatial_features")
		if err != nil {
			return err
		}

		// ---------------------------------------------------------
		// 1. CORE ENTITIES
		// ---------------------------------------------------------

		// logistic_locations
		locs := core.NewBaseCollection("logistic_locations")
		locs.Fields.Add(&core.RelationField{Name: "spatial_feature_id", CollectionId: spatial.Id})
		locs.Fields.Add(&core.SelectField{
			Name: "location_type",
			Values: []string{"PORT", "CLIENT_FARM", "OWN_FACILITY", "GARAGE"},
		})
		locs.Fields.Add(&core.TextField{Name: "official_name"})
		locs.Fields.Add(&core.TextField{Name: "contact_name"})
		locs.Fields.Add(&core.TextField{Name: "contact_phone"})
		applyAuthRules(locs)
		if err := app.Save(locs); err != nil { return err }

		// trucks
		trucks := core.NewBaseCollection("trucks")
		trucks.Fields.Add(&core.TextField{Name: "license_plate"})
		trucks.Fields.Add(&core.TextField{Name: "brand"})
		trucks.Fields.Add(&core.NumberField{Name: "model_year"})
		trucks.Fields.Add(&core.TextField{Name: "color"})
		trucks.Fields.Add(&core.SelectField{
			Name: "status",
			Values: []string{"OPERATIONAL", "IN_SHOP", "OUT_OF_SERVICE"},
		})
		applyAuthRules(trucks)
		if err := app.Save(trucks); err != nil { return err }

		// pastures
		pastures := core.NewBaseCollection("pastures")
		pastures.Fields.Add(&core.RelationField{Name: "spatial_feature_id", CollectionId: spatial.Id})
		pastures.Fields.Add(&core.TextField{Name: "pasture_name"})
		pastures.Fields.Add(&core.SelectField{
			Name: "pasture_status",
			Values: []string{"OPEN", "CLOSED", "RESTING", "UNDER_MAINTENANCE"},
		})
		pastures.Fields.Add(&core.NumberField{Name: "total_area_hectares"})
		applyAuthRules(pastures)
		if err := app.Save(pastures); err != nil { return err }

		// registry_moko_cases
		moko := core.NewBaseCollection("registry_moko_cases")
		moko.Fields.Add(&core.DateField{Name: "detected_at"})
		moko.Fields.Add(&core.TextField{Name: "work_area_id"})
		moko.Fields.Add(&core.NumberField{Name: "initial_plants_flagged"})
		moko.Fields.Add(&core.SelectField{
			Name: "containment_status",
			Values: []string{"QUARANTINED", "MONITORING", "CLEARED"},
		})
		moko.Fields.Add(&core.TextField{Name: "internal_location_notes"})
		applyAuthRules(moko)
		if err := app.Save(moko); err != nil { return err }

		// logistic_routes (Depends on logistic_locations)
		routes := core.NewBaseCollection("logistic_routes")
		routes.Fields.Add(&core.TextField{Name: "route_name"})
		routes.Fields.Add(&core.RelationField{Name: "destiny_location_id", CollectionId: locs.Id})
		routes.Fields.Add(&core.RelationField{Name: "start_location_id", CollectionId: locs.Id})
		routes.Fields.Add(&core.NumberField{Name: "total_km_roundtrip"})
		applyAuthRules(routes)
		if err := app.Save(routes); err != nil { return err }

		// drivers (Depends on users)
		users, _ := app.FindCollectionByNameOrId("users")
		drivers := core.NewBaseCollection("drivers")
		if users != nil {
			drivers.Fields.Add(&core.RelationField{Name: "user_id", CollectionId: users.Id, Required: false})
		}
		drivers.Fields.Add(&core.TextField{Name: "full_name"})
		drivers.Fields.Add(&core.TextField{Name: "national_id"})
		drivers.Fields.Add(&core.DateField{Name: "date_of_birth"})
		drivers.Fields.Add(&core.TextField{Name: "license_type"})
		drivers.Fields.Add(&core.DateField{Name: "license_expiration"})
		applyAuthRules(drivers)
		if err := app.Save(drivers); err != nil { return err }

		// ---------------------------------------------------------
		// 2. STANDALONE OPERATIONAL LEDGERS (Immutable)
		// ---------------------------------------------------------

		// ledger_material_inputs
		matInputs := core.NewBaseCollection("ledger_material_inputs")
		matInputs.Fields.Add(&core.DateField{Name: "applied_at"})
		matInputs.Fields.Add(&core.TextField{Name: "spatial_target_id"})
		matInputs.Fields.Add(&core.TextField{Name: "spatial_target_type"})
		matInputs.Fields.Add(&core.SelectField{
			Name: "category",
			Values: []string{"NUTRITION", "FUNGICIDE", "HERBICIDE", "BIOLOGICAL", "SANITIZATION"},
		})
		matInputs.Fields.Add(&core.TextField{Name: "product_name_raw"})
		matInputs.Fields.Add(&core.NumberField{Name: "quantity_spent"})
		matInputs.Fields.Add(&core.TextField{Name: "unit_type"})
		matInputs.Fields.Add(&core.NumberField{Name: "safety_entry_days"})
		matInputs.Fields.Add(&core.RelationField{Name: "associated_moko_case_id", CollectionId: moko.Id, Required: false})
		matInputs.Fields.Add(&core.TextField{Name: "internal_location_notes"})
		applyAuthRules(matInputs)
		if err := app.Save(matInputs); err != nil { return err }

		// ledger_field_labor
		fieldLabor := core.NewBaseCollection("ledger_field_labor")
		fieldLabor.Fields.Add(&core.DateField{Name: "verified_at"})
		fieldLabor.Fields.Add(&core.TextField{Name: "spatial_target_id"})
		fieldLabor.Fields.Add(&core.SelectField{
			Name: "labor_type",
			Values: []string{"DEFOLIATION", "PRUNING_HARVESTED", "SELECTION", "WEED_CLEARING", "MATERIAL_STAGING", "BRACING_CANE", "BRACING_STRAP", "FRUIT_PROTECTION", "CALIBRATION", "SOIL_PREPARATION"},
		})
		fieldLabor.Fields.Add(&core.NumberField{Name: "metric_value"})
		fieldLabor.Fields.Add(&core.TextField{Name: "metric_unit"})
		fieldLabor.Fields.Add(&core.RelationField{Name: "associated_moko_case_id", CollectionId: moko.Id, Required: false})
		fieldLabor.Fields.Add(&core.TextField{Name: "internal_location_notes"})
		applyAuthRules(fieldLabor)
		if err := app.Save(fieldLabor); err != nil { return err }

		// ledger_crop_seeding
		seeding := core.NewBaseCollection("ledger_crop_seeding")
		seeding.Fields.Add(&core.DateField{Name: "planted_at"})
		seeding.Fields.Add(&core.TextField{Name: "spatial_target_id"})
		seeding.Fields.Add(&core.TextField{Name: "crop_variety"})
		seeding.Fields.Add(&core.NumberField{Name: "quantity_planted"})
		seeding.Fields.Add(&core.TextField{Name: "unit_type"})
		applyAuthRules(seeding)
		if err := app.Save(seeding); err != nil { return err }

		// ledger_pasture_rotation
		rotation := core.NewBaseCollection("ledger_pasture_rotation")
		rotation.Fields.Add(&core.RelationField{Name: "pasture_id", CollectionId: pastures.Id})
		rotation.Fields.Add(&core.SelectField{
			Name: "event_type",
			Values: []string{"OPENED", "CLOSED"},
		})
		rotation.Fields.Add(&core.DateField{Name: "event_timestamp"})
		rotation.Fields.Add(&core.NumberField{Name: "livestock_count"})
		rotation.Fields.Add(&core.TextField{Name: "animal_group_reference"})
		applyAuthRules(rotation)
		if err := app.Save(rotation); err != nil { return err }

		// ledger_logistics_trips
		trips := core.NewBaseCollection("ledger_logistics_trips")
		trips.Fields.Add(&core.RelationField{Name: "route_id", CollectionId: routes.Id})
		trips.Fields.Add(&core.RelationField{Name: "truck_id", CollectionId: trucks.Id})
		trips.Fields.Add(&core.RelationField{Name: "driver_id", CollectionId: drivers.Id})
		trips.Fields.Add(&core.DateField{Name: "started_at"})
		trips.Fields.Add(&core.DateField{Name: "ended_at"})
		trips.Fields.Add(&core.TextField{Name: "client_farm_name"})
		applyAuthRules(trips)
		if err := app.Save(trips); err != nil { return err }

		// ledger_asset_events
		assets := core.NewBaseCollection("ledger_asset_events")
		assets.Fields.Add(&core.DateField{Name: "logged_at"})
		assets.Fields.Add(&core.SelectField{
			Name: "asset_type",
			Values: []string{"VEHICLE", "FARM_MACHINERY", "FACILITY_INFRASTRUCTURE"},
		})
		assets.Fields.Add(&core.TextField{Name: "asset_reference_name"})
		assets.Fields.Add(&core.NumberField{Name: "current_odometer_reading"})
		assets.Fields.Add(&core.SelectField{
			Name: "maintenance_type",
			Values: []string{"OIL_CHANGE", "TIRE_REPLACEMENT", "ROUTINE_SERVICE", "REPAIR"},
		})
		assets.Fields.Add(&core.TextField{Name: "action_summary"})
		applyAuthRules(assets)
		if err := app.Save(assets); err != nil { return err }

		return nil
	}, func(app core.App) error {
		// Revert in reverse order of dependencies
		collectionsToDrop := []string{
			"ledger_asset_events", "ledger_logistics_trips", "ledger_pasture_rotation",
			"ledger_crop_seeding", "ledger_field_labor", "ledger_material_inputs",
			"drivers", "logistic_routes", "registry_moko_cases", "pastures", "trucks", "logistic_locations",
		}

		for _, name := range collectionsToDrop {
			if col, err := app.FindCollectionByNameOrId(name); err == nil {
				if err := app.Delete(col); err != nil {
					return err
				}
			}
		}
		
		// Note: Intentionally leaving spatial_features intact during rollback 
		// if it was created here, as it might be bound to other operational schemas.

		return nil
	})
}