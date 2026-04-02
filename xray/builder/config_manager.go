package builder

import (
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/xray"
)

type ConfigManager struct {
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// GenerateFullConfig builds a full Xray configuration from all enabled
// Xray-based inbounds in the database. VPN protocols (AmneziaWG, OpenVPN XOR)
// are skipped because they are managed by their own adapters.
func (m *ConfigManager) GenerateFullConfig() (*xray.Config, error) {
	db := database.GetDB()
	var inbounds []model.Inbound
	err := db.Find(&inbounds).Error
	if err != nil {
		return nil, err
	}

	builder := NewXrayConfigBuilder()

	for _, ib := range inbounds {
		if !ib.Enable {
			continue
		}

		// Skip VPN protocols — they are not handled by Xray
		if m.isVPN(ib.Protocol) {
			continue
		}

		// Use the model's own method to generate the Xray inbound config.
		// This correctly uses the Settings JSON (which already contains
		// the "clients" array) rather than trying to read clients from
		// a non-existent GORM relationship.
		xrayIb := ib.GenXrayInboundConfig()
		builder.config.InboundConfigs = append(builder.config.InboundConfigs, *xrayIb)
	}

	return builder.Build(), nil
}

func (m *ConfigManager) isVPN(p model.Protocol) bool {
	return p == model.AmneziaWGv1 || p == model.AmneziaWGv2 || p == model.OpenVPNXOR
}
