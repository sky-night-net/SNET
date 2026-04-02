package vpn

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/util/sys"
	"github.com/sky-night-net/snet/vpn/adapters"
	"github.com/sky-night-net/snet/xray"
)

type ProcessManager struct {
	mu     sync.Mutex
	fw     *sys.FirewallManager
	ctx    context.Context
	cancel context.CancelFunc
}

func NewProcessManager() *ProcessManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ProcessManager{
		fw:     sys.NewFirewallManager(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// StartReconciler starts a background loop that ensures DB state matches running state.
func (m *ProcessManager) StartReconciler() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.reconcile()
			}
		}
	}()
}

func (m *ProcessManager) reconcile() {
	m.mu.Lock()
	defer m.mu.Unlock()

	db := database.GetDB()
	var inbounds []*model.Inbound
	// Preload ClientStats to get traffic info
	if err := db.Preload("ClientStats").Find(&inbounds).Error; err != nil {
		return
	}

	for _, ib := range inbounds {
		adapter, err := adapters.GetAdapter(ib.Protocol)
		if err != nil {
			continue // Skip Xray protocols
		}

		// Extract settings
		settings := struct {
			Clients []model.Client `json:"clients"`
		}{}
		json.Unmarshal([]byte(ib.Settings), &settings)

		// Pre-filter valid clients
		var validClients []model.Client
		now := time.Now().Unix() * 1000 // model uses ms
		
		// Build map for quick access to stats
		statsMap := make(map[string]xray.ClientTraffic)
		for _, stat := range ib.ClientStats {
			statsMap[stat.Email] = stat
		}

		for _, client := range settings.Clients {
			if !client.Enable {
				continue
			}
			// Check Expiry
			if client.ExpiryTime > 0 && client.ExpiryTime < now {
				continue
			}
			// Check Traffic Limit
			stat, hasStat := statsMap[client.Email]
			if client.TotalGB > 0 && hasStat {
				totalLimit := client.TotalGB * 1024 * 1024 * 1024
				if (stat.Up + stat.Down) >= totalLimit {
					continue
				}
			}
			validClients = append(validClients, client)
		}

		isRunning := adapter.IsRunning(ib)

		if ib.Enable && !isRunning && len(validClients) > 0 {
			logger.Infof("Reconciler: starting VPN inbound %d (%s)", ib.Id, ib.Protocol)
			if err := adapter.Start(ib); err == nil {
				m.fw.SetupNAT(m.ifaceName(ib))
			}
		} else if (!ib.Enable || len(validClients) == 0) && isRunning {
			logger.Infof("Reconciler: stopping VPN inbound %d (%s)", ib.Id, ib.Protocol)
			if err := adapter.Stop(ib); err == nil {
				m.fw.CleanupNAT(m.ifaceName(ib))
			}
		}
	}
}

func (m *ProcessManager) Stop() {
	m.cancel()
}

// Low-level overrides still available for UI triggers
func (m *ProcessManager) StartInbound(ib *model.Inbound) error {
	adapter, err := adapters.GetAdapter(ib.Protocol)
	if err != nil {
		return err
	}
	if err := adapter.Start(ib); err != nil {
		return err
	}
	return m.fw.SetupNAT(m.ifaceName(ib))
}

func (m *ProcessManager) StopInbound(ib *model.Inbound) error {
	adapter, err := adapters.GetAdapter(ib.Protocol)
	if err != nil {
		return err
	}
	if err := adapter.Stop(ib); err != nil {
		return err
	}
	m.fw.CleanupNAT(m.ifaceName(ib))
	return nil
}

func (m *ProcessManager) RestartInbound(ib *model.Inbound) error {
	m.StopInbound(ib)
	return m.StartInbound(ib)
}

func (m *ProcessManager) GetTraffic(inbound *model.Inbound) (map[string]adapters.Traffic, error) {
	adapter, err := adapters.GetAdapter(inbound.Protocol)
	if err != nil {
		return nil, err
	}
	return adapter.GetTraffic(inbound)
}

func (m *ProcessManager) ifaceName(ib *model.Inbound) string {
	switch ib.Protocol {
	case model.AmneziaWG, model.AmneziaWGv1, model.AmneziaWGv2:
		return fmt.Sprintf("awg%d", ib.Id)
	case model.OpenVPNXOR:
		return fmt.Sprintf("tun%d", ib.Id)
	}
	return ""
}
