package service

import (
	"sync"
	"github.com/sky-night-net/snet/vpn"
)

type VpnService struct {
	manager *vpn.ProcessManager
}

var (
	vpnService *VpnService
	vpnOnce    sync.Once
)

func GetVpnService() *VpnService {
	vpnOnce.Do(func() {
		vpnService = &VpnService{
			manager: vpn.NewProcessManager(),
		}
	})
	return vpnService
}

func (s *VpnService) Start() {
	s.manager.StartReconciler()
}

func (s *VpnService) Stop() {
	s.manager.Stop()
}

func (s *VpnService) GetManager() *vpn.ProcessManager {
	return s.manager
}
