package adapters

import (
	"fmt"
	"github.com/sky-night-net/snet/database/model"
)

var (
	registry = make(map[model.Protocol]VPNAdapter)
)

func Register(protocol model.Protocol, adapter VPNAdapter) {
	registry[protocol] = adapter
}

func GetAdapter(protocol model.Protocol) (VPNAdapter, error) {
	adapter, ok := registry[protocol]
	if !ok {
		return nil, fmt.Errorf("no adapter registered for protocol: %s", protocol)
	}
	return adapter, nil
}

func GetProtocols() []model.Protocol {
	protocols := make([]model.Protocol, 0, len(registry))
	for p := range registry {
		protocols = append(protocols, p)
	}
	return protocols
}
