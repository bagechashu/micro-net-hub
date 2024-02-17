package main

import (
	"bytes"
	"encoding/base64"
	"html/template"

	"github.com/gin-gonic/gin"
)

var (
	viewTemplate  *template.Template
	errorTemplate *template.Template
)

func renderViewTemplate(c *gin.Context, status int, text []byte, fileData []byte, fileName string) {
	var tpl bytes.Buffer
	viewTemplate.Execute(&tpl, gin.H{
		"Text":     string(text),
		"FileData": base64.StdEncoding.EncodeToString(fileData),
		"FileName": fileName,
	})
	c.Data(status, "text/html", tpl.Bytes())
}

func renderErrorTemplate(c *gin.Context, status int, message string) {
	var tpl bytes.Buffer
	errorTemplate.Execute(&tpl, gin.H{
		"Message": message,
	})
	c.Data(status, "text/html", tpl.Bytes())
}
func init() {
	// 定义 HTML 模板
	viewTemplateStr := `
    <html>
    <head>
        <title>Paste</title>
    </head>
    <body>
        <h1>Paste Content</h1>
        <p>{{.Text}}</p>
        <h2>File</h2>
        <a href="data:application/octet-stream;base64,{{.FileData}}" download="{{.FileName}}">{{.FileName}}</a>
    </body>
    </html>
    `
	errorTemplateStr := `
    <html>
    <head>
        <title>Error</title>
    </head>
    <body>
        <h1>Error</h1>
        <p>{{.Message}}</p>
    </body>
    </html>
    `

	// 解析 HTML 模板
	viewTemplate = template.Must(template.New("view").Parse(viewTemplateStr))
	errorTemplate = template.Must(template.New("error").Parse(errorTemplateStr))
}
