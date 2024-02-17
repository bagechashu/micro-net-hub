package main

// TODO: online clipboard service
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	mutex     sync.Mutex
	pastes    = make(map[string]pasteData)
	expireDur = time.Minute * 5 // 设置默认的过期时间为5分钟
	key       = []byte("examplekey16bytes")
)

type pasteData struct {
	Text      []byte
	FileData  []byte
	FileName  string
	ExpiresAt time.Time
}

func encrypt(plainText []byte, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	ciphertext := make([]byte, aes.BlockSize+len(plainText))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plainText)
	return ciphertext
}

func decrypt(ciphertext []byte, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext
}

func handlePaste(c *gin.Context) {
	text := c.PostForm("text")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "No file provided")
		return
	}
	var fileData []byte
	if file != nil {
		defer file.Close()
		fileData, err = io.ReadAll(file)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read file")
			return
		}
	}

	mutex.Lock()
	defer mutex.Unlock()
	id := generateID()
	encryptedText := encrypt([]byte(text), key)
	encryptedFileData := encrypt(fileData, key)
	paste := pasteData{
		Text:      encryptedText,
		FileData:  encryptedFileData,
		FileName:  header.Filename,
		ExpiresAt: time.Now().Add(expireDur),
	}
	pastes[id] = paste

	c.String(http.StatusOK, "https://yourdomain.com/paste/%s", id)
}

func handleView(c *gin.Context) {
	id := c.Param("id")
	mutex.Lock()
	paste, ok := pastes[id]
	mutex.Unlock()
	if !ok {
		renderErrorTemplate(c, http.StatusNotFound, "Paste not found")
		return
	}
	if paste.ExpiresAt.Before(time.Now()) {
		mutex.Lock()
		delete(pastes, id)
		mutex.Unlock()
		renderErrorTemplate(c, http.StatusGone, "Paste has expired")
		return
	}
	decryptedText := decrypt(paste.Text, key)
	decryptedFileData := decrypt(paste.FileData, key)
	renderViewTemplate(c, http.StatusOK, decryptedText, decryptedFileData, paste.FileName)
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func cleanupExpiredPastes() {
	for {
		time.Sleep(time.Minute) // 每分钟检查一次过期的粘贴
		mutex.Lock()
		for id, paste := range pastes {
			if paste.ExpiresAt.Before(time.Now()) {
				delete(pastes, id)
			}
		}
		mutex.Unlock()
	}
}

func main() {
	go cleanupExpiredPastes() // 启动清理过期粘贴的协程

	router := gin.Default()
	router.POST("/new", handlePaste)
	router.GET("/paste/:id", handleView)
	router.Run(":8080")
}
