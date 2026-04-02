# 🌟 SNET 3.2.0: Professional Native VPN Panel

<div align="center">
  <p>
    <a href="README.md"><b>English (EN)</b></a> | 
    <a href="README_RU.md"><b>Russian (RU)</b></a>
  </p>
  
  [![Build and Release SNET](https://github.com/sky-night-net/SNET/actions/workflows/release.yml/badge.svg)](https://github.com/sky-night-net/SNET/actions/workflows/release.yml)
  [![Version](https://img.shields.io/badge/version-3.2.0-blue.svg)](https://github.com/sky-night-net/SNET/releases)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
</div>

---

**SNET 3.2.0** is a high-performance native VPN infrastructure management panel designed for professional security, blazing speed, and a premium modern user experience.

> [!IMPORTANT]
> Version 3.2.0 is a milestone release featuring built-in Firewall management and full Internationalization (i18n) support.

---

## 🔥 Key Features

### 🎨 Design & Experience
- **Premium Interface**: Modern Glassmorphism design with Dark Mode and high responsiveness.
- **🌍 Multi-language (i18n)**: One-click switching between **English** and **Russian**.
- **📊 Real-time Telemetry**: Live CPU, RAM, and Network traffic charts (updated every 2 sec).

### 🚀 Protocols & Security
- **Native Support**: AmneziaWG (v1/v2), OpenVPN XOR, VLESS, VMess, Trojan, Shadowsocks.
- **🛡️ Built-in Firewall**: Direct `iptables` management through the web interface.
- **Reality & TLS**: Advanced traffic cloaking methods with automatic key generation.
- **Xray-core**: Automated delivery and management of the latest Xray binary.

### ⚙️ Administration
- **CLI Utility**: Powerful `snet` tool for service management via terminal.
- **Autonomous**: Everything is bundled into a single binary. No Docker or heavy dependencies.
- **Sync**: Instant application of configurations with zero downtime.

---

## 🚀 Quick Start (Installation)

Supported OS: **Ubuntu 22.04 / 24.04 (amd64, arm64)**.

1. **Download Installer**:
   ```bash
   wget https://raw.githubusercontent.com/sky-night-net/SNET/main/install.sh
   chmod +x install.sh
   ```

2. **Run Installation**:
   ```bash
   sudo ./install.sh
   ```

3. **Access the Panel**:
   - URL: `http://your-server-ip:8080`
   - User: `admin`
   - Pass: `admin` (change after first login!)

---

## 🛡️ Web Firewall

Manage your server security directly from the browser. All rules are persisted in the database and applied to `iptables` on startup.

> [!CAUTION]
> Before changing the panel port in settings, ensure you first allow the new port in the Firewall section to maintain access.

---

## 🛠️ CLI Commands (`snet`)

After installation, use the `snet` command for rapid management:

| Command | Description |
| :--- | :--- |
| `snet start` | Start the panel |
| `snet stop` | Stop the panel |
| `snet status` | Check service status |
| `snet logs` | View real-time logs |

---

## 🤝 Contribution
Contributions are welcome! Feel free to open issues or submit pull requests to help improve the project.

---

> Built with Go, React, and a commitment to a free internet. 🌌
