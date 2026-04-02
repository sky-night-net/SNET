#!/bin/bash

# SNET 3.0: 🌟 Native Installation Script
# Supports: Ubuntu 22.04 / 24.04 (amd64, arm64)

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}     SNET 3.0: NATIVE VPN PANEL          ${NC}"
echo -e "${BLUE}=========================================${NC}"

# 1. Check Root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}Этот скрипт должен быть запущен от имени root${NC}" 
   exit 1
fi

# 2. Install dependencies
echo -e "${GREEN}[1/5] Установка зависимостей...${NC}"
apt-get update -q
apt-get install -y -q wireguard-tools openvpn iptables iproute2 curl jq wget ca-certificates

# 3. Setup Directories
echo -e "${GREEN}[2/5] Настройка директорий...${NC}"
mkdir -p /etc/snet
mkdir -p /usr/local/snet
mkdir -p /etc/amnezia/amneziawg

# 4. Install SNET Binary (Assuming it's in the same directory for this test)
# In production, this would download from GitHub Releases
echo -e "${GREEN}[3/5] Установка бинарного файла...${NC}"
if [ -f "./snet" ]; then
    cp ./snet /usr/local/snet/snet
    chmod +x /usr/local/snet/snet
else
    echo -e "${RED}Ошибка: бинарный файл 'snet' не найден в текущей директории.${NC}"
    exit 1
fi

# 5. Setup Firewall & IPv4 Forwarding
echo -e "${GREEN}[4/5] Настройка сети...${NC}"
sysctl -w net.ipv4.ip_forward=1
echo "net.ipv4.ip_forward=1" > /etc/sysctl.d/99-snet.conf

# 6. Create Systemd Service
echo -e "${GREEN}[5/5] Создание системной службы...${NC}"
cat <<EOF > /etc/systemd/system/snet.service
[Unit]
Description=SNET 3.0 Native VPN Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/usr/local/snet
ExecStart=/usr/local/snet/snet
Restart=always
RestartSec=5
Environment=PORT=8080
Environment=SNET_DB_PATH=/etc/snet/snet.db

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable snet
systemctl start snet

echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Установка SNET 3.0 завершена успешно!${NC}"
echo -e "Панель доступна по адресу: ${BLUE}http://$(curl -s https://api.ipify.org):8080${NC}"
echo -e "Логин/Пароль по умолчанию: ${BLUE}admin / admin${NC}"
echo -e "${BLUE}=========================================${NC}"
