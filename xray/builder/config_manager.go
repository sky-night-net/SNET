package builder

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/xray"
)

type ConfigManager struct {
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

func (m *ConfigManager) GenerateFullConfig() (*xray.Config, error) {
	db := database.GetDB()
	var inbounds []model.Inbound
	err := db.Preload("Clients").Find(&inbounds).Error
	if err != nil {
		return nil, err
	}

	builder := NewXrayConfigBuilder()
	now := time.Now().Unix()

	for _, ib := range inbounds {
		if !ib.Enable {
			continue
		}

		// Skip VPN protocols
		if m.isVPN(ib.Protocol) {
			continue
		}

		// Build Xray settings based on protocol
		settings := m.buildSettings(&ib, now)
		builder.AddInbound(string(ib.Protocol), ib.Port, ib.Tag, settings)
	}

	return builder.Build(), nil
}

func (m *ConfigManager) isVPN(p model.Protocol) bool {
	return p == model.AmneziaWGv1 || p == model.AmneziaWGv2 || p == model.OpenVPNXOR
}

func (m *ConfigManager) buildSettings(ib *model.Inbound, now int64) interface{} {
	// Filter enabled and non-expired clients
	var clients []VLESSClient
	for _, c := range ib.Clients {
		if !c.Enable {
			continue
		}
		if c.ExpiryTime > 0 && c.ExpiryTime < now {
			continue
		}
		if c.Total > 0 && (c.Up+c.Down) >= c.Total {
			continue
		}

		clients = append(clients, VLESSClient{
			ID:    c.UUID, // Assume UUID field exists or use ID
			Email: c.Email,
		})
	}

	switch ib.Protocol {
	case "vless":
		return VLESSSettings{
			Clients:    clients,
			Decryption: "none",
		}
	// Add other protocols (vmess, trojan, shadowsocks) here
	default:
		return map[string]interface{}{"clients": clients}
	}
}
