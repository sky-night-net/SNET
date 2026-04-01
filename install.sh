#!/bin/bash

# SNET v2 Production Installer
# Supports Ubuntu 22.04+ / Debian 11+

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🌌 SNET v2 Installer${NC}"

# 1. Check Root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi

# 2. Install Dependencies
echo -e "${GREEN}1/4 Installing system dependencies...${NC}"
apt-get update && apt-get install -y docker.io docker-compose curl git jq

# 3. Clone Repository (if not already in a repo)
if [ ! -d ".git" ]; then
    echo -e "${GREEN}2/4 Cloning SNET v2 repository...${NC}"
    rm -rf /opt/snet-v2
    git clone https://github.com/sky-night-net/snet-v2.git /opt/snet-v2
    cd /opt/snet-v2
else
    echo -e "${GREEN}2/4 Using local repository source...${NC}"
fi

# 4. Pull Protocol Images & Build Panel
echo -e "${GREEN}3/4 Preparing protocol images and building panel...${NC}"
# Pre-pull base images for faster build
docker pull golang:1.24-alpine
docker pull alpine:latest

# Build and start services
docker-compose up -d --build

# 5. Success
echo -e "${BLUE}==================================================${NC}"
echo -e "${GREEN}✅ SNET v2 was successfully installed!${NC}"
echo -e "Access the panel at: ${BLUE}http://$(curl -s ifconfig.me):2053${NC}"
echo -e "Default credentials: ${BLUE}admin / admin${NC}"
echo -e "${BLUE}==================================================${NC}"
echo -e "Logs can be viewed with: ${BLUE}docker-compose logs -f${NC}"
