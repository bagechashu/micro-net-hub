package web

import (
	"log"
	"net/http"
	"time"

	"ocserv-users/internal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func RunWebServer(addr string, cfg *internal.Config) {
	if err := InitAuth(&cfg.Auth); err != nil {
		log.Fatalf("[auth] 初始化失败: %v", err)
	}

	// Initialize chi router
	r := chi.NewRouter()

	// Apply global middlewares
	r.Use(httprate.LimitByIP(60, 5*time.Second))
	r.Use(middleware.Timeout(time.Second * 60))
	r.Use(middleware.Recoverer)

	// Apply CORS middleware using chi/cors
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		MaxAge:           300,
		AllowCredentials: true, // Allow cookies
	}))

	// Apply security headers middleware
	r.Use(securityHeadersMiddleware)

	// register Static file routes
	registerStatic(r)

	// Core routes (no auth required)
	r.Post("/core/nft", nftCheckTriggerHandler)
	r.Post("/core/vpnaccess", vpnAccessCheckTriggerHandler)

	// index page
	r.Get("/", indexWebHandler)
	r.Get("/partials/ocusers-index.html", indexOcUsersPartialWebHandler)

	// Authentication routes (always available)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Post("/api/auth/login", loginAPIHandler)
		r.Get("/api/auth/session", SessionInfoHandler)
		r.Post("/api/auth/logout", logoutAPIHandler)
	})

	// Protected routes (require auth)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(authMiddleware)

		// ocUsers management
		r.Get("/ocusers.html", ocUsersWebHandler)
		r.Get("/partials/ocusers.html", ocUsersPartialWebHandler)

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(adminMiddleware)

			// NFT management
			r.Get("/nft.html", nftWebHandler)
			r.Get("/partials/nftlistruleset.html", nftPartialWebHandler)

			// Config pages
			r.Get("/config.html", configWebHandler)
			r.Get("/config-editor.html", configEditorWebHandler)
		})

		// API routes
		r.Route("/api", func(r chi.Router) {
			// Violations API
			r.Get("/violations", GetViolationsHandler)
			r.Get("/violations/stats", GetViolationStatsHandler)
			r.Get("/violations/users", GetViolationUsersHandler)

			// Admin-only routes
			r.Group(func(r chi.Router) {
				r.Use(adminMiddleware)

				// OCCTL APIs
				r.Post("/occtl/disconnect/{id}", occtlDisconnectUserHandler)

				// Violations API
				r.Post("/violations/clearold", ClearViolationOldDataHandler)
				r.Post("/violations/vacuum", VacuumViolationDBHandler)

				// Config APIs
				r.Get("/config/view", ConfigViewHandler)
				r.Get("/config/export", ExportConfigHandler)
				r.Post("/config/validate", ConfigValidateHandler)
				r.Post("/config/preview", ConfigPreviewHandler)
				r.Post("/config/save", ConfigSaveHandler)
				r.Get("/config/save-status/{id}", ConfigSaveStatusHandler)
			})
		})
	})

	// Custom 404 handler
	r.NotFound(notFoundHandler)

	// Custom 405 handler
	r.MethodNotAllowed(methodNotAllowedHandler)

	go func() {
		log.Printf("[web] 服务运行中: http://%s", addr)
		if err := http.ListenAndServe(addr, r); err != nil {
			log.Fatal(err)
		}
	}()
}
