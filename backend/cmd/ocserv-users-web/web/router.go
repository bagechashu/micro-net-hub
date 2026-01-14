package web

import (
	"log"
	"net/http"
)

func RunWebServer(addr string) {
	// corefunc
	http.HandleFunc("/nftables", nftablesHandler)
	http.HandleFunc("/vpnaccess", vpnAccessHandler)

	// static
	registerStatic()

	// index
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/partials/users.html", usersPartialHandler)

	// config pages
	http.HandleFunc("/config.html", configHandler)
	http.HandleFunc("/config-editor.html", configEditorPageHandler)

	// Export APIs
	http.HandleFunc("/api/config/export", ExportConfigHandler)

	// Unified Config APIs (Phase 1 & 2)
	http.HandleFunc("/api/config/view", ConfigViewHandler)
	http.HandleFunc("/api/config/validate", ConfigValidateHandler)
	http.HandleFunc("/api/config/preview", ConfigPreviewHandler)
	http.HandleFunc("/api/config/save", ConfigSaveHandler)

	go func() {
		log.Printf("[web] 服务运行中: http://%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()
}
