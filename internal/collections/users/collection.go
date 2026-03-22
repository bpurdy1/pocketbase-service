package users

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/collections/roles"
)

// EnsureCollection adds custom fields to the built-in users collection if they don't exist.
func EnsureCollection(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			log.Printf("Failed to find users collection: %v", err)
			return e.Next()
		}

		changed := false

		if users.Fields.GetByName("phone") == nil {
			users.Fields.Add(&core.TextField{
				Name:     "phone",
				Required: true,
				Pattern:  `^\+?[1-9]\d{1,14}$`,
			})
			changed = true
		}

		if users.Fields.GetByName("role") == nil {
			users.Fields.Add(&core.SelectField{
				Name:      "role",
				Required:  true,
				MaxSelect: 1,
				Values:    roles.AllPlatform,
			})
			changed = true
		}

		if changed {
			if err := app.Save(users); err != nil {
				log.Printf("Failed to update users collection: %v", err)
			} else {
				log.Println("Updated users collection fields")
			}
		}

		return e.Next()
	})
}
