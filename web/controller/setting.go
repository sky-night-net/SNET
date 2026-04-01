package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/web/entity"
	"github.com/sky-night-net/snet/web/service"
)

type SettingController struct {
	BaseController
	settingService service.SettingService
}

func NewSettingController(g *gin.RouterGroup) *SettingController {
	c := &SettingController{}
	g.GET("/all", c.GetAll)
	g.POST("/update", c.Update)
	return c
}

func (c *SettingController) GetAll(ctx *gin.Context) {
	settings := entity.AllSetting{}
	port, _ := c.settingService.GetPort()
	basePath, _ := c.settingService.GetBasePath()
	serverIP, _ := c.settingService.GetServerIP()
	
	settings.WebPort = port
	settings.WebBasePath = basePath
	settings.ServerIP = serverIP
	
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Obj: settings})
}

func (c *SettingController) Update(ctx *gin.Context) {
	var settings map[string]string
	if err := ctx.ShouldBindJSON(&settings); err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	
	for k, v := range settings {
		err := c.settingService.SetSetting(k, v)
		if err != nil {
			ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
			return
		}
	}
	
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Settings updated"})
}
