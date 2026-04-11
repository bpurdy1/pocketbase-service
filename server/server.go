package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/cronjobs"
	"pocketbase-server/internal/database"
	"pocketbase-server/internal/logging"
	"pocketbase-server/pb/collections/auth"
	"pocketbase-server/pb/collections/notifications"
	"pocketbase-server/pb/collections/organizations"
	"pocketbase-server/pb/collections/photos"
	"pocketbase-server/pb/collections/realestate"
	"pocketbase-server/pb/collections/tenancy"
	"pocketbase-server/pb/collections/users"
	"pocketbase-server/server/admin"
	"pocketbase-server/server/router"
	"pocketbase-server/service"
)

type Config struct {
	Addr       string `env:"PB_ADDR" envDefault:"0.0.0.0:8090"`
	DataDir    string `env:"PB_DATA_DIR" envDefault:"./db/pb_data"`
	Dev        bool   `env:"PB_DEV" envDefault:"false"`
	AdminEmail string `env:"PB_ADMIN_EMAIL" envDefault:""`
	AdminPass  string `env:"PB_ADMIN_PASS" envDefault:""`

	// e.g., "http://localhost:8080" for local, "libsql://xxx.turso.io" for cloud
	LibSQLURL      string        `env:"LIBSQL_URL" envDefault:"http://localhost:8080"`
	LibSQLToken    string        `env:"LIBSQL_AUTH_TOKEN" envDefault:""`
	LibSQLInterval time.Duration `env:"LIBSQL_SYNC_INTERVAL" envDefault:"30s"`

	// S3 file storage (optional — leave blank to use local disk)
	S3Bucket         string `env:"S3_BUCKET" envDefault:""`
	S3Region         string `env:"S3_REGION" envDefault:""`
	S3Endpoint       string `env:"S3_ENDPOINT" envDefault:""`
	S3AccessKey      string `env:"S3_ACCESS_KEY" envDefault:""`
	S3Secret         string `env:"S3_SECRET" envDefault:""`
	S3ForcePathStyle bool   `env:"S3_FORCE_PATH_STYLE" envDefault:"false"`
}

func (cfg *Config) httpFlag() string {
	return fmt.Sprintf("--http=%s", cfg.Addr)
}

func NewConfig() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}

type Server struct {
	app  *pocketbase.PocketBase
	cfg  *Config
	conn *database.LibSQLConnection
}

type Option func(*pocketbase.Config)

