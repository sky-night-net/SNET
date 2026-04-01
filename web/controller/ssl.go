package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/util/sys"
	"github.com/sky-night-net/snet/web/entity"
)

type SSLController struct {
	BaseController
	manager *sys.SSLManager
}

func NewSSLController(g *gin.RouterGroup) *SSLController {
	c := &SSLController{
		manager: sys.NewSSLManager(),
	}
	g.POST("/issue", c.Issue)
	g.GET("/status", c.Status)
	return c
}

func (c *SSLController) Issue(ctx *gin.Context) {
	domain := ctx.PostForm("domain")
	email := ctx.PostForm("email")
	
	if domain == "" || email == "" {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: "Domain and Email are required"})
		return
	}

	err := c.manager.IssueCertificate(domain, email)
	if err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Certificate issued successfully"})
}

func (c *SSLController) Status(ctx *gin.Context) {
	exists := c.manager.IsCertExists()
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Obj: gin.H{"has_cert": exists}})
}
