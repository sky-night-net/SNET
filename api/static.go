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
	// Root of our static files
	sub, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		panic(err)
	}

	// Disable Gin's internal redirecting for the frontend routes
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 1. Skip API routes
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "API route not found"})
			return
		}

		// 2. Clean path for FS lookup
		filePath := strings.TrimPrefix(path, "/")
		
		// 3. If it's root or empty, serve index.html immediately
		if filePath == "" || filePath == "/" {
			c.FileFromFS("index.html", http.FS(sub))
			return
		}

		// 4. Try to open the file in the embedded FS
		f, err := sub.Open(filePath)
		if err == nil {
			f.Close()
			// If it's a directory, don't serve it directly (prevents 301 loops)
			stat, _ := fs.Stat(sub, filePath)
			if stat != nil && !stat.IsDir() {
				c.FileFromFS(filePath, http.FS(sub))
				return
			}
		}

		// 5. Fallback to index.html for all other routes (React SPA)
		c.FileFromFS("index.html", http.FS(sub))
	})
}
