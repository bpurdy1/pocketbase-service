package realestate

import (
	"log"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/collections/patch"
	"pocketbase-server/pb/rules"
)

func EnsureSavedPropertiesOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureSavedProperties(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsureSavedProperties(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("saved_properties")
	if existing != nil {
		return patch.Collection(app, "saved_properties",
			patch.AutodateFields(),
		)
	}

	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	propertiesCol, err := app.FindCollectionByNameOrId("properties")
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection("saved_properties")

	ownerRule := rules.Ptr(rules.OwnRecord("user"))
	collection.ListRule = ownerRule
	collection.ViewRule = ownerRule
	collection.CreateRule = ownerRule
	collection.UpdateRule = ownerRule
	collection.DeleteRule = ownerRule

	collection.Fields.Add(
		&core.RelationField{
			Name:          "user",
			CollectionId:  usersCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.RelationField{
			Name:          "property",
			CollectionId:  propertiesCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.TextField{Name: "notes"},
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)

	// Prevent duplicate saves of the same property by the same user
	collection.AddIndex("idx_saved_properties_user_property", true, "user, property", "")
	collection.AddIndex("idx_saved_properties_user", false, "user", "")

	return app.Save(collection)
}

func EnsureSavedPropertyHistoryOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureSavedPropertyHistory(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsureSavedPropertyHistory(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("saved_property_history")
	if existing != nil {
		return patch.Collection(app, "saved_property_history",
			patch.AutodateFields(),
		)
	}

	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	propertiesCol, err := app.FindCollectionByNameOrId("properties")
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection("saved_property_history")

	ownerRule := rules.Ptr(rules.OwnRecord("user"))
	collection.ListRule = ownerRule
	collection.ViewRule = ownerRule
	// history is written by hooks only — no direct create/update/delete from clients
	collection.CreateRule = nil
	collection.UpdateRule = nil
	collection.DeleteRule = nil

	collection.Fields.Add(
		&core.RelationField{
			Name:         "user",
			CollectionId: usersCol.Id,
			Required:     true,
			MaxSelect:    1,
		},
		&core.RelationField{
			Name:         "property",
			CollectionId: propertiesCol.Id,
			Required:     true,
			MaxSelect:    1,
		},
		&core.SelectField{
			Name:      "action",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"saved", "unsaved"},
		},
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)

	collection.AddIndex("idx_saved_property_history_user", false, "user", "")
	collection.AddIndex("idx_saved_property_history_property", false, "property", "")

	return app.Save(collection)
}

// RegisterSavedPropertyHooks writes a history record whenever a property is saved or unsaved.
func RegisterSavedPropertyHooks(app core.App) {
	app.OnRecordCreate("saved_properties").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		writeHistory(app, e.Record.GetString("user"), e.Record.GetString("property"), "saved")
		return nil
	})

	app.OnRecordDelete("saved_properties").BindFunc(func(e *core.RecordEvent) error {
		// Capture before deletion
		userId := e.Record.GetString("user")
		propertyId := e.Record.GetString("property")

		if err := e.Next(); err != nil {
			return err
		}

		writeHistory(app, userId, propertyId, "unsaved")
		return nil
	})
}

func writeHistory(app core.App, userId, propertyId, action string) {
	col, err := app.FindCollectionByNameOrId("saved_property_history")
	if err != nil {
		log.Printf("saved_property_history collection not found: %v", err)
		return
	}

	record := core.NewRecord(col)
	record.Set("user", userId)
	record.Set("property", propertyId)
	record.Set("action", action)

	if err := app.Save(record); err != nil {
		log.Printf("Failed to write saved_property_history (%s): %v", action, err)
	}
}
