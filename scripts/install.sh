#!/bin/bash

# Скрипт установки systemd сервиса для Linux

SERVICE_NAME="service-boilerplate"
SERVICE_FILE="service.service"
INSTALL_DIR="/opt/${SERVICE_NAME}"
CONFIG_DIR="/etc/${SERVICE_NAME}"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Installing ${SERVICE_NAME}...${NC}"

# Проверяем права root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

# Создаем директории
mkdir -p "${INSTALL_DIR}"
mkdir -p "${CONFIG_DIR}/configs"
mkdir -p "/var/log/${SERVICE_NAME}"

# Копируем бинарник (предполагается, что он уже скомпилирован)
if [ -f "../service-boilerplate" ]; then
    cp "../service-boilerplate" "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/service-boilerplate"
else
    echo -e "${RED}Binary not found. Please build first: go build -o service-boilerplate ./cmd/service-boilerplate${NC}"
    exit 1
fi

# Копируем конфиг
if [ -f "../configs/config.yaml" ]; then
    cp "../configs/config.yaml" "${CONFIG_DIR}/configs/"
else
    echo -e "${RED}Config not found${NC}"
    exit 1
fi

# Устанавливаем systemd unit
if [ -f "${SERVICE_FILE}" ]; then
    cp "${SERVICE_FILE}" "/etc/systemd/system/${SERVICE_NAME}.service"
    systemctl daemon-reload
    echo -e "${GREEN}Systemd unit installed${NC}"
else
    echo -e "${RED}Service file not found: ${SERVICE_FILE}${NC}"
    exit 1
fi

# Устанавливаем права
chown -R root:root "${INSTALL_DIR}"
chown -R root:root "${CONFIG_DIR}"
chmod 755 "${INSTALL_DIR}"
chmod 755 "${CONFIG_DIR}"

echo -e "${GREEN}Installation completed!${NC}"
echo ""
echo "Usage:"
echo "  systemctl start ${SERVICE_NAME}    # Start service"
echo "  systemctl stop ${SERVICE_NAME}     # Stop service"
echo "  systemctl status ${SERVICE_NAME}   # Check status"
echo "  journalctl -u ${SERVICE_NAME} -f   # View logs"
