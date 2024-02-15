package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"

	qrcode "github.com/skip2/go-qrcode"
)

func generateQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		http.Error(w, "Missing 'data' parameter, eg: http://localhost:8080/qrcode?data=qrcodestring", http.StatusBadRequest)
		return
	}

	qrCode, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCode)

	w.Header().Set("Content-Type", "text/html")
	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<title>QR Code Generator</title>
</head>
<body>
	<h3>QR Code for "{{.Data}}"</h3>
	<img src="data:image/png;base64,{{.QRCode}}" alt="QR Code">
</body>
</html>
`
	t, err := template.New("qr").Parse(tmpl)
	if err != nil {
		http.Error(w, "Failed to create HTML template", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, struct {
		Data   string
		QRCode string
	}{data, qrCodeBase64})
	if err != nil {
		http.Error(w, "Failed to render HTML template", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/qrcode", http.StatusFound)
	})
	http.HandleFunc("/qrcode", generateQRCodeHandler)

	fmt.Println("Starting server at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
