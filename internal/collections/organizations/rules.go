package organizations

import (
	"log"

	"github.com/pocketbase/pocketbase/core"
)

// ApplyRules sets access rules on organizations and org_members.
// Must run after both collections have been created (Phase 2).
func ApplyRules(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		applyOrganizationsRules(app)
		applyOrgMembersRules(app)
		return e.Next()
	})
}

func applyOrganizationsRules(app core.App) {
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

// ApplyOrgSettingsRules sets access rules on org_settings.
// Owners/admins can read+write, members can read-only.
func ApplyOrgSettingsRules(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		applyOrgSettingsRules(app)
		return e.Next()
	})
}

func applyOrgSettingsRules(app core.App) {
	collection, err := app.FindCollectionByNameOrId("org_settings")
	if err != nil || collection.ListRule != nil {
		return
	}

	// Members can view settings for their org
	memberRule := "@request.auth.id != '' && organization.id ?= @collection.org_members.organization && @request.auth.id ?= @collection.org_members.user"
	collection.ListRule = &memberRule
	collection.ViewRule = &memberRule

	// Owners/admins can update settings
	adminRule := "@request.auth.id != '' && organization.id ?= @collection.org_members.organization && @request.auth.id ?= @collection.org_members.user && (@collection.org_members.role = 'owner' || @collection.org_members.role = 'admin')"
	collection.UpdateRule = &adminRule

	// Only system/hooks create settings (auto-created on org creation)
	collection.CreateRule = nil
	collection.DeleteRule = nil

	if err := app.Save(collection); err != nil {
		log.Printf("Failed to apply org_settings rules: %v", err)
	} else {
		log.Println("Applied org_settings access rules")
	}
}

func applyOrgMembersRules(app core.App) {
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
