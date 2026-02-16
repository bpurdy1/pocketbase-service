package collections

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// EnsureOrganizations creates the organizations collection if it doesn't exist.
// Rules are applied separately by ApplyOrgRules after all collections exist.
func EnsureOrganizations(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		existing, _ := app.FindCollectionByNameOrId("organizations")
		if existing != nil {
			return e.Next()
		}

		collection := core.NewBaseCollection("organizations")
		collection.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "slug", Required: true},
			&core.URLField{Name: "website"},
			&core.TextField{Name: "phone"},
			&core.TextField{Name: "address"},
			&core.TextField{Name: "city"},
			&core.TextField{Name: "state", Max: 2},
			&core.TextField{Name: "zip_code"},
		)

		// Only set create rule (no cross-collection dependency)
		authRule := "@request.auth.id != ''"
		collection.CreateRule = &authRule

		collection.AddIndex("idx_organizations_slug", true, "slug", "")

		if err := app.Save(collection); err != nil {
			log.Printf("Failed to create organizations collection: %v", err)
		} else {
			log.Println("Created organizations collection")
		}

		return e.Next()
	})

	// Auto-create org_members entry with "owner" role when an org is created
	app.OnRecordCreateRequest("organizations").BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		members, err := app.FindCollectionByNameOrId("org_members")
		if err != nil {
			log.Printf("org_members collection not found: %v", err)
			return nil
		}

		member := core.NewRecord(members)
		member.Set("user", e.Auth.Id)
		member.Set("organization", e.Record.Id)
		member.Set("role", OrgRoleOwner)

		if err := app.Save(member); err != nil {
			log.Printf("Failed to create org owner membership: %v", err)
		}

		return nil
	})
}
