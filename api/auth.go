package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/util/crypto"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid request"})
		return
	}

	db := database.GetDB()
	var user model.User
	err := db.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		log.Printf("User %s not found: %v", req.Username, err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "Неверный логин или пароль"})
		return
	}

	if !crypto.CheckPasswordHash(user.Password, req.Password) {
		log.Printf("Invalid password for user %s", req.Username)
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "Неверный логин или пароль"})
		return
	}

	// For SNET 3.0, we use a simple mock token for frontend dev, 
	// but real implementation should generate a JWT here.
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "Login successful",
		"token":   "mock-jwt-token",
		"obj":     user,
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != "Bearer mock-jwt-token" {
			// Real app will decode JWT here.
			// c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		}
		c.Next()
	}
}
