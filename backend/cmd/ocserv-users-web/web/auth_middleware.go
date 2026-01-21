package web

import (
	"log"
	"net/http"
	"time"

	"ocserv-users/internal"
)

// AuthMiddleware checks if the user has a valid session
// It wraps the next handler and ensures authentication is performed if enabled
func AuthMiddleware(sessionStore *internal.AuthSessionStore, sessionCookieName string, requireAdmin bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get session cookie
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				if requireAdmin {
					http.Redirect(w, r, "/?redirect="+r.URL.Path, http.StatusSeeOther)
					return
				}
				// If auth is not required but enabled, proceed with guest status
				next(w, r)
				return
			}

			// Validate session
			session, err := sessionStore.GetSession(cookie.Value)
			if err != nil {
				log.Printf("[auth] Invalid session ID: %v", err)
				if requireAdmin {
					http.Redirect(w, r, "/?redirect="+r.URL.Path, http.StatusSeeOther)
					return
				}
				next(w, r)
				return
			}

			// Check admin requirement
			if requireAdmin && !session.IsAdmin {
				w.WriteHeader(http.StatusForbidden)
				renderWithLayout(w, "403.html", nil)
				return
			}

			// Renew session
			if err := sessionStore.RenewSession(cookie.Value); err != nil {
				log.Printf("[auth] Session renewal failed: %v", err)
			}

			next(w, r)
		}
	}
}

// CreateSessionCookie creates a session cookie
func CreateSessionCookie(sessionID string, cookieName string, timeout time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(timeout),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,  // 改为 Strict 以支持同站点导航时的 cookie
	}
}

// DeleteSessionCookie creates a cookie to delete the session
func DeleteSessionCookie(cookieName string) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}
