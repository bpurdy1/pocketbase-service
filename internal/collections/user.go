package collections

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// EnsureUserFields adds custom fields to the users collection if they don't exist.
func EnsureUserFields(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			log.Printf("Failed to find users collection: %v", err)
			return e.Next()
		}

		changed := false

		// Add phone field if it doesn't exist
		if users.Fields.GetByName("phone") == nil {
			users.Fields.Add(&core.TextField{
				Name:     "phone",
				Required: true,
				Pattern:  `^\+?[1-9]\d{1,14}$`,
			})
			changed = true
		}

		// Add role field if it doesn't exist
		if users.Fields.GetByName("role") == nil {
			users.Fields.Add(&core.SelectField{
				Name:      "role",
				Required:  true,
				MaxSelect: 1,
				Values:    AllPlatformRoles,
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

	// Auto-assign "user" role on signup (before save)
	app.OnRecordCreateRequest("users").BindFunc(func(e *core.RecordRequestEvent) error {
		if e.Record.GetString("role") == "" {
			e.Record.Set("role", RoleUser)
		}
		return e.Next()
	})

	// Auto-create settings record after user is created
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		settingsCol, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			log.Printf("settings collection not found: %v", err)
			return nil
		}

		settings := core.NewRecord(settingsCol)
		settings.Set("user", e.Record.Id)
		settings.Set("email_notifications", true)
		settings.Set("sms_notifications", false)
		settings.Set("theme", "system")

		if err := app.Save(settings); err != nil {
			log.Printf("Failed to create settings for user %s: %v", e.Record.Id, err)
		}

		return nil
	})
}
