package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		log.Println("[Migration] Starting Personnel & Qualifications upgrade...")
		db := app.DB()

		// =====================================================================
		// STEP 1: Add New Fields (Initially non-required to avoid validation crash)
		// =====================================================================
		log.Println("[Migration] 1/3: Expanding schemas...")

		// Personnel Updates
		p, err := app.FindCollectionByNameOrId("personnel")
		if err != nil {
			return fmt.Errorf("failed to find 'personnel': %w", err)
		}
		p.Fields.Add(&core.TextField{Name: "internal_code"})
		p.Fields.Add(&core.SelectField{Name: "type", Values: []string{"FULL_TIME", "DAY_LABORER", "CONTRACTOR", "SALARY"}})
		p.Fields.Add(&core.SelectField{Name: "status", Values: []string{"ACTIVE", "INACTIVE", "ON_LEAVE"}})
		p.Fields.Add(&core.DateField{Name: "hire_date"})
		p.Fields.Add(&core.DateField{Name: "termination_date"})
		p.Fields.Add(&core.TextField{Name: "contact_phone"})
		if err := app.Save(p); err != nil {
			return fmt.Errorf("failed saving expanded personnel schema: %w", err)
		}

		// Qualifications Updates
		pq, err := app.FindCollectionByNameOrId("personnel_qualifications")
		if err != nil {
			return fmt.Errorf("failed to find 'personnel_qualifications': %w", err)
		}
		pq.Fields.Add(&core.SelectField{
			Name:   "qualification_tag",
			Values: []string{"COMMERCIAL_DRIVING", "HEAVY_MACHINERY", "FIELD_LABOR", "CHEMICAL_APPLICATION", "SUPERVISOR", "OFFICE_ADMIN", "OTHER"},
		})
		pq.Fields.Add(&core.TextField{Name: "license_reference"})
		pq.Fields.RemoveByName("issue_date")

		if err := app.Save(pq); err != nil {
			return fmt.Errorf("failed saving expanded qualifications schema: %w", err)
		}

		// =====================================================================
		// STEP 2: Raw SQL Data Backfill & Normalization
		// =====================================================================
		log.Println("[Migration] 2/3: Backfilling legacy data via SQL...")

		// Map personnel active bool to status, default type, and map hire_date to created date
		_, err = db.NewQuery("UPDATE personnel SET status = CASE WHEN active = 1 OR active = 'true' THEN 'ACTIVE' ELSE 'INACTIVE' END").Execute()
		if err != nil {
			return err
		}

		_, err = db.NewQuery("UPDATE personnel SET type = 'DAY_LABORER' WHERE type = '' OR type IS NULL").Execute()
		if err != nil {
			return err
		}

		_, err = db.NewQuery("UPDATE personnel SET hire_date = created WHERE hire_date = '' OR hire_date IS NULL").Execute()
		if err != nil {
			return err
		}

		// Smart-map legacy text qualification types to the new Enum tags
		_, err = db.NewQuery(`
			UPDATE personnel_qualifications
			SET qualification_tag = CASE
				WHEN UPPER(qualification_type) IN ('COMMERCIAL_DRIVING', 'HEAVY_MACHINERY', 'FIELD_LABOR', 'CHEMICAL_APPLICATION', 'SUPERVISOR', 'OFFICE_ADMIN', 'OTHER')
					THEN UPPER(qualification_type)
				WHEN UPPER(qualification_type) LIKE '%DRIV%' THEN 'COMMERCIAL_DRIVING'
				WHEN UPPER(qualification_type) LIKE '%MACHIN%' OR UPPER(qualification_type) LIKE '%TRACTOR%' THEN 'HEAVY_MACHINERY'
				WHEN UPPER(qualification_type) LIKE '%CHEM%' OR UPPER(qualification_type) LIKE '%FUMIGAT%' THEN 'CHEMICAL_APPLICATION'
				ELSE 'OTHER'
			END
		`).Execute()
		if err != nil {
			return err
		}

		// =====================================================================
		// STEP 3: Cleanup & Enforce Constraints
		// =====================================================================
		log.Println("[Migration] 3/3: Enforcing strict constraints and Cascade rules...")

		// Enforce Personnel
		p, _ = app.FindCollectionByNameOrId("personnel")
		p.Fields.RemoveByName("active") // Drop deprecated field

		if f, ok := p.Fields.GetByName("type").(*core.SelectField); ok {
			f.Required = true
		}
		if f, ok := p.Fields.GetByName("status").(*core.SelectField); ok {
			f.Required = true
		}
		if f, ok := p.Fields.GetByName("hire_date").(*core.DateField); ok {
			f.Required = true
		}

		if err := app.Save(p); err != nil {
			return err
		}

		// Enforce Qualifications
		pq, _ = app.FindCollectionByNameOrId("personnel_qualifications")
		pq.Fields.RemoveByName("qualification_type") // Drop deprecated field

		if f, ok := pq.Fields.GetByName("qualification_tag").(*core.SelectField); ok {
			f.Required = true
		}
		if exp, ok := pq.Fields.GetByName("expiration_date").(*core.DateField); ok {
			exp.Required = false
		} // Make Nullable
		if rel, ok := pq.Fields.GetByName("personnel_id").(*core.RelationField); ok {
			rel.CascadeDelete = true
		}

		if err := app.Save(pq); err != nil {
			return err
		}

		log.Println("[Migration] Personnel schemas upgraded successfully!")
		return nil

	}, func(app core.App) error {
		log.Println("[Migration Rollback] Rolling back Personnel schemas...")
		db := app.DB()

		// 1. Revert Personnel
		p, err := app.FindCollectionByNameOrId("personnel")
		if err == nil {
			p.Fields.Add(&core.BoolField{Name: "active"})

			// Save to create column, then backfill
			_ = app.Save(p)
			db.NewQuery("UPDATE personnel SET active = (status = 'ACTIVE')").Execute()

			p, _ = app.FindCollectionByNameOrId("personnel")
			p.Fields.RemoveByName("internal_code")
			p.Fields.RemoveByName("type")
			p.Fields.RemoveByName("status")
			p.Fields.RemoveByName("hire_date")
			p.Fields.RemoveByName("termination_date")
			p.Fields.RemoveByName("contact_phone")
			_ = app.Save(p)
		}

		// 2. Revert Qualifications
		pq, err := app.FindCollectionByNameOrId("personnel_qualifications")
		if err == nil {
			pq.Fields.Add(&core.TextField{Name: "qualification_type", Required: true})

			// Save to create column, then backfill
			_ = app.Save(pq)
			db.NewQuery("UPDATE personnel_qualifications SET qualification_type = qualification_tag").Execute()

			pq, _ = app.FindCollectionByNameOrId("personnel_qualifications")
			pq.Fields.RemoveByName("qualification_tag")

			if exp, ok := pq.Fields.GetByName("expiration_date").(*core.DateField); ok {
				exp.Required = true
			}
			if rel, ok := pq.Fields.GetByName("personnel_id").(*core.RelationField); ok {
				rel.CascadeDelete = false
			}

			_ = app.Save(pq)
		}

		return nil
	})
}
