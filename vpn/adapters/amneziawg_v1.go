package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/util/sys"
)

var DEFAULT_AWG_V1_OBFUSCATION = map[string]interface{}{
	"Jc":   5,
	"Jmin": 50,
	"Jmax": 1000,
	"S1":   69,
	"S2":   115,
	"H1":   924883749,
	"H2":   16843009,
	"H3":   305419896,
	"H4":   878082202,
}

type AmneziaWGv1Adapter struct {
	protocol model.Protocol
}

func NewAmneziaWGv1Adapter() *AmneziaWGv1Adapter {
	return &AmneziaWGv1Adapter{
		protocol: model.AmneziaWGv1,
	}
}

func (a *AmneziaWGv1Adapter) Protocol() model.Protocol {
	return a.protocol
}

func (a *AmneziaWGv1Adapter) Start(inbound *model.Inbound) error {
	iface := a.ifaceName(inbound)

	// Ensure config dir exists
	configDir := "/etc/amnezia/amneziawg"
	os.MkdirAll(configDir, 0755)

	// Generate and write server config
	conf, _ := a.GenerateServerConfig(inbound)
	confPath := fmt.Sprintf("%s/%s.conf", configDir, iface)
	os.WriteFile(confPath, []byte(conf), 0600)

	// Proactive cleanup
	a.Stop(inbound)

	// Start natively using awg-quick
	_, err := sys.Execute(fmt.Sprintf("awg-quick up %s", confPath))
	return err
}

func (a *AmneziaWGv1Adapter) Stop(inbound *model.Inbound) error {
	iface := a.ifaceName(inbound)
	configDir := "/etc/amnezia/amneziawg"
	confPath := fmt.Sprintf("%s/%s.conf", configDir, iface)

	// Stop natively
	sys.Execute(fmt.Sprintf("awg-quick down %s", confPath))
	sys.Execute(fmt.Sprintf("ip link delete %s", iface))

	return nil
}

func (a *AmneziaWGv1Adapter) IsRunning(inbound *model.Inbound) bool {
	iface := a.ifaceName(inbound)
	res, err := sys.Execute(fmt.Sprintf("ip link show %s", iface))
	return err == nil && strings.Contains(res, iface)
}

func (a *AmneziaWGv1Adapter) AddClient(inbound *model.Inbound, client *model.Client) error {
	iface := a.ifaceName(inbound)

	cmd := fmt.Sprintf("awg set %s peer %s allowed-ips %s",
		iface, client.PublicKey, client.AllowedIPs)

	if client.PresharedKey != "" {
		pskPath := fmt.Sprintf("/etc/amnezia/amneziawg/psk_%s.key", client.Email)
		os.WriteFile(pskPath, []byte(client.PresharedKey+"\n"), 0600)
		defer os.Remove(pskPath)

		cmd = fmt.Sprintf("awg set %s peer %s allowed-ips %s preshared-key %s",
			iface, client.PublicKey, client.AllowedIPs, pskPath)
	}

	_, err := sys.Execute(cmd)
	return err
}

func (a *AmneziaWGv1Adapter) RemoveClient(inbound *model.Inbound, client *model.Client) error {
	iface := a.ifaceName(inbound)
	_, err := sys.Execute(fmt.Sprintf("awg set %s peer %s remove", iface, client.PublicKey))
	return err
}

func (a *AmneziaWGv1Adapter) GenerateKeypair() (KeyPair, error) {
	priv, err := sys.Execute("awg genkey")
	if err != nil {
		return KeyPair{}, err
	}
	priv = strings.TrimSpace(priv)

	pub, err := sys.Execute(fmt.Sprintf("echo '%s' | awg pubkey", priv))
	if err != nil {
		return KeyPair{}, err
	}
	pub = strings.TrimSpace(pub)

	psk, err := sys.Execute("awg genpsk")
	if err != nil {
		return KeyPair{}, err
	}
	psk = strings.TrimSpace(psk)

	return KeyPair{
		PrivateKey:   priv,
		PublicKey:    pub,
		PresharedKey: psk,
	}, nil
}

func (a *AmneziaWGv1Adapter) GenerateServerConfig(inbound *model.Inbound) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)

	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.StreamSettings), &obfs)
	for k, v := range DEFAULT_AWG_V1_OBFUSCATION {
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
		"# AmneziaWG v1 Obfuscation Parameters",
		fmt.Sprintf("S1 = %v", obfs["S1"]),
		fmt.Sprintf("S2 = %v", obfs["S2"]),
		fmt.Sprintf("H1 = %v", obfs["H1"]),
		fmt.Sprintf("H2 = %v", obfs["H2"]),
		fmt.Sprintf("H3 = %v", obfs["H3"]),
		fmt.Sprintf("H4 = %v", obfs["H4"]),
	}

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

func (a *AmneziaWGv1Adapter) GenerateClientConfig(inbound *model.Inbound, client *model.Client) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)

	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.StreamSettings), &obfs)
	for k, v := range DEFAULT_AWG_V1_OBFUSCATION {
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
		"# AmneziaWG v1 Obfuscation (Junk packets: client-only)",
		fmt.Sprintf("Jc = %v", obfs["Jc"]),
		fmt.Sprintf("Jmin = %v", obfs["Jmin"]),
		fmt.Sprintf("Jmax = %v", obfs["Jmax"]),
		fmt.Sprintf("S1 = %v", obfs["S1"]),
		fmt.Sprintf("S2 = %v", obfs["S2"]),
		fmt.Sprintf("H1 = %v", obfs["H1"]),
		fmt.Sprintf("H2 = %v", obfs["H2"]),
		fmt.Sprintf("H3 = %v", obfs["H3"]),
		fmt.Sprintf("H4 = %v", obfs["H4"]),
		"",
		"[Peer]",
		fmt.Sprintf("PublicKey = %s", serverPub),
		fmt.Sprintf("Endpoint = %s:%d", serverIP, port),
		"AllowedIPs = 0.0.0.0/0, ::/0",
		"PersistentKeepalive = 25",
	}
	if client.PresharedKey != "" {
		lines = append(lines, fmt.Sprintf("PresharedKey = %s", client.PresharedKey))
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func (a *AmneziaWGv1Adapter) GetTraffic(inbound *model.Inbound) (map[string]Traffic, error) {
	iface := a.ifaceName(inbound)

	res, err := sys.Execute(fmt.Sprintf("awg show %s dump", iface))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(res), "\n")
	trafficMap := make(map[string]Traffic)

	// Skip first line (interface info)
	for i := 1; i < len(lines); i++ {
		parts := strings.Split(lines[i], "\t")
		if len(parts) >= 7 {
			pubKey := parts[0]
			var rx, tx int64
			fmt.Sscanf(parts[5], "%d", &rx)
			fmt.Sscanf(parts[6], "%d", &tx)

			trafficMap[pubKey] = Traffic{
				Up:   tx,
				Down: rx,
			}
		}
	}

	return trafficMap, nil
}

func (a *AmneziaWGv1Adapter) CheckPrerequisites() error {
	_, err := sys.Execute("awg-quick --version")
	if err != nil {
		return fmt.Errorf("amneziawg-tools (awg-quick) is not installed: %v", err)
	}
	return nil
}

func (a *AmneziaWGv1Adapter) ifaceName(inbound *model.Inbound) string {
	return fmt.Sprintf("awg%d", inbound.Id)
}

func init() {
	Register(model.AmneziaWGv1, NewAmneziaWGv1Adapter())
}
