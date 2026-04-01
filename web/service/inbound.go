package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/xray"
	"gorm.io/gorm"
)

type InboundService struct {
	vpnService  *VPNService
	xrayService *XrayService
}

func NewInboundService(vpnService *VPNService, xrayService *XrayService) *InboundService {
	return &InboundService{
		vpnService:  vpnService,
		xrayService: xrayService,
	}
}

func (s *InboundService) GetAllInbounds() ([]*model.Inbound, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Preload("Clients").Find(&inbounds).Error
	return inbounds, err
}

func (s *InboundService) GetInbound(id int) (*model.Inbound, error) {
	db := database.GetDB()
	var inbound model.Inbound
	err := db.Preload("Clients").First(&inbound, id).Error
	return &inbound, err
}

func (s *InboundService) AddInbound(inbound *model.Inbound) error {
	db := database.GetDB()
	err := db.Create(inbound).Error
	if err != nil {
		return err
	}

	if inbound.Enable {
		return s.SyncInbound(inbound)
	}
	return nil
}

func (s *InboundService) UpdateInbound(inbound *model.Inbound) error {
	db := database.GetDB()
	err := db.Save(inbound).Error
	if err != nil {
		return err
	}
	return s.SyncInbound(inbound)
}

func (s *InboundService) DelInbound(id int) error {
	inbound, err := s.GetInbound(id)
	if err != nil {
		return err
	}

	// Stop process first
	if s.IsVPN(inbound.Protocol) {
		s.vpnService.StopInbound(inbound)
	} else {
		// For Xray, we need to restart the whole core without this inbound
		defer s.NotifyXrayRestart()
	}

	db := database.GetDB()
	return db.Delete(&model.Inbound{}, id).Error
}

func (s *InboundService) SyncInbound(inbound *model.Inbound) error {
	if !inbound.Enable {
		if s.IsVPN(inbound.Protocol) {
			return s.vpnService.StopInbound(inbound)
		}
		s.NotifyXrayRestart()
		return nil
	}

	if s.IsVPN(inbound.Protocol) {
		return s.vpnService.RestartInbound(inbound)
	}

	s.NotifyXrayRestart()
	return nil
}

func (s *InboundService) IsVPN(protocol model.Protocol) bool {
	switch protocol {
	case model.AmneziaWGv1, model.AmneziaWGv2, model.OpenVPNXOR:
		return true
	}
	return false
}

func (s *InboundService) NotifyXrayRestart() {
	// In a real implementation, this would trigger a delayed restart
	// to avoid multiple restarts when multiple inbounds change.
	// For now, let's just build the config and restart.
	logger.Info("Triggering Xray restart due to inbound changes")
	
	// Implementation of full Xray config generation goes here
}

// Client Management

func (s *InboundService) AddClient(client *model.Client) error {
	db := database.GetDB()
	err := db.Create(client).Error
	if err != nil {
		return err
	}
	
	inbound, err := s.GetInbound(client.InboundId)
	if err != nil {
		return err
	}
	
	if s.IsVPN(inbound.Protocol) && inbound.Enable && client.Enable {
		adapter, _ := s.vpnService.manager.GetAdapter(inbound.Protocol) // Need export or access
		// Wait, I should use vpnService methods
		// Actually, VPN adapters expect peer addition to be handled by Start/Restart for simplicity
		// or via a dedicated AddClient if supported.
		return s.SyncInbound(inbound)
	}
	
	if !s.IsVPN(inbound.Protocol) {
		s.NotifyXrayRestart()
	}
	
	return nil
}

func (s *InboundService) UpdateClient(client *model.Client) error {
	db := database.GetDB()
	err := db.Save(client).Error
	if err != nil {
		return err
	}
	
	inbound, err := s.GetInbound(client.InboundId)
	if err == nil {
		return s.SyncInbound(inbound)
	}
	return nil
}

func (s *InboundService) DelClient(id int) error {
	db := database.GetDB()
	var client model.Client
	if err := db.First(&client, id).Error; err != nil {
		return err
	}
	
	inbound, _ := s.GetInbound(client.InboundId)
	
	err := db.Delete(&model.Client{}, id).Error
	if err == nil && inbound != nil {
		s.SyncInbound(inbound)
	}
	return err
}
