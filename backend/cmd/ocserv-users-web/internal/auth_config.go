package internal

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Enable authentication (default: true, set at LoadConfig)
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`

	// RADIUS authentication configuration
	Radius RadiusConfig `json:"radius,omitempty" yaml:"radius,omitempty"`

	// Session configuration
	Session AuthSessionConfig `json:"session,omitempty" yaml:"session,omitempty"`

	// Admin users configuration
	Admins AdminsConfig `json:"admins,omitempty" yaml:"admins,omitempty"`
}

// RadiusConfig holds RADIUS server configuration
type RadiusConfig struct {
	// Path to radiusclient.conf
	ConfigFile string `json:"config_file,omitempty" yaml:"config_file,omitempty"`

	// Path to servers file (alternative to reading from config_file)
	ServersFile string `json:"servers_file,omitempty" yaml:"servers_file,omitempty"`
}

// AuthSessionConfig holds session configuration
type AuthSessionConfig struct {
	// Session timeout in minutes (default: 60 = 1 hour)
	TimeoutMinutes int `json:"timeout_minutes,omitempty" yaml:"timeout_minutes,omitempty"`

	// Cookie name for session ID (default: "ocauth_session")
	CookieName string `json:"cookie_name,omitempty" yaml:"cookie_name,omitempty"`
}

// AdminsConfig holds administrator configuration
type AdminsConfig struct {
	// List of usernames that have admin privileges
	Users []string `json:"users,omitempty" yaml:"users,omitempty"`
}

func (ac *AuthConfig) setDefaults() {
	if ac.Enabled == nil {
		defaultEnabled := true
		ac.Enabled = &defaultEnabled
	}
	if ac.Session.TimeoutMinutes <= 0 {
		ac.Session.TimeoutMinutes = 60
	}
	if ac.Session.CookieName == "" {
		ac.Session.CookieName = "ocauth_session"
	}
}
