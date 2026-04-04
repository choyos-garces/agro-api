// agro-api is a PocketBase-extended HTTP server with custom auth endpoints.
//
// Usage:
//
//	agro-api serve [--config=config.yaml]
//
// The server starts PocketBase with additional custom routes:
//
//	POST /api/auth/login   – authenticate and receive a session cookie
//	GET  /api/auth/verify  – verify the current session
//	POST /api/auth/logout  – destroy the session cookie
package main

import (
	"flag"
	"log"
	"strings"

	"github.com/choyos-garces/agro-api/config"
	"github.com/choyos-garces/agro-api/internal/auth"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	app := pocketbase.New()

	// ------------------------------------------------------------------
	// CORS middleware – applied before all routes.
	// ------------------------------------------------------------------
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.BindFunc(corsMiddleware(cfg))
		return se.Next()
	})

	// ------------------------------------------------------------------
	// Register custom auth routes.
	// ------------------------------------------------------------------
	authHandlers := auth.New(cfg)
	authHandlers.RegisterRoutes(app)

	// ------------------------------------------------------------------
	// Optional: auto-register the migrate command so that
	//   agro-api migrate up / down
	// works out of the box.
	// ------------------------------------------------------------------
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// ------------------------------------------------------------------
	// Override the default PocketBase address with what is in config.yaml.
	// PocketBase's serve sub-command honours --http flag; we set the
	// address via the app's settings before starting.
	// ------------------------------------------------------------------
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		log.Printf("agro-api listening on %s", cfg.Server.Addr())
		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// corsMiddleware returns a PocketBase hook function that injects CORS headers
// on every response and short-circuits OPTIONS pre-flight requests.
func corsMiddleware(cfg *config.Config) func(*core.RequestEvent) error {
	allowedOrigins := cfg.CORS.AllowedOrigins
	allowedMethods := joinStrings(cfg.CORS.AllowedMethods, ", ")
	allowedHeaders := joinStrings(cfg.CORS.AllowedHeaders, ", ")

	return func(e *core.RequestEvent) error {
		origin := e.Request.Header.Get("Origin")

		// Only set ACAO header when there is an Origin (i.e. a CORS request).
		if origin != "" {
			allowed := allowedOrigins[0] == "*"
			if !allowed {
				for _, o := range allowedOrigins {
					if o == origin {
						allowed = true
						break
					}
				}
			}

			if allowed {
				if allowedOrigins[0] == "*" && !cfg.CORS.AllowCredentials {
					e.Response.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					e.Response.Header().Set("Access-Control-Allow-Origin", origin)
					e.Response.Header().Set("Vary", "Origin")
				}

				if cfg.CORS.AllowCredentials {
					e.Response.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				e.Response.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				e.Response.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			}
		}

		// Short-circuit pre-flight requests.
		if e.Request.Method == "OPTIONS" {
			e.Response.WriteHeader(204)
			return nil
		}

		return e.Next()
	}
}

// joinStrings joins a slice of strings with the given separator.
func joinStrings(ss []string, sep string) string {
	return strings.Join(ss, sep)
}
