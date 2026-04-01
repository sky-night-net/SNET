package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/web/session"
)

func Auth(c *gin.Context) {
	if !session.IsLogin(c) {
		if isAPI(c) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"msg":     "Login failed",
			})
			c.Abort()
		} else {
			// Redirect to login if accessing HTML
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
		}
		return
	}
	c.Next()
}

func isAPI(c *gin.Context) bool {
	path := c.Request.URL.Path
	return strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/panel/api")
}
