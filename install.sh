#!/bin/bash

# SNET 3.0: 🌌 Ultimate Native Installer
# Supports: Ubuntu 22.04 / 24.04+ (amd64, arm64)

set -e

# --- Цвета ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# --- Константы ---
SNET_DIR="/usr/local/snet"
CONF_DIR="/etc/snet"
SNET_BIN="/usr/local/snet/snet"
CLI_BIN="/usr/bin/snet"

# --- Функции ---
log() { echo -e "${BLUE}[SNET]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}[ERROR]${NC} Этот скрипт должен быть запущен от имени root (используйте sudo)"
        exit 1
    fi
}

optimize_system() {
    log "Оптимизация параметров ядра (BBR, limits)..."
    if ! grep -q "net.core.default_qdisc=fq" /etc/sysctl.conf; then
        echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf
        echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf
        sysctl -p > /dev/null || true
    fi
    cat <<EOF > /etc/security/limits.d/snet.conf
* soft nofile 1048576
* hard nofile 1048576
root soft nofile 1048576
root hard nofile 1048576
EOF
    sysctl -w net.ipv4.ip_forward=1 > /dev/null || true
    echo "net.ipv4.ip_forward=1" > /etc/sysctl.d/99-snet.conf
    success "Система оптимизирована"
}

setup_cli() {
    log "Создание утилиты управления 'snet'..."
    cat <<EOF > $CLI_BIN
#!/bin/bash
case "\$1" in
    start) systemctl start snet ;;
    stop) systemctl stop snet ;;
    restart) systemctl restart snet ;;
    status) systemctl status snet ;;
    logs) journalctl -u snet -f -o cat ;;
    *) echo "Использование: snet {start|stop|restart|status|logs}" ;;
esac
EOF
    chmod +x $CLI_BIN
}

install() {
    check_root
    echo -e "${PURPLE}=========================================${NC}"
    echo -e "${PURPLE}     SNET 3.0: ULTIMATE INSTALLER        ${NC}"
    echo -e "${PURPLE}=========================================${NC}"

    log "Настройка конфигурации..."
    read -p "Введите порт для панели [8080]: " PANEL_PORT
    PANEL_PORT=${PANEL_PORT:-8080}
    
    AUTO_IP=$(curl -s --max-time 5 https://api.ipify.org || echo "0.0.0.0")
    read -p "Введите внешний IP сервера [$AUTO_IP]: " EXTERNAL_IP
    EXTERNAL_IP=${EXTERNAL_IP:-$AUTO_IP}

    log "Установка зависимостей..."
    apt-get update -q && apt-get install -y -q wireguard-tools openvpn easy-rsa iptables iproute2 curl jq wget ca-certificates

    mkdir -p $SNET_DIR $CONF_DIR /etc/amnezia/amneziawg

    log "Загрузка бинарных файлов (SNET & Xray)..."
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)  TARGET="linux-amd64"; XRAY_ARCH="64" ;;
        aarch64) TARGET="linux-arm64"; XRAY_ARCH="arm64-v8a" ;;
        *)       error "Неподдерживаемая архитектура: $ARCH" ;;
    esac

    # 1. Загрузка SNET Panel
    REPO="sky-night-net/SNET"
    LATEST_URL=$(curl -s https://api.github.com/repos/$REPO/releases/latest | jq -r ".assets[] | select(.name | contains(\"$TARGET\")) | .browser_download_url" | head -n 1)
    wget -qO snet_dist.tar.gz "$LATEST_URL"
    tar -xzf snet_dist.tar.gz
    mv snet-$TARGET $SNET_BIN
    
    # Check if openvpn-xor is in the distribution, if not, we use the one we might have uploaded
    if [ -f "openvpn-xor" ]; then
        mv openvpn-xor $SNET_DIR/bin/openvpn-xor
        chmod +x $SNET_DIR/bin/openvpn-xor
    fi
    
    chmod +x $SNET_BIN
    rm snet_dist.tar.gz

    # 2. Загрузка Xray-core
    log "Установка Xray-core..."
    mkdir -p $SNET_DIR/bin
    XRAY_LATEST=$(curl -s https://api.github.com/repos/XTLS/Xray-core/releases/latest | jq -r .tag_name)
    wget -qO xray.zip "https://github.com/XTLS/Xray-core/releases/download/$XRAY_LATEST/Xray-linux-$XRAY_ARCH.zip"
    apt-get install -y unzip
    unzip -o xray.zip -d $SNET_DIR/bin/
    rm xray.zip
    chmod +x $SNET_DIR/bin/xray
    success "Xray-core установлен в $SNET_DIR/bin/xray"

    # 3. Подготовка AmneziaWG
    log "Установка AmneziaWG..."
    # We rely on system wireguard-tools and our amneziawg logic in the binary
    
    optimize_system
    
    cat <<EOF > $CONF_DIR/snet.conf
PORT=$PANEL_PORT
EXTERNAL_IP=$EXTERNAL_IP
DB_PATH=$CONF_DIR/snet.db
XRAY_PATH=$SNET_DIR/bin/xray
OPENVPN_XOR_PATH=$SNET_DIR/bin/openvpn-xor
EOF

    log "Настройка systemd..."
    cat <<EOF > /etc/systemd/system/snet.service
[Unit]
Description=SNET 3.0 Native VPN Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$SNET_DIR
ExecStart=$SNET_BIN
Restart=always
RestartSec=5
EnvironmentFile=$CONF_DIR/snet.conf

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable snet
    systemctl restart snet
    setup_cli

    echo -e "${PURPLE}=========================================${NC}"
    echo -e "${GREEN}SNET 3.0 успешно установлена!${NC}"
    echo -e "Панель:   ${CYAN}http://$EXTERNAL_IP:$PANEL_PORT${NC}"
    echo -e "${PURPLE}=========================================${NC}"
}

uninstall() {
    check_root
    log "Удаление SNET..."
    systemctl stop snet || true
    systemctl disable snet || true
    rm -f /etc/systemd/system/snet.service
    rm -rf $SNET_DIR $CONF_DIR $CLI_BIN
    success "SNET 3.0 удалена."
}

# --- Главный вход ---
case "$1" in
    uninstall) uninstall ;;
    *) install ;;
esac
