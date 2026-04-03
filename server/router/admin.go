package router

import (
	"encoding/json"
	"log"

	"github.com/pocketbase/pocketbase/core"
)

// bindAdminRoutes registers admin-only API endpoints.
func (r *Router) bindAdminRoutes(e *core.ServeEvent) {
	// POST /api/admin/users — platform admin creates a new user
	e.Router.POST("/api/admin/users", func(re *core.RequestEvent) error {
		if re.Auth == nil {
			return re.JSON(401, map[string]any{"error": "authentication required"})
		}

		// Must be a platform admin or superuser
		isSuperuser := re.Auth.Collection().Name == "_superusers"
		isAdmin := re.Auth.GetString("role") == "admin"
		if !isSuperuser && !isAdmin {
			return re.JSON(403, map[string]any{"error": "platform admin access required"})
		}

		var body struct {
			Email          string `json:"email"`
			Password       string `json:"password"`
			Username       string `json:"username"`
			Phone          string `json:"phone"`
			Role           string `json:"role"`
			OrganizationId string `json:"organization_id"`
			OrgRole        string `json:"org_role"`
		}
		if err := json.NewDecoder(re.Request.Body).Decode(&body); err != nil {
			return re.JSON(400, map[string]any{"error": "invalid request body"})
		}
		if body.Email == "" || body.Password == "" {
			return re.JSON(400, map[string]any{"error": "email and password are required"})
		}

		usersCol, err := r.app.FindCollectionByNameOrId("users")
		if err != nil {
			return re.JSON(500, map[string]any{"error": "users collection not found"})
		}

		user := core.NewRecord(usersCol)
		user.Set("email", body.Email)
		user.SetPassword(body.Password)
		user.Set("verified", true)

		if body.Username != "" {
			user.Set("username", body.Username)
		}
		if body.Phone != "" {
			user.Set("phone", body.Phone)
		}
		if body.Role != "" {
			user.Set("role", body.Role)
		}

		// Save triggers OnRecordCreate hooks (auto-create settings + personal org)
		if err := r.app.Save(user); err != nil {
			log.Printf("Admin user creation failed: %v", err)
			return re.JSON(400, map[string]any{"error": err.Error()})
		}

		response := map[string]any{
			"id":    user.Id,
			"email": user.GetString("email"),
			"role":  user.GetString("role"),
		}

		// Optionally add user to an organization
		if body.OrganizationId != "" {
			orgRole := body.OrgRole
			if orgRole == "" {
				orgRole = "member"
			}

			membersCol, err := r.app.FindCollectionByNameOrId("org_members")
			if err == nil {
				member := core.NewRecord(membersCol)
				member.Set("user", user.Id)
				member.Set("organization", body.OrganizationId)
				member.Set("role", orgRole)

				if err := r.app.Save(member); err != nil {
					log.Printf("Failed to add admin-created user to org: %v", err)
					response["org_warning"] = "user created but failed to add to organization: " + err.Error()
				} else {
					response["organization_id"] = body.OrganizationId
					response["org_role"] = orgRole
				}
			}
		}

		return re.JSON(201, response)
	})
}
