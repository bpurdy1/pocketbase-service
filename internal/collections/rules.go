package collections

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ApplyOrgRules sets access rules on organizations and org_members.
// Must run after both collections have been created.
func ApplyOrgRules(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		applyOrganizationsRules(app)
		applyOrgMembersRules(app)
		return e.Next()
	})
}

func applyOrganizationsRules(app *pocketbase.PocketBase) {
	collection, err := app.FindCollectionByNameOrId("organizations")
	if err != nil || collection.ListRule != nil {
		return
	}

	memberRule := "@request.auth.id != '' && @request.auth.id ?= @collection.org_members.user && id ?= @collection.org_members.organization"
	collection.ListRule = &memberRule
	collection.ViewRule = &memberRule

	adminRule := "@request.auth.id != '' && @request.auth.id ?= @collection.org_members.user && id ?= @collection.org_members.organization && (@collection.org_members.role = 'owner' || @collection.org_members.role = 'admin')"
	collection.UpdateRule = &adminRule

	ownerRule := "@request.auth.id != '' && @request.auth.id ?= @collection.org_members.user && id ?= @collection.org_members.organization && @collection.org_members.role = 'owner'"
	collection.DeleteRule = &ownerRule

	if err := app.Save(collection); err != nil {
		log.Printf("Failed to apply organizations rules: %v", err)
	} else {
		log.Println("Applied organizations access rules")
	}
}

func applyOrgMembersRules(app *pocketbase.PocketBase) {
	collection, err := app.FindCollectionByNameOrId("org_members")
	if err != nil || collection.ListRule != nil {
		return
	}

	memberRule := "@request.auth.id != '' && organization.id ?= @collection.org_members.organization && @request.auth.id ?= @collection.org_members.user"
	collection.ListRule = &memberRule
	collection.ViewRule = &memberRule

	adminRule := "@request.auth.id != '' && organization.id ?= @collection.org_members.organization && @request.auth.id ?= @collection.org_members.user && (@collection.org_members.role = 'owner' || @collection.org_members.role = 'admin')"
	collection.CreateRule = &adminRule
	collection.UpdateRule = &adminRule
	collection.DeleteRule = &adminRule

	if err := app.Save(collection); err != nil {
		log.Printf("Failed to apply org_members rules: %v", err)
	} else {
		log.Println("Applied org_members access rules")
	}
}
