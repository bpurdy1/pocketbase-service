package router

import (
	"encoding/json"
	"log"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// bindInviteRoutes registers custom invite endpoints.
func (r *Router) bindInviteRoutes(e *core.ServeEvent) {
	// GET /api/invites/verify?token=... — public, returns invite details
	e.Router.GET("/api/invites/verify", func(re *core.RequestEvent) error {
		token := re.Request.URL.Query().Get("token")
		if token == "" {
			return re.JSON(400, map[string]any{"error": "token is required"})
		}

		invite, err := r.app.FindFirstRecordByFilter(
			"org_invites",
			"token = {:token}",
			dbx.Params{"token": token},
		)
		if err != nil {
			return re.JSON(404, map[string]any{"valid": false, "reason": "invite not found"})
		}

		orgName := ""
		if org, err := r.app.FindRecordById("organizations", invite.GetString("organization")); err == nil {
			orgName = org.GetString("name")
		}

		expired := false
		expiresAt := invite.GetDateTime("expires_at")
		if !expiresAt.IsZero() && expiresAt.Time().Before(time.Now()) {
			expired = true
		}

		status := invite.GetString("status")
		valid := status == "pending" && !expired

		return re.JSON(200, map[string]any{
			"valid":    valid,
			"expired":  expired,
			"status":   status,
			"email":    invite.GetString("email"),
			"role":     invite.GetString("role"),
			"org_name": orgName,
			"org_id":   invite.GetString("organization"),
		})
	})

	// POST /api/invites/accept — authenticated, accepts {"token": "..."}
	e.Router.POST("/api/invites/accept", func(re *core.RequestEvent) error {
		if re.Auth == nil {
			return re.JSON(401, map[string]any{"error": "authentication required"})
		}

		var body struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(re.Request.Body).Decode(&body); err != nil || body.Token == "" {
			return re.JSON(400, map[string]any{"error": "token is required"})
		}

		invite, err := r.app.FindFirstRecordByFilter(
			"org_invites",
			"token = {:token}",
			dbx.Params{"token": body.Token},
		)
		if err != nil {
			return re.JSON(404, map[string]any{"error": "invite not found"})
		}

		if invite.GetString("status") != "pending" {
			return re.JSON(400, map[string]any{"error": "invite is no longer pending", "status": invite.GetString("status")})
		}

		// Check expiry
		expiresAt := invite.GetDateTime("expires_at")
		if !expiresAt.IsZero() && expiresAt.Time().Before(time.Now()) {
			invite.Set("status", "expired")
			r.app.Save(invite)
			return re.JSON(400, map[string]any{"error": "invite has expired"})
		}

		// Verify email matches
		if re.Auth.GetString("email") != invite.GetString("email") {
			return re.JSON(403, map[string]any{"error": "invite email does not match your account"})
		}

		// Set status to accepted — the existing OnRecordUpdate hook
		// in invites.go will create the org_member record
		invite.Set("status", "accepted")
		if err := r.app.Save(invite); err != nil {
			log.Printf("Failed to accept invite: %v", err)
			return re.JSON(500, map[string]any{"error": "failed to accept invite"})
		}

		return re.JSON(200, map[string]any{
			"success": true,
			"org_id":  invite.GetString("organization"),
		})
	})

	// POST /api/orgs/:orgId/invites/bulk — authenticated org admin/owner
	e.Router.POST("/api/orgs/{orgId}/invites/bulk", func(re *core.RequestEvent) error {
		if re.Auth == nil {
			return re.JSON(401, map[string]any{"error": "authentication required"})
		}

		orgId := re.Request.PathValue("orgId")

		// Verify caller is an owner/admin of this org
		_, err := r.app.FindFirstRecordByFilter(
			"org_members",
			"user = {:userId} && organization = {:orgId} && (role = 'owner' || role = 'admin')",
			dbx.Params{"userId": re.Auth.Id, "orgId": orgId},
		)
		if err != nil {
			return re.JSON(403, map[string]any{"error": "you must be an org owner or admin"})
		}

		var body struct {
			Emails []string `json:"emails"`
			Role   string   `json:"role"`
		}
		if err := json.NewDecoder(re.Request.Body).Decode(&body); err != nil {
			return re.JSON(400, map[string]any{"error": "invalid request body"})
		}
		if len(body.Emails) == 0 {
			return re.JSON(400, map[string]any{"error": "emails array is required"})
		}
		if body.Role == "" {
			body.Role = "member"
		}

		invitesCol, err := r.app.FindCollectionByNameOrId("org_invites")
		if err != nil {
			return re.JSON(500, map[string]any{"error": "org_invites collection not found"})
		}

		type result struct {
			Email  string `json:"email"`
			Status string `json:"status"`
			Error  string `json:"error,omitempty"`
		}

		var results []result
		sent := 0

		for _, email := range body.Emails {
			invite := core.NewRecord(invitesCol)
			invite.Set("organization", orgId)
			invite.Set("email", email)
			invite.Set("role", body.Role)
			// token, status, expires_at, invited_by are set by RegisterInviteHooks

			// Use the request-aware save to trigger OnRecordCreateRequest hooks
			if err := r.app.Save(invite); err != nil {
				results = append(results, result{Email: email, Status: "failed", Error: err.Error()})
			} else {
				results = append(results, result{Email: email, Status: "sent"})
				sent++
			}
		}

		return re.JSON(200, map[string]any{
			"sent":    sent,
			"total":   len(body.Emails),
			"results": results,
		})
	})
}
