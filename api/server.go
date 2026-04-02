package api

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
}

func NewServer() *Server {
	engine := gin.Default()

	// Setup CORS
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Adjust for production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	s := &Server{
		Engine: engine,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	api := s.Engine.Group("/api")

	// Auth
	authController := NewAuthController()
	api.POST("/login", authController.Login)

	// Protected routes using AuthMiddleware (stubbed for now)
	protected := api.Group("/")
	protected.Use(AuthMiddleware())
	
	// Inbounds
	inboundController := NewInboundController()
	protected.GET("/inbounds", inboundController.GetInbounds)
	protected.POST("/inbounds", inboundController.CreateInbound)
	protected.PUT("/inbounds/:id", inboundController.UpdateInbound)
	protected.DELETE("/inbounds/:id", inboundController.DeleteInbound)


	// Clients
	clientController := NewClientController()
	protected.POST("/inbounds/:id/clients", clientController.AddClient)
	protected.DELETE("/inbounds/:id/clients/:clientId", clientController.RemoveClient)
	protected.POST("/clients/keygen/:protocol", clientController.Keygen)

	// System
	systemController := NewSystemController()
	protected.GET("/system/status", systemController.GetStatus)

	// Settings
	settingsController := NewSettingsController()
	protected.GET("/settings", settingsController.GetAll)
	protected.PUT("/settings", settingsController.Update)
	protected.POST("/settings/password", settingsController.ChangePassword)
	protected.GET("/settings/backup", settingsController.DownloadBackup)

	// Frontend
	ServeFrontend(s.Engine)
}

func (s *Server) Start(addr string) error {
	log.Printf("Starting SNET 3.0 API Server on %s", addr)
	return s.Engine.Run(addr)
}
