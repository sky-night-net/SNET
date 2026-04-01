package xray

import (
	"bytes"

	"github.com/sky-night-net/snet/util/json_util"
)

type Config struct {
	LogConfig        json_util.RawMessage `json:"log"`
	RouterConfig     json_util.RawMessage `json:"routing"`
	DNSConfig        json_util.RawMessage `json:"dns"`
	InboundConfigs   []InboundConfig      `json:"inbounds"`
	OutboundConfigs  json_util.RawMessage `json:"outbounds"`
	Transport        json_util.RawMessage `json:"transport"`
	Policy           json_util.RawMessage `json:"policy"`
	API              json_util.RawMessage `json:"api"`
	Stats            json_util.RawMessage `json:"stats"`
	Reverse          json_util.RawMessage `json:"reverse"`
	FakeDNS          json_util.RawMessage `json:"fakedns"`
	Observatory      json_util.RawMessage `json:"observatory"`
	BurstObservatory json_util.RawMessage `json:"burstObservatory"`
	Metrics          json_util.RawMessage `json:"metrics"`
}

func (c *Config) Equals(other *Config) bool {
	if len(c.InboundConfigs) != len(other.InboundConfigs) {
		return false
	}
	for i, inbound := range c.InboundConfigs {
		if !inbound.Equals(&other.InboundConfigs[i]) {
			return false
		}
	}
	if !bytes.Equal(c.LogConfig, other.LogConfig) {
		return false
	}
	if !bytes.Equal(c.RouterConfig, other.RouterConfig) {
		return false
	}
	if !bytes.Equal(c.DNSConfig, other.DNSConfig) {
		return false
	}
	if !bytes.Equal(c.OutboundConfigs, other.OutboundConfigs) {
		return false
	}
	return true
}
