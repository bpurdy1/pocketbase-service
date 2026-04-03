package realestate

import (
	"log"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/collections/patch"
	"pocketbase-server/internal/collections/tenancy"
)

// ensurePropertyFields adds any missing fields to an existing properties collection.
func ensurePropertyFields(app core.App, col *core.Collection) {
	newFields := []core.Field{
		&core.NumberField{Name: "lat"},
		&core.NumberField{Name: "lng"},
		&core.NumberField{Name: "price"},
		&core.NumberField{Name: "bedrooms"},
		&core.NumberField{Name: "bathrooms"},
		&core.NumberField{Name: "sqft"},
		&core.TextField{Name: "property_type"},
		&core.TextField{Name: "notes"},
	}

	changed := false
	for _, f := range newFields {
		if col.Fields.GetByName(f.GetName()) == nil {
			col.Fields.Add(f)
			changed = true
		}
	}

	if changed {
		if err := app.Save(col); err != nil {
			log.Printf("Failed to update properties collection fields: %v", err)
		} else {
			log.Println("Updated properties collection with new fields")
		}
	}
}

func init() {
	tenancy.RegisterPublicRead("properties", "organization")
}

// EnsurePropertiesOnBeforeServe registers the properties collection setup on server start.
func EnsurePropertiesOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureProperties(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

// EnsureProperties creates the properties collection if it doesn't exist.
// Access rules are applied automatically by tenancy.EnforceTenancy.
func EnsureProperties(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("properties")
	if existing != nil {
		ensurePropertyFields(app, existing)
		return patch.Collection(app, "properties",
			patch.AutodateFields(),
			patch.Index("idx_properties_created", false, "created"),
			patch.ClearRules(),
			patch.RelationField("organization", func(f *core.RelationField) bool {
				if !f.Required {
					return false
				}
				f.Required = false
				return true
			}),
		)
	}

	orgsCol, err := app.FindCollectionByNameOrId("organizations")
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection("properties")
	collection.Fields.Add(
		&core.RelationField{
			Name:          "organization",
			CollectionId:  orgsCol.Id,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.TextField{Name: "property_name", Required: true},
		&core.TextField{Name: "address", Required: true},
		&core.TextField{Name: "city", Required: true},
		&core.TextField{Name: "state", Max: 2},
		&core.TextField{Name: "zip_code", Pattern: `^\d{5}(-\d{4})?$`},
		&core.TextField{Name: "county"},
		&core.NumberField{Name: "lat"},
		&core.NumberField{Name: "lng"},
		&core.NumberField{Name: "price"},
		&core.NumberField{Name: "bedrooms"},
		&core.NumberField{Name: "bathrooms"},
		&core.NumberField{Name: "sqft"},
		&core.NumberField{Name: "year_built"},
		&core.NumberField{Name: "number_of_units"},
		&core.NumberField{Name: "building_sf"},
		&core.NumberField{Name: "lot_sf"},
		&core.TextField{Name: "property_type"},
		&core.TextField{Name: "notes"},
	)

	collection.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
	collection.AddIndex("idx_properties_org", false, "organization", "")
	collection.AddIndex("idx_properties_created", false, "created", "")

	return app.Save(collection)
}
