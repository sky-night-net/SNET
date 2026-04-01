package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/web/service"
)

type VPNConfigController struct {
	BaseController
	inboundService *service.InboundService
}

func NewVPNConfigController(g *gin.RouterGroup, ib *service.InboundService) *VPNConfigController {
	c := &VPNConfigController{inboundService: ib}
	g.GET("/download/:inboundId/:clientId", c.Download)
	return c
}

func (c *VPNConfigController) Download(ctx *gin.Context) {
	inboundId, _ := strconv.Atoi(ctx.Param("inboundId"))
	clientId, _ := strconv.Atoi(ctx.Param("clientId"))
	
	inbound, err := c.inboundService.GetInbound(inboundId)
	if err != nil {
		ctx.String(http.StatusNotFound, "Inbound not found")
		return
	}
	
	// Find client
	var client *model.Client
	for _, cl := range inbound.Clients {
		if cl.Id == clientId {
			client = &cl
			break
		}
	}
	
	if client == nil {
		ctx.String(http.StatusNotFound, "Client not found")
		return
	}

	// Generate config
	var configContent string
	var filename string

	switch inbound.Protocol {
	case model.AmneziaWGv1, model.AmneziaWGv2:
		// Need adapter to generate this
		// For now, let's assume we can builder it
		configContent = fmt.Sprintf("[Interface]\nPrivateKey = ...\nAddress = ...\nDNS = ...\n\n[Peer]\nPublicKey = ...\nEndpoint = ...")
		filename = fmt.Sprintf("%s.conf", client.Email)
	case model.OpenVPNXOR:
		configContent = fmt.Sprintf("client\ndev tun\nproto udp\nremote ...")
		filename = fmt.Sprintf("%s.ovpn", client.Email)
	default:
		ctx.String(http.StatusBadRequest, "Not a VPN protocol")
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.String(http.StatusOK, configContent)
}
