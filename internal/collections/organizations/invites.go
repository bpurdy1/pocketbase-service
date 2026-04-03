package organizations

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"

	"pocketbase-server/internal/collections/patch"
	"pocketbase-server/internal/collections/roles"
)

// EnsureInvitesOnBeforeServe registers the org_invites collection setup on server start.
func EnsureInvitesOnBeforeServe(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureInvites(e.App); err != nil {
			return err
		}
		return e.Next()
	})
}

// EnsureInvites creates the org_invites collection if it doesn't exist.
func EnsureInvites(app core.App) error {
	existing, _ := app.FindCollectionByNameOrId("org_invites")
	if existing != nil {
		return patch.Collection(app, "org_invites",
			patch.AutodateFields(),
		)
	}

	orgsCol, err := app.FindCollectionByNameOrId("organizations")
	if err != nil {
		return err
	}

	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection("org_invites")
	collection.Fields.Add(
		&core.RelationField{
			Name:          "organization",
			CollectionId:  orgsCol.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.EmailField{
			Name:     "email",
			Required: true,
		},
		&core.SelectField{
			Name:      "role",
			Required:  true,
			MaxSelect: 1,
			Values:    roles.AllOrg,
		},
		&core.TextField{
			Name:     "token",
			Required: true,
		},
		&core.SelectField{
			Name:      "status",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"pending", "accepted", "expired", "revoked"},
		},
		&core.DateField{
			Name:     "expires_at",
			Required: true,
		},
		&core.RelationField{
			Name:         "invited_by",
			CollectionId: usersCol.Id,
			MaxSelect:    1,
		},
	)

	collection.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
	collection.AddIndex("idx_org_invites_token", true, "token", "")
	collection.AddIndex("idx_org_invites_email_org", false, "email, organization", "")

	return app.Save(collection)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RegisterInviteHooks sets up hooks for the invite lifecycle:
//   - On create: generate token, set defaults, send invite email
//   - On update: when status becomes "accepted", create the org_member record
func RegisterInviteHooks(app core.App) {
	// Before create: fill in token, status, expiry
	app.OnRecordCreateRequest("org_invites").BindFunc(func(e *core.RecordRequestEvent) error {
		e.Record.Set("token", generateToken())
		e.Record.Set("status", "pending")
		e.Record.Set("expires_at", time.Now().Add(7*24*time.Hour).UTC().Format(time.RFC3339))

		if e.Auth != nil {
			e.Record.Set("invited_by", e.Auth.Id)
		}

		return e.Next()
	})

	// After create: send the invite email
	app.OnRecordCreate("org_invites").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		email := e.Record.GetString("email")
		token := e.Record.GetString("token")
		orgId := e.Record.GetString("organization")

		// Look up org name for a nicer email
		orgName := orgId
		if org, err := app.FindRecordById("organizations", orgId); err == nil {
			orgName = org.GetString("name")
		}

		inviterName := "Someone"
		if invitedBy := e.Record.GetString("invited_by"); invitedBy != "" {
			if user, err := app.FindRecordById("users", invitedBy); err == nil {
				inviterName = user.GetString("username")
				if inviterName == "" {
					inviterName = user.GetString("email")
				}
			}
		}

		// Build accept URL — the frontend handles this route
		acceptURL := fmt.Sprintf("/invite/accept?token=%s", token)

		message := &mailer.Message{
			To:      []mail.Address{{Address: email}},
			Subject: fmt.Sprintf("You've been invited to %s", orgName),
			HTML: fmt.Sprintf(`
				<h2>You've been invited!</h2>
				<p><strong>%s</strong> has invited you to join <strong>%s</strong>.</p>
				<p>Click the link below to accept the invitation:</p>
				<p><a href="%s">Accept Invitation</a></p>
				<p>This invitation expires in 7 days.</p>
			`, inviterName, orgName, acceptURL),
			Text: fmt.Sprintf(
				"%s has invited you to join %s.\n\nAccept here: %s\n\nThis invitation expires in 7 days.",
				inviterName, orgName, acceptURL,
			),
		}

		if err := app.NewMailClient().Send(message); err != nil {
			log.Printf("Failed to send invite email to %s: %v", email, err)
		} else {
			log.Printf("Sent invite email to %s for org %s", email, orgName)
		}

		return nil
	})

	// On update: if status changed back to "pending" (resend), regenerate token + expiry and re-send email
	app.OnRecordUpdateRequest("org_invites").BindFunc(func(e *core.RecordRequestEvent) error {
		oldStatus := e.Record.Original().GetString("status")
		newStatus := e.Record.GetString("status")

		// Resend: any non-pending status being set back to pending
		if newStatus == "pending" && oldStatus != "pending" {
			e.Record.Set("token", generateToken())
			e.Record.Set("expires_at", time.Now().Add(7*24*time.Hour).UTC().Format(time.RFC3339))
		}

		return e.Next()
	})

	// After update: if status was changed to "pending" (resend), re-send the email
	app.OnRecordUpdate("org_invites").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		oldStatus := e.Record.Original().GetString("status")
		newStatus := e.Record.GetString("status")

		if newStatus == "pending" && oldStatus != "pending" {
			email := e.Record.GetString("email")
			token := e.Record.GetString("token")
			orgId := e.Record.GetString("organization")

			orgName := orgId
			if org, err := app.FindRecordById("organizations", orgId); err == nil {
				orgName = org.GetString("name")
			}

			acceptURL := fmt.Sprintf("/invite/accept?token=%s", token)

			message := &mailer.Message{
				To:      []mail.Address{{Address: email}},
				Subject: fmt.Sprintf("Reminder: You've been invited to %s", orgName),
				HTML: fmt.Sprintf(`
					<h2>Invitation Reminder</h2>
					<p>You have a pending invitation to join <strong>%s</strong>.</p>
					<p>Click the link below to accept:</p>
					<p><a href="%s">Accept Invitation</a></p>
					<p>This invitation expires in 7 days.</p>
				`, orgName, acceptURL),
				Text: fmt.Sprintf(
					"Reminder: You have a pending invitation to join %s.\n\nAccept here: %s\n\nThis invitation expires in 7 days.",
					orgName, acceptURL,
				),
			}

			if err := app.NewMailClient().Send(message); err != nil {
				log.Printf("Failed to resend invite email to %s: %v", email, err)
			} else {
				log.Printf("Resent invite email to %s for org %s", email, orgName)
			}
		}

		return nil
	})

	// On update: if status changed to "accepted", create org_member
	app.OnRecordUpdate("org_invites").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		if e.Record.GetString("status") != "accepted" {
			return nil
		}

		// Check expiry
		expiresAt := e.Record.GetDateTime("expires_at")
		if !expiresAt.IsZero() && expiresAt.Time().Before(time.Now()) {
			e.Record.Set("status", "expired")
			app.Save(e.Record)
			log.Printf("Invite %s expired", e.Record.Id)
			return nil
		}

		email := e.Record.GetString("email")
		orgId := e.Record.GetString("organization")
		role := e.Record.GetString("role")

		// Find the user by email
		user, err := app.FindAuthRecordByEmail("users", email)
		if err != nil {
			log.Printf("Invite accepted but no user found for email %s: %v", email, err)
			return nil
		}

		// Check if already a member
		existing, _ := app.FindFirstRecordByFilter(
			"org_members",
			"user = {:userId} && organization = {:orgId}",
			dbx.Params{"userId": user.Id, "orgId": orgId},
		)
		if existing != nil {
			log.Printf("User %s already a member of org %s, skipping", user.Id, orgId)
			return nil
		}

		// Create org_member
		membersCol, err := app.FindCollectionByNameOrId("org_members")
		if err != nil {
			log.Printf("org_members collection not found: %v", err)
			return nil
		}

		member := core.NewRecord(membersCol)
		member.Set("user", user.Id)
		member.Set("organization", orgId)
		member.Set("role", role)

		if err := app.Save(member); err != nil {
			log.Printf("Failed to create org_member on invite accept: %v", err)
		} else {
			log.Printf("Added user %s to org %s with role %s via invite", user.Id, orgId, role)
		}

		return nil
	})
}

// ApplyInviteRules sets access rules on org_invites.
// Org owners/admins can create and manage invites.
func ApplyInviteRules(app core.App) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		collection, err := app.FindCollectionByNameOrId("org_invites")
		if err != nil || collection.ListRule != nil {
			return e.Next()
		}

		// Org admins/owners can list and view invites for their org
		adminRule := "@request.auth.id != '' && organization.id ?= @collection.org_members.organization && @request.auth.id ?= @collection.org_members.user && (@collection.org_members.role = 'owner' || @collection.org_members.role = 'admin')"
		collection.ListRule = &adminRule
		collection.ViewRule = &adminRule
		collection.CreateRule = &adminRule
		collection.UpdateRule = &adminRule
		collection.DeleteRule = &adminRule

		if err := app.Save(collection); err != nil {
			log.Printf("Failed to apply org_invites rules: %v", err)
		} else {
			log.Println("Applied org_invites access rules")
		}

		return e.Next()
	})
}
