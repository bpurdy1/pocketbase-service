package organizations

import (
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/collections/patch"
)

func EnsureOrgSettingsOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureOrgSettings(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsureOrgSettings(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("org_settings")
	if existing != nil {
		return patch.Collection(app, "org_settings",
			patch.AutodateFields(),
		)
	}

	orgsCol, err := app.FindCollectionByNameOrId("organizations")
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection("org_settings")
	collection.Fields.Add(
		&core.RelationField{
			Name:          "organization",
			CollectionId:  orgsCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.EmailField{
			Name: "billing_email",
		},
		&core.SelectField{
			Name:      "billing_plan",
			MaxSelect: 1,
			Values:    []string{"free", "starter", "pro", "enterprise"},
		},
		&core.JSONField{
			Name:    "features",
			MaxSize: 65536,
		},
		&core.JSONField{
			Name:    "notification_preferences",
			MaxSize: 65536,
		},
	)

	collection.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
	collection.AddIndex("idx_org_settings_org", true, "organization", "")

	return app.Save(collection)
}
