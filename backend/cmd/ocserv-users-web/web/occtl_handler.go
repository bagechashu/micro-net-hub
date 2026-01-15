package web

import (
	"net/http"
	"ocserv-users/internal"
)

func occtlDisconnectUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/occtl/disconnect/"):]

	err := internal.OcctlDisconnectUserByID(id)
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to disconnect user")
		return	
	}

	sendSuccess(w, "user disconnected", nil)
}
