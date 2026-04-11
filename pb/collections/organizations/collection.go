package organizations

import (
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/collections/patch"
)

func EnsureCollectionOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureCollection(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsureCollection(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("organizations")
	if existing != nil {
		return patch.Collection(app, "organizations",
			patch.AutodateFields(),
		)
	}

	collection := core.NewBaseCollection("organizations")
	collection.Fields.Add(
		&core.TextField{Name: "name", Required: true},
		&core.TextField{Name: "slug", Required: true},
		&core.URLField{Name: "website"},
		&core.TextField{Name: "phone"},
		&core.TextField{Name: "address"},
		&core.TextField{Name: "city"},
		&core.TextField{Name: "state", Max: 2},
		&core.TextField{Name: "zip_code"},
	)

	authRule := "@request.auth.id != ''"
	collection.CreateRule = &authRule

	collection.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
	collection.AddIndex("idx_organizations_slug", true, "slug", "")

	return app.Save(collection)
}
