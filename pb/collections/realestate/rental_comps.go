package realestate

import (
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/collections/patch"
	"pocketbase-server/pb/rules"
)

func EnsureRentalCompsOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureRentalComps(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsureRentalComps(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("rental_comps")
	if existing != nil {
		return patch.Collection(app, "rental_comps",
			patch.AutodateFields(),
			patch.Index("idx_rental_comps_created", false, "created"),
		)
	}

	photosCol, err := app.FindCollectionByNameOrId("photos")
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection("rental_comps")

	collection.ListRule = rules.Ptr(rules.Public)
	collection.ViewRule = rules.Ptr(rules.Public)
	collection.CreateRule = rules.Ptr(rules.AuthOnly)
	collection.UpdateRule = rules.Ptr(rules.AuthOnly)
	collection.DeleteRule = rules.Ptr(rules.AuthOnly)

	collection.Fields.Add(
		&core.TextField{Name: "address", Required: true},
		&core.NumberField{Name: "lat"},
		&core.NumberField{Name: "lng"},
		&core.NumberField{Name: "price"},
		&core.NumberField{Name: "sqft"},
		&core.NumberField{Name: "price_per_sqft"},
		&core.TextField{Name: "building"},
		&core.URLField{Name: "website"},
		// Link to the original listing
		&core.URLField{Name: "listing_url"},
		// Google Maps or similar
		&core.URLField{Name: "address_url"},
		// e.g. Zillow, Redfin, CoStar, LoopNet
		&core.TextField{Name: "data_source"},
		// Relation to photos collection — supports multiple photos
		&core.RelationField{
			Name:         "photos",
			CollectionId: photosCol.Id,
			MaxSelect:    20,
		},
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)

	collection.AddIndex("idx_rental_comps_created", false, "created", "")

	return app.Save(collection)
}
