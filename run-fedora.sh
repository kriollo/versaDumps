#!/usr/bin/env bash
# Script para ejecutar VersaDumps en Fedora
# Configura PKG_CONFIG_PATH para resolver webkit2gtk-4.0 -> webkit2gtk-4.1

set -e

# Asegurar que el symlink existe
PKGCONFIG_DIR="$HOME/.local/lib/pkgconfig"
mkdir -p "$PKGCONFIG_DIR"

if [ ! -L "$PKGCONFIG_DIR/webkit2gtk-4.0.pc" ]; then
    echo "Creando symlink webkit2gtk-4.0 -> webkit2gtk-4.1..."
    ln -sf /usr/lib64/pkgconfig/webkit2gtk-4.1.pc "$PKGCONFIG_DIR/webkit2gtk-4.0.pc"
fi

# Configurar PKG_CONFIG_PATH
export PKG_CONFIG_PATH="$PKGCONFIG_DIR:$PKG_CONFIG_PATH"

# Verificar que wails está instalado
if [ ! -f "$HOME/go/bin/wails" ]; then
    echo "Error: Wails CLI no está instalado"
    echo "Instálalo con: go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0"
    exit 1
fi

# Navegar al directorio app y ejecutar
cd "$(dirname "$0")/app"

echo "Iniciando VersaDumps en modo desarrollo..."
echo "Frontend: http://localhost:5173/"
echo "DevServer: http://localhost:34115"
echo ""

"$HOME/go/bin/wails" dev
