package main

import (
	"log"
	"os"

	"github.com/choyos-garces/agro-api/config"
	"github.com/choyos-garces/agro-api/internal/auth"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	// 1. CONFIGURATION LOADING
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

	// 2. INITIALIZE POCKETBASE
	app := pocketbase.New()

	app.RootCmd.PersistentFlags().String("config", "config.yaml", "path to the base YAML config file")

	// 3. SERVER SETUP HOOK
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		// Replace default CORS with our config-driven CORS
		se.Router.Unbind(apis.DefaultCorsMiddlewareId)
		se.Router.Bind(apis.CORS(apis.CORSConfig{
			AllowOrigins:     cfg.CORS.AllowedOrigins,
			AllowMethods:     cfg.CORS.AllowedMethods,
			AllowHeaders:     cfg.CORS.AllowedHeaders,
			AllowCredentials: cfg.CORS.AllowCredentials,
		}))

		// Register Custom Routes (using the updated se.Router signature)
		authHandlers := auth.New(cfg)
		authHandlers.RegisterRoutes(se.Router)

		return se.Next()
	})

	// 4. AUTO-MIGRATIONS
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: cfg.Dev,
	})

	// 5. INJECT SERVER BINDING (Safely)
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

	// 6. START APPLICATION
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
