package web

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"ocserv-users/internal"
)

// Global variables for auth (initialized in main)
var (
	authConfig          *internal.AuthConfig
	radiusAuthenticator *internal.RadiusAuthenticator
	globalSessionStore  *internal.AuthSessionStore
	adminUsersList      map[string]bool
)

// InitAuth initializes the authentication system.
func InitAuth(cfg *internal.AuthConfig) error {
	// Set global auth config
	authConfig = cfg

	// Initialize global session store
	globalSessionStore = internal.NewAuthSessionStore(time.Duration(cfg.Session.TimeoutMinutes) * time.Minute)

	// If authentication is disabled, skip further initialization.
	// Must after setting up authConfig and globalSessionStore, or other code may panic.
	if !*cfg.Enabled {
		log.Println("[auth] authentication is disabled")
		return nil
	}

	// Build admin users map for quick lookup
	adminUsersList = make(map[string]bool)
	for _, user := range cfg.Admins.Users {
		adminUsersList[user] = true
	}

	// Initialize RADIUS authenticator if configured
	if cfg.Radius.ConfigFile != "" || cfg.Radius.ServersFile != "" {
		authenticator, err := internal.NewRadiusAuthenticator(cfg.Radius)
		if err != nil {
			return err
		}
		radiusAuthenticator = authenticator
		log.Println("[auth] RADIUS 认证已初始化")
	}

	log.Println("[auth] 身份认证系统已初始化")
	return nil
}

// loginAPIHandler handles login API requests
func loginAPIHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		sendError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	// Authenticate with RADIUS
	if radiusAuthenticator == nil {
		sendError(w, http.StatusInternalServerError, "authentication system not ready")
		return
	}

	authenticated, err := radiusAuthenticator.Authenticate(req.Username, req.Password)
	if err != nil {
		log.Printf("[auth] authentication error: %v", err)
		sendError(w, http.StatusInternalServerError, "login failed, please try again")
		return
	}

	if !authenticated {
		log.Printf("[auth] user %s authentication failed", req.Username)
		sendError(w, http.StatusUnauthorized, "username or password is incorrect")
		return
	}

	// Check if user is admin
	isAdmin := adminUsersList[req.Username]

	// Create session
	session, err := globalSessionStore.CreateSession(req.Username, isAdmin)
	if err != nil {
		log.Printf("[auth] failed to create session: %v", err)
		sendError(w, http.StatusInternalServerError, "login failed, please try again")
		return
	}

	// session cookie
	if authConfig.Session.CookieName == "" {
		sendError(w, http.StatusInternalServerError, "session cookie name is not configured")
		return
	}

	timeout := time.Duration(authConfig.Session.TimeoutMinutes) * time.Minute
	http.SetCookie(w, createSessionCookie(session.ID, authConfig.Session.CookieName, timeout))

	log.Printf("[auth] user %s logged in (isAdmin: %v)", req.Username, isAdmin)

	sendSuccess(w, "login successful", map[string]interface{}{
		"session": map[string]interface{}{
			"id":       session.ID,
			"username": session.Username,
			"isAdmin":  session.IsAdmin,
		},
	})
}

// logoutAPIHandler handles logout API requests
func logoutAPIHandler(w http.ResponseWriter, r *http.Request) {
	// session cookie
	if authConfig.Session.CookieName == "" {
		sendError(w, http.StatusInternalServerError, "session cookie name is not configured")
		return
	}

	cookie, err := r.Cookie(authConfig.Session.CookieName)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "not logged in")
		return
	}

	username, err := globalSessionStore.DeleteSession(cookie.Value)
	if err != nil {
		log.Printf("[auth] failed to delete '%s' session: %v", username, err)
	}

	// Delete session cookie
	http.SetCookie(w, deleteSessionCookie(authConfig.Session.CookieName))

	log.Printf("[auth] user %s logged out", username)

	sendSuccess(w, "logged out successfully", nil)
}

// SessionInfoHandler returns current session info
func SessionInfoHandler(w http.ResponseWriter, r *http.Request) {
	// session cookie
	if authConfig.Session.CookieName == "" {
		sendError(w, http.StatusInternalServerError, "session cookie name is not configured")
		return
	}

	// Get session cookie
	cookie, err := r.Cookie(authConfig.Session.CookieName)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "not logged in")
		return
	}

	// Get session from store
	session, err := globalSessionStore.GetSession(cookie.Value)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "session invalid or expired")
		return
	}

	sendSuccess(w, "account info", map[string]interface{}{
		"username": session.Username,
		"isAdmin":  session.IsAdmin,
	})
}

// CreateSessionCookie creates a session cookie
func createSessionCookie(sessionID string, cookieName string, timeout time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(timeout),
		HttpOnly: true,
		Secure:   false,                   // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode, // 改为 Strict 以支持同站点导航时的 cookie
	}
}

// DeleteSessionCookie creates a cookie to delete the session
func deleteSessionCookie(cookieName string) *http.Cookie {
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
