package web

import (
	"log"
	"net/http"
	"strings"
)

func RunWebServer(addr string) {
	// corefunc
	http.HandleFunc("/core/nft", securityHeadersMiddleware(nftCheckTriggerHandler))
	http.HandleFunc("/core/vpnaccess", securityHeadersMiddleware(vpnAccessCheckTriggerHandler))

	// static
	registerStatic()

	// index
	http.HandleFunc("/", securityHeadersMiddleware(indexWebHandler))
	http.HandleFunc("/partials/users.html", securityHeadersMiddleware(usersPartialWebHandler))

	// nft
	http.HandleFunc("/nft.html", securityHeadersMiddleware(nftWebHandler))
	http.HandleFunc("/partials/nftlistruleset.html", securityHeadersMiddleware(nftPartialWebHandler))

	// occtl APIs
	http.HandleFunc("/api/occtl/disconnect/{id}", securityHeadersMiddleware(corsMiddleware(occtlDisconnectUserHandler)))

	// config pages
	http.HandleFunc("/config.html", securityHeadersMiddleware(configWebHandler))
	http.HandleFunc("/config-editor.html", securityHeadersMiddleware(configEditorWebHandler))

	// Config Export APIs
	http.HandleFunc("/api/config/export", securityHeadersMiddleware(corsMiddleware(ExportConfigHandler)))

	// Config APIs (Phase 1 & 2)
	http.HandleFunc("/api/config/view", securityHeadersMiddleware(corsMiddleware(ConfigViewHandler)))
	http.HandleFunc("/api/config/validate", securityHeadersMiddleware(corsMiddleware(ConfigValidateHandler)))
	http.HandleFunc("/api/config/preview", securityHeadersMiddleware(corsMiddleware(ConfigPreviewHandler)))
	http.HandleFunc("/api/config/save", securityHeadersMiddleware(corsMiddleware(ConfigSaveHandler)))

	go func() {
		log.Printf("[web] 服务运行中: http://%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()
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
