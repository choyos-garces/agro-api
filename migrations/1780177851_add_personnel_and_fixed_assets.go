package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Starting Master Polymorphic Architecture Transition...")

		applyAuthRules := func(c *core.Collection) {
			rule := "@request.auth.id != \"\""
			c.ListRule, c.ViewRule, c.CreateRule, c.UpdateRule = &rule, &rule, &rule, &rule
		}

		// =====================================================================
		// STEP 1: CREATE NEW COLLECTIONS
		// =====================================================================
		log.Println("[Migration] 1/5: Creating new polymorphic collections...")

		// 1A. personnel
		personnel := core.NewBaseCollection("personnel")
		personnel.Fields.Add(&core.TextField{Name: "full_name", Required: true})
		personnel.Fields.Add(&core.TextField{Name: "national_id", Required: true})
		personnel.Fields.Add(&core.DateField{Name: "date_of_birth", Required: true})
		personnel.Fields.Add(&core.BoolField{Name: "active"}) // Default false natively, adjust in UI
		applyAuthRules(personnel)
		if err := app.Save(personnel); err != nil {
			return err
		}

		// 1B. personnel_qualifications
		quals := core.NewBaseCollection("personnel_qualifications")
		quals.Fields.Add(&core.RelationField{Name: "personnel_id", CollectionId: personnel.Id})
		quals.Fields.Add(&core.TextField{Name: "qualification_type", Required: true})
		quals.Fields.Add(&core.DateField{Name: "issue_date", Required: true})
		quals.Fields.Add(&core.DateField{Name: "expiration_date", Required: true})
		applyAuthRules(quals)
		if err := app.Save(quals); err != nil {
			return err
		}

		// 1C. fixed_assets
		assets := core.NewBaseCollection("fixed_assets")
		assets.Fields.Add(&core.TextField{Name: "internal_code", Required: true})
		assets.Fields.Add(&core.SelectField{
			Name:   "asset_category",
			Values: []string{"VEHICLE", "MACHINERY", "INFRASTRUCTURE"},
		})
		assets.Fields.Add(&core.TextField{Name: "brand_model", Required: true})
		assets.Fields.Add(&core.SelectField{
			Name:   "status",
			Values: []string{"OPERATIONAL", "IN_SHOP", "OUT_OF_SERVICE"},
		})
		assets.Fields.Add(&core.BoolField{Name: "active"})
		applyAuthRules(assets)
		if err := app.Save(assets); err != nil {
			return err
		}

		// 1D. field_observations
		tracts, _ := app.FindCollectionByNameOrId("tracts")
		spatial, _ := app.FindCollectionByNameOrId("spatial_features")

		obs := core.NewBaseCollection("field_observations")
		if tracts != nil {
			obs.Fields.Add(&core.RelationField{Name: "tract_id", CollectionId: tracts.Id})
		}
		if spatial != nil {
			obs.Fields.Add(&core.RelationField{Name: "spatial_feature_id", CollectionId: spatial.Id})
		}
		obs.Fields.Add(&core.TextField{Name: "observation_type", Required: true})
		obs.Fields.Add(&core.SelectField{Name: "severity", Values: []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}})
		obs.Fields.Add(&core.TextField{Name: "notes", Required: true})
		obs.Fields.Add(&core.FileField{Name: "photos", MaxSelect: 10, MimeTypes: []string{"image/jpeg", "image/png", "image/webp"}})
		obs.Fields.Add(&core.DateField{Name: "recorded_at", Required: true})
		applyAuthRules(obs)
		if err := app.Save(obs); err != nil {
			return err
		}

		// =====================================================================
		// STEP 2: MIGRATE DATA WITH ID PRESERVATION
		// =====================================================================
		log.Println("[Migration] 2/5: Migrating legacy data (Trucks -> Assets, Drivers -> Personnel)...")

		// Migrate Trucks -> Fixed Assets
		if trucksCol, err := app.FindCollectionByNameOrId("trucks"); err == nil {
			trucks, _ := app.FindAllRecords(trucksCol)
			for _, t := range trucks {
				asset := core.NewRecord(assets)
				asset.Set("id", t.Id) // CRITICAL: Preserves relations in ledgers
				asset.Set("internal_code", t.GetString("license_plate"))
				asset.Set("asset_category", "VEHICLE")
				asset.Set("brand_model", fmt.Sprintf("%s %s", t.GetString("brand"), t.GetString("model_year")))
				asset.Set("status", t.GetString("status"))
				asset.Set("active", t.GetBool("active"))
				app.Save(asset)
			}
		}

		// Migrate Drivers -> Personnel
		if driversCol, err := app.FindCollectionByNameOrId("drivers"); err == nil {
			drivers, _ := app.FindAllRecords(driversCol)
			for _, d := range drivers {
				p := core.NewRecord(personnel)
				p.Set("id", d.Id) // CRITICAL: Preserves relations in ledgers
				p.Set("full_name", d.GetString("full_name"))
				p.Set("national_id", d.GetString("national_id"))
				p.Set("date_of_birth", d.GetDateTime("date_of_birth"))
				p.Set("active", d.GetBool("active"))
				app.Save(p)

				// Extract License to Qualifications
				if d.GetString("license_type") != "" {
					q := core.NewRecord(quals)
					q.Set("personnel_id", p.Id)
					q.Set("qualification_type", "COMMERCIAL_DRIVING")
					q.Set("issue_date", d.GetDateTime("date_of_birth")) // Placeholder issue date
					q.Set("expiration_date", d.GetDateTime("license_expiration"))
					app.Save(q)
				}
			}
		}

		// =====================================================================
		// STEP 3: RAW SQL DATA NORMALIZATION (Enums)
		// =====================================================================
		log.Println("[Migration] 3/5: Normalizing legacy enums via SQL...")
		db := app.DB()

		// Logistic Locations
		db.NewQuery("UPDATE logistic_locations SET location_type = 'CLIENT' WHERE location_type = 'CLIENT_FARM'").Execute()

		// Infrastructure mappings
		db.NewQuery("UPDATE infrastructure SET type = 'ROAD' WHERE type = 'parking'").Execute()
		db.NewQuery("UPDATE infrastructure SET type = 'BUILDING' WHERE type = 'facility'").Execute()
		db.NewQuery("UPDATE infrastructure SET type = UPPER(type)").Execute() // Capitalize the rest

		// Work Areas mappings
		db.NewQuery("UPDATE work_areas SET category = 'PHYTOSANITARY' WHERE category IN ('diseased', 'treatment_active', 'fallow', 'replanted', 'reseeded')").Execute()
		db.NewQuery("UPDATE work_areas SET category = 'OPERATIONAL_DELAY' WHERE category NOT IN ('PHYTOSANITARY', 'ENVIRONMENTAL_DAMAGE', 'TRIAL_NURSERY')").Execute()

		// =====================================================================
		// STEP 4: APPLY SCHEMA MODIFICATIONS
		// =====================================================================
		log.Println("[Migration] 4/5: Updating field schemas and relation targets...")

		// 4A. logistic_locations
		if locs, err := app.FindCollectionByNameOrId("logistic_locations"); err == nil {
			if f, ok := locs.Fields.GetByName("location_type").(*core.SelectField); ok {
				f.Name = "type"
				f.Values = []string{"PORT", "CLIENT", "OWN_FACILITY", "GARAGE"}
			}
			app.Save(locs)
		}

		// 4B. logistic_routes
		if routes, err := app.FindCollectionByNameOrId("logistic_routes"); err == nil {
			if f, ok := routes.Fields.GetByName("start_location_id").(*core.RelationField); ok {
				f.Name = "origin_location_id"
			}
			routes.Fields.Add(&core.NumberField{Name: "expected_toll_count"})
			routes.Fields.Add(&core.NumberField{Name: "expected_toll_cost"})
			app.Save(routes)
		}

		// 4C. spatial_features
		if sf, err := app.FindCollectionByNameOrId("spatial_features"); err == nil {
			sf.Fields.Add(&core.NumberField{Name: "area"})
			sf.Fields.Add(&core.BoolField{Name: "archived"})
			app.Save(sf)
		}

		// 4D. infrastructure
		if infra, err := app.FindCollectionByNameOrId("infrastructure"); err == nil {
			if f, ok := infra.Fields.GetByName("type").(*core.SelectField); ok {
				f.Values = []string{"BUILDING", "WATER_BODY", "ROAD", "EATING_AREA", "PACKING_PLANT", "FEEDER", "WAREHOUSE", "CABLEWAY", "PUMP_STATION", "DRAINAGE_CANAL"}
			}
			app.Save(infra)
		}

		// 4E. work_areas
		if wa, err := app.FindCollectionByNameOrId("work_areas"); err == nil {
			if f, ok := wa.Fields.GetByName("category").(*core.SelectField); ok {
				f.Name = "type"
				f.Values = []string{"PHYTOSANITARY", "OPERATIONAL_DELAY", "ENVIRONMENTAL_DAMAGE", "TRIAL_NURSERY"}
			}
			wa.Fields.Add(&core.SelectField{Name: "status", Values: []string{"ACTIVE", "RESOLVED", "ARCHIVED"}})
			app.Save(wa)
		}

		// 4F. ledger_logistics_trips (Update Targets)
		if trips, err := app.FindCollectionByNameOrId("ledger_logistics_trips"); err == nil {
			if f, ok := trips.Fields.GetByName("truck_id").(*core.RelationField); ok {
				f.Name = "asset_id"
				f.CollectionId = assets.Id // Point to new collection
			}
			if f, ok := trips.Fields.GetByName("driver_id").(*core.RelationField); ok {
				f.Name = "personnel_id"
				f.CollectionId = personnel.Id // Point to new collection
			}
			trips.Fields.Add(&core.NumberField{Name: "start_odometer"})
			trips.Fields.Add(&core.NumberField{Name: "end_odometer"})
			trips.Fields.Add(&core.NumberField{Name: "actual_tolls_paid"})
			app.Save(trips)
		}

		// 4G. ledger_crop_seeding
		if seeding, err := app.FindCollectionByNameOrId("ledger_crop_seeding"); err == nil {
			seeding.Fields.Add(&core.SelectField{Name: "origin_type", Values: []string{"INTERNAL_TRANSFER", "EXTERNAL_PURCHASE"}})
			seeding.Fields.Add(&core.TextField{Name: "origin_reference"})
			app.Save(seeding)
		}

		// =====================================================================
		// STEP 5: DROP LEGACY COLLECTIONS
		// =====================================================================
		log.Println("[Migration] 5/5: Dropping deprecated collections...")

		for _, dropName := range []string{"trucks", "drivers"} {
			if col, err := app.FindCollectionByNameOrId(dropName); err == nil {
				app.Delete(col)
			}
		}

		log.Println("[Migration] Master transition completed successfully!")
		return nil

	}, func(app core.App) error {
		// ROLLBACK
		log.Println("[Migration] Rollback initiated. Warning: Dropped table data cannot be perfectly restored via schema rollback.")

		// Note: A true data rollback here would require recreating `trucks` and `drivers`
		// and re-running a reverse SQL data injection from `fixed_assets` and `personnel`.
		// Given the complexity of this state change, structural rollback focuses on
		// dropping the new tables to unblock the compiler.

		collectionsToDrop := []string{
			"field_observations",
			"fixed_assets",
			"personnel_qualifications",
			"personnel",
		}

		for _, name := range collectionsToDrop {
			if col, err := app.FindCollectionByNameOrId(name); err == nil {
				app.Delete(col)
			}
		}

		return nil
	})
}
