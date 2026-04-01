# 🌌 SNET v2 — Next-Gen VPN Control Panel

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Enabled-blue.svg)](https://www.docker.com/)

**SNET v2** is a professional, high-performance VPN control panel written in Go. It builds upon the battle-tested architecture of 3x-ui, extending it with unique support for specialized stealth protocols and a robust, self-healing core.

---

## ⚡ Installation on Clean Ubuntu (22.04 / 24.04)

### One-Click Installation (Recommended)

Run the following command on your server to install all dependencies and start the panel:

```bash
curl -fsSL https://raw.githubusercontent.com/sky-night-net/SNET/main/install.sh | sudo bash
```

### Manual Installation

If you prefer to set up the repository manually, follow these steps:

1. **Install Base Tools**:
   ```bash
   sudo apt update && sudo apt install -y curl git docker.io docker-compose
   ```

2. **Clone and Run**:
   ```bash
   git clone https://github.com/sky-night-net/SNET.git
   cd SNET
   sudo docker-compose up -d --build
   ```

---

## 🚀 Key Features

- **Multi-Protocol Excellence**: 
  - **Xray-Core**: VLESS, VMess, Trojan, Shadowsocks, Shadowsocks-2022, WireGuard.
  - **Stealth VPNs**: Native AmneziaWG (v1/v2) and OpenVPN+XOR (in isolated Docker containers).
- **Self-Healing Reconciler**: A background state-machine continuously ensures all services and firewall rules match the database state.
- **Enterprise-Grade Traffic Management**: Automatic disconnection of clients on data limit or expiration.
- **Integrated SSL Manager**: One-click certificate issuance via Let's Encrypt (acme.sh).
- **Zero-Dependency Binary**: All frontend assets and translations are embedded into a single Go executable.
- **Premium Interface**: A stunning "Glassmorphism" Dark Mode dashboard based on Vue 3 and Ant Design.

---

## 🌍 Supported Languages

SNET v2 currently supports the most essential languages for its core user base:
- 🇺🇸 **English** (en)
- 🇷🇺 **Russian** (ru)

*Adding new languages is as simple as adding a single JSON file to `web/translation/` and recompiling the binary.*

---

## 🛠 Architecture

- **Backend**: Go 1.24, Gin (Web), GORM (Database), Docker SDK (VPN Management).
- **Database**: SQLite (Relational structure for high-speed client lookups).
- **Networking**: Automated `iptables` management with NAT/Forwarding rules.
- **Process Control**: Isolated Docker environments for heterogeneous VPN protocols to prevent kernel conflicts.

---

## 📦 Deployment Options

### Manual Installation
1. Ensure Docker is installed.
2. Download the latest binary from Releases.
3. Run `./snet-v2 run`.

### Docker Compose
```yaml
version: '3'
services:
  snet:
    image: snet-v2:latest
    network_mode: host
    privileged: true
    volumes:
      - /etc/snet:/etc/snet
      - /var/run/docker.sock:/var/run/docker.sock
    restart: always
```

---
> [!TIP]
> **Default Credentials**: `admin / admin`
> Change them immediately after the first login in the System Settings tab.
