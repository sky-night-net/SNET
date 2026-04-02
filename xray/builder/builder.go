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

type StreamSettings struct {
	Network  string          `json:"network"`
	Security string          `json:"security"`
	TLS      *TLSConfig      `json:"tlsSettings,omitempty"`
	Reality  *RealityConfig  `json:"realitySettings,omitempty"`
	WS       *WSConfig       `json:"wsSettings,omitempty"`
	GRPC     *GRPCConfig     `json:"grpcSettings,omitempty"`
	TCP      *TCPConfig      `json:"tcpSettings,omitempty"`
}

type TLSConfig struct {
	ServerName   string   `json:"serverName,omitempty"`
	Alpn         []string `json:"alpn,omitempty"`
	Certificates []Cert   `json:"certificates,omitempty"`
}

type Cert struct {
	CertificateFile string   `json:"certificateFile,omitempty"`
	KeyFile         string   `json:"keyFile,omitempty"`
	Usage           string   `json:"usage,omitempty"`
}

type RealityConfig struct {
	Show        bool              `json:"show"`
	Dest        string            `json:"dest"`
	Xver        int               `json:"xver"`
	ServerNames []string          `json:"serverNames"`
	PrivateKey  string            `json:"privateKey"`
	MinClient   string            `json:"minClient"`
	MaxClient   string            `json:"maxClient"`
	ShortIds    []string          `json:"shortIds"`
}

type WSConfig struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

type GRPCConfig struct {
	ServiceName string `json:"serviceName"`
}

type TCPConfig struct {
	Header json.RawMessage `json:"header,omitempty"`
}

