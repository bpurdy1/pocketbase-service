package organizations

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// EnsureCollection creates the organizations collection if it doesn't exist.
func EnsureCollection(app *pocketbase.PocketBase) {
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
}
