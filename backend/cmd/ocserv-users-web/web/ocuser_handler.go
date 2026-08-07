package web

import (
	"net/http"

	"ocserv-users/internal"
)

func ocUsersWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "ocusers.html", ocUsersTable["ocusers"], "partials/ocusers-table.html")
}

func ocUsersPartialWebHandler(w http.ResponseWriter, r *http.Request) {
	sessions, err := internal.OcctlGetSessions()
	if err != nil {
		http.Error(w, "failed to load sessions", 500)
		return
	}

	render(w, "partials/ocusers.html", sessions)
}

func indexWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "index.html", ocUsersTable["index"], "partials/ocusers-table.html")
}

func indexOcUsersPartialWebHandler(w http.ResponseWriter, r *http.Request) {
	sessions, err := internal.OcctlGetSessions()
	if err != nil {
		http.Error(w, "failed to load sessions", 500)
		return
	}

	render(w, "partials/ocusers-index.html", sessions)
}

// TableConfig 定义表格配置
type ocUsersTableTemplateData struct {
	Title        string   `json:"title"`
	Columns      []string `json:"columns"`
	DataEndpoint string   `json:"dataEndpoint"`
}

// ocUsersTable 在线用户表格配置集
var ocUsersTable = map[string]ocUsersTableTemplateData{
	"index": {
		Title:        "在线用户",
		Columns:      []string{"Username", "IPAddr", "ConnectedFor", "ConnectedAt", "UserAgent"},
		DataEndpoint: "/partials/ocusers-index.html",
	},
	"ocusers": {
		Title:        "在线用户",
		Columns:      []string{"Action", "ID", "Username", "IPAddr", "ConnectedFor", "ConnectedAt", "UserAgent", "RX", "TX", "AverageRX", "AverageTX"},
		DataEndpoint: "/partials/ocusers.html",
	},
}

// Columns: []string{"Action", "ID", "Username", "Groupname", "State", "IPAddr", "ConnectedFor", "ConnectedAt", "UserAgent", "RX", "TX", "AverageRX", "AverageTX"}
