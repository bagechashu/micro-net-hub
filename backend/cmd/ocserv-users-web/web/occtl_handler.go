package web

import (
	"net/http"
	"ocserv-users/internal"
)

func occtlDisconnectUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/occtl/disconnect/"):]

	// Validate session ID format before processing
	if err := internal.ValidateSessionID(id); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := internal.OcctlDisconnectUserByID(id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, "user disconnected", nil)
}
