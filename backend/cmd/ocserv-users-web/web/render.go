package web

import (
	"embed"
	"encoding/json"
	"html"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*/*
var staticFS embed.FS

// HTMLEscape safely escapes JSON output for use in HTML context
// Converts a value to JSON and then HTML-escapes the result
func htmlEscapeJSON(v interface{}) (template.HTML, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	// HTML-escape the JSON string for safe inclusion in HTML
	escaped := html.EscapeString(string(data))
	return template.HTML(escaped), nil
}

// Template functions map
var funcMap = template.FuncMap{
	// safeJSON provides HTML-escaped JSON for safe template usage
	"safeJSON": htmlEscapeJSON,
}

// <!-- 安全方式 1：在 HTML 属性中 -->
// <div data-config="{{ safeJSON .UserData }}"></div>

// <!-- 安全方式 2：在脚本中用 html 过滤器 -->
// <script>
//   var config = {{ .Config | html }};
// </script>

// <!-- 安全方式 3：在文本内容中 -->
// <pre>{{ safeJSON .RuleData }}</pre>

// renderWithLayout 渲染带布局的模板
// name: 主模板文件名 (如 "index.html")
// data: 传递给模板的数据
// tmplPaths: 可选的额外模板路径 (如 "partials/custom.html")
func renderWithLayout(w http.ResponseWriter, name string, data any, tmplPaths ...string) {
	paths := []string{
		"templates/layout.html",
		"templates/" + name,
	}

	for _, path := range tmplPaths {
		paths = append(paths, "templates/"+path)
	}

	tmpl := template.Must(
		template.New("").Funcs(funcMap).ParseFS(templatesFS, paths...),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// render 渲染单个模板
// name: 模板文件名 (如 "partials/ocusers.html")
// data: 传递给模板的数据
// tmplPaths: 可选的额外模板路径
func render(w http.ResponseWriter, name string, data any, tmplPaths ...string) {
	paths := []string{
		"templates/" + name,
	}

	for _, path := range tmplPaths {
		paths = append(paths, "templates/"+path)
	}

	tmpl := template.Must(
		template.New(filepath.Base(name)).Funcs(funcMap).ParseFS(templatesFS, paths...),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// registerStatic registers static file routes for chi
func registerStatic(r chi.Router) {
	fsys, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("failed to create static sub-filesystem: " + err.Error())
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(fsys))))
}

// Response helper functions
func sendJSON(w http.ResponseWriter, code int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[web] JSON encode error: %v", err)
	}
}

func sendError(w http.ResponseWriter, code int, message string) {
	sendJSON(w, code, Response{Code: code, Message: message})
}

func sendSuccess(w http.ResponseWriter, message string, data interface{}) {
	sendJSON(w, http.StatusOK, Response{Code: 0, Message: message, Data: data})
}
