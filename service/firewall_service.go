package service

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
)

type FirewallService struct {
}

var firewallService *FirewallService

func GetFirewallService() *FirewallService {
	if firewallService == nil {
		firewallService = &FirewallService{}
	}
	return firewallService
}

// Sync applies all enabled rules from the database to the system iptables.
func (s *FirewallService) Sync() error {
	db := database.GetDB()
	var rules []model.FirewallRule
	if err := db.Find(&rules, "enable = ?", true).Error; err != nil {
		return err
	}

	for _, rule := range rules {
		s.ApplyRule(&rule)
	}
	return nil
}

func (s *FirewallService) AddPort(port int, proto string, remark string) error {
	db := database.GetDB()
	var rule model.FirewallRule
	// Check if already exists
	if err := db.Where("port = ? AND protocol = ?", port, proto).First(&rule).Error; err == nil {
		rule.Enable = true
		db.Save(&rule)
		return s.ApplyRule(&rule)
	}

	rule = model.FirewallRule{
		Action:   "allow",
		Port:     port,
		Protocol: proto,
		Ip:       "0.0.0.0/0",
		Remark:   remark,
		Enable:   true,
	}
	if err := db.Create(&rule).Error; err != nil {
		return err
	}
	return s.ApplyRule(&rule)
}

func (s *FirewallService) RemovePort(port int, proto string) error {
	db := database.GetDB()
	var rule model.FirewallRule
	if err := db.Where("port = ? AND protocol = ?", port, proto).First(&rule).Error; err == nil {
		s.RemoveRule(&rule)
		return db.Delete(&rule).Error
	}
	return nil
}
// ScanSystemRules parses 'iptables -S INPUT' and syncs visible rules to our DB.
func (s *FirewallService) ScanSystemRules() error {
	cmd := exec.Command("iptables", "-S", "INPUT")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	db := database.GetDB()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "-A INPUT") {
			continue
		}

		// Example: -A INPUT -p tcp -m tcp --dport 80 -j ACCEPT
		var proto string
		var port int
		var action string

		// Protocol
		if strings.Contains(line, "-p tcp") {
			proto = "tcp"
		} else if strings.Contains(line, "-p udp") {
			proto = "udp"
		}

		// Port
		if strings.Contains(line, "--dport") {
			parts := strings.Split(line, " ")
			for i, p := range parts {
				if p == "--dport" && i+1 < len(parts) {
					port, _ = strconv.Atoi(parts[i+1])
				}
			}
		}

		// Action
		if strings.Contains(line, "-j ACCEPT") {
			action = "allow"
		} else if strings.Contains(line, "-j REJECT") || strings.Contains(line, "-j DROP") {
			action = "deny"
		}

		if port > 0 && proto != "" && action != "" {
			// Check if we already have it
			var existing model.FirewallRule
			if err := db.Where("port = ? AND protocol = ?", port, proto).First(&existing).Error; err != nil {
				// Not found, add new one as "Imported"
				newRule := model.FirewallRule{
					Action:   action,
					Port:     port,
					Protocol: proto,
					Ip:       "0.0.0.0/0",
					Remark:   fmt.Sprintf("Imported: %s/%d", proto, port),
					Enable:   true,
				}
				db.Create(&newRule)
			}
		}
	}
	return nil
}

// ApplyRule translates a database rule into an iptables command.
func (s *FirewallService) ApplyRule(rule *model.FirewallRule) error {
	action := "-A" // Append
	if rule.Action == "deny" {
		// For simplicity, we use REJECT for deny
		return s.runIptables(action, rule.Protocol, rule.Port, rule.Ip, "REJECT")
	}
	return s.runIptables(action, rule.Protocol, rule.Port, rule.Ip, "ACCEPT")
}

// RemoveRule removes a rule from system iptables.
func (s *FirewallService) RemoveRule(rule *model.FirewallRule) error {
	action := "-D" // Delete
	target := "ACCEPT"
	if rule.Action == "deny" {
		target = "REJECT"
	}
	return s.runIptables(action, rule.Protocol, rule.Port, rule.Ip, target)
}

func (s *FirewallService) runIptables(action, proto string, port int, ip string, target string) error {
	protos := []string{proto}
	if proto == "both" || proto == "" {
		protos = []string{"tcp", "udp"}
	}

	for _, p := range protos {
		args := []string{action, "INPUT"}
		
		if p != "" {
			args = append(args, "-p", p)
		}
		
		if port > 0 {
			if p == "" {
				// Cannot use --dport without protocol, but we just set it above
			}
			args = append(args, "--dport", strconv.Itoa(port))
		}
		
		if ip != "" && ip != "0.0.0.0/0" {
			args = append(args, "-s", ip)
		}
		
		args = append(args, "-j", target)
		
		cmd := exec.Command("iptables", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// If rule already exists/doesn't exist, we might get an error but we often want to ignore it for idempotency
			outStr := string(output)
			if strings.Contains(outStr, "already exists") || strings.Contains(outStr, "Bad rule") || strings.Contains(outStr, "No chain/target/match by that name") {
				continue
			}
			return fmt.Errorf("iptables error (%s): %v, output: %s", p, err, outStr)
		}
	}
	return nil
}
