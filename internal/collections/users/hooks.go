package users

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/collections/roles"
)

// RegisterHooks sets up lifecycle hooks for the users collection:
//   - Auto-assign "user" role on signup
//   - Auto-create settings record after user creation
func RegisterHooks(app *pocketbase.PocketBase) {
	// Auto-assign "user" role on signup (before save)
	app.OnRecordCreateRequest("users").BindFunc(func(e *core.RecordRequestEvent) error {
		if e.Record.GetString("role") == "" {
			e.Record.Set("role", roles.User)
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
