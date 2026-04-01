package controller

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sky-night-net/snet/web/entity"
)

type ServerController struct {
	BaseController
}

func NewServerController(g *gin.RouterGroup) *ServerController {
	c := &ServerController{}
	g.GET("/status", c.GetStatus)
	return c
}

func (c *ServerController) GetStatus(ctx *gin.Context) {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(time.Second, false)
	h, _ := host.Info()

	status := gin.H{
		"cpu":    c[0],
		"mem":    v.UsedPercent,
		"uptime": h.Uptime,
		"os":     runtime.GOOS,
		"arch":   runtime.GOARCH,
	}

	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Obj: status})
}
