package web

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"ocserv-users/internal"
)

// -------------------- Web handler --------------------

//go:embed index.html
var Static embed.FS

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(Static, "static/index.html"))
	sessions, err := internal.GetSessions()
	if err != nil {
		http.Error(w, "无法获取用户数据: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = tmpl.Execute(w, sessions)
}

func RunWebServer(addr string) {
	http.HandleFunc("/", indexHandler)
	go func() {
		log.Printf("[web] 服务运行中: http://%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()
}
