package sys

import (
	"os/exec"
	"strings"
)

// GetDefaultInterface returns the name of the default network interface by parsing 'ip route'.
func GetDefaultInterface() string {
	cmd := exec.Command("sh", "-c", "ip route | grep default | awk '{print $5}' | head -n 1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "eth0" // Fallback
	}
	iface := strings.TrimSpace(string(output))
	if iface == "" {
		return "eth0" 
	}
	return iface
}
