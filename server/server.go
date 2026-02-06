package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/tursodatabase/go-libsql"

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
	app       *pocketbase.PocketBase
	cfg       *Config
	connector *libsql.Connector
}

type Option func(*pocketbase.Config)

func newLibsqlConnection(cfg *Config) (*libsql.Connector, error) {
	if cfg.LibSQLInterval == 0 {
		cfg.LibSQLInterval = time.Minute * 5
	}

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	localDBPath := filepath.Join(cfg.DataDir, "data.db")

	// Build connector options
	opts := []libsql.Option{
		libsql.WithSyncInterval(cfg.LibSQLInterval),
	}

	// Only add auth token if provided (not needed for local libSQL)
	if cfg.LibSQLToken != "" {
		opts = append(opts, libsql.WithAuthToken(cfg.LibSQLToken))
	}

	if cfg.LibSQLURL != "" {
		fmt.Println(localDBPath, cfg.LibSQLURL)
		fmt.Println(opts)
		return libsql.NewEmbeddedReplicaConnector(localDBPath, cfg.LibSQLURL, opts...)
	}
	return libsql.NewEmbeddedReplicaConnector(localDBPath, cfg.LibSQLURL, opts...)
}

func New() (*Server, error) {
	cfg := NewConfig()

	connector, err := newLibsqlConnection(cfg)
	if err != nil {
		panic(err)
	}
	sqlDB := sql.OpenDB(connector)

	var pbcfg = pocketbase.Config{
		DefaultDataDir: cfg.DataDir,
		DefaultDev:     cfg.Dev,
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			// Use libSQL connector for the main data.db
			if strings.HasSuffix(dbPath, "data.db") {
				return dbx.NewFromDB(sqlDB, "sqlite"), nil
			}
			// Use default SQLite for auxiliary databases (logs, etc.)
			return core.DefaultDBConnect(dbPath)
		},
	}

	app := pocketbase.NewWithConfig(pbcfg)

	s := &Server{
		app:       app,
		cfg:       cfg,
		connector: connector,
	}

	router := router.NewRouter(s.App())
	router.RegisterBind()
	admin.BindSyncFunc(s.App(), s)
	admin.RedirectAdminUI(s.App())
	admin.EnsureAdmin(s.App(), cfg.AdminEmail, cfg.AdminPass)

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
	if s.connector != nil {
		_, err := s.connector.Sync()
		return err
	}
	return nil
}
func (s *Server) Close() error {
	if s.connector != nil {
		return s.connector.Close()
	}
	return nil
}
