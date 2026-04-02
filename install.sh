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
    [[ $EUID -ne 0 ]] && error "Этот скрипт должен быть запущен от имени root"
}

optimize_system() {
    log "Оптимизация параметров ядра (BBR, limits)..."
    
    # Включение BBR
    if ! grep -q "net.core.default_qdisc=fq" /etc/sysctl.conf; then
        echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf
        echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf
        sysctl -p > /dev/null
    fi

    # Лимиты открытых файлов
    cat <<EOF > /etc/security/limits.d/snet.conf
* soft nofile 1048576
* hard nofile 1048576
root soft nofile 1048576
root hard nofile 1048576
EOF
    
    # IPv4 Forwarding
    sysctl -w net.ipv4.ip_forward=1 > /dev/null
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
    version) /usr/local/snet/snet -version ;;
    *) echo "Использование: snet {start|stop|restart|status|logs|version}" ;;
esac
EOF
    chmod +x $CLI_BIN
    success "Утилита 'snet' добавлена (попробуйте в терминале: snet status)"
}

install() {
    check_root
    
    echo -e "${PURPLE}=========================================${NC}"
    echo -e "${PURPLE}     SNET 3.0: ULTIMATE INSTALLER        ${NC}"
    echo -e "${PURPLE}=========================================${NC}"

    # 1. Сбор параметров
    log "Настройка конфигурации..."
    DEF_PORT=8080
    read -p "Введите порт для панели [$DEF_PORT]: " PANEL_PORT
    PANEL_PORT=${PANEL_PORT:-$DEF_PORT}

    # Внешний IP
    AUTO_IP=$(curl -s https://api.ipify.org || echo "0.0.0.0")
    read -p "Введите внешний IP сервера [$AUTO_IP]: " EXTERNAL_IP
    EXTERNAL_IP=${EXTERNAL_IP:-$AUTO_IP}

    # 2. Установка зависимостей
    log "Установка системных пакетов..."
    apt-get update -q
    apt-get install -y -q wireguard-tools openvpn iptables iproute2 curl jq wget ca-certificates
    
    # 3. Подготовка директорий
    mkdir -p $SNET_DIR
    mkdir -p $CONF_DIR
    mkdir -p /etc/amnezia/amneziawg

    # 4. Копирование бинарника (или скачивание)
    if [ -f "./snet" ]; then
        cp ./snet $SNET_BIN
        chmod +x $SNET_BIN
    else
        warn "Бинарник 'snet' не найден локально. Попытка скачивания последней версии..."
        # TODO: Добавить скачивание с GitHub Releases в будущем
        error "Файл snet не найден. Сначала соберите проект: go build -o snet main.go"
    fi

    # 5. Оптимизация
    optimize_system

    # 6. Сохранение конфигурации
    cat <<EOF > $CONF_DIR/snet.conf
PORT=$PANEL_PORT
EXTERNAL_IP=$EXTERNAL_IP
DB_PATH=$CONF_DIR/snet.db
EOF

    # 7. Systemd
    log "Настройка службы snet.service..."
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
    echo -e "Команда:  ${CYAN}snet logs${NC} (просмотр логов)"
    echo -e "БД:       ${CYAN}$CONF_DIR/snet.db${NC}"
    echo -e "${PURPLE}=========================================${NC}"
}

uninstall() {
    check_root
    log "Начинается полное удаление SNET..."
    
    systemctl stop snet || true
    systemctl disable snet || true
    rm -f /etc/systemd/system/snet.service
    systemctl daemon-reload
    
    rm -rf $SNET_DIR
    rm -rf $CONF_DIR
    rm -f $CLI_BIN
    
    success "SNET 3.0 полностью удалена."
}

# --- Main ---
case "$1" in
    uninstall) uninstall ;;
    *) install ;;
esac
