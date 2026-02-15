#!/usr/bin/env bash
set -euo pipefail

# setup-icons.sh [source-png]
# If no argument is provided, defaults to app/build/appicon.png

SRC=${1:-app/build/appicon.png}
BUILD_DIR="app/build"
WIN_DIR="$BUILD_DIR/windows"
MAC_DIR="$BUILD_DIR/darwin"
LINUX_DIR="$BUILD_DIR/linux"

echo "🔧 Configurando iconos para todas las plataformas usando: $SRC"

if [ ! -f "$SRC" ]; then
    echo "❌ Error: archivo fuente no encontrado: $SRC"
    exit 1
fi

mkdir -p "$WIN_DIR" "$MAC_DIR" "$LINUX_DIR"

echo "✅ Archivo fuente encontrado: $SRC"

# Helper: check for ImageMagick
has_magick() {
    command -v magick >/dev/null 2>&1
}

echo "🖼️  Generando icono Windows (ICO) con múltiples tamaños..."
if has_magick; then
    # Generate ICO containing multiple sizes without relying on shell parentheses
    TMPICONDIR=$(mktemp -d)
    magick "$SRC" -resize 16x16  "$TMPICONDIR/icon_16.png"
    magick "$SRC" -resize 32x32  "$TMPICONDIR/icon_32.png"
    magick "$SRC" -resize 48x48  "$TMPICONDIR/icon_48.png"
    magick "$SRC" -resize 64x64  "$TMPICONDIR/icon_64.png"
    magick "$SRC" -resize 128x128 "$TMPICONDIR/icon_128.png"
    magick "$SRC" -resize 256x256 "$TMPICONDIR/icon_256.png"
    magick "$TMPICONDIR/icon_16.png" "$TMPICONDIR/icon_32.png" \
        "$TMPICONDIR/icon_48.png" "$TMPICONDIR/icon_64.png" \
        "$TMPICONDIR/icon_128.png" "$TMPICONDIR/icon_256.png" \
        "$WIN_DIR/icon.ico"
    rm -rf "$TMPICONDIR"
    echo "✅ $WIN_DIR/icon.ico created with ImageMagick"
else
    # Fallback: try go tool if available, otherwise use single-size PNG copy
    if command -v go >/dev/null 2>&1 && [ -f "app/tools/convert-icon.go" ]; then
        echo "ℹ️  ImageMagick not found — using Go converter (app/tools/convert-icon.go)"
        # convert-icon.go expects input at ../build/appicon.png by default; copy temp then run
        TMPDIR=$(mktemp -d)
        cp "$SRC" "$TMPDIR/appicon.png"
        (cd app/tools && GO111MODULE=on go run convert-icon.go "$TMPDIR/appicon.png" "$WIN_DIR/icon.ico") || true
        rm -rf "$TMPDIR"
        if [ -f "$WIN_DIR/icon.ico" ]; then
            echo "✅ $WIN_DIR/icon.ico created with Go converter"
        else
            echo "⚠️  No se pudo generar ICO automáticamente, usando PNG como fallback"
            cp "$SRC" "$WIN_DIR/icon.ico"
        fi
    else
        echo "⚠️  ImageMagick ni Go converter disponibles — copiando PNG a ICO como fallback"
        cp "$SRC" "$WIN_DIR/icon.ico"
    fi
fi

echo "🍎 Generando icono macOS (ICNS si iconutil disponible, fallback PNG)..."
# Create Icon.iconset structure if iconutil available
if command -v iconutil >/dev/null 2>&1 && has_magick; then
    ICONSET_DIR=$(mktemp -d)/Icon.iconset
    mkdir -p "$ICONSET_DIR"
    # sizes required by iconutil
    magick "$SRC" -resize 16x16  "$ICONSET_DIR/icon_16x16.png"
    magick "$SRC" -resize 32x32  "$ICONSET_DIR/icon_16x16@2x.png"
    magick "$SRC" -resize 32x32  "$ICONSET_DIR/icon_32x32.png"
    magick "$SRC" -resize 64x64  "$ICONSET_DIR/icon_32x32@2x.png"
    magick "$SRC" -resize 128x128 "$ICONSET_DIR/icon_128x128.png"
    magick "$SRC" -resize 256x256 "$ICONSET_DIR/icon_128x128@2x.png"
    magick "$SRC" -resize 256x256 "$ICONSET_DIR/icon_256x256.png"
    magick "$SRC" -resize 512x512 "$ICONSET_DIR/icon_256x256@2x.png"
    magick "$SRC" -resize 512x512 "$ICONSET_DIR/icon_512x512.png"
    magick "$SRC" -resize 1024x1024 "$ICONSET_DIR/icon_512x512@2x.png"
    # build icns
    iconutil -c icns "$ICONSET_DIR" -o "$MAC_DIR/icon.icns"
    rm -rf "$(dirname "$ICONSET_DIR")"
    echo "✅ $MAC_DIR/icon.icns created"
else
    # fallback: copy PNG
    cp "$SRC" "$MAC_DIR/icon.png"
    echo "⚠️  iconutil or ImageMagick not available — copied PNG to $MAC_DIR/icon.png"
fi

echo "🐧 Generando iconos Linux (varias resoluciones)..."
SIZES=(16 24 32 48 64 128 256 512)
for s in "${SIZES[@]}"; do
    magick "$SRC" -resize ${s}x${s} "$LINUX_DIR/icon-${s}x${s}.png" 2>/dev/null || cp "$SRC" "$LINUX_DIR/icon-${s}x${s}.png"
done
# Also provide a canonical 256x256 icon.png for packaging spec
if [ -f "$LINUX_DIR/icon-256x256.png" ]; then
    cp "$LINUX_DIR/icon-256x256.png" "$LINUX_DIR/icon.png"
fi

echo "🎉 ¡Configuración de iconos completada!"
echo ""
echo "📁 Archivos generados:" 
echo "   - Windows: $WIN_DIR/icon.ico"
if [ -f "$MAC_DIR/icon.icns" ]; then
    echo "   - macOS:   $MAC_DIR/icon.icns"
else
    echo "   - macOS:   $MAC_DIR/icon.png"
fi
echo "   - Linux:   $LINUX_DIR/icon.png + icon-<size>x<size>.png"
echo ""
echo "💡 Reemplaza en 'app/build' tu icono base y ejecuta 'wails build' para que los instaladores lo incluyan."
