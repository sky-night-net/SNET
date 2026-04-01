package vpn

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/util/sys"
	"github.com/sky-night-net/snet/vpn/adapters"
)

type ProcessManager struct {
	mu       sync.Mutex
	fw       *sys.FirewallManager
	ctx      context.Context
	cancel   context.CancelFunc
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
		ticker := time.NewTicker(5 * time.Second)
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
	var inbounds []model.Inbound
	if err := db.Find(&inbounds).Error; err != nil {
		return
	}

	for _, ib := range inbounds {
		adapter, err := adapters.GetAdapter(ib.Protocol)
		if err != nil {
			continue // Skip Xray protocols
		}

		// Pre-filter valid clients (not expired, not over-limit)
		var validClients []model.Client
		now := time.Now().Unix()
		for _, client := range ib.Clients {
			if !client.Enable {
				continue
			}
			// Check Expiry
			if client.ExpiryTime > 0 && client.ExpiryTime < now {
				logger.Infof("Client %s expired, skipping", client.Email)
				continue
			}
			// Check Traffic Limit
			if client.Total > 0 && (client.Up+client.Down) >= client.Total {
				logger.Infof("Client %s reached traffic limit, skipping", client.Email)
				continue
			}
			validClients = append(validClients, client)
		}
		
		// Update ib with only valid clients for adapter.Start
		ib.Clients = validClients

		isRunning := adapter.IsRunning(&ib)

		if ib.Enable && !isRunning {
			logger.Infof("Reconciler: starting missing VPN inbound %d (%s)", ib.Id, ib.Protocol)
			if err := adapter.Start(&ib); err == nil {
				m.fw.SetupNAT(m.ifaceName(&ib))
			}
		} else if !ib.Enable && isRunning {
			logger.Infof("Reconciler: stopping unwanted VPN inbound %d (%s)", ib.Id, ib.Protocol)
			if err := adapter.Stop(&ib); err == nil {
				m.fw.CleanupNAT(m.ifaceName(&ib))
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
	case model.AmneziaWGv1, model.AmneziaWGv2:
		return fmt.Sprintf("awg%d", ib.Id)
	case model.OpenVPNXOR:
		return fmt.Sprintf("tun%d", ib.Id)
	}
	return ""
}
