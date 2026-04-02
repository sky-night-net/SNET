package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
)

type InboundController struct{}

func NewInboundController() *InboundController {
	return &InboundController{}
}

func (c *InboundController) GetInbounds(ctx *gin.Context) {
	var inbounds []model.Inbound
	db := database.GetDB()
	
	// Preload client statistics
	err := db.Preload("ClientStats").Find(&inbounds).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"obj":     inbounds,
	})
}

func (c *InboundController) CreateInbound(ctx *gin.Context) {
	var inbound model.Inbound
	if err := ctx.ShouldBindJSON(&inbound); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}

	db := database.GetDB()
	
	// Set a default Tag if empty
	if inbound.Tag == "" {
		inbound.Tag = "inbound-" + strconv.Itoa(inbound.Port)
	}

	if err := db.Create(&inbound).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	// TODO: Trigger VpnMgr or XrayService to apply changes

	ctx.JSON(http.StatusOK, gin.H{"success": true, "msg": "Inbound created", "obj": inbound})
}

func (c *InboundController) DeleteInbound(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid ID"})
		return
	}

	db := database.GetDB()
	if err := db.Delete(&model.Inbound{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	// TODO: Trigger cleanup of interfaces/processes

	ctx.JSON(http.StatusOK, gin.H{"success": true, "msg": "Inbound deleted"})
}
