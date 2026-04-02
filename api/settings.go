package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/util/crypto"
)

type SettingsController struct{}

func NewSettingsController() *SettingsController {
	return &SettingsController{}
}

func (c *SettingsController) GetAll(ctx *gin.Context) {
	var settings []model.Setting
	database.GetDB().Find(&settings)
	
	// Convert to map for easier frontend use
	res := make(map[string]string)
	for _, s := range settings {
		res[s.Key] = s.Value
	}
	
	ctx.JSON(http.StatusOK, gin.H{"success": true, "obj": res})
}

func (c *SettingsController) Update(ctx *gin.Context) {
	var payload map[string]string
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}

	db := database.GetDB()
	for k, v := range payload {
		var setting model.Setting
		if err := db.Where("key = ?", k).First(&setting).Error; err == nil {
			setting.Value = v
			db.Save(&setting)
		} else {
			db.Create(&model.Setting{Key: k, Value: v})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (c *SettingsController) ChangePassword(ctx *gin.Context) {
	var payload struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}

	db := database.GetDB()
	var user model.User
	// Assuming single user for now or get from auth context
	db.First(&user)

	if !crypto.CheckPasswordHash(user.Password, payload.OldPassword) {
		ctx.JSON(http.StatusForbidden, gin.H{"success": false, "msg": "Invalid old password"})
		return
	}

	hashed, _ := crypto.HashPasswordAsBcrypt(payload.NewPassword)
	user.Password = hashed
	db.Save(&user)

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (c *SettingsController) DownloadBackup(ctx *gin.Context) {
	dbPath := "snet.db" // In production, this should come from config
	ctx.FileAttachment(dbPath, "snet_backup.db")
}
