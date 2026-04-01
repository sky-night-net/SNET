package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/web/entity"
	"github.com/sky-night-net/snet/web/service"
	"github.com/sky-night-net/snet/web/session"
)

type IndexController struct {
	BaseController
	userService service.UserService
}

func NewIndexController(g *gin.RouterGroup) *IndexController {
	c := &IndexController{}
	g.POST("/login", c.Login)
	g.POST("/logout", c.Logout)
	return c
}

func (c *IndexController) Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	
	user, err := c.userService.Login(username, password)
	if err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	if user == nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: "Invalid username or password"})
		return
	}

	err = session.Set(ctx, session.UserKey, user)
	if err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: "Session error"})
		return
	}

	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Login success"})
}

func (c *IndexController) Logout(ctx *gin.Context) {
	session.Clear(ctx)
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Logout success"})
}

type BaseController struct{}
