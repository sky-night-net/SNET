package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/service"
)

type FirewallController struct {
	firewallService *service.FirewallService
}

func NewFirewallController() *FirewallController {
	return &FirewallController{
		firewallService: service.GetFirewallService(),
	}
}

func (c *FirewallController) GetAll(ctx *gin.Context) {
	db := database.GetDB()
	var rules []model.FirewallRule
	db.Find(&rules)
	ctx.JSON(http.StatusOK, gin.H{"success": true, "obj": rules})
}

func (c *FirewallController) Create(ctx *gin.Context) {
	var rule model.FirewallRule
	if err := ctx.ShouldBindJSON(&rule); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}

	db := database.GetDB()
	if err := db.Create(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	if rule.Enable {
		err := c.firewallService.ApplyRule(&rule)
		if err != nil {
			fmt.Printf("Firewall rule application error: %v\n", err)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (c *FirewallController) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	db := database.GetDB()
	var rule model.FirewallRule
	if err := db.First(&rule, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "Rule not found"})
		return
	}

	// Remove from system if it was enabled
	if rule.Enable {
		c.firewallService.RemoveRule(&rule)
	}

	db.Delete(&rule)
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
