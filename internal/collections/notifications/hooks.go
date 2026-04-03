package notifications

import (
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/notifications"
)

func RegisterHooks(app core.App) {
	client := notifications.NewClient(app)

	// --- Invite Hooks ---
	app.OnRecordCreate("org_invites").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		email := e.Record.GetString("email")
		orgId := e.Record.GetString("organization")

		user, err := app.FindAuthRecordByEmail("users", email)
		if err != nil {
			return nil // User doesn't exist, skip notification
		}

		_, err = client.Send(
			notifications.NotificationOpts{
				Recipient:    user.Id,
				Organization: orgId,
				Type:         notifications.TypeInfo,
				Title:        "New Invitation",
				Message:      "You've been invited to join an organization.",
				Data:         map[string]any{"invite_id": e.Record.Id},
			})

		return nil
	})

	// --- Member Hooks ---
	app.OnRecordCreate("org_members").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		userId := e.Record.GetString("user")
		orgId := e.Record.GetString("organization")

		// Notify admins about the new member
		admins, _ := app.FindRecordsByFilter(
			"org_members",
			"organization = {:orgId} && (role = 'owner' || role = 'admin') && user != {:userId}",
			"", 0, 0,
			map[string]any{"orgId": orgId, "userId": userId},
		)

		for _, admin := range admins {
			client.Send(notifications.NotificationOpts{
				Recipient:    admin.GetString("user"),
				Organization: orgId,
				Type:         notifications.TypeInfo,
				Title:        "New Member",
				Message:      "A new member has joined your organization.",
			})
		}

		return nil
	})

	// --- Property Hooks ---
	app.OnRecordCreate("properties").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		// Similar logic: find admins and call client.Send()
		return nil
	})
}
