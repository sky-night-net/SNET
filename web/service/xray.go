package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/xray"
	"go.uber.org/atomic"
)

var (
	xrayProcess       *xray.Process
	xrayLock          sync.Mutex
	isNeedXrayRestart atomic.Bool
	isManuallyStopped atomic.Bool
)

type XrayService struct {
	settingService SettingService
	xrayAPI        xray.XrayAPI
}

func (s *XrayService) IsXrayRunning() bool {
	return xrayProcess != nil && xrayProcess.IsRunning()
}

func (s *XrayService) GetXrayConfig() (*xray.Config, error) {
	manager := builder.NewConfigManager()
	return manager.GenerateFullConfig()
}

func (s *XrayService) RestartXray(xrayConfig *xray.Config) error {
	xrayLock.Lock()
	defer xrayLock.Unlock()
	
	isManuallyStopped.Store(false)
	
	if s.IsXrayRunning() {
		xrayProcess.Stop()
		time.Sleep(500 * time.Millisecond) // Give time to cleanup
	}

	xrayProcess = xray.NewProcess(xrayConfig)
	err := xrayProcess.Start()
	if err != nil {
		return err
	}
	
	// Wait a bit for API to become available
	time.Sleep(1 * time.Second)
	
	return nil
}

func (s *XrayService) StopXray() error {
	xrayLock.Lock()
	defer xrayLock.Unlock()
	isManuallyStopped.Store(true)
	if s.IsXrayRunning() {
		return xrayProcess.Stop()
	}
	return errors.New("xray is not running")
}

func (s *XrayService) GetXrayTraffic() ([]*xray.Traffic, []*xray.ClientTraffic, error) {
    if !s.IsXrayRunning() {
        return nil, nil, errors.New("xray is not running")
    }
    
    apiPort := xrayProcess.GetAPIPort()
    if apiPort == 0 {
    	// Look up in config if not set in process yet
    	for _, ib := range xrayProcess.GetConfig().InboundConfigs {
    		if ib.Tag == "api" {
    			apiPort = ib.Port
    			break
    		}
    	}
    }
    
    if apiPort == 0 {
    	return nil, nil, errors.New("xray API port not found")
    }

    s.xrayAPI.Init(apiPort)
    defer s.xrayAPI.Close()

    return s.xrayAPI.GetTraffic(true)
}
