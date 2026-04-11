package notifications

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/notifications"
)

// RegisterRoutes sets up custom HTTP endpoints for notifications.
func RegisterRoutes(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Initialize the client
		client := notifications.NewClient(app)

		// Endpoint to manually send a notification (e.g., System Alerts)
		e.Router.POST("/api/custom/notifications/send", func(e *core.RequestEvent) error {
			// 1. You could parse a JSON body here into NotificationOpts
			// For now, let's assume a simple manual alert
			opts := notifications.NotificationOpts{
				Recipient: "TARGET_USER_ID", // Usually from request body
				Type:      notifications.TypeSystem,
				Title:     "System Maintenance",
				Message:   "The system will be down for 5 minutes.",
			}

			// 2. Use the client to save and get the struct back
			notif, err := client.Send(opts)
			if err != nil {
				return err
			}

			// 3. Return the struct directly to the frontend
			return e.JSON(200, notif)
		})

		return e.Next()
	})
}
