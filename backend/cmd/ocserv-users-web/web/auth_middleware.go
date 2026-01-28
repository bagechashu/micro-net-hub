package web

import (
	"context"
	"log"
	"net/http"
)

var ctxKeyIsAdmin = "isAdmin"

// AuthMiddleware checks if the user has a valid session
// It wraps the next handler and ensures authentication is performed if enabled
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If authentication is disabled, skip checks
		if !*authConfig.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Get session cookie
		cookie, err := r.Cookie(authConfig.Session.CookieName)
		if err != nil {
			log.Printf("[auth] Invalid cookie: %v", err)
			w.WriteHeader(http.StatusForbidden)
			// If this is an HTMX request, redirect instead of rendering error page
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/")
				return
			}
			renderWithLayout(w, "403.html", nil)
			return
		}

		// Validate session
		session, err := globalSessionStore.GetSession(cookie.Value)
		if err != nil {
			log.Printf("[auth] Invalid session ID: %v", err)
			w.WriteHeader(http.StatusForbidden)
			// If this is an HTMX request, redirect instead of rendering error page
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/")
				return
			}
			renderWithLayout(w, "403.html", nil)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyIsAdmin, session.IsAdmin)

		// Renew session
		if err := globalSessionStore.RenewSession(cookie.Value); err != nil {
			log.Printf("[auth] Session renewal failed: %v", err)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin := r.Context().Value(ctxKeyIsAdmin)
		if isAdmin == nil || !isAdmin.(bool) {
			w.WriteHeader(http.StatusForbidden)
			renderWithLayout(w, "403.html", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
