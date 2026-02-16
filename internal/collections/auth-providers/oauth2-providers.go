package authproviders

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type AuthProvidersConfig struct {
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
	GithubClientID     string `env:"GITHUB_CLIENT_ID"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET"`
}

func NewAuthProvidersConfig() AuthProvidersConfig {
	var cfg AuthProvidersConfig
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return cfg
}

// EnsureOAuth2Providers configures OAuth2 providers on the users auth collection.
func EnsureOAuth2Providers(app *pocketbase.PocketBase) {
	cfg := NewAuthProvidersConfig()

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			log.Printf("Failed to find users collection for OAuth2: %v", err)
			return e.Next()
		}

		var providers []core.OAuth2ProviderConfig

		// Preserve any existing providers configured via admin UI
		providers = append(providers, users.OAuth2.Providers...)

		// Google OAuth2
		if cfg.GoogleClientID != "" {
			if _, exists := users.OAuth2.GetProviderConfig("google"); !exists {
				providers = append(providers, core.OAuth2ProviderConfig{
					Name:         "google",
					ClientId:     cfg.GoogleClientID,
					ClientSecret: cfg.GoogleClientSecret,
				})
			}
		}

		// GitHub OAuth2
		if cfg.GithubClientID != "" {
			if _, exists := users.OAuth2.GetProviderConfig("github"); !exists {
				providers = append(providers, core.OAuth2ProviderConfig{
					Name:         "github",
					ClientId:     cfg.GithubClientID,
					ClientSecret: cfg.GithubClientSecret,
				})
			}
		}

		if len(providers) == 0 {
			return e.Next()
		}

		users.OAuth2.Enabled = true
		users.OAuth2.Providers = providers

		if err := app.Save(users); err != nil {
			log.Printf("Failed to save OAuth2 providers: %v", err)
		} else {
			log.Println("OAuth2 providers configured")
		}

		return e.Next()
	})
}
