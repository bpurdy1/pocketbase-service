package tenancy

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// OrgScoped defines how a collection is tied to an organization.
type OrgScoped struct {
	// Collection name in PocketBase
	Collection string
	// OrgField is the name of the relation field pointing to organizations (e.g. "organization")
	OrgField string
}

// registered holds all org-scoped collections
var registered []OrgScoped

// Register adds a collection to the org-scoped tenancy system.
// Call this from your Ensure* functions before EnforceTenancy runs.
func Register(collection, orgField string) {
	registered = append(registered, OrgScoped{Collection: collection, OrgField: orgField})
}

// EnforceTenancy auto-applies org-scoped access rules to all registered collections.
// Must run in Phase 2 (after all collections are created).
//
// Rules applied:
//   - List/View: user must be a member of the record's org
//   - Create/Update/Delete: user must be an owner or admin of the record's org
//   - Platform admins (role="admin") bypass via @request.auth.role = "admin"
func EnforceTenancy(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		for _, scope := range registered {
			applyOrgScopedRules(app, scope)
		}
		return e.Next()
	})
}

func applyOrgScopedRules(app *pocketbase.PocketBase, scope OrgScoped) {
	collection, err := app.FindCollectionByNameOrId(scope.Collection)
	if err != nil {
		log.Printf("tenancy: collection %q not found, skipping rules", scope.Collection)
		return
	}

	// Skip if rules are already set (e.g. configured via admin UI)
	if collection.ListRule != nil {
		return
	}

	orgField := scope.OrgField

	// Members of the org can list and view
	memberRule := "@request.auth.id != '' && " +
		orgField + ".id ?= @collection.org_members.organization && " +
		"@request.auth.id ?= @collection.org_members.user"

	// Owners and admins of the org can create, update, delete
	adminRule := memberRule + " && " +
		"(@collection.org_members.role = 'owner' || @collection.org_members.role = 'admin')"

	// Platform admins bypass all rules
	memberRuleWithAdmin := "(@request.auth.role = 'admin') || (" + memberRule + ")"
	adminRuleWithAdmin := "(@request.auth.role = 'admin') || (" + adminRule + ")"

	collection.ListRule = &memberRuleWithAdmin
	collection.ViewRule = &memberRuleWithAdmin
	collection.CreateRule = &adminRuleWithAdmin
	collection.UpdateRule = &adminRuleWithAdmin
	collection.DeleteRule = &adminRuleWithAdmin

	if err := app.Save(collection); err != nil {
		log.Printf("tenancy: failed to apply rules to %q: %v", scope.Collection, err)
	} else {
		log.Printf("tenancy: applied org-scoped rules to %q", scope.Collection)
	}
}
