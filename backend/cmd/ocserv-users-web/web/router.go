package web

import (
	"log"
	"net/http"
)

func RunWebServer(addr string) {
	// corefunc
	http.HandleFunc("/core/nft", nftCheckTriggerHandler)
	http.HandleFunc("/core/vpnaccess", vpnAccessCheckTriggerHandler)

	// static
	registerStatic()

	// index
	http.HandleFunc("/", indexWebHandler)
	http.HandleFunc("/partials/users.html", usersPartialWebHandler)

	// nft
	http.HandleFunc("/nft.html", nftWebHandler)
	http.HandleFunc("/partials/nftlistruleset.html", nftPartialWebHandler)

	// occtl APIs
	http.HandleFunc("/api/occtl/disconnect/{id}", occtlDisconnectUserHandler)

	// config pages
	http.HandleFunc("/config.html", configWebHandler)
	http.HandleFunc("/config-editor.html", configEditorWebHandler)

	// Config Export APIs
	http.HandleFunc("/api/config/export", ExportConfigHandler)

	// Config APIs (Phase 1 & 2)
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
