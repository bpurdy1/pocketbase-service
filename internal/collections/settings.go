package collections

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// EnsureSettings creates the user_settings collection if it doesn't exist.
func EnsureSettings(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		existing, _ := app.FindCollectionByNameOrId("settings")
		if existing != nil {
			return e.Next()
		}

		usersCol, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			log.Printf("Failed to find users collection: %v", err)
			return e.Next()
		}

		collection := core.NewBaseCollection("settings")
		collection.Fields.Add(
			&core.RelationField{
				Name:          "user",
				CollectionId:  usersCol.Id,
				Required:      true,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.BoolField{Name: "email_notifications"},
			&core.BoolField{Name: "sms_notifications"},
			&core.SelectField{
				Name:      "theme",
				MaxSelect: 1,
				Values:    []string{"light", "dark", "system"},
			},
			&core.TextField{Name: "timezone"},
			&core.JSONField{Name: "preferences"},
		)

		// Only the owner can read/update their own settings
		ownerRule := "@request.auth.id = user || @request.auth.role = \"admin\""
		collection.ListRule = &ownerRule
		collection.ViewRule = &ownerRule
		collection.UpdateRule = &ownerRule
		collection.CreateRule = &ownerRule
		collection.DeleteRule = &ownerRule

		// One settings record per user
		collection.AddIndex("idx_user_settings_user", true, "user", "")

		if err := app.Save(collection); err != nil {
			log.Printf("Failed to create user_settings collection: %v", err)
		} else {
			log.Println("Created user_settings collection")
		}

		return e.Next()
	})
}
