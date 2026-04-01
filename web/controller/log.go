package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sky-night-net/snet/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type LogController struct {
	BaseController
}

func NewLogController(g *gin.RouterGroup) *LogController {
	c := &LogController{}
	g.GET("/logs", c.StreamLogs)
	return c
}

func (c *LogController) StreamLogs(ctx *gin.Context) {
	ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// Initial burst of logs from buffer
	buffer := logger.GetLogs(5000, "")
	for _, line := range buffer {
		msg, _ := json.Marshal(map[string]string{"type": "log", "msg": line})
		if ws.WriteMessage(websocket.TextMessage, msg) != nil {
			return
		}
	}

	// Stream new logs
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastIndex := len(buffer)
	for {
		select {
		case <-ticker.C:
			currentBuffer := logger.GetLogs(5000, "")
			if len(currentBuffer) > lastIndex {
				for i := lastIndex; i < len(currentBuffer); i++ {
					msg, _ := json.Marshal(map[string]string{"type": "log", "msg": currentBuffer[i]})
					if ws.WriteMessage(websocket.TextMessage, msg) != nil {
						return
					}
				}
				lastIndex = len(currentBuffer)
			}
		}
	}
}
