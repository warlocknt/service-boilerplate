#!/bin/bash

# Скрипт удаления systemd сервиса для Linux

SERVICE_NAME="service-boilerplate"
INSTALL_DIR="/opt/${SERVICE_NAME}"
CONFIG_DIR="/etc/${SERVICE_NAME}"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Uninstalling ${SERVICE_NAME}...${NC}"

# Проверяем права root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

# Останавливаем сервис
if systemctl is-active --quiet "${SERVICE_NAME}"; then
    echo "Stopping service..."
    systemctl stop "${SERVICE_NAME}"
fi

# Отключаем сервис
if systemctl is-enabled --quiet "${SERVICE_NAME}" 2>/dev/null; then
    echo "Disabling service..."
    systemctl disable "${SERVICE_NAME}"
fi

# Удаляем systemd unit
if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    systemctl daemon-reload
    echo -e "${GREEN}Systemd unit removed${NC}"
fi

# Удаляем директории
if [ -d "${INSTALL_DIR}" ]; then
    rm -rf "${INSTALL_DIR}"
    echo "Removed: ${INSTALL_DIR}"
fi

if [ -d "${CONFIG_DIR}" ]; then
    rm -rf "${CONFIG_DIR}"
    echo "Removed: ${CONFIG_DIR}"
fi

echo -e "${GREEN}Uninstallation completed!${NC}"
