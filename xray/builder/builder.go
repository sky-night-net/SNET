package builder

import (
	"encoding/json"

	"github.com/sky-night-net/snet/util/json_util"
	"github.com/sky-night-net/snet/xray"
)

type XrayConfigBuilder struct {
	config *xray.Config
}

func NewXrayConfigBuilder() *XrayConfigBuilder {
	return &XrayConfigBuilder{
		config: &xray.Config{
			LogConfig: json.RawMessage(`{"loglevel": "warning"}`),
			InboundConfigs: []xray.InboundConfig{
				{
					Protocol: "vless",
					Port:     10085, // Default API port for 3x-ui style
					Listen:   json_util.RawMessage(`"127.0.0.1"`),
					Tag:      "api",
					Settings: json.RawMessage(`{}`),
				},
			},
			OutboundConfigs: json.RawMessage(`[{"protocol": "freedom"}]`),
		},
	}
}

func (b *XrayConfigBuilder) AddInbound(protocol string, port int, tag string, settings any) *XrayConfigBuilder {
	sJson, _ := json.Marshal(settings)
	ib := xray.InboundConfig{
		Protocol: protocol,
		Port:     port,
		Tag:      tag,
		Settings: sJson,
	}
	b.config.InboundConfigs = append(b.config.InboundConfigs, ib)
	return b
}

func (b *XrayConfigBuilder) Build() *xray.Config {
	return b.config
}

func (b *XrayConfigBuilder) BuildJSON() (string, error) {
	data, err := json.MarshalIndent(b.config, "", "  ")
	return string(data), err
}

// Specialized settings helpers

type VLESSSettings struct {
	Clients    []VLESSClient `json:"clients"`
	Decryption string        `json:"decryption"`
	Fallback   any           `json:"fallback,omitempty"`
}

type VLESSClient struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Flow  string `json:"flow,omitempty"`
}
