package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/vpn/adapters"
	"github.com/sky-night-net/snet/database/model"
)

type ClientController struct{}

func NewClientController() *ClientController {
	return &ClientController{}
}

func (c *ClientController) AddClient(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": true, "msg": "Client added"})
}

func (c *ClientController) RemoveClient(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": true, "msg": "Client removed"})
}

func (c *ClientController) Keygen(ctx *gin.Context) {
	protocol := model.Protocol(ctx.Param("protocol"))

	adapter, err := adapters.GetAdapter(protocol)
	if err != nil {
		log.Printf("Adapter %s not found: %v", protocol, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	keys, err := adapter.GenerateKeypair()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"obj":     keys,
	})
}
