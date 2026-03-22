package roles

// Platform-level roles (on the user record itself)
const (
	User  = "user"
	Agent = "agent"
	Admin = "admin"
)

// AllPlatform are the valid values for the users.role field.
var AllPlatform = []string{User, Agent, Admin}

// Organization-level roles (on the org_members join table)
const (
	OrgOwner  = "owner"
	OrgAdmin  = "admin"
	OrgMember = "member"
)

// AllOrg are the valid values for the org_members.role field.
var AllOrg = []string{OrgOwner, OrgAdmin, OrgMember}
