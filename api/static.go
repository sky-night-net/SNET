package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var frontendFS embed.FS

func ServeFrontend(r *gin.Engine) {
	sub, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		panic(err)
	}

	// Read index.html into memory to avoid any redirect logic from http.FS
	indexHTML, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		panic(err)
	}

	// Serve raw index.html data to prevent 301/307 redirects
	serveIndexRaw := func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}

	// Explicitly handle root and index.html
	r.GET("/", serveIndexRaw)
	r.GET("/index.html", serveIndexRaw)

	// Handle all other files in dist/
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "API route not found"})
			return
		}

		filePath := strings.TrimPrefix(path, "/")
		
		// Serve static file if it exists
		f, err := sub.Open(filePath)
		if err == nil {
			f.Close()
			c.FileFromFS(filePath, http.FS(sub))
			return
		}

		// Fallback to React app
		serveIndexRaw(c)
	})
}
