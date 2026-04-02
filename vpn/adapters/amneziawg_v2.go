package adapters

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sky-night-net/snet/database/model"
)

var DEFAULT_AWG_V2_OBFUSCATION = map[string]interface{}{
	"Jc":   5,
	"Jmin": 50,
	"Jmax": 1000,
	"S1":   69,
	"S2":   115,
	"S3":   69,
	"S4":   69,
	"H1":   "10000000-20000000",
	"H2":   "20000001-30000000",
	"H3":   "30000001-40000000",
	"H4":   "40000001-50000000",
	"I1":   "<b 0xabcd010000010000000000000562656c657402746d0000010001>",
	"I2":   "<b 0x1234010000010000000000000573706565640562656c6574026d650000010001>",
	"I3":   "",
	"I4":   "",
	"I5":   "",
}

type AmneziaWGv2Adapter struct {
	AmneziaWGv1Adapter
}

func NewAmneziaWGv2Adapter() *AmneziaWGv2Adapter {
	return &AmneziaWGv2Adapter{
		AmneziaWGv1Adapter: *NewAmneziaWGv1Adapter(),
	}
}

func (a *AmneziaWGv2Adapter) Protocol() model.Protocol {
	return model.AmneziaWGv2
}

func (a *AmneziaWGv2Adapter) GenerateServerConfig(inbound *model.Inbound) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)

	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.StreamSettings), &obfs)
	for k, v := range DEFAULT_AWG_V2_OBFUSCATION {
		if _, ok := obfs[k]; !ok {
			obfs[k] = v
		}
	}

	priv := settings["private_key"].(string)
	addr := settings["address"].(string)
	port := inbound.Port
	mtu := 1420
	if m, ok := settings["mtu"].(float64); ok {
		mtu = int(m)
	}

	lines := []string{
		"[Interface]",
		fmt.Sprintf("PrivateKey = %s", priv),
		fmt.Sprintf("Address = %s", addr),
		fmt.Sprintf("ListenPort = %d", port),
		fmt.Sprintf("MTU = %d", mtu),
		"",
		"# AmneziaWG v2 Obfuscation Parameters",
		fmt.Sprintf("S1 = %v", obfs["S1"]),
		fmt.Sprintf("S2 = %v", obfs["S2"]),
		fmt.Sprintf("S3 = %v", obfs["S3"]),
		fmt.Sprintf("S4 = %v", obfs["S4"]),
		fmt.Sprintf("H1 = %v", obfs["H1"]),
		fmt.Sprintf("H2 = %v", obfs["H2"]),
		fmt.Sprintf("H3 = %v", obfs["H3"]),
		fmt.Sprintf("H4 = %v", obfs["H4"]),
	}

	for _, iKey := range []string{"I1", "I2", "I3", "I4", "I5"} {
		if val, ok := obfs[iKey].(string); ok && val != "" {
			lines = append(lines, fmt.Sprintf("%s = %s", iKey, val))
		}
	}

	// Add peers
	// Extract clients from settings
	var clients []model.Client
	if s, ok := settings["clients"].([]interface{}); ok {
		bs, _ := json.Marshal(s)
		json.Unmarshal(bs, &clients)
	}

	for _, client := range clients {
		if !client.Enable {
			continue
		}
		lines = append(lines, "", fmt.Sprintf("# Client: %s", client.Email), "[Peer]",
			fmt.Sprintf("PublicKey = %s", client.PublicKey),
			fmt.Sprintf("AllowedIPs = %s", client.AllowedIPs))
		if client.PresharedKey != "" {
			lines = append(lines, fmt.Sprintf("PresharedKey = %s", client.PresharedKey))
		}
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func (a *AmneziaWGv2Adapter) GenerateClientConfig(inbound *model.Inbound, client *model.Client) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)

	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.StreamSettings), &obfs)
	for k, v := range DEFAULT_AWG_V2_OBFUSCATION {
		if _, ok := obfs[k]; !ok {
			obfs[k] = v
		}
	}

	serverPub := settings["public_key"].(string)
	serverIP := settings["server_ip"].(string)
	port := inbound.Port
	dns := "1.1.1.1, 8.8.8.8"
	if d, ok := settings["dns"].(string); ok {
		dns = d
	}
	mtu := 1420
	if m, ok := settings["mtu"].(float64); ok {
		mtu = int(m)
	}

	lines := []string{
		"[Interface]",
		fmt.Sprintf("PrivateKey = %s", client.PrivateKey),
		fmt.Sprintf("Address = %s", client.AllowedIPs),
		fmt.Sprintf("DNS = %s", dns),
		fmt.Sprintf("MTU = %d", mtu),
		"",
		"# AmneziaWG v2 Obfuscation",
		fmt.Sprintf("Jc = %v", obfs["Jc"]),
		fmt.Sprintf("Jmin = %v", obfs["Jmin"]),
		fmt.Sprintf("Jmax = %v", obfs["Jmax"]),
		fmt.Sprintf("S1 = %v", obfs["S1"]),
		fmt.Sprintf("S2 = %v", obfs["S2"]),
		fmt.Sprintf("S3 = %v", obfs["S3"]),
		fmt.Sprintf("S4 = %v", obfs["S4"]),
		fmt.Sprintf("H1 = %v", obfs["H1"]),
		fmt.Sprintf("H2 = %v", obfs["H2"]),
		fmt.Sprintf("H3 = %v", obfs["H3"]),
		fmt.Sprintf("H4 = %v", obfs["H4"]),
	}

	for _, iKey := range []string{"I1", "I2", "I3", "I4", "I5"} {
		if val, ok := obfs[iKey].(string); ok && val != "" {
			lines = append(lines, fmt.Sprintf("%s = %s", iKey, val))
		}
	}

	lines = append(lines, "", "[Peer]",
		fmt.Sprintf("PublicKey = %s", serverPub),
		fmt.Sprintf("Endpoint = %s:%d", serverIP, port),
		"AllowedIPs = 0.0.0.0/0, ::/0",
		"PersistentKeepalive = 25")

	if client.PresharedKey != "" {
		lines = append(lines, fmt.Sprintf("PresharedKey = %s", client.PresharedKey))
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func init() {
	Register(model.AmneziaWGv2, NewAmneziaWGv2Adapter())
}
