package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/choyos-garces/agro-api/config"
	"github.com/choyos-garces/agro-api/internal/hooks"
	_ "github.com/choyos-garces/agro-api/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	// CONFIGURATION LOADING
	// We now default to "config.yaml".
	// The config.Load() function will automatically find and merge "config.local.yaml"!
	configPath := "config.yaml"
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
		} else if len(arg) > 9 && arg[:9] == "--config=" {
			configPath = arg[9:]
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// INITIALIZE POCKETBASE
	app := pocketbase.New()

	app.RootCmd.PersistentFlags().String("config", "config.yaml", "path to the base YAML config file")

	// SERVER SETUP HOOK
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		// Replace default CORS with our config-driven CORS
		customCors := apis.CORS(apis.CORSConfig{
			AllowOrigins:     cfg.CORS.AllowedOrigins,
			AllowMethods:     cfg.CORS.AllowedMethods,
			AllowHeaders:     cfg.CORS.AllowedHeaders,
			AllowCredentials: cfg.CORS.AllowCredentials,
		})

		customCors.Id = apis.DefaultCorsMiddlewareId // Override the default CORS middleware
		se.Router.Bind(customCors)

		usersCollection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		usersCollection.OAuth2 = core.OAuth2Config{
			Enabled: true,
			Providers: []core.OAuth2ProviderConfig{
				{
					Name:         "google",
					ClientId:     cfg.Google.ClientID,     // From your config.yaml
					ClientSecret: cfg.Google.ClientSecret, // From your config.yaml
				},
			},
		}

		// 3. Save the collection to persist the changes in the DB
		if err := app.Save(usersCollection); err != nil {
			return err
		}

		return se.Next()
	})

	// OAUTH2 DOMAIN RESTRICTION HOOK
	app.OnRecordAuthWithOAuth2Request("users").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
		// Check to make sure users is comming from an allowed domain. This is really not
		// necessary since the client is set a `internal` provider, but might as well be safe.
		if e.OAuth2User != nil {
			email := e.OAuth2User.Email

			allowedDomains := []string{"@hoyosgarces.com", "@hygagro.com"}
			domainAllowed := false
			for _, domain := range allowedDomains {
				if strings.HasSuffix(email, domain) {
					domainAllowed = true
					break
				}
			}

			// If the email didn't match any domain in the list, reject it
			if !domainAllowed {
				return errors.New("unauthorized domain: please use your company email")
			}

		}

		// Continue the authentication flow
		return e.Next()
	})

	// REGISTER HOOKS (Database and Record events)
	hooks.Register(app)

	// AUTO-MIGRATIONS
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: cfg.Dev,
	})

	// INJECT SERVER BINDING (Safely)
	// If running the `serve` command, append the host:port from config.
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		hasHttpFlag := false

		// Check if the user manually typed --http in the terminal
		for _, arg := range os.Args {
			if arg == "--http" || (len(arg) > 7 && arg[:7] == "--http=") {
				hasHttpFlag = true
				break
			}
		}

		// Only inject the config address if the user didn't provide one manually
		if !hasHttpFlag {
			os.Args = append(os.Args, "--http", cfg.Server.Addr())
		}
	}

	// START APPLICATION
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
