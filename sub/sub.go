package sub

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/web/service"
)

type Server struct {
	httpServer     *http.Server
	listener       net.Listener
	inboundService *service.InboundService
	settingService *service.SettingService

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(ib *service.InboundService, st *service.SettingService) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ctx:            ctx,
		cancel:         cancel,
		inboundService: ib,
		settingService: st,
	}
}

func (s *Server) initRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	engine.GET("/sub/:token", s.GetSubscription)
	
	return engine
}

func (s *Server) GetSubscription(ctx *gin.Context) {
	token := ctx.Param("token")
	
	// 1. Find client by subscription token
	// For now, let's assume token is SubID or client email for lookup
	// In a real implementation, we'd have a separate SubToken field
	
	// For demo: build a unified config base64
	builder := NewConfigBuilder(s.inboundService)
	content, err := builder.BuildBase64(token)
	if err != nil {
		ctx.String(http.StatusNotFound, "Subscription not found")
		return
	}
	
	ctx.Header("Subscription-Userinfo", "upload=0;download=0;total=0;expire=0")
	ctx.String(http.StatusOK, content)
}

func (s *Server) Start() error {
	portVal, _ := s.settingService.GetSetting("sub_port")
	port, _ := strconv.Atoi(portVal)
	if port == 0 {
		port = 2054 // Default sub port
	}

	listenAddr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	s.listener = listener

	engine := s.initRouter()
	s.httpServer = &http.Server{
		Handler: engine,
	}

	go func() {
		logger.Infof("Subscription Server running on %s", listenAddr)
		_ = s.httpServer.Serve(listener)
	}()

	return nil
}

func (s *Server) Stop() error {
	s.cancel()
	if s.httpServer != nil {
		return s.httpServer.Shutdown(s.ctx)
	}
	return nil
}
