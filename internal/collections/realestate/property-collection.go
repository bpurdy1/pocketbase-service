package realestate

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// EnsureProperties creates the properties collection if it doesn't exist.
// Rules are applied separately by ApplyPropertyRules after all collections exist.
func EnsureProperties(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		existing, _ := app.FindCollectionByNameOrId("properties")
		if existing != nil {
			return e.Next()
		}

		orgsCol, err := app.FindCollectionByNameOrId("organizations")
		if err != nil {
			log.Printf("Failed to find organizations collection: %v", err)
			return e.Next()
		}

		collection := core.NewBaseCollection("properties")
		collection.Fields.Add(
			&core.RelationField{
				Name:          "organization",
				CollectionId:  orgsCol.Id,
				Required:      true,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.TextField{Name: "property_name", Required: true},
			&core.TextField{Name: "address", Required: true},
			&core.TextField{Name: "city", Required: true},
			&core.TextField{Name: "state", Max: 2},
			&core.TextField{Name: "zip_code", Pattern: `^\d{5}(-\d{4})?$`},
			&core.TextField{Name: "county"},
			&core.NumberField{Name: "year_built"},
			&core.NumberField{Name: "number_of_units"},
			&core.NumberField{Name: "building_sf"},
			&core.NumberField{Name: "lot_sf"},
		)

		collection.AddIndex("idx_properties_org", false, "organization", "")

		if err := app.Save(collection); err != nil {
			log.Printf("Failed to create properties collection: %v", err)
		} else {
			log.Println("Created properties collection")
		}

		return e.Next()
	})
}

// ApplyPropertyRules sets access rules on properties. Must run after org_members exists.
func ApplyPropertyRules(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		collection, err := app.FindCollectionByNameOrId("properties")
		if err != nil || collection.ListRule != nil {
			return e.Next()
		}

		memberRule := "@request.auth.id != '' && organization.id ?= @collection.org_members.organization && @request.auth.id ?= @collection.org_members.user"
		collection.ListRule = &memberRule
		collection.ViewRule = &memberRule

		adminRule := "@request.auth.id != '' && organization.id ?= @collection.org_members.organization && @request.auth.id ?= @collection.org_members.user && (@collection.org_members.role = 'owner' || @collection.org_members.role = 'admin')"
		collection.CreateRule = &adminRule
		collection.UpdateRule = &adminRule
		collection.DeleteRule = &adminRule

		if err := app.Save(collection); err != nil {
			log.Printf("Failed to apply properties rules: %v", err)
		} else {
			log.Println("Applied properties access rules")
		}

		return e.Next()
	})
}
