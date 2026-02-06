package pb

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type PocketBaseConfig struct {
	Dev        bool   `env:"PB_DEV" envDefault:"false"`
	Enabled    bool   `env:"PB_ENABLED" envDefault:"true"`
	AdminEmail string `env:"PB_ADMIN_EMAIL" envDefault:""`
	AdminPass  string `env:"PB_ADMIN_PASS" envDefault:""`
	Addr       string `env:"PB_ADDR" envDefault:"0.0.0.0:8080"`
	DataDir    string `env:"PB_DATA_DIR" envDefault:"./db/pb_data"`
}

func (cfg *PocketBaseConfig) HttpFlag() string {
	fmt.Println(cfg.Addr)
	return fmt.Sprintf("--http=%s", cfg.Addr)
}

func NewConfig() *PocketBaseConfig {
	cfg := &PocketBaseConfig{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
	return cfg
}

// PocketBaseConfig holds PocketBase configuration

// PocketBaseGroup handles PocketBase integration
type PocketBaseGroup struct {
	app     *pocketbase.PocketBase
	cfg     *PocketBaseConfig
	enabled bool
}

// NewPocketBaseGroup creates a new PocketBase handler group
func NewPocketBaseGroup(cfg *PocketBaseConfig) *PocketBaseGroup {
	if cfg == nil {
		cfg = &PocketBaseConfig{
			DataDir: "./db/pb_data",
			Dev:     false,
			Enabled: false,
		}
	}

	if !cfg.Enabled {
		return &PocketBaseGroup{cfg: cfg, enabled: false}
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: cfg.DataDir,
		DefaultDev:     cfg.Dev,
	})

	g := &PocketBaseGroup{
		app:     app,
		cfg:     cfg,
		enabled: cfg.Enabled,
	}

	g.app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		e.Router.GET("/admin", func(re *core.RequestEvent) error {
			return re.Redirect(http.StatusTemporaryRedirect, "/_/")
		})
		e.Router.GET("/admin/{path...}", func(re *core.RequestEvent) error {
			path := re.Request.PathValue("path")
			return re.Redirect(http.StatusTemporaryRedirect, "/_/"+path)
		})
		return e.Next()
	})

	return g
}

func (g *PocketBaseGroup) Start() error {
	if !g.enabled || g.app == nil {
		return nil
	}

	if g.cfg.AdminEmail != "" && g.cfg.AdminPass != "" {
		g.app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			superusers, err := g.app.FindAllRecords("_superusers")
			if err != nil {
				fmt.Println("Failed to find superusers:", err)
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
				// Create default superuser
				collection, err := g.app.FindCollectionByNameOrId("_superusers")
				if err != nil {
					log.Printf("Failed to find _superusers collection: %v", err)
					return e.Next()
				}

				superuser := core.NewRecord(collection)
				superuser.Set("email", g.cfg.AdminEmail)
				superuser.SetPassword(g.cfg.AdminPass)

				if err := g.app.Save(superuser); err != nil {
					log.Printf("Failed to create default superuser: %v", err)
				} else {
					log.Printf("Created default superuser: %s", g.cfg.AdminEmail)
				}
			}

			return e.Next()
		})
	}

	// Set args to simulate "serve" command
	args := []string{"pocketbase", "serve", g.cfg.HttpFlag()}
	if g.cfg.Dev {
		args = append(args, "--dev")
	}
	os.Args = args

	return g.app.Start()
}

func (g *PocketBaseGroup) App() *pocketbase.PocketBase {
	return g.app
}

func (g *PocketBaseGroup) RegisterRoutes(mux *http.ServeMux) {
	// PocketBase runs on :8090, not on the main mux
}

func (g *PocketBaseGroup) IsEnabled() bool {
	return g.enabled
}
