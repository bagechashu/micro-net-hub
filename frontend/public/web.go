package web

import "embed"

//go:embed index.html favicon.ico static
var Static embed.FS
