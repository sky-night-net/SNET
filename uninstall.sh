#!/bin/bash

# SNET v2 Uninstaller
# This script will stop and remove all SNET-related containers, images, and data.

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🌌 SNET v2 Uninstaller${NC}"

# Check Root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi

read -p "Are you sure you want to completely remove SNET? All data and VPN configs will be lost! (y/n): " confirm
if [[ $confirm != "y" ]]; then
    echo "Uninstallation cancelled."
    exit 0
fi

echo -e "${GREEN}1/4 Stopping and removing SNET panel...${NC}"
if [ -f "docker-compose.yml" ]; then
    docker-compose down || true
elif [ -d "/opt/SNET" ]; then
    cd /opt/SNET && docker-compose down || true
fi

echo -e "${GREEN}2/4 Removing protocol containers and local images...${NC}"
# Remove any containers started by the panel (they prefix with snet-local)
docker ps -a --filter name=snet-local -q | xargs -r docker rm -f

# Remove SNET panel image
docker rmi snet-v2:latest 2>/dev/null || true

# Remove local protocol images
docker images -q snet-local/* | xargs -r docker rmi -f

echo -e "${GREEN}3/4 Deleting SNET data and logs...${NC}"
rm -rf /etc/snet
rm -rf /var/log/snet

echo -e "${GREEN}4/4 Optional: removing source directory...${NC}"
read -p "Do you want to remove the source directory /opt/SNET? (y/n): " remove_source
if [[ $remove_source == "y" ]]; then
    rm -rf /opt/SNET
    echo "Source directory removed."
fi

echo -e "${BLUE}==================================================${NC}"
echo -e "${GREEN}✅ SNET v2 has been successfully uninstalled.${NC}"
echo -e "${BLUE}==================================================${NC}"
