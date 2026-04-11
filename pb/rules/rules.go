// Package rules provides reusable PocketBase access rule strings and builders.
// Rules are plain strings — use Ptr() when assigning to collection rule fields.
package rules

import "fmt"

// ---- Primitives ----

// Public allows anyone to access a resource with no authentication.
const Public = ""

// AuthOnly requires a valid authenticated session.
const AuthOnly = "@request.auth.id != ''"

// SystemOnly blocks all API access (server-side only via app.Save).
// Assign nil directly to a collection rule field for this behaviour.
// This constant is provided for documentation purposes only.
const SystemOnly = "" // use nil pointer on the collection field

// PlatformAdmin matches users with the platform-level admin role.
const PlatformAdmin = "@request.auth.role = 'admin'"

// ---- User / record ownership ----

// OwnRecord returns a rule matching users who own a record via a given field.
//
//	OwnRecord("user")  →  "@request.auth.id = user || @request.auth.role = 'admin'"
func OwnRecord(field string) string {
	return fmt.Sprintf("(@request.auth.id = %s || @request.auth.role = 'admin')", field)
}

// OwnUser is a shorthand for records where the PK is the user's own ID.
//
//	"@request.auth.id = id || @request.auth.role = 'admin'"
const OwnUser = "(@request.auth.id = id || @request.auth.role = 'admin')"

// RecipientOnly restricts list/view to records where recipient = the caller.
//
//	Used for notifications and similar per-user inboxes.
func RecipientOnly(recipientField string) string {
	return fmt.Sprintf("@request.auth.id != '' && %s = @request.auth.id", recipientField)
}

// ---- Org membership ----

// OrgMember returns a rule allowing any member of the record's org.
// orgField is the collection field that points to the organizations collection.
//
//	OrgMember("organization")
func OrgMember(orgField string) string {
	return fmt.Sprintf(
		"@request.auth.id != '' && %s.id ?= @collection.org_members.organization && @request.auth.id ?= @collection.org_members.user",
		orgField,
	)
}

// OrgAdmin returns a rule allowing only org owners and admins.
func OrgAdmin(orgField string) string {
	return OrgMember(orgField) + " && (@collection.org_members.role = 'owner' || @collection.org_members.role = 'admin')"
}

// OrgOwner returns a rule allowing only the org owner.
func OrgOwner(orgField string) string {
	return OrgMember(orgField) + " && @collection.org_members.role = 'owner'"
}

// WithPlatformAdmin wraps any rule so platform admins always bypass it.
//
//	WithPlatformAdmin(OrgMember("organization"))
//	→  "(@request.auth.role = 'admin') || (<orgMemberRule>)"
func WithPlatformAdmin(rule string) string {
	return fmt.Sprintf("(%s) || (%s)", PlatformAdmin, rule)
}

// ---- Direct org collection rules (for the organizations table itself) ----

// DirectOrgMember matches users who are members of the org record being accessed
// (used on the organizations collection where the PK is the org ID).
const DirectOrgMember = "@request.auth.id != '' && @request.auth.id ?= @collection.org_members.user && id ?= @collection.org_members.organization"

// DirectOrgAdmin matches org owners/admins on the organizations collection.
const DirectOrgAdmin = DirectOrgMember + " && (@collection.org_members.role = 'owner' || @collection.org_members.role = 'admin')"

// DirectOrgOwner matches only the org owner on the organizations collection.
const DirectOrgOwner = DirectOrgMember + " && @collection.org_members.role = 'owner'"

// ---- Helpers ----

// Ptr returns a pointer to s, required when assigning to PocketBase rule fields.
// Pass nil directly when you want SystemOnly (no API access).
func Ptr(s string) *string {
	return &s
}
