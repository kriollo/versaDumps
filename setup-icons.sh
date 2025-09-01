#!/bin/bash
# Script para configurar iconos multiplataforma

echo "🔧 Configurando iconos para todas las plataformas..."

# Directorio base
BUILD_DIR="app/build"
APPICON_PNG="$BUILD_DIR/appicon.png"

# Verificar que existe el archivo appicon.png
if [ ! -f "$APPICON_PNG" ]; then
    echo "❌ Error: No se encuentra $APPICON_PNG"
    exit 1
fi

echo "✅ Archivo fuente encontrado: $APPICON_PNG"

# Para Windows - convertir PNG a ICO
echo "🖼️  Configurando icono para Windows..."
if command -v magick &> /dev/null; then
    # Si ImageMagick está disponible
    magick "$APPICON_PNG" -resize 256x256 "$BUILD_DIR/windows/icon.ico"
    echo "✅ Icono Windows generado con ImageMagick"
elif command -v convert &> /dev/null; then
    # Si convert está disponible (versión antigua de ImageMagick)
    convert "$APPICON_PNG" -resize 256x256 "$BUILD_DIR/windows/icon.ico"
    echo "✅ Icono Windows generado con convert"
else
    # Usar PowerShell en Windows o fallback en otros sistemas
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        powershell -Command "
            Add-Type -AssemblyName System.Drawing
            \$png = [System.Drawing.Image]::FromFile((Get-Item '$APPICON_PNG').FullName)
            \$bitmap = New-Object System.Drawing.Bitmap(\$png)
            \$iconHandle = \$bitmap.GetHicon()
            \$icon = [System.Drawing.Icon]::FromHandle(\$iconHandle)
            \$iconStream = New-Object System.IO.FileStream('$BUILD_DIR/windows/icon.ico', [System.IO.FileMode]::Create)
            \$icon.Save(\$iconStream)
            \$iconStream.Close()
            \$icon.Dispose()
            \$bitmap.Dispose()
            \$png.Dispose()
        "
        echo "✅ Icono Windows generado con PowerShell"
    else
        echo "⚠️  No se pudo generar icono ICO, usando PNG como fallback"
        cp "$APPICON_PNG" "$BUILD_DIR/windows/icon.ico"
    fi
fi

# Para macOS - copiar PNG
echo "🍎 Configurando icono para macOS..."
cp "$APPICON_PNG" "$BUILD_DIR/darwin/icon.png"
echo "✅ Icono macOS configurado"

# Para Linux - copiar PNG (algunos sistemas usan iconos PNG)
echo "🐧 Configurando icono para Linux..."
if [ ! -d "$BUILD_DIR/linux" ]; then
    mkdir -p "$BUILD_DIR/linux"
fi
cp "$APPICON_PNG" "$BUILD_DIR/linux/icon.png"
echo "✅ Icono Linux configurado"

echo "🎉 ¡Configuración de iconos completada!"
echo ""
echo "📁 Archivos generados:"
echo "   - Windows: $BUILD_DIR/windows/icon.ico"
echo "   - macOS:   $BUILD_DIR/darwin/icon.png"  
echo "   - Linux:   $BUILD_DIR/linux/icon.png"
echo ""
echo "💡 Ahora puedes compilar tu aplicación con 'wails build' y tendrá los iconos configurados."
