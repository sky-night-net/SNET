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

	// Helper to serve index.html
	serveIndex := func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(sub))
	}

	// Explicitly handle root and index.html to avoid any Gin internal redirects
	r.GET("/", serveIndex)
	r.GET("/index.html", serveIndex)

	// Handle all other files in dist/
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "API route not found"})
			return
		}

		filePath := strings.TrimPrefix(path, "/")
		
		// Check if file exists in the embedded FS
		f, err := sub.Open(filePath)
		if err == nil {
			f.Close()
			c.FileFromFS(filePath, http.FS(sub))
			return
		}

		// Fallback to index.html for React SPA
		serveIndex(c)
	})
}
