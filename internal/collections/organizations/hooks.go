package organizations

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/collections/roles"
)

// RegisterHooks sets up lifecycle hooks for organizations:
//   - Auto-create org_members entry with "owner" role when an org is created
func RegisterHooks(app *pocketbase.PocketBase) {
	app.OnRecordCreateRequest("organizations").BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		members, err := app.FindCollectionByNameOrId("org_members")
		if err != nil {
			log.Printf("org_members collection not found: %v", err)
			return nil
		}

		member := core.NewRecord(members)
		member.Set("user", e.Auth.Id)
		member.Set("organization", e.Record.Id)
		member.Set("role", roles.OrgOwner)

		if err := app.Save(member); err != nil {
			log.Printf("Failed to create org owner membership: %v", err)
		}

		return nil
	})
}
