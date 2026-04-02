package service

import (
	"sync"

	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/xray"
	"github.com/sky-night-net/snet/xray/builder"
)

type XrayService struct {
	process *xray.Process
	api     *xray.XrayAPI
	lock    sync.Mutex
}

var (
	xrayInstance *XrayService
	xrayOnce     sync.Once
)

func GetXrayService() *XrayService {
	xrayOnce.Do(func() {
		xrayInstance = &XrayService{
			api: &xray.XrayAPI{},
		}
	})
	return xrayInstance
}

// ApplyConfig regenerates the full Xray configuration from the database
// and restarts the Xray core process.
func (s *XrayService) ApplyConfig() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	configMgr := builder.NewConfigManager()
	cfg, err := configMgr.GenerateFullConfig()
	if err != nil {
		logger.Errorf("Failed to generate Xray config: %v", err)
		return err
	}

	// Stop existing process if running
	if s.process != nil && s.process.IsRunning() {
		s.process.Stop()
	}

	// Start new process
	s.process = xray.NewProcess(cfg)
	err = s.process.Start()
	if err != nil {
		logger.Errorf("Failed to start Xray process: %v", err)
		return err
	}

	// Wait a moment for the process to initialize its API
	go func() {
		// Small delay to ensure the API port is listening
		// In a production environment, we'd use a retry loop or readiness check
		_ = s.api.Init(s.process.GetAPIPort())
	}()

	return nil
}

func (s *XrayService) GetAPI() *xray.XrayAPI {
	return s.api
}

func (s *XrayService) IsRunning() bool {
	return s.process != nil && s.process.IsRunning()
}
