#!/usr/bin/env bash
# Script para ejecutar VersaDumps en Fedora
# Configura PKG_CONFIG_PATH para resolver webkit2gtk-4.0 -> webkit2gtk-4.1

set -e

# Asegurar que el symlink existe
PKGCONFIG_DIR="$HOME/.local/lib/pkgconfig"
mkdir -p "$PKGCONFIG_DIR"

if [ ! -L "$PKGCONFIG_DIR/webkit2gtk-4.0.pc" ]; then
    echo "Creando symlinks para compatibilidad (4.0 -> 4.1)..."
    ln -sf /usr/lib64/pkgconfig/webkit2gtk-4.1.pc "$PKGCONFIG_DIR/webkit2gtk-4.0.pc"
    ln -sf /usr/lib64/pkgconfig/javascriptcoregtk-4.1.pc "$PKGCONFIG_DIR/javascriptcoregtk-4.0.pc"
fi

# Configurar PKG_CONFIG_PATH y PATH de Go
export PKG_CONFIG_PATH="$PKGCONFIG_DIR:$PKG_CONFIG_PATH"
export PATH="$HOME/go/bin:$PATH"

# Verificar que wails está instalado (busca en múltiples ubicaciones)
find_wails() {
    local paths=(
        "$HOME/go/bin/wails"
        "$HOME/.local/bin/wails"
        "$HOME/.asdf/installs/golang/"*/"packages/bin/wails"
        "/usr/local/bin/wails"
        "/usr/bin/wails"
    )
    for p in "${paths[@]}"; do
        # Expandir globs si existen
        for expanded in $p; do
            if [ -x "$expanded" ]; then
                echo "$expanded"
                return 0
            fi
        done
    done
    which wails 2>/dev/null
}

WAILS_BIN=$(find_wails)
if [ -z "$WAILS_BIN" ]; then
    echo "Error: Wails CLI no está instalado"
    echo "Instálalo con: go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0"
    exit 1
fi

MODE="${1:-dev}"
if [ "$MODE" = "build" ]; then
    echo "Construyendo versión final (Linux) con Wails..."
    cd "$(dirname "$0")/app"
    "$WAILS_BIN" build
    echo "Build final completado. Artefactos en: $(pwd)/build/bin"
    exit 0
fi

# Navegar al directorio app y ejecutar
cd "$(dirname "$0")/app"

echo "Iniciando VersaDumps en modo desarrollo..."
echo "Frontend: http://localhost:5173/"
echo "DevServer: http://localhost:34115"
echo ""

"$WAILS_BIN" dev
