package service

import (
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/vpn"
	"github.com/sky-night-net/snet/vpn/adapters"
)

type VPNService struct {
	manager *vpn.ProcessManager
}

func NewVPNService(manager *vpn.ProcessManager) *VPNService {
	return &VPNService{
		manager: manager,
	}
}

func (s *VPNService) StartInbound(inbound *model.Inbound) error {
	return s.manager.StartInbound(inbound)
}

func (s *VPNService) StopInbound(inbound *model.Inbound) error {
	return s.manager.StopInbound(inbound)
}

func (s *VPNService) RestartInbound(inbound *model.Inbound) error {
	return s.manager.RestartInbound(inbound)
}

func (s *VPNService) GetTraffic(inbound *model.Inbound) (map[string]adapters.Traffic, error) {
	return s.manager.GetTraffic(inbound)
}
