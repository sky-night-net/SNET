package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/service"
)

type InboundController struct{}

func NewInboundController() *InboundController {
	return &InboundController{}
}

func (c *InboundController) applyService(ib *model.Inbound) {
	if ib.Protocol == "amneziawg" || ib.Protocol == "amneziawg-v1" || ib.Protocol == "amneziawg-v2" || ib.Protocol == "openvpn-xor" {
		vpnSvc := service.GetVpnService()
		_ = vpnSvc.GetManager().RestartInbound(ib)
	} else {
		xraySvc := service.GetXrayService()
		_ = xraySvc.ApplyConfig()
	}
}

func (c *InboundController) GetInbounds(ctx *gin.Context) {
	var inbounds []model.Inbound
	database.GetDB().Preload("ClientStats").Find(&inbounds)
	ctx.JSON(http.StatusOK, gin.H{"success": true, "obj": inbounds})
}

func (c *InboundController) CreateInbound(ctx *gin.Context) {
	var inbound model.Inbound
	if err := ctx.ShouldBindJSON(&inbound); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}

	if inbound.Tag == "" {
		inbound.Tag = "inbound-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	if err := database.GetDB().Create(&inbound).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	c.applyService(&inbound)
	ctx.JSON(http.StatusOK, gin.H{"success": true, "obj": inbound})
}

func (c *InboundController) UpdateInbound(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var inbound model.Inbound
	if err := database.GetDB().First(&inbound, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "Inbound not found"})
		return
	}

	if err := ctx.ShouldBindJSON(&inbound); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}

	database.GetDB().Save(&inbound)
	c.applyService(&inbound)
	ctx.JSON(http.StatusOK, gin.H{"success": true, "obj": inbound})
}

func (c *InboundController) DeleteInbound(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var inbound model.Inbound
	if err := database.GetDB().First(&inbound, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "Inbound not found"})
		return
	}

	database.GetDB().Delete(&inbound)

	if inbound.Protocol == "amneziawg" || inbound.Protocol == "amneziawg-v1" || inbound.Protocol == "amneziawg-v2" || inbound.Protocol == "openvpn-xor" {
		service.GetVpnService().GetManager().StopInbound(&inbound)
	} else {
		service.GetXrayService().ApplyConfig()
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "msg": "Inbound deleted"})
}
