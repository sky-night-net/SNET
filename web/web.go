package web

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/sky-night-net/snet/config"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/vpn"
	"github.com/sky-night-net/snet/web/controller"
	"github.com/sky-night-net/snet/web/job"
	"github.com/sky-night-net/snet/web/middleware"
	"github.com/sky-night-net/snet/web/service"
	"github.com/sky-night-net/snet/web/session"
)

type Server struct {
	httpServer *http.Server
	listener   net.Listener

	inboundService *service.InboundService
	xrayService    *service.XrayService
	vpnService     *service.VPNService
	settingService *service.SettingService
	userService    *service.UserService

	cron *cron.Cron

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	vpnMgr := vpn.NewProcessManager()
	
	s := &Server{
		ctx:            ctx,
		cancel:         cancel,
		settingService: &service.SettingService{},
		userService:    &service.UserService{},
		xrayService:    &service.XrayService{},
		vpnService:     service.NewVPNService(vpnMgr),
	}
	s.inboundService = service.NewInboundService(s.vpnService, s.xrayService, s.settingService)
	
	// Start the self-healing reconciler
	vpnMgr.StartReconciler()
	
	return s
}

func (s *Server) GetInboundService() *service.InboundService {
	return s.inboundService
}

func (s *Server) GetSettingService() *service.SettingService {
	return s.settingService
}

func (s *Server) initRouter() (*gin.Engine, error) {
	if config.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()
	engine.Use(gzip.Gzip(gzip.DefaultCompression))
	
	store := session.NewStore()
	engine.Use(sessions.Sessions("snet-v2", store))
	
	basePath, _ := s.settingService.GetBasePath()
	
	// API routes (protected)
	api := engine.Group(basePath + "api")
	api.Use(middleware.Auth)
	{
		controller.NewServerController(api)
		controller.NewSettingController(api)
		controller.NewSSLController(api)
		controller.NewLogController(api)
		
		ibCtrl := controller.NewInboundController(api)
		ibCtrl.SetInboundService(s.inboundService)
		ibCtrl.RegisterRoutes(api)
		
		controller.NewVPNConfigController(api, s.inboundService)
	}
	
	// Public routes
	public := engine.Group(basePath)
	controller.NewIndexController(public)
	
	public.GET("/", func(c *gin.Context) {
		if session.IsLogin(c) {
			c.Redirect(http.StatusTemporaryRedirect, basePath+"panel")
		} else {
			c.Redirect(http.StatusTemporaryRedirect, basePath+"login")
		}
	})
	
	// Setup HTML rendering from embedded FS
	// In production we would use LoadHTMLFS, for now we will adapt the GET handlers
	
	public.GET("/login", func(c *gin.Context) {
		data, _ := htmlFS.ReadFile("html/login.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	
	panel := public.Group("panel")
	panel.Use(middleware.Auth)
	panel.GET("/", func(c *gin.Context) {
		data, _ := htmlFS.ReadFile("html/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	return engine, nil
}

func (s *Server) startTask() {
	s.cron.AddJob("@every 10s", job.NewTrafficJob(s.inboundService, s.xrayService, s.vpnService))
	
	// Auto-start enabled inbounds
	go func() {
		time.Sleep(1 * time.Second)
		// s.vpnService.SyncAll() // Deprecated: Reconciler handles this
		// s.xrayService.SyncAll() // To be implemented
	}()
}

func (s *Server) Start() error {
	loc, _ := s.settingService.GetTimeLocation()
	s.cron = cron.New(cron.WithLocation(loc), cron.WithSeconds())
	s.cron.Start()

	engine, err := s.initRouter()
	if err != nil {
		return err
	}

	port, _ := s.settingService.GetPort()
	bind := config.GetEnvBindAddress()
	listenAddr := net.JoinHostPort(bind, strconv.Itoa(port))
	
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	s.listener = listener

	s.httpServer = &http.Server{
		Handler: engine,
	}

	go func() {
		logger.Infof("SNET v2 Web Server running on %s", listenAddr)
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("Web server error:", err)
		}
	}()

	s.startTask()

	return nil
}

func (s *Server) Stop() error {
	s.cancel()
	if s.cron != nil {
		s.cron.Stop()
	}
	if s.httpServer != nil {
		return s.httpServer.Shutdown(s.ctx)
	}
	return nil
}
