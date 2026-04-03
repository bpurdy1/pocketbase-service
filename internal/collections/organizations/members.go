package organizations

import (
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/collections/patch"
	"pocketbase-server/internal/collections/roles"
)

func EnsureMembersOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureMembers(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

func EnsureMembers(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("org_members")
	if existing != nil {
		return patch.Collection(app, "org_members",
			patch.AutodateFields(),
			patch.Index("idx_org_members_created", false, "created"),
		)
	}

	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}
	orgsCol, err := app.FindCollectionByNameOrId("organizations")
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection("org_members")
	collection.Fields.Add(
		&core.RelationField{
			Name:          "user",
			CollectionId:  usersCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.RelationField{
			Name:          "organization",
			CollectionId:  orgsCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.SelectField{
			Name:      "role",
			Required:  true,
			MaxSelect: 1,
			Values:    roles.AllOrg,
		},
	)

	collection.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
	collection.AddIndex("idx_org_members_unique", true, "user, organization", "")
	collection.AddIndex("idx_org_members_org", false, "organization", "")
	collection.AddIndex("idx_org_members_created", false, "created", "")

	return app.Save(collection)
}
