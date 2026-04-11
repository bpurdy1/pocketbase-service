package tenancy

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/pb/rules"
)

// OrgScoped defines how a collection is tied to an organization.
type OrgScoped struct {
	// Collection name in PocketBase
	Collection string
	// OrgField is the name of the relation field pointing to organizations (e.g. "organization")
	OrgField string
	// PublicRead allows anyone to list/view records without authentication.
	// Write operations (create/update/delete) still require org membership.
	PublicRead bool
}

// registered holds all org-scoped collections
var registered []OrgScoped

// Register adds a collection to the org-scoped tenancy system.
// Call this from your Ensure* functions before EnforceTenancy runs.
func Register(collection, orgField string) {
	registered = append(registered, OrgScoped{Collection: collection, OrgField: orgField})
}

// RegisterPublicRead adds a collection with public list/view but org-gated writes.
func RegisterPublicRead(collection, orgField string) {
	registered = append(registered, OrgScoped{Collection: collection, OrgField: orgField, PublicRead: true})
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

	writeRule := rules.WithPlatformAdmin(rules.OrgAdmin(scope.OrgField))

	if scope.PublicRead {
		collection.ListRule = rules.Ptr(rules.Public)
		collection.ViewRule = rules.Ptr(rules.Public)
	} else {
		readRule := rules.WithPlatformAdmin(rules.OrgMember(scope.OrgField))
		collection.ListRule = rules.Ptr(readRule)
		collection.ViewRule = rules.Ptr(readRule)
	}

	collection.CreateRule = rules.Ptr(writeRule)
	collection.UpdateRule = rules.Ptr(writeRule)
	collection.DeleteRule = rules.Ptr(writeRule)

	if err := app.Save(collection); err != nil {
		log.Printf("tenancy: failed to apply rules to %q: %v", scope.Collection, err)
	} else {
		log.Printf("tenancy: applied org-scoped rules to %q", scope.Collection)
	}
}