func New() (*Server, error) {
	cfg := NewConfig()
	conn, err := database.NewLibSQLConnection(
		&database.LibSQLConfig{
			DataDir:  cfg.DataDir,
			URL:      cfg.LibSQLURL,
			Token:    cfg.LibSQLToken,
			Interval: cfg.LibSQLInterval,
		},
	)
	if err != nil {
		panic(err)
	}

	var pbcfg = pocketbase.Config{
		DefaultDataDir: cfg.DataDir,
		DefaultDev:     cfg.Dev,
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			// Use libSQL connector for the main data.db
			if strings.HasSuffix(dbPath, "data.db") {
				return dbx.NewFromDB(conn.DB, "sqlite"), nil
			}
			// Use default SQLite for auxiliary databases (logs, etc.)
			return core.DefaultDBConnect(dbPath)
		},
	}

	app := pocketbase.NewWithConfig(pbcfg)

	s := &Server{
		app:  app,
		cfg:  cfg,
		conn: conn,
	}

	// Request logging middleware
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		e.Router.BindFunc(func(re *core.RequestEvent) error {
			logging.Infof("%s %s", re.Request.Method, re.Request.URL.Path)
			return re.Next()
		})
		return e.Next()
	})

	router := router.NewRouter(s.App())
	router.RegisterBind()
	admin.BindSyncFunc(s.App(), s)
	admin.RedirectAdminUI(s.App())
	admin.EnsureAdmin(s.App(), cfg.AdminEmail, cfg.AdminPass)

	// S3 file storage (only applied if S3_BUCKET is set)
	if cfg.S3Bucket != "" {
		app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			settings := app.Settings()
			settings.S3.Enabled = true
			settings.S3.Bucket = cfg.S3Bucket
			settings.S3.Region = cfg.S3Region
			settings.S3.Endpoint = cfg.S3Endpoint
			settings.S3.AccessKey = cfg.S3AccessKey
			settings.S3.Secret = cfg.S3Secret
			settings.S3.ForcePathStyle = cfg.S3ForcePathStyle
			if err := app.Save(settings); err != nil {
				return fmt.Errorf("failed to configure S3: %w", err)
			}
			return e.Next()
		})
	}

	// Phase 1: Create collections (no cross-collection rules)
	users.EnsureCollectionOnBeforeServe(s.App())
	users.EnsureSettingsOnBeforeServe(s.App())
	users.RegisterHooks(s.App())
	organizations.EnsureCollectionOnBeforeServe(s.App())
	organizations.EnsureMembersOnBeforeServe(s.App())
	organizations.EnsureOrgSettingsOnBeforeServe(s.App())
	organizations.EnsureInvitesOnBeforeServe(s.App())
	organizations.RegisterHooks(s.App())
	organizations.RegisterInviteHooks(s.App())
	notifications.EnsureCollectionOnBeforeServe(s.App())
	notifications.RegisterHooks(s.App())
	photos.EnsureCollectionOnBeforeServe(s.App())
	realestate.EnsurePropertiesOnBeforeServe(s.App())
	realestate.EnsurePropertyDetailsOnBeforeServe(s.App())
	realestate.EnsurePropertySaleHistoryOnBeforeServe(s.App())
	realestate.EnsurePropertyTaxHistoryOnBeforeServe(s.App())
	realestate.EnsurePropertyContactsOnBeforeServe(s.App())
	realestate.EnsureRentalCompsOnBeforeServe(s.App())
	realestate.EnsureSavedPropertiesOnBeforeServe(s.App())
	realestate.EnsureSavedPropertyHistoryOnBeforeServe(s.App())
	realestate.RegisterSavedPropertyHooks(s.App())

	// Phase 2: Apply access rules (all collections now exist)
	organizations.ApplyRules(s.App())
	organizations.ApplyInviteRules(s.App())
	organizations.ApplyOrgSettingsRules(s.App())
	tenancy.EnforceTenancy(s.App())
	auth.EnsureOAuth2Providers(s.App())

	// Cron jobs
	cronjobs.RegisterExpireInvites(s.App())

	service.NewService(s.App())

	return s, nil
}

func (s *Server) Start() error {
	args := []string{"pocketbase", "serve", s.cfg.httpFlag()}
	if s.cfg.Dev {
		args = append(args, "--dev")
	}
	os.Args = args

	return s.app.Start()
}

func (s *Server) App() *pocketbase.PocketBase {
	return s.app
}
func (s *Server) Config() *Config {
	return s.cfg
}

type RouteHandler func(e *core.RequestEvent) error

func (s *Server) AddRoute(method, path string, handler RouteHandler) {
	s.app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		switch method {
		case http.MethodGet:
			e.Router.GET(path, handler)
		case http.MethodPost:
			e.Router.POST(path, handler)
		case http.MethodPut:
			e.Router.PUT(path, handler)
		case http.MethodPatch:
			e.Router.PATCH(path, handler)
		case http.MethodDelete:
			e.Router.DELETE(path, handler)
		}
		return e.Next()
	})
}

// AddHTTPHandler wraps a standard http.HandlerFunc for use with PocketBase
func (s *Server) AddHTTPHandler(method, path string, handler http.HandlerFunc) {
	s.AddRoute(method, path, func(e *core.RequestEvent) error {
		handler(e.Response, e.Request)
		return nil
	})
}

func (s *Server) Sync() error {
	return s.conn.Sync()
}
func (s *Server) Close() error {
	return s.conn.Close()
}
