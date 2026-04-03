package users

import (
	"log"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/collections/patch"
)

// EnsureSettingsOnBeforeServe registers the settings collection setup on server start.
func EnsureSettingsOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureSettings(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

// EnsureSettings creates the settings collection if it doesn't exist.
func EnsureSettings(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("settings")
	if existing != nil {
		ensureSettingsFields(app, existing)
		return patch.Collection(app, "settings",
			patch.AutodateFields(),
		)
	}

	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
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

	ownerRule := "@request.auth.id = user || @request.auth.role = \"admin\""
	collection.ListRule = &ownerRule
	collection.ViewRule = &ownerRule
	collection.UpdateRule = &ownerRule
	collection.CreateRule = &ownerRule
	collection.DeleteRule = &ownerRule

	collection.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
	collection.AddIndex("idx_user_settings_user", true, "user", "")

	return app.Save(collection)
}

// ensureSettingsFields updates the theme SelectField to include all valid values
// if the collection was created by an older version that was missing some.
func ensureSettingsFields(app core.App, col *core.Collection) {
	themeField := col.Fields.GetByName("theme")
	if themeField == nil {
		return
	}

	sel, ok := themeField.(*core.SelectField)
	if !ok {
		return
	}

	expected := []string{"light", "dark", "system"}
	missing := false
	have := map[string]bool{}
	for _, v := range sel.Values {
		have[v] = true
	}
	for _, v := range expected {
		if !have[v] {
			missing = true
			sel.Values = append(sel.Values, v)
		}
	}

	if missing {
		if err := app.Save(col); err != nil {
			log.Printf("Failed to update settings theme field: %v", err)
		} else {
			log.Println("Updated settings theme field with missing values")
		}
	}
}
