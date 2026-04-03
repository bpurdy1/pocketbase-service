package tests_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pocketbase-server/internal/collections/organizations"
	"pocketbase-server/internal/collections/users"
)

// bootstrapApp creates a real SQLite-backed PocketBase test app in a local
// temp directory. It creates all custom collections and registers hooks.
//
// The SQLite file lives at <os.TempDir>/pb_test_<random>/pb_data/data.db
// so you can inspect it after a test run if needed.
//
// The returned cleanup func removes the temp directory.
func bootstrapApp(t *testing.T) (*pbtests.TestApp, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "pb_test_*")
	require.NoError(t, err)

	t.Logf("Test DB: %s/pb_data/data.db", dir)

	app, err := pbtests.NewTestApp(dir)
	require.NoError(t, err)

	// Phase 1: create collections in dependency order
	// (users is built-in — EnsureCollection just adds custom fields)
	require.NoError(t, users.EnsureCollection(app), "users.EnsureCollection")
	require.NoError(t, users.EnsureSettings(app), "users.EnsureSettings")
	require.NoError(t, organizations.EnsureCollection(app), "organizations.EnsureCollection")
	require.NoError(t, organizations.EnsureMembers(app), "organizations.EnsureMembers")
	require.NoError(t, organizations.EnsureOrgSettings(app), "organizations.EnsureOrgSettings")
	require.NoError(t, organizations.EnsureInvites(app), "organizations.EnsureInvites")

	// Phase 2: register hooks
	// NOTE: OnRecordCreateRequest / OnRecordUpdateRequest are HTTP-only and
	// won't fire on direct app.Save() calls — only OnRecordCreate /
	// OnRecordUpdate fire in these tests.
	users.RegisterHooks(app)
	organizations.RegisterHooks(app)
	organizations.RegisterInviteHooks(app)

	return app, func() {
		app.Cleanup()
		os.RemoveAll(dir)
	}
}

// --------------------------------------------------------------------------
// Schema tests
// --------------------------------------------------------------------------

func TestCollectionSchemas(t *testing.T) {
	app, cleanup := bootstrapApp(t)
	defer cleanup()

	t.Run("users has custom fields", func(t *testing.T) {
		col, err := app.FindCollectionByNameOrId("users")
		require.NoError(t, err)
		assert.NotNil(t, col.Fields.GetByName("phone"), "phone field")
		assert.NotNil(t, col.Fields.GetByName("deactivated"), "deactivated field")
		assert.NotNil(t, col.Fields.GetByName("role"), "role field")
		assert.NotNil(t, col.ListRule, "list rule should be set")
	})

	t.Run("settings collection", func(t *testing.T) {
		col, err := app.FindCollectionByNameOrId("settings")
		require.NoError(t, err)
		assert.NotNil(t, col.Fields.GetByName("user"), "user relation")
		assert.NotNil(t, col.Fields.GetByName("theme"), "theme field")
		assert.NotNil(t, col.Fields.GetByName("email_notifications"), "email_notifications field")
		assert.NotNil(t, col.Fields.GetByName("sms_notifications"), "sms_notifications field")
		assert.NotNil(t, col.Fields.GetByName("preferences"), "preferences field")
		assert.NotNil(t, col.ListRule, "list rule should be set")
	})

	t.Run("organizations collection", func(t *testing.T) {
		col, err := app.FindCollectionByNameOrId("organizations")
		require.NoError(t, err)
		assert.NotNil(t, col.Fields.GetByName("name"), "name field")
		assert.NotNil(t, col.Fields.GetByName("slug"), "slug field")
	})

	t.Run("org_members collection", func(t *testing.T) {
		col, err := app.FindCollectionByNameOrId("org_members")
		require.NoError(t, err)
		assert.NotNil(t, col.Fields.GetByName("user"), "user relation")
		assert.NotNil(t, col.Fields.GetByName("organization"), "organization relation")
		assert.NotNil(t, col.Fields.GetByName("role"), "role field")
		// unique index on (user, organization)
		assert.NotEmpty(t, col.GetIndex("idx_org_members_unique"), "unique index")
	})

	t.Run("org_settings collection", func(t *testing.T) {
		col, err := app.FindCollectionByNameOrId("org_settings")
		require.NoError(t, err)
		assert.NotNil(t, col.Fields.GetByName("organization"), "organization relation")
		assert.NotNil(t, col.Fields.GetByName("billing_plan"), "billing_plan field")
		assert.NotNil(t, col.Fields.GetByName("features"), "features field")
		assert.NotNil(t, col.Fields.GetByName("notification_preferences"), "notification_preferences field")
	})

	t.Run("org_invites collection", func(t *testing.T) {
		col, err := app.FindCollectionByNameOrId("org_invites")
		require.NoError(t, err)
		assert.NotNil(t, col.Fields.GetByName("organization"), "organization relation")
		assert.NotNil(t, col.Fields.GetByName("email"), "email field")
		assert.NotNil(t, col.Fields.GetByName("role"), "role field")
		assert.NotNil(t, col.Fields.GetByName("token"), "token field")
		assert.NotNil(t, col.Fields.GetByName("status"), "status field")
		assert.NotNil(t, col.Fields.GetByName("expires_at"), "expires_at field")
		assert.NotNil(t, col.Fields.GetByName("invited_by"), "invited_by field")
		// token must be unique
		assert.NotEmpty(t, col.GetIndex("idx_org_invites_token"), "token unique index")
	})
}

// --------------------------------------------------------------------------
// User signup auto-creates settings + personal org
// --------------------------------------------------------------------------

