package web

import (
	"net/http"

	"ocserv-users/internal"

	"github.com/go-chi/chi/v5"
)

func occtlDisconnectUserHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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
