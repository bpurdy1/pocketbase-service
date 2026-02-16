package collections

// Platform-level roles (on the user record itself)
const (
	RoleUser  = "user"
	RoleAgent = "agent"
	RoleAdmin = "admin"
)

// AllPlatformRoles are the valid values for the users.role field.
var AllPlatformRoles = []string{RoleUser, RoleAgent, RoleAdmin}

// Organization-level roles (on the org_members join table)
const (
	OrgRoleOwner  = "owner"
	OrgRoleAdmin  = "admin"
	OrgRoleMember = "member"
)

// AllOrgRoles are the valid values for the org_members.role field.
var AllOrgRoles = []string{OrgRoleOwner, OrgRoleAdmin, OrgRoleMember}
