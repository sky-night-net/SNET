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

	staticServer := http.FileServer(http.FS(sub))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// If requesting an API route that doesn't exist, return 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "API route not found"})
			return
		}

		// Serve static file if it exists in the embedded FS
		_, err := sub.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			staticServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Otherwise, serve index.html (for SPA routing)
		c.FileFromFS("index.html", http.FS(sub))
	})
}
