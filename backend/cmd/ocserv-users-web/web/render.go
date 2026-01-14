package web

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Marshal converts an interface to JSON string
func marshal(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	return string(data), err
}

// Template functions map
var funcMap = template.FuncMap{
	"marshal": marshal,
}

// render 渲染模板
func renderWithLayout(w http.ResponseWriter, name string, data any) {
	tmpl := template.Must(
		template.New("").Funcs(funcMap).ParseFS(
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
		template.New(filepath.Base(name)).Funcs(funcMap).ParseFS(
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
