package admin

import (
	"log"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Sync interface {
	Sync() error
}

func RedirectAdminUI(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(
		func(e *core.ServeEvent) error {
			e.Router.GET("/admin", func(re *core.RequestEvent) error {
				return re.Redirect(http.StatusTemporaryRedirect, "/_/")
			})
			e.Router.GET("/admin/{path...}", func(re *core.RequestEvent) error {
				path := re.Request.PathValue("path")
				return re.Redirect(http.StatusTemporaryRedirect, "/_/"+path)
			})
			return e.Next()
		},
	)
}

func BindSyncFunc(app *pocketbase.PocketBase, s Sync) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// POST /api/admin/sync - requires superuser auth
		e.Router.POST("/api/admin/sync", func(re *core.RequestEvent) error {
			// Check if user is authenticated as superuser
			if re.Auth == nil || re.Auth.Collection().Name != "_superusers" {
				return re.JSON(401, map[string]any{
					"error": "unauthorized",
				})
			}

			start := time.Now()
			if err := s.Sync(); err != nil {
				log.Printf("Manual sync failed: %v", err)
				return re.JSON(500, map[string]any{
					"error":   "sync failed",
					"message": err.Error(),
				})
			}
			duration := time.Since(start)
			log.Printf("Manual sync completed in %v", duration)
			return re.JSON(200, map[string]any{
				"success":  true,
				"message":  "sync completed",
				"duration": duration.String(),
			})
		})

		return e.Next()
	})
}

// EnsureAdmin creates an admin if one doesn't exist.
// Uses OnServe because DB must be initialized before we can query records.
func EnsureAdmin(app *pocketbase.PocketBase, adminEmail, adminPass string) {
	if adminEmail == "" || adminPass == "" {
		return
	}

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		superusers, err := app.FindAllRecords("_superusers")
		if err != nil {
			log.Printf("Failed to find superusers: %v", err)
			return e.Next()
		}

		hasRealAdmin := false
		for _, su := range superusers {
			email := su.GetString("email")
			if email != "" && email != "__pbinstaller@example.com" {
				hasRealAdmin = true
				break
			}
		}

		if !hasRealAdmin {
			collection, err := app.FindCollectionByNameOrId("_superusers")
			if err != nil {
				log.Printf("Failed to find _superusers collection: %v", err)
				return e.Next()
			}

			superuser := core.NewRecord(collection)
			superuser.Set("email", adminEmail)
			superuser.SetPassword(adminPass)

			if err := app.Save(superuser); err != nil {
				log.Printf("Failed to create default superuser: %v", err)
			} else {
				log.Printf("Created default superuser: %s", adminEmail)
			}
		}

		return e.Next()
	})
}
