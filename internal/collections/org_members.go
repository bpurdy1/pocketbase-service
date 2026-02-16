package collections

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// EnsureOrgMembers creates the org_members join table if it doesn't exist.
// Rules are applied separately by ApplyOrgRules after all collections exist.
func EnsureOrgMembers(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		existing, _ := app.FindCollectionByNameOrId("org_members")
		if existing != nil {
			return e.Next()
		}

		usersCol, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			log.Printf("Failed to find users collection: %v", err)
			return e.Next()
		}
		orgsCol, err := app.FindCollectionByNameOrId("organizations")
		if err != nil {
			log.Printf("Failed to find organizations collection: %v", err)
			return e.Next()
		}

		collection := core.NewBaseCollection("org_members")
		collection.Fields.Add(
			&core.RelationField{
				Name:          "user",
				CollectionId:  usersCol.Id,
				Required:      true,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.RelationField{
				Name:          "organization",
				CollectionId:  orgsCol.Id,
				Required:      true,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.SelectField{
				Name:      "role",
				Required:  true,
				MaxSelect: 1,
				Values:    AllOrgRoles,
			},
		)

		// One membership per user per org
		collection.AddIndex("idx_org_members_unique", true, "user, organization", "")
		collection.AddIndex("idx_org_members_org", false, "organization", "")

		if err := app.Save(collection); err != nil {
			log.Printf("Failed to create org_members collection: %v", err)
		} else {
			log.Println("Created org_members collection")
		}

		return e.Next()
	})
}
