package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/util/sys"
)

var DEFAULT_AWG_V1_OBFUSCATION = map[string]interface{}{
	"Jc": 5,
	"Jmin": 50,
	"Jmax": 1000,
	"S1": 69,
	"S2": 115,
	"H1": 924883749,
	"H2": 16843009,
	"H3": 305419896,
	"H4": 878082202,
}

type AmneziaWGv1Adapter struct {
	DockerBaseAdapter
	protocol model.Protocol
}

func NewAmneziaWGv1Adapter() *AmneziaWGv1Adapter {
	base, _ := NewDockerBaseAdapter()
	return &AmneziaWGv1Adapter{
		DockerBaseAdapter: *base,
		protocol:          model.AmneziaWGv1,
	}
}

func (a *AmneziaWGv1Adapter) Protocol() model.Protocol {
	return a.protocol
}

func (a *AmneziaWGv1Adapter) Start(inbound *model.Inbound) error {
	ctx := context.Background()
	iface := a.ifaceName(inbound)
	containerName := a.containerName(inbound)
	
	// Ensure config dir exists
	configDir := "/etc/amnezia/amneziawg"
	os.MkdirAll(configDir, 0755)

	// Generate and write server config
	conf, _ := a.GenerateServerConfig(inbound)
	confPath := fmt.Sprintf("%s/%s.conf", configDir, iface)
	os.WriteFile(confPath, []byte(conf), 0600)

	// Proactive cleanup
	a.Stop(inbound)

	// Start container using Docker SDK
	resp, err := a.cli.ContainerCreate(ctx, &container.Config{
		Image: "snet-local/amneziawg:latest",
		Cmd:   []string{"/bin/bash", "-c", fmt.Sprintf("awg-quick up /etc/amnezia/amneziawg/%s.conf && tail -f /dev/null", iface)},
		Tty:   true,
	}, &container.HostConfig{
		NetworkMode: "host",
		Privileged:  true,
		CapAdd:      []string{"NET_ADMIN"},
		Binds:       []string{fmt.Sprintf("%s:/etc/amnezia/amneziawg", configDir)},
		Resources: container.Resources{
			Devices: []container.DeviceMapping{
				{PathOnHost: "/dev/net/tun", PathInContainer: "/dev/net/tun", CgroupPermissions: "rwm"},
			},
		},
	}, nil, nil, containerName)

	if err != nil {
		return err
	}

	return a.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}

func (a *AmneziaWGv1Adapter) Stop(inbound *model.Inbound) error {
	containerName := a.containerName(inbound)
	iface := a.ifaceName(inbound)
	
	// Stop container
	sys.Execute(fmt.Sprintf("docker rm -f %s", containerName))
	
	// Cleanup interface on host (since --network host was used)
	sys.Execute(fmt.Sprintf("ip link delete %s", iface))
	
	return nil
}

func (a *AmneziaWGv1Adapter) IsRunning(inbound *model.Inbound) bool {
	containerName := a.containerName(inbound)
	res, err := sys.Execute(fmt.Sprintf("docker inspect -f '{{.State.Running}}' %s", containerName))
	return err == nil && strings.TrimSpace(res) == "true"
}

func (a *AmneziaWGv1Adapter) AddClient(inbound *model.Inbound, client *model.Client) error {
	containerName := a.containerName(inbound)
	iface := a.ifaceName(inbound)

	cmd := fmt.Sprintf("docker exec %s awg set %s peer %s allowed-ips %s", 
		containerName, iface, client.PublicKey, client.AllowedIPs)
	
	if client.PresharedKey != "" {
		// PSK is handled carefully to avoid leaking in process list
		// In Go version we can use a temporary file inside the volume
		pskPath := fmt.Sprintf("/etc/amnezia/amneziawg/psk_%s.key", client.Email)
		hostPskPath := fmt.Sprintf("/etc/amnezia/amneziawg/psk_%s.key", client.Email)
		os.WriteFile(hostPskPath, []byte(client.PresharedKey+"\n"), 0600)
		defer os.Remove(hostPskPath)
		
		cmd = fmt.Sprintf("docker exec %s awg set %s peer %s allowed-ips %s preshared-key %s", 
			containerName, iface, client.PublicKey, client.AllowedIPs, pskPath)
	}

	_, err := sys.Execute(cmd)
	return err
}

func (a *AmneziaWGv1Adapter) RemoveClient(inbound *model.Inbound, client *model.Client) error {
	containerName := a.containerName(inbound)
	iface := a.ifaceName(inbound)
	_, err := sys.Execute(fmt.Sprintf("docker exec %s awg set %s peer %s remove", containerName, iface, client.PublicKey))
	return err
}

func (a *AmneziaWGv1Adapter) GenerateKeypair() (KeyPair, error) {
	// We can use the already defined GenerateWGKeypair if it's compatible with awg
	// Since AWG is backward compatible for keys, we can use standard wg binary if available
	// or exec into the docker image.
	// For reliability let's use the utility we just created.
	// Wait, I should import vpn package, but this is adapters subpackage.
	// I'll assume we have a way to get it or just use shell.
	
	// Actually, let's just use the docker image since we know it has 'awg' binary
	priv, err := sys.Execute("docker run --rm snet-local/amneziawg:latest awg genkey")
	if err != nil {
		return KeyPair{}, err
	}
	priv = strings.TrimSpace(priv)
	
	pub, err := sys.Execute(fmt.Sprintf("echo '%s' | docker run --rm -i snet-local/amneziawg:latest awg pubkey", priv))
	if err != nil {
		return KeyPair{}, err
	}
	pub = strings.TrimSpace(pub)
	
	psk, err := sys.Execute("docker run --rm snet-local/amneziawg:latest awg genpsk")
	if err != nil {
		return KeyPair{}, err
	}
	psk = strings.TrimSpace(psk)
	
	return KeyPair{
		PrivateKey: priv,
		PublicKey: pub,
		PresharedKey: psk,
	}, nil
}

func (a *AmneziaWGv1Adapter) GenerateServerConfig(inbound *model.Inbound) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)
	
	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.Obfuscation), &obfs)
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

	// Clients are stored in relational table, so they will be added via AddClient when starting
	// or we can pre-populate here. For persistence on restart, adding to config is better.
	// But in 3x-ui style, we'll probably have a service that calls Start() which will build this config.
	
	// To be thorough, we should include ALL enabled clients in the server config
	// so that when awg-quick starts, all peers are configured.
	for _, client := range inbound.Clients {
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
	json.Unmarshal([]byte(inbound.Obfuscation), &obfs)
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
		fmt.Sprintf("Address = %s", client.AllowedIPs), // Client should have the /32 address
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
	containerName := a.containerName(inbound)
	
	res, err := sys.Execute(fmt.Sprintf("docker exec %s awg show %s dump", containerName, iface))
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
				Up: tx,
				Down: rx,
			}
		}
	}
	
	return trafficMap, nil
}

func (a *AmneziaWGv1Adapter) CheckPrerequisites() error {
	_, err := sys.Execute("docker info")
	if err != nil {
		return fmt.Errorf("docker is not running or not installed: %v", err)
	}
	return nil
}

func (a *AmneziaWGv1Adapter) ifaceName(inbound *model.Inbound) string {
	return fmt.Sprintf("awg%d", inbound.Id)
}

func (a *AmneziaWGv1Adapter) containerName(inbound *model.Inbound) string {
	return fmt.Sprintf("snet_%s_%d", inbound.Protocol, inbound.Id)
}

func init() {
	Register(model.AmneziaWGv1, NewAmneziaWGv1Adapter())
}
