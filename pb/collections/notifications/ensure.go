package notifications

import (
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/collections/patch"
	"pocketbase-server/pb/rules"
)

func EnsureCollectionOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// 1. Run your custom logic
		if err := EnsureCollection(e.App); err != nil {
			return err // If this fails, the server SHOULD stop
		}

		// 2. CRITICAL: Continue the PocketBase lifecycle
		return e.Next()
	})
}

func EnsureCollection(app core.App) error {
	// 1. Check if it already exists
	_, err := app.FindCollectionByNameOrId("notifications")
	if err == nil {
		return patch.Collection(app, "notifications",
			patch.AutodateFields(),
			patch.Index("idx_notifications_created", false, "created"),
		)
	}

	// 2. Define the collection using the new v0.23 field names (lowercase)
	collection := core.NewBaseCollection("notifications")

	recipientOnly := rules.RecipientOnly("recipient")
	collection.ListRule = rules.Ptr(recipientOnly)
	collection.ViewRule = rules.Ptr(recipientOnly)
	collection.UpdateRule = rules.Ptr(recipientOnly)
	collection.DeleteRule = rules.Ptr(recipientOnly)
	collection.CreateRule = nil // system/hooks only

	// 3. Add fields using the new v0.23 field types
	collection.Fields.Add(&core.TextField{
		Name:     "recipient",
		Required: true,
	})
	collection.Fields.Add(&core.TextField{
		Name: "owner",
	})
	collection.Fields.Add(&core.TextField{
		Name: "organization",
	})
	collection.Fields.Add(&core.TextField{
		Name:     "type",
		Required: true,
	})
	collection.Fields.Add(&core.TextField{
		Name:     "title",
		Required: true,
	})
	collection.Fields.Add(&core.TextField{
		Name: "message",
	})
	collection.Fields.Add(&core.BoolField{
		Name: "dismissed",
	})
	collection.Fields.Add(&core.JSONField{
		Name: "data",
	})

	collection.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
	collection.AddIndex("idx_notifications_created", false, "created", "")

	// 4. Save using the app's DAO
	return app.Save(collection)
}
