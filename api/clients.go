package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/service"
	"github.com/sky-night-net/snet/vpn/adapters"
)

type ClientController struct{}

func NewClientController() *ClientController {
	return &ClientController{}
}

// getServerHost returns the public server IP/host to use in client config generation.
// Priority: EXTERNAL_IP env var > Request.Host (stripped of port).
func getServerHost(ctx *gin.Context) string {
	if extIP := os.Getenv("EXTERNAL_IP"); extIP != "" {
		return extIP
	}
	host := ctx.Request.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return host
}

func (c *ClientController) AddClient(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	var client model.Client
	if err := ctx.ShouldBindJSON(&client); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}

	db := database.GetDB()
	var inbound model.Inbound
	if err := db.First(&inbound, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "Inbound not found"})
		return
	}

	// Parse settings and add client
	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	if settings == nil {
		settings = make(map[string]any)
	}

	clients, ok := settings["clients"].([]any)
	if !ok {
		clients = make([]any, 0)
	}
	clients = append(clients, client)
	settings["clients"] = clients

	updatedSettings, _ := json.Marshal(settings)
	inbound.Settings = string(updatedSettings)
	db.Save(&inbound)

	// Update VPN adapter if needed (for certificates etc.)
	adapter, err := adapters.GetAdapter(inbound.Protocol)
	if err == nil {
		_ = adapter.AddClient(&inbound, &client)
		// If it's a VPN protocol, trigger manager sync
		if inbound.Protocol == "amneziawg" || inbound.Protocol == "amneziawg-v1" || inbound.Protocol == "openvpn" {
			vpnSvc := service.GetVpnService()
			_ = vpnSvc.GetManager().RestartInbound(&inbound)
		}
	}

	// Live update via gRPC and config refresh for Xray
	xraySvc := service.GetXrayService()
	if xraySvc.IsRunning() {
		userMap := make(map[string]any)
		userJson, _ := json.Marshal(client)
		json.Unmarshal(userJson, &userMap)
		_ = xraySvc.GetAPI().AddUser(string(inbound.Protocol), inbound.Tag, userMap)
	}
	_ = xraySvc.ApplyConfig()

	ctx.JSON(http.StatusOK, gin.H{"success": true, "msg": "Client added", "obj": client})
}

func (c *ClientController) RemoveClient(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)
	email := ctx.Param("clientId") // clientId is email in our frontend

	db := database.GetDB()
	var inbound model.Inbound
	if err := db.First(&inbound, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "Inbound not found"})
		return
	}

	// Find the client before removing to have its data
	var targetClient model.Client
	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	if clients, ok := settings["clients"].([]any); ok {
		newClients := make([]any, 0)
		for _, cl := range clients {
			if m, ok := cl.(map[string]any); ok {
				mEmail, _ := m["email"].(string)
				mId, _ := m["id"].(string)
				if mEmail == email || mId == email {
					jb, _ := json.Marshal(cl)
					json.Unmarshal(jb, &targetClient)
				} else {
					newClients = append(newClients, cl)
				}
			}
		}
		settings["clients"] = newClients
		updatedSettings, _ := json.Marshal(settings)
		inbound.Settings = string(updatedSettings)
		db.Save(&inbound)
	}

	// Update VPN adapter if needed
	adapter, err := adapters.GetAdapter(inbound.Protocol)
	if err == nil {
		_ = adapter.RemoveClient(&inbound, &targetClient)
		if inbound.Protocol == "amneziawg" || inbound.Protocol == "amneziawg-v1" || inbound.Protocol == "openvpn" {
			vpnSvc := service.GetVpnService()
			_ = vpnSvc.GetManager().RestartInbound(&inbound)
		}
	}

	// Live update via gRPC and config refresh for Xray
	xraySvc := service.GetXrayService()
	if xraySvc.IsRunning() {
		_ = xraySvc.GetAPI().RemoveUser(inbound.Tag, email)
	}
	_ = xraySvc.ApplyConfig()

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

func (c *ClientController) GetConfig(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)
	email := ctx.Param("clientId")

	db := database.GetDB()
	var inbound model.Inbound
	if err := db.First(&inbound, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "Inbound not found"})
		return
	}

	// Find client in settings
	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	clients, _ := settings["clients"].([]any)
	
	var targetClient *model.Client
	for _, cl := range clients {
		if m, ok := cl.(map[string]any); ok {
			mEmail, _ := m["email"].(string)
			mId, _ := m["id"].(string)
			if mEmail == email || mId == email {
				// Re-marshal to model.Client
				jb, _ := json.Marshal(cl)
				json.Unmarshal(jb, &targetClient)
				break
			}
		}
	}

	if targetClient == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "Client not found"})
		return
	}

	adapter, err := adapters.GetAdapter(inbound.Protocol)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	// Use EXTERNAL_IP env var for the host - critical for correct client configs
	host := getServerHost(ctx)

	config, err := adapter.GenerateClientConfig(&inbound, targetClient, host)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"obj":     config,
	})
}

