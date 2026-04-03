package sys

import (
	"fmt"

	"github.com/sky-night-net/snet/logger"
)

type FirewallManager struct {
}

func NewFirewallManager() *FirewallManager {
	return &FirewallManager{}
}

func (m *FirewallManager) SetupNAT(iface string) error {
	logger.Infof("Setting up NAT for interface %s", iface)
	
	// 1. Enable IPv4 forwarding
	Execute("sysctl -w net.ipv4.ip_forward=1")

	// 2. Add MASQUERADE rule
	// Get default interface instead of hardcoding eth0
	publicIface := GetDefaultInterface()
	err := m.addRuleOnlyIfMissing("nat", "POSTROUTING", fmt.Sprintf("-o %s -j MASQUERADE", publicIface))
	if err != nil {
		// Fallback to a wider rule if specific interface fails
		m.addRuleOnlyIfMissing("nat", "POSTROUTING", "-j MASQUERADE")
	}

	// 3. Add FORWARD rules
	m.addRuleOnlyIfMissing("filter", "FORWARD", fmt.Sprintf("-i %s -j ACCEPT", iface))
	m.addRuleOnlyIfMissing("filter", "FORWARD", fmt.Sprintf("-o %s -j ACCEPT", iface))

	return nil
}

func (m *FirewallManager) CleanupNAT(iface string) {
	logger.Infof("Cleaning up NAT for interface %s", iface)
	Execute(fmt.Sprintf("iptables -D FORWARD -i %s -j ACCEPT", iface))
	Execute(fmt.Sprintf("iptables -D FORWARD -o %s -j ACCEPT", iface))
}

func (m *FirewallManager) addRuleOnlyIfMissing(table, chain, rule string) error {
	checkCmd := fmt.Sprintf("iptables -t %s -C %s %s", table, chain, rule)
	_, err := Execute(checkCmd)
	if err != nil {
		// Rule doesn't exist, add it
		addCmd := fmt.Sprintf("iptables -t %s -A %s %s", table, chain, rule)
		_, err = Execute(addCmd)
		return err
	}
	return nil
}

func (m *FirewallManager) FlushAllSNETRules() {
	// Logic to cleanup all rules tagged with a specific comment if we use comments
	// For now, simple cleanup is enough
}
