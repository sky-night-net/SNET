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
	args := []string{"INPUT", action}
	
	if proto != "both" && proto != "" {
		args = append(args, "-p", proto)
	}
	
	if port > 0 {
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
		if strings.Contains(string(output), "already exists") || strings.Contains(string(output), "Bad rule") {
			return nil
		}
		return fmt.Errorf("iptables error: %v, output: %s", err, string(output))
	}
	return nil
}
