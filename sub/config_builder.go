package sub

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/web/service"
)

type ConfigBuilder struct {
	inboundService *service.InboundService
}

func NewConfigBuilder(ib *service.InboundService) *ConfigBuilder {
	return &ConfigBuilder{
		inboundService: ib,
	}
}

func (b *ConfigBuilder) BuildBase64(token string) (string, error) {
	// 1. Get all enabled inbounds
	inbounds, err := b.inboundService.GetAllInbounds()
	if err != nil {
		return "", err
	}

	var configs []string

	for _, ib := range inbounds {
		if !ib.Enable {
			continue
		}
		
		// For each inbound, find the client matching the token
		for _, client := range ib.Clients {
			if client.Email == token || client.SubID == token {
				if !client.Enable {
					continue
				}
				
				config := b.GenerateConfigLink(ib, &client)
				if config != "" {
					configs = append(configs, config)
				}
			}
		}
	}

	if len(configs) == 0 {
		return "", fmt.Errorf("no configs found for token")
	}

	// 3x-ui standard: join with newlines and base64 encode
	raw := strings.Join(configs, "\n")
	return base64.StdEncoding.EncodeToString([]byte(raw)), nil
}

func (b *ConfigBuilder) GenerateConfigLink(ib *model.Inbound, client *model.Client) string {
	// Identify protocol and generate SIP002-like link
	// Example: vmess://... or vless://...
	
	// VPN protocols:
	switch ib.Protocol {
	case model.AmneziaWGv1, model.AmneziaWGv2:
		return fmt.Sprintf("amneziawg://%s@%s:%d/?title=%s", client.Email, "your-server-ip", ib.Port, ib.Remark)
	case model.OpenVPNXOR:
		return fmt.Sprintf("openvpn://%s@%s:%d/?title=%s", client.Email, "your-server-ip", ib.Port, ib.Remark)
	}
	
	// Xray protocols: (Stub for now, would use 3x-ui link logic)
	return fmt.Sprintf("%s://%s@%s:%d#%s", string(ib.Protocol), client.Email, "your-server-ip", ib.Port, ib.Remark)
}
