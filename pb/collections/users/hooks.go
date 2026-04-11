package users

import (
	"errors"
	"log"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/notifications"
	"pocketbase-server/pb/collections/roles"
)

// RegisterHooks sets up lifecycle hooks for the users collection:
//   - Block authentication for deactivated users
//   - Auto-assign "user" role on signup
//   - Auto-create settings record after user creation
func RegisterHooks(app core.App) {
	// Block deactivated users from authenticating
	app.OnRecordAuthRequest("users").BindFunc(func(e *core.RecordAuthRequestEvent) error {
		if e.Record.GetBool("deactivated") {
			return errors.New("this account has been deactivated")
		}
		return e.Next()
	})
	// Auto-assign "user" role on signup (before save)
	app.OnRecordCreateRequest("users").BindFunc(func(e *core.RecordRequestEvent) error {
		if e.Record.GetString("role") == "" {
			e.Record.Set("role", roles.User)
		}
		return e.Next()
	})

	// After user is created: auto-create settings + personal organization
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		// Auto-create settings record (skip if one already exists)
		settingsCol, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			log.Printf("settings collection not found: %v", err)
		} else {
			exists, _ := app.FindFirstRecordByFilter("settings", "user = {:userId}", map[string]any{"userId": e.Record.Id})
			if exists == nil {
				settings := core.NewRecord(settingsCol)
				settings.Set("user", e.Record.Id)
				settings.Set("email_notifications", true)
				settings.Set("sms_notifications", false)
				settings.Set("theme", "system")

				if err := app.Save(settings); err != nil {
					log.Printf("Failed to create settings for user %s: %v", e.Record.Id, err)
				}
			}
		}

		// Auto-create a personal organization and add user as owner
		orgsCol, err := app.FindCollectionByNameOrId("organizations")
		if err != nil {
			log.Printf("organizations collection not found: %v", err)
			return nil
		}

		username := e.Record.GetString("username")
		if username == "" {
			username = e.Record.GetString("email")
		}

		org := core.NewRecord(orgsCol)
		org.Set("name", username+"'s Organization")
		org.Set("slug", e.Record.Id)

		if err := app.Save(org); err != nil {
			log.Printf("Failed to create personal org for user %s: %v", e.Record.Id, err)
			return nil
		}
		log.Printf("Created personal org %s for user %s", org.Id, e.Record.Id)

		// Add user as owner of the new org
		membersCol, err := app.FindCollectionByNameOrId("org_members")
		if err != nil {
			log.Printf("org_members collection not found: %v", err)
			return nil
		}

		member := core.NewRecord(membersCol)
		member.Set("user", e.Record.Id)
		member.Set("organization", org.Id)
		member.Set("role", roles.OrgOwner)

		if err := app.Save(member); err != nil {
			log.Printf("Failed to create org owner membership for user %s: %v", e.Record.Id, err)
		}

		// Send welcome notification
		notifClient := notifications.NewClient(app)
		if _, err := notifClient.Send(notifications.NotificationOpts{
			Recipient:    e.Record.Id,
			Owner:        e.Record.Id,
			Organization: org.Id,
			Type:         notifications.TypeSystem,
			Title:        "Welcome!",
			Message:      "Your account is ready. Start by exploring your dashboard.",
		}); err != nil {
			log.Printf("Failed to send welcome notification for user %s: %v", e.Record.Id, err)
		}

		return nil
	})
}
