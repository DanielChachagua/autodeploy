#!/usr/bin/env bash

# Evitar continuar si ocurre un error
set -e

# Colores para la consola
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # Sin color

# Configuración del repositorio (Cambia esto por tu repo real cuando lo publiques)
GITHUB_REPO="DanielChachagua/autodeploy"
BINARY_NAME="autodeploy"
DEST_DIR="/usr/bin"

echo -e "${BLUE}===============================================${NC}"
echo -e "${BLUE}     Instalador de AutoDeploy CLI (Linux)      ${NC}"
echo -e "${BLUE}===============================================${NC}"

# 1. Validar Sistema Operativo
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$OS" != "linux" ]; then
    echo -e "${RED}Error: Este instalador solo es compatible con Linux.${NC}"
    exit 1
fi

# 2. Detectar Arquitectura
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        BIN_ARCH="amd64"
        ;;
    aarch64|arm64)
        BIN_ARCH="arm64"
        ;;
    i386|i686)
        BIN_ARCH="386"
        ;;
    *)
        echo -e "${RED}Error: Arquitectura de CPU no soportada (${ARCH}).${NC}"
        exit 1
        ;;
esac

echo -e "Sistema: ${GREEN}Linux${NC} (${ARCH})"

# 3. Obtener la última versión desde GitHub Releases
echo -e "${YELLOW}Buscando la última versión en GitHub...${NC}"

# Obtener tag de la última versión
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)

if [ -z "$LATEST_RELEASE" ]; then
    echo -e "${YELLOW}Advertencia: No se pudo determinar la última versión desde GitHub (¿repositorio privado o sin releases?).${NC}"
    echo -e "Usando versión por defecto: ${BLUE}v1.0.0${NC}"
    LATEST_RELEASE="v1.0.0"
else
    echo -e "Versión encontrada: ${GREEN}${LATEST_RELEASE}${NC}"
fi

# Construir URL de descarga del binario precompilado
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_RELEASE}/${BINARY_NAME}-linux-${BIN_ARCH}"

# Si el repositorio no está configurado, avisamos al usuario
if [[ "$GITHUB_REPO" == "usuario/autodeploy" ]]; then
    echo -e "\n${YELLOW}[!] NOTA: El script está usando el repositorio de ejemplo 'usuario/autodeploy'.${NC}"
    echo -e "Para usarlo de forma real, asegúrate de reemplazar GITHUB_REPO con tu repositorio en este script."
    echo -e "Presiona Ctrl+C para cancelar, o Enter para continuar intentando descargar...${NC}"
    read -r
fi

# 4. Descargar el binario
TEMP_FILE="/tmp/${BINARY_NAME}"
echo -e "${YELLOW}Descargando binario desde:${NC}"
echo -e "   ${BLUE}${DOWNLOAD_URL}${NC}"

# Descargar usando curl o wget
if command -v curl &> /dev/null; then
    curl -L -f -o "$TEMP_FILE" "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget -q -O "$TEMP_FILE" "$DOWNLOAD_URL"
else
    echo -e "${RED}Error: Se requiere 'curl' o 'wget' para descargar el binario.${NC}"
    exit 1
fi

echo -e "${GREEN}Descarga completada con éxito.${NC}"

# 5. Instalar en /usr/bin
echo -e "${YELLOW}Instalando el ejecutable en ${DEST_DIR}/${BINARY_NAME}...${NC}"

# Verificar si se necesitan privilegios de administrador
if [ -w "$DEST_DIR" ]; then
    mv "$TEMP_FILE" "${DEST_DIR}/${BINARY_NAME}"
    chmod +x "${DEST_DIR}/${BINARY_NAME}"
else
    echo -e "${YELLOW}Se requieren permisos de administrador (sudo) para escribir en ${DEST_DIR}.${NC}"
    sudo mv "$TEMP_FILE" "${DEST_DIR}/${BINARY_NAME}"
    sudo chmod +x "${DEST_DIR}/${BINARY_NAME}"
fi

# 6. Validar instalación
if command -v "$BINARY_NAME" &> /dev/null; then
    echo -e "\n${GREEN}===============================================${NC}"
    echo -e "${GREEN} ¡AutoDeploy CLI se ha instalado correctamente! ${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo -e "Puedes ejecutar la herramienta desde cualquier parte usando:"
    echo -e "  ${BLUE}autodeploy --help${NC}"
    echo -e "\nRecuerda ejecutar con privilegios de root para despliegues reales:"
    echo -e "  ${BLUE}sudo autodeploy run${NC}"
else
    echo -e "\n${RED}Error: El binario fue copiado pero no se pudo encontrar en el PATH.${NC}"
    echo -e "Por favor, asegúrate de que '${DEST_DIR}' está en tu variable de entorno PATH.${NC}"
    exit 1
fi
