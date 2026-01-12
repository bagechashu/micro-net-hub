package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// render 渲染模板
func renderWithLayout(w http.ResponseWriter, name string, data any) {
	tmpl := template.Must(
		template.ParseFS(
			templatesFS,
			"templates/layout.html",
			"templates/"+name,
		),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// render 渲染模板
func render(w http.ResponseWriter, name string, data any) {
	tmpl := template.Must(
		template.ParseFS(
			templatesFS,
			"templates/"+name,
		),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 注册静态文件路由
func registerStatic() {
	fsys, _ := fs.Sub(staticFS, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fsys))))
}
