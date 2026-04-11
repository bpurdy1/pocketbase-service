package realestate

import (
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/collections/patch"
	"pocketbase-server/pb/rules"
)

// ── Property Details ──────────────────────────────────────────────────────────
// Public-record physical facts that rarely change. One row per property.

func EnsurePropertyDetailsOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsurePropertyDetails(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsurePropertyDetails(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("property_details")
	if existing != nil {
		return patch.Collection(app, "property_details",
			patch.AutodateFields(),
			patch.Index("idx_property_details_apn", false, "apn"),
		)
	}

	propertiesCol, err := app.FindCollectionByNameOrId("properties")
	if err != nil {
		return err
	}

	col := core.NewBaseCollection("property_details")

	// Org members can read; writes are system-only (scraped / imported data)
	col.ListRule = rules.Ptr(rules.AuthOnly)
	col.ViewRule = rules.Ptr(rules.AuthOnly)
	col.CreateRule = nil
	col.UpdateRule = nil
	col.DeleteRule = nil

	col.Fields.Add(
		&core.RelationField{
			Name:          "property",
			CollectionId:  propertiesCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},

		// Public record identifiers
		&core.TextField{Name: "apn"},                // Assessor Parcel Number — stable cross-source ID
		&core.TextField{Name: "zoning"},             // e.g. NR3, MF, C1
		&core.TextField{Name: "property_condition"}, // Very Good / Good / Fair / Poor

		// Structure
		&core.NumberField{Name: "stories"},
		&core.NumberField{Name: "sqft_finished"},
		&core.NumberField{Name: "sqft_unfinished"},
		&core.NumberField{Name: "year_renovated"},
		&core.TextField{Name: "foundation"},             // Poured Concrete / Crawl Space / Slab
		&core.TextField{Name: "construction_materials"}, // Cement Plank, Wood
		&core.TextField{Name: "roof_type"},              // Composition / Metal / Flat
		&core.TextField{Name: "basement"},               // Daylight Finished / None
		&core.BoolField{Name: "has_view"},

		// Lot
		&core.NumberField{Name: "lot_size_acres"},

		// Systems
		&core.TextField{Name: "heating"},       // Forced Air, Heat Pump, Radiant
		&core.TextField{Name: "cooling"},       // Heat Pump / None
		&core.TextField{Name: "energy_source"}, // Electric, Natural Gas

		// Parking
		&core.NumberField{Name: "parking_spaces"},
		&core.TextField{Name: "parking_type"}, // Off Street / Garage / Carport

		// Source
		&core.TextField{Name: "data_source"}, // Redfin / Zillow / County / Manual
		&core.TextField{Name: "source_url"},

		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)

	col.AddIndex("idx_property_details_property", true, "property", "") // one row per property
	col.AddIndex("idx_property_details_apn", false, "apn", "")

	return app.Save(col)
}

// ── Property Sale History ─────────────────────────────────────────────────────
// Every recorded sale/listing event (from MLS, public records, etc.).

func EnsurePropertySaleHistoryOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsurePropertySaleHistory(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsurePropertySaleHistory(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("property_sale_history")
	if existing != nil {
		return patch.Collection(app, "property_sale_history",
			patch.AutodateFields(),
			patch.Index("idx_psh_property", false, "property"),
			patch.Index("idx_psh_event_date", false, "event_date"),
		)
	}

	propertiesCol, err := app.FindCollectionByNameOrId("properties")
	if err != nil {
		return err
	}

	col := core.NewBaseCollection("property_sale_history")

	col.ListRule = rules.Ptr(rules.AuthOnly)
	col.ViewRule = rules.Ptr(rules.AuthOnly)
	col.CreateRule = nil
	col.UpdateRule = nil
	col.DeleteRule = nil

	col.Fields.Add(
		&core.RelationField{
			Name:          "property",
			CollectionId:  propertiesCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.TextField{Name: "event_date", Required: true}, // YYYY-MM-DD
		&core.SelectField{
			Name:      "event_type",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"sold", "listed", "pending", "contingent", "relisted", "off_market", "delisted"},
		},
		&core.NumberField{Name: "price"},
		&core.NumberField{Name: "price_per_sqft"},
		&core.TextField{Name: "source"},    // NWMLS, PublicRecord, Redfin, etc.
		&core.TextField{Name: "source_id"}, // MLS grid # or APN

		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)

	col.AddIndex("idx_psh_property", false, "property", "")
	col.AddIndex("idx_psh_event_date", false, "event_date", "")

	return app.Save(col)
}

// ── Property Tax History ──────────────────────────────────────────────────────
// Annual tax assessments. One row per property per year.

func EnsurePropertyTaxHistoryOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsurePropertyTaxHistory(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsurePropertyTaxHistory(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("property_tax_history")
	if existing != nil {
		return patch.Collection(app, "property_tax_history",
			patch.AutodateFields(),
			patch.Index("idx_pth_property", false, "property"),
		)
	}

	propertiesCol, err := app.FindCollectionByNameOrId("properties")
	if err != nil {
		return err
	}

	col := core.NewBaseCollection("property_tax_history")

	col.ListRule = rules.Ptr(rules.AuthOnly)
	col.ViewRule = rules.Ptr(rules.AuthOnly)
	col.CreateRule = nil
	col.UpdateRule = nil
	col.DeleteRule = nil

	col.Fields.Add(
		&core.RelationField{
			Name:          "property",
			CollectionId:  propertiesCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.NumberField{Name: "tax_year", Required: true},
		&core.NumberField{Name: "tax_amount"},
		&core.NumberField{Name: "assessed_value"},
		&core.NumberField{Name: "land_value"},
		&core.NumberField{Name: "improvement_value"},
		&core.TextField{Name: "data_source"},

		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)

	col.AddIndex("idx_pth_property", false, "property", "")
	col.AddIndex("idx_pth_property_year", true, "property, tax_year", "") // one row per year

	return app.Save(col)
}

// ── Property Contacts ─────────────────────────────────────────────────────────
// Listing agents, buyer agents, property managers, etc.

func EnsurePropertyContactsOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsurePropertyContacts(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsurePropertyContacts(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("property_contacts")
	if existing != nil {
		return patch.Collection(app, "property_contacts",
			patch.AutodateFields(),
			patch.Index("idx_pc_property", false, "property"),
		)
	}

	propertiesCol, err := app.FindCollectionByNameOrId("properties")
	if err != nil {
		return err
	}

	col := core.NewBaseCollection("property_contacts")

	col.ListRule = rules.Ptr(rules.AuthOnly)
	col.ViewRule = rules.Ptr(rules.AuthOnly)
	col.CreateRule = nil
	col.UpdateRule = nil
	col.DeleteRule = nil

	col.Fields.Add(
		&core.RelationField{
			Name:          "property",
			CollectionId:  propertiesCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.SelectField{
			Name:      "role",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"listing_agent", "buyer_agent", "property_manager", "owner"},
		},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "brokerage"},
		&core.TextField{Name: "phone"},
		&core.EmailField{Name: "email"},
		&core.TextField{Name: "license_number"},
		&core.TextField{Name: "data_source"}, // NWMLS / Redfin / Manual

		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)

	col.AddIndex("idx_pc_property", false, "property", "")
	col.AddIndex("idx_pc_property_role", false, "property, role", "")

	return app.Save(col)
}
