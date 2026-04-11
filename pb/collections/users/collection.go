package users

import (
	"log"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/collections/roles"
	"pocketbase-server/pb/rules"
)

// EnsureCollectionOnBeforeServe registers the users collection setup on server start.
func EnsureCollectionOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureCollection(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

// EnsureCollection adds custom fields to the built-in users collection if they don't exist.
func EnsureCollection(app core.App) error {
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	changed := false

	if f := users.Fields.GetByName("phone"); f == nil {
		users.Fields.Add(&core.TextField{
			Name:    "phone",
			Pattern: `^\+?[1-9]\d{1,14}$`,
		})
		changed = true
	} else if tf, ok := f.(*core.TextField); ok && tf.Required {
		tf.Required = false
		changed = true
	}

	if users.Fields.GetByName("deactivated") == nil {
		users.Fields.Add(&core.BoolField{
			Name: "deactivated",
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

	// Apply access rules if not already set
	if users.ListRule == nil {
		users.ListRule = rules.Ptr(rules.OwnUser)
		users.ViewRule = rules.Ptr(rules.OwnUser)
		users.UpdateRule = rules.Ptr(rules.OwnUser)
		users.DeleteRule = rules.Ptr(rules.OwnUser)
		changed = true
	}

	if changed {
		if err := app.Save(users); err != nil {
			log.Printf("Failed to update users collection: %v", err)
		} else {
			log.Println("Updated users collection fields")
		}
	}

	return nil
}
