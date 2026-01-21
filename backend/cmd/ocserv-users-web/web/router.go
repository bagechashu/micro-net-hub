package web

import (
	"log"
	"net/http"
	"strings"

	"ocserv-users/internal"
)

func RunWebServer(addr string, cfg *internal.Config) {
	if err := InitAuth(&cfg.Auth); err != nil {
		log.Fatalf("[auth] 初始化失败: %v", err)
	}

	// Create middleware wrapper for protected routes
	protectedMiddleware := protectedMiddlewareWrapper(globalSessionStore)

	// corefunc (登录时的触发器, 暂时不需要认证)
	http.HandleFunc("/core/nft", securityHeadersMiddleware(nftCheckTriggerHandler))
	http.HandleFunc("/core/vpnaccess", securityHeadersMiddleware(vpnAccessCheckTriggerHandler))

	// Authentication routes (always available)
	http.HandleFunc("/api/auth/login", securityHeadersMiddleware(corsMiddleware(loginAPIHandler)))
	http.HandleFunc("/api/auth/session", securityHeadersMiddleware(corsMiddleware(SessionInfoHandler)))
	// http.HandleFunc("/api/auth/logout", securityHeadersMiddleware(corsMiddleware(logoutAPIHandler)))
	http.HandleFunc("/api/auth/logout", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(logoutAPIHandler)(w, r) }, false))

	// static
	registerStatic()

	// index (公开)
	http.HandleFunc("/", protectedMiddleware(indexWebHandler, false))
	http.HandleFunc("/partials/users.html", protectedMiddleware(usersPartialWebHandler, false))

	// violations (规则违规日志) - 需要管理员权限
	http.HandleFunc("/api/violations", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(GetViolationsHandler)(w, r) }, true))
	http.HandleFunc("/api/violations/stats", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(GetViolationStatsHandler)(w, r) }, true))
	http.HandleFunc("/api/violations/users", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(GetViolationUsersHandler)(w, r) }, true))
	http.HandleFunc("/api/violations/clearold", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(ClearViolationOldDataHandler)(w, r) }, true))
	http.HandleFunc("/api/violations/vacuum", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(VacuumViolationDBHandler)(w, r) }, true))

	// nft (需要管理员权限)
	http.HandleFunc("/nft.html", protectedMiddleware(nftWebHandler, true))
	http.HandleFunc("/partials/nftlistruleset.html", protectedMiddleware(nftPartialWebHandler, true))

	// occtl APIs (需要管理员权限)
	http.HandleFunc("/api/occtl/disconnect/{id}", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(occtlDisconnectUserHandler)(w, r) }, true))

	// config pages (需要管理员权限)
	http.HandleFunc("/config.html", protectedMiddleware(configWebHandler, true))
	http.HandleFunc("/config-editor.html", protectedMiddleware(configEditorWebHandler, true))

	// Config Export APIs (需要管理员权限)
	http.HandleFunc("/api/config/export", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(ExportConfigHandler)(w, r) }, true))

	// Config APIs (需要管理员权限)
	http.HandleFunc("/api/config/view", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(ConfigViewHandler)(w, r) }, true))
	http.HandleFunc("/api/config/validate", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(ConfigValidateHandler)(w, r) }, true))
	http.HandleFunc("/api/config/preview", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(ConfigPreviewHandler)(w, r) }, true))
	http.HandleFunc("/api/config/save", protectedMiddleware(
		func(w http.ResponseWriter, r *http.Request) { corsMiddleware(ConfigSaveHandler)(w, r) }, true))

	go func() {
		log.Printf("[web] 服务运行中: http://%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()
}

// protectedMiddlewareWrapper returns a middleware wrapper for protected routes
func protectedMiddlewareWrapper(sessionStore *internal.AuthSessionStore) (protectedMiddleware func(http.HandlerFunc, bool) http.HandlerFunc) {
	if sessionStore != nil {
		return func(handler http.HandlerFunc, requireAdmin bool) http.HandlerFunc {
			authMW := AuthMiddleware(sessionStore, authConfig.Session.CookieName, requireAdmin)
			return authMW(securityHeadersMiddleware(handler))
		}
	}
	// If auth is disabled, just apply security headers
	return func(handler http.HandlerFunc, _ bool) http.HandlerFunc {
		return securityHeadersMiddleware(handler)
	}

}

// securityHeadersMiddleware adds essential security headers to HTTP responses
func securityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent browsers from caching sensitive data
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Referrer policy: minimize referrer information leakage
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy: restrict resource loading
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")

		next(w, r)
	}
}

// corsMiddleware handles CORS for API endpoints
// Only allows same-origin requests and handles preflight requests
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Only allow requests from the same origin or localhost in development
		// In production, restrict to specific allowed origins
		allowedOrigins := map[string]bool{
			"":     true, // same-origin requests have empty Origin header
			"null": true, // file:// URLs have Origin: null
		}

		// For development/internal use, you can add specific origins:
		// allowedOrigins["https://trusted-domain.com"] = true

		if isLocalhost(origin) {
			// Allow localhost in development
			allowedOrigins[origin] = true
		}

		if allowedOrigins[origin] && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Specify allowed methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Specify allowed headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		// Preflight cache duration (5 minutes)
		w.Header().Set("Access-Control-Max-Age", "300")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// isLocalhost checks if the origin is a localhost URL
func isLocalhost(origin string) bool {
	if origin == "" {
		return false
	}
	localhostPatterns := []string{
		"http://localhost",
		"http://127.0.0.1",
		"http://[::1]",
		"https://localhost",
		"https://127.0.0.1",
		"https://[::1]",
	}
	for _, pattern := range localhostPatterns {
		if strings.HasPrefix(origin, pattern) {
			return true
		}
	}
	return false
}
