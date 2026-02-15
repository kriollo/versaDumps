# Script para configurar iconos multiplataforma en Windows
param(
    [string]$AppiconPath = "app\build\appicon.png"
)

Write-Host "🔧 Configurando iconos para todas las plataformas..." -ForegroundColor Cyan

if (-not (Test-Path $AppiconPath)) {
    Write-Error "❌ Error: No se encuentra $AppiconPath"
    exit 1
}

Write-Host "✅ Archivo fuente encontrado: $AppiconPath" -ForegroundColor Green

$BuildDir = "app\build"
$WindowsDir = "$BuildDir\windows"
$DarwinDir = "$BuildDir\darwin"
$LinuxDir = "$BuildDir\linux"

@($WindowsDir, $DarwinDir, $LinuxDir) | ForEach-Object {
    if (-not (Test-Path $_)) { New-Item -ItemType Directory -Path $_ -Force | Out-Null }
}

function Has-Command([string]$cmd) { return (Get-Command $cmd -ErrorAction SilentlyContinue) -ne $null }

Write-Host "🖼️  Generando icono Windows (ICO) y otros formatos..." -ForegroundColor Yellow

if (Has-Command magick) {
    # Use ImageMagick if available
    magick @( 
        ("$AppiconPath"), ("$AppiconPath"), ("$AppiconPath")) -resize 256x256 "$WindowsDir\icon.ico" 2>$null;
    # Better to call magick with multiple resized inputs; use a simple approach here
    magick ("$AppiconPath" -resize 16x16) ("$AppiconPath" -resize 32x32) ("$AppiconPath" -resize 48x48) ("$AppiconPath" -resize 64x64) ("$AppiconPath" -resize 128x128) ("$AppiconPath" -resize 256x256) "$WindowsDir\icon.ico"
    Write-Host "✅ Icono Windows generado con ImageMagick: $WindowsDir\icon.ico" -ForegroundColor Green
} elseif (Has-Command go -and (Test-Path "app\tools\convert-icon.go")) {
    Write-Host "ℹ️  ImageMagick no disponible: usando convert-icon.go (Go)" -ForegroundColor Yellow
    $tmp = Join-Path (Get-Location) "tmp_icon_src.png"
    Copy-Item $AppiconPath $tmp -Force
    Push-Location app\tools
    try { & go run convert-icon.go $tmp "$WindowsDir\icon.ico" } catch { }
    Pop-Location
    Remove-Item $tmp -ErrorAction SilentlyContinue
    if (Test-Path "$WindowsDir\icon.ico") { Write-Host "✅ Icono Windows generado con Go converter" -ForegroundColor Green } else { Write-Warning "⚠️  No se pudo generar .ico, se copiará PNG como fallback"; Copy-Item $AppiconPath "$WindowsDir\icon.ico" -Force }
} else {
    Write-Warning "⚠️  No hay ImageMagick ni herramienta Go; copiando PNG como fallback"
    Copy-Item $AppiconPath "$WindowsDir\icon.ico" -Force
}

Write-Host "🍎 macOS: creando ICNS si es posible (iconutil), fallback: PNG" -ForegroundColor Yellow
if (Has-Command iconutil -and Has-Command magick) {
    $iconset = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString() + ".iconset")
    New-Item -ItemType Directory -Path $iconset | Out-Null
    magick "$AppiconPath" -resize 16x16 "$iconset\icon_16x16.png"
    magick "$AppiconPath" -resize 32x32 "$iconset\icon_16x16@2x.png"
    magick "$AppiconPath" -resize 32x32 "$iconset\icon_32x32.png"
    magick "$AppiconPath" -resize 64x64 "$iconset\icon_32x32@2x.png"
    magick "$AppiconPath" -resize 128x128 "$iconset\icon_128x128.png"
    magick "$AppiconPath" -resize 256x256 "$iconset\icon_128x128@2x.png"
    magick "$AppiconPath" -resize 256x256 "$iconset\icon_256x256.png"
    magick "$AppiconPath" -resize 512x512 "$iconset\icon_256x256@2x.png"
    magick "$AppiconPath" -resize 512x512 "$iconset\icon_512x512.png"
    magick "$AppiconPath" -resize 1024x1024 "$iconset\icon_512x512@2x.png"
    & iconutil -c icns $iconset -o "$DarwinDir\icon.icns"
    Remove-Item $iconset -Recurse -Force
    Write-Host "✅ ICNS creado: $DarwinDir\icon.icns" -ForegroundColor Green
} else {
    Copy-Item $AppiconPath "$DarwinDir\icon.png" -Force
    Write-Warning "⚠️ iconutil o ImageMagick no disponibles: copiado PNG a $DarwinDir\icon.png" -ForegroundColor Yellow
}

Write-Host "🐧 Linux: generando PNGs en varias resoluciones" -ForegroundColor Yellow
$sizes = @(16,24,32,48,64,128,256,512)
foreach ($s in $sizes) {
    if (Has-Command magick) {
        magick "$AppiconPath" -resize "${s}x${s}" "$LinuxDir\icon-${s}x${s}.png" 2>$null
    } else {
        Copy-Item $AppiconPath "$LinuxDir\icon-${s}x${s}.png" -Force
    }
}
Copy-Item "$LinuxDir\icon-256x256.png" "$LinuxDir\icon.png" -Force -ErrorAction SilentlyContinue

Write-Host ""; Write-Host "🎉 ¡Configuración de iconos completada!" -ForegroundColor Magenta
Write-Host ""; Write-Host "📁 Archivos generados:" -ForegroundColor Cyan
Write-Host "   - Windows: $WindowsDir\icon.ico" -ForegroundColor White
if (Test-Path "$DarwinDir\icon.icns") { Write-Host "   - macOS:   $DarwinDir\icon.icns" -ForegroundColor White } else { Write-Host "   - macOS:   $DarwinDir\icon.png" -ForegroundColor White }
Write-Host "   - Linux:   $LinuxDir\icon.png + icon-<size>x<size>.png" -ForegroundColor White
Write-Host ""; Write-Host "💡 Recompila con 'wails build' para incluir iconos en los instaladores." -ForegroundColor Green
