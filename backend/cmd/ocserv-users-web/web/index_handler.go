package web

import (
	"net/http"
	"ocserv-users/internal"
)

func indexWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "index.html", nil)
}

func usersPartialWebHandler(w http.ResponseWriter, r *http.Request) {
	sessions, err := internal.GetSessions()
	if err != nil {
		http.Error(w, "failed to load sessions", 500)
		return
	}

	render(w, "partials/users.html", sessions)
}
