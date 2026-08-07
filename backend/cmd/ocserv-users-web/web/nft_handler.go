package web

import (
	"net/http"

	"ocserv-users/internal"
)

func nftWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "nft.html", nil)
}

func nftPartialWebHandler(w http.ResponseWriter, r *http.Request) {
	out, err := internal.GetNftAllRules()
	if err != nil {
		http.Error(w, "failed to load sessions", 500)
		return
	}

	render(w, "partials/nftlistruleset.html", out)
}
