package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/sky-night-net/snet/service"
)

type SystemController struct {
	startTime time.Time
}

func NewSystemController() *SystemController {
	return &SystemController{
		startTime: time.Now(),
	}
}

func (c *SystemController) GetStatus(ctx *gin.Context) {
	v, _ := mem.VirtualMemory()
	cUsage, _ := cpu.Percent(time.Second, false)
	l, _ := load.Avg()
	h, _ := host.Info()
	n, _ := net.IOCounters(false)

	var cpuPercent float64
	if len(cUsage) > 0 {
		cpuPercent = cUsage[0]
	}

	history := service.GetStatsService().GetHistory()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"obj": gin.H{
			"cpu":     cpuPercent,
			"mem":     gin.H{"current": v.Used, "total": v.Total},
			"load":    l.Load1,
			"uptime":  h.Uptime,
			"net":     n[0],
			"go":      gin.H{"goroutines": runtime.NumGoroutine(), "version": runtime.Version()},
			"history": history,
		},
	})
}

