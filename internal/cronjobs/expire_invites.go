package cronjobs

import (
	"log"
	"time"

	"github.com/pocketbase/pocketbase"
)

// RegisterExpireInvites registers an hourly cron job that marks
// pending invites past their expires_at as "expired".
func RegisterExpireInvites(app *pocketbase.PocketBase) {
	app.Cron().MustAdd("expire_invites", "0 * * * *", func() {
		now := time.Now().UTC().Format("2006-01-02 15:04:05.000Z")

		records, err := app.FindRecordsByFilter(
			"org_invites",
			"status = 'pending' && expires_at < {:now}",
			"",
			0, 0,
			map[string]any{"now": now},
		)
		if err != nil {
			log.Printf("expire_invites: failed to query: %v", err)
			return
		}

		if len(records) == 0 {
			return
		}

		for _, record := range records {
			record.Set("status", "expired")
			if err := app.Save(record); err != nil {
				log.Printf("expire_invites: failed to expire invite %s: %v", record.Id, err)
			}
		}

		log.Printf("expire_invites: expired %d invite(s)", len(records))
	})
}
