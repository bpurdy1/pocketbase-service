package organizations

import (
	"log"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/collections/roles"
)

// RegisterHooks sets up lifecycle hooks for organizations:
//   - Auto-create org_members entry with "owner" role when an org is created
//   - Auto-create org_settings record when an org is created
func RegisterHooks(app core.App) {
	app.OnRecordCreateRequest("organizations").BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		// Auto-create org_member with owner role
		members, err := app.FindCollectionByNameOrId("org_members")
		if err != nil {
			log.Printf("org_members collection not found: %v", err)
			return nil
		}

		member := core.NewRecord(members)
		member.Set("user", e.Auth.Id)
		member.Set("organization", e.Record.Id)
		member.Set("role", roles.OrgOwner)

		if err := app.Save(member); err != nil {
			log.Printf("Failed to create org owner membership: %v", err)
		}

		// Auto-create org_settings with defaults
		settingsCol, err := app.FindCollectionByNameOrId("org_settings")
		if err != nil {
			log.Printf("org_settings collection not found: %v", err)
			return nil
		}

		settings := core.NewRecord(settingsCol)
		settings.Set("organization", e.Record.Id)
		settings.Set("billing_plan", "free")
		settings.Set("features", map[string]any{
			"max_members":    5,
			"max_properties": 100,
		})
		settings.Set("notification_preferences", map[string]any{
			"member_joined":    true,
			"member_removed":   true,
			"property_changed": true,
			"invite_accepted":  true,
		})

		if err := app.Save(settings); err != nil {
			log.Printf("Failed to create org_settings for org %s: %v", e.Record.Id, err)
		}

		return nil
	})
}
