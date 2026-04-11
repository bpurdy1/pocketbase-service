package organizations

import (
	"log"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/rules"
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

	collection.ListRule = rules.Ptr(rules.DirectOrgMember)
	collection.ViewRule = rules.Ptr(rules.DirectOrgMember)
	collection.UpdateRule = rules.Ptr(rules.DirectOrgAdmin)
	collection.DeleteRule = rules.Ptr(rules.DirectOrgOwner)

	if err := app.Save(collection); err != nil {
		log.Printf("Failed to apply organizations rules: %v", err)
	} else {
		log.Println("Applied organizations access rules")
	}
}

// ApplyOrgSettingsRules sets access rules on org_settings.
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

	collection.ListRule = rules.Ptr(rules.OrgMember("organization"))
	collection.ViewRule = rules.Ptr(rules.OrgMember("organization"))
	collection.UpdateRule = rules.Ptr(rules.OrgAdmin("organization"))
	collection.CreateRule = nil // system/hooks only
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

	collection.ListRule = rules.Ptr(rules.OrgMember("organization"))
	collection.ViewRule = rules.Ptr(rules.OrgMember("organization"))
	collection.CreateRule = rules.Ptr(rules.OrgAdmin("organization"))
	collection.UpdateRule = rules.Ptr(rules.OrgAdmin("organization"))
	collection.DeleteRule = rules.Ptr(rules.OrgAdmin("organization"))

	if err := app.Save(collection); err != nil {
		log.Printf("Failed to apply org_members rules: %v", err)
	} else {
		log.Println("Applied org_members access rules")
	}
}
