# Script para configurar iconos multiplataforma en Windows
param(
    [string]$AppiconPath = "app\build\appicon.png"
)

Write-Host "🔧 Configurando iconos para todas las plataformas..." -ForegroundColor Cyan

# Verificar que existe el archivo appicon.png
if (-not (Test-Path $AppiconPath)) {
    Write-Error "❌ Error: No se encuentra $AppiconPath"
    exit 1
}

Write-Host "✅ Archivo fuente encontrado: $AppiconPath" -ForegroundColor Green

# Directorios
$BuildDir = "app\build"
$WindowsDir = "$BuildDir\windows"
$DarwinDir = "$BuildDir\darwin" 
$LinuxDir = "$BuildDir\linux"

# Crear directorios si no existen
@($WindowsDir, $DarwinDir, $LinuxDir) | ForEach-Object {
    if (-not (Test-Path $_)) {
        New-Item -ItemType Directory -Path $_ -Force | Out-Null
    }
}

# Para Windows - convertir PNG a ICO usando la herramienta Go
Write-Host "🖼️  Configurando icono para Windows..." -ForegroundColor Yellow

try {
    # Usar la herramienta Go para crear ICO válido
    $originalDir = Get-Location
    Set-Location "app\tools"
    $result = & go run convert-icon.go 2>&1
    Set-Location $originalDir
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Icono Windows generado con herramienta Go: $WindowsDir\icon.ico" -ForegroundColor Green
    } else {
        throw "Error en herramienta Go: $result"
    }
} catch {
    Write-Warning "⚠️  Error generando ICO con Go: $($_.Exception.Message)"
    Write-Host "📋 Intentando con PowerShell como fallback..." -ForegroundColor Yellow
    
    try {
        Add-Type -AssemblyName System.Drawing
        
        $png = [System.Drawing.Image]::FromFile((Get-Item $AppiconPath).FullName)
        $bitmap = New-Object System.Drawing.Bitmap($png)
        $iconHandle = $bitmap.GetHicon()
        $icon = [System.Drawing.Icon]::FromHandle($iconHandle)
        
        $iconPath = Join-Path $WindowsDir "icon.ico"
        $iconStream = New-Object System.IO.FileStream($iconPath, [System.IO.FileMode]::Create)
        $icon.Save($iconStream)
        $iconStream.Close()
        
        # Limpiar recursos
        $icon.Dispose()
        $bitmap.Dispose() 
        $png.Dispose()
        
        Write-Host "✅ Icono Windows generado con PowerShell: $iconPath" -ForegroundColor Green
    } catch {
        Write-Warning "⚠️  Error generando ICO con PowerShell: $($_.Exception.Message)"
        Write-Host "📋 Copiando PNG como fallback..." -ForegroundColor Yellow
        Copy-Item $AppiconPath "$WindowsDir\icon.ico"
    }
}

# Para macOS - copiar PNG
Write-Host "🍎 Configurando icono para macOS..." -ForegroundColor Yellow
$macIconPath = Join-Path $DarwinDir "icon.png"
Copy-Item $AppiconPath $macIconPath
Write-Host "✅ Icono macOS configurado: $macIconPath" -ForegroundColor Green

# Para Linux - copiar PNG
Write-Host "🐧 Configurando icono para Linux..." -ForegroundColor Yellow
$linuxIconPath = Join-Path $LinuxDir "icon.png"
Copy-Item $AppiconPath $linuxIconPath
Write-Host "✅ Icono Linux configurado: $linuxIconPath" -ForegroundColor Green

Write-Host ""
Write-Host "🎉 ¡Configuración de iconos completada!" -ForegroundColor Magenta
Write-Host ""
Write-Host "📁 Archivos generados:" -ForegroundColor Cyan
Write-Host "   - Windows: $WindowsDir\icon.ico" -ForegroundColor White
Write-Host "   - macOS:   $DarwinDir\icon.png" -ForegroundColor White
Write-Host "   - Linux:   $LinuxDir\icon.png" -ForegroundColor White
Write-Host ""
Write-Host "💡 Ahora puedes compilar tu aplicación con 'wails build' y tendrá los iconos configurados." -ForegroundColor Green