func TestUserSignupAutoCreates(t *testing.T) {
	app, cleanup := bootstrapApp(t)
	defer cleanup()

	usersCol, err := app.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	// Create a new user — the OnRecordCreate hook fires automatically.
	// Note: OnRecordCreateRequest (HTTP-only) won't fire here, so we set
	// the default role manually just like that hook would.
	user := core.NewRecord(usersCol)
	user.SetEmail("alice@example.com")
	user.SetPassword("password1234!")
	user.Set("emailVisibility", true)
	user.Set("role", "user")
	require.NoError(t, app.Save(user), "save user")

	t.Logf("Created user id=%s", user.Id)

	t.Run("settings record created", func(t *testing.T) {
		settings, err := app.FindFirstRecordByFilter(
			"settings",
			"user = {:userId}",
			dbx.Params{"userId": user.Id},
		)
		require.NoError(t, err, "settings record should exist")
		assert.Equal(t, user.Id, settings.GetString("user"))
		assert.Equal(t, "system", settings.GetString("theme"), "default theme")
		assert.True(t, settings.GetBool("email_notifications"), "email notifications default on")
	})

	t.Run("personal organization created", func(t *testing.T) {
		org, err := app.FindFirstRecordByFilter(
			"organizations",
			"slug = {:slug}",
			dbx.Params{"slug": user.Id},
		)
		require.NoError(t, err, "personal org should exist")
		t.Logf("Personal org id=%s name=%q", org.Id, org.GetString("name"))
		assert.Contains(t, org.GetString("name"), "Organization", "org name should contain 'Organization'")
	})

	t.Run("user added as org owner", func(t *testing.T) {
		// Find the personal org first
		org, err := app.FindFirstRecordByFilter(
			"organizations",
			"slug = {:slug}",
			dbx.Params{"slug": user.Id},
		)
		require.NoError(t, err)

		member, err := app.FindFirstRecordByFilter(
			"org_members",
			"user = {:userId} && organization = {:orgId}",
			dbx.Params{"userId": user.Id, "orgId": org.Id},
		)
		require.NoError(t, err, "org_member should exist")
		assert.Equal(t, "owner", member.GetString("role"), "should be owner")
	})
}

// --------------------------------------------------------------------------
// Invite flow: create invite → accept → org_member created
// --------------------------------------------------------------------------

func TestInviteFlow(t *testing.T) {
	app, cleanup := bootstrapApp(t)
	defer cleanup()

	usersCol, err := app.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	// --- Create org owner (user A) ---
	userA := core.NewRecord(usersCol)
	userA.SetEmail("owner@example.com")
	userA.SetPassword("password1234!")
	userA.Set("role", "user")
	require.NoError(t, app.Save(userA), "save owner")

	// Get the personal org auto-created for user A
	org, err := app.FindFirstRecordByFilter(
		"organizations",
		"slug = {:slug}",
		dbx.Params{"slug": userA.Id},
	)
	require.NoError(t, err, "personal org for owner should exist")
	t.Logf("Org id=%s", org.Id)

	// --- Create invited user (user B) ---
	inviteEmail := "invited@example.com"
	userB := core.NewRecord(usersCol)
	userB.SetEmail(inviteEmail)
	userB.SetPassword("password1234!")
	userB.Set("role", "user")
	require.NoError(t, app.Save(userB), "save invited user")
	t.Logf("Invited user id=%s", userB.Id)

	// --- Create the invite record ---
	// OnRecordCreateRequest won't fire (HTTP-only), so we set token/status manually.
	invitesCol, err := app.FindCollectionByNameOrId("org_invites")
	require.NoError(t, err)

	token := fmt.Sprintf("test-token-%d", time.Now().UnixNano())
	invite := core.NewRecord(invitesCol)
	invite.Set("organization", org.Id)
	invite.Set("email", inviteEmail)
	invite.Set("role", "member")
	invite.Set("token", token)
	invite.Set("status", "pending")
	invite.Set("expires_at", time.Now().Add(7*24*time.Hour).UTC().Format(time.RFC3339))
	invite.Set("invited_by", userA.Id)
	require.NoError(t, app.Save(invite), "save invite")
	t.Logf("Invite id=%s token=%s", invite.Id, token)

	t.Run("invite is pending", func(t *testing.T) {
		loaded, err := app.FindRecordById("org_invites", invite.Id)
		require.NoError(t, err)
		assert.Equal(t, "pending", loaded.GetString("status"))
	})

	// --- Verify user B is NOT yet a member ---
	t.Run("user B is not yet a member", func(t *testing.T) {
		_, err := app.FindFirstRecordByFilter(
			"org_members",
			"user = {:userId} && organization = {:orgId}",
			dbx.Params{"userId": userB.Id, "orgId": org.Id},
		)
		assert.Error(t, err, "user B should not be a member yet")
	})

	// --- Accept the invite ---
	// Load the invite fresh from DB so Original() has the old "pending" status,
	// then change to "accepted". OnRecordUpdate fires and creates the org_member.
	fresh, err := app.FindRecordById("org_invites", invite.Id)
	require.NoError(t, err)
	fresh.Set("status", "accepted")
	require.NoError(t, app.Save(fresh), "accept invite")

	// --- Assert org_member was created ---
	t.Run("user B is now an org member", func(t *testing.T) {
		member, err := app.FindFirstRecordByFilter(
			"org_members",
			"user = {:userId} && organization = {:orgId}",
			dbx.Params{"userId": userB.Id, "orgId": org.Id},
		)
		require.NoError(t, err, "org_member should be created after invite acceptance")
		assert.Equal(t, "member", member.GetString("role"), "role should be 'member'")
		t.Logf("Member id=%s role=%s", member.Id, member.GetString("role"))
	})
}
