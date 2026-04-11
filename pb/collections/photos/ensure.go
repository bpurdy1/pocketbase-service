package photos

import (
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/collections/patch"
	"pocketbase-server/pb/rules"
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
	existing, _ := app.FindCollectionByNameOrId("photos")
	if existing != nil {
		return patch.Collection(app, "photos",
			patch.AutodateFields(),
		)
	}

	collection := core.NewBaseCollection("photos")

	collection.ListRule = rules.Ptr(rules.Public)
	collection.ViewRule = rules.Ptr(rules.Public)
	collection.CreateRule = rules.Ptr(rules.AuthOnly)
	collection.UpdateRule = rules.Ptr(rules.AuthOnly)
	collection.DeleteRule = rules.Ptr(rules.AuthOnly)

	collection.Fields.Add(
		// The actual file — PocketBase serves this from local disk or S3 transparently
		&core.FileField{
			Name:      "file",
			Required:  true,
			MaxSelect: 1,
			MaxSize:   15 * 1024 * 1024, // 15MB
			MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
		},
		&core.TextField{Name: "alt_text"},
		// Original URL if the photo was scraped from a listing site
		&core.URLField{Name: "source_url"},
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)

	collection.AddIndex("idx_photos_created", false, "created", "")

	return app.Save(collection)
}
