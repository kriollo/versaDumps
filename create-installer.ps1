# Script para crear un instalador personalizado de VersaDumps
param(
    [string]$Version = "1.1.10",
    [string]$OutputPath = "versaDumps-setup-$Version.exe"
)

Write-Host "🚀 Creando instalador personalizado de VersaDumps v$Version" -ForegroundColor Cyan

# Verificar que existe el ejecutable compilado
$AppExe = "app\build\bin\VersaDumps.exe"
$ConfigFile = "app\config.yml"

if (-not (Test-Path $AppExe)) {
    Write-Error "❌ No se encuentra $AppExe. Compila la aplicación primero con 'wails build'"
    exit 1
}

Write-Host "✅ Ejecutable encontrado: $AppExe" -ForegroundColor Green

# Crear directorio temporal para el instalador
$TempDir = "$env:TEMP\VersaDumps-Installer"
if (Test-Path $TempDir) {
    Remove-Item $TempDir -Recurse -Force
}
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

Write-Host "📁 Directorio temporal: $TempDir" -ForegroundColor Yellow

# Copiar archivos necesarios
Copy-Item $AppExe "$TempDir\VersaDumps.exe"
if (Test-Path $ConfigFile) {
    Copy-Item $ConfigFile "$TempDir\config.yml"
    Write-Host "✅ Archivo de configuración copiado" -ForegroundColor Green
}

# Copiar icono
$IconFile = "app\build\windows\icon.ico"
if (Test-Path $IconFile) {
    Copy-Item $IconFile "$TempDir\icon.ico"
    Write-Host "✅ Icono copiado" -ForegroundColor Green
}

# Crear script de instalación
$InstallScript = @"
# VersaDumps Installer Script
param([string]`$InstallPath = "`$env:ProgramFiles\VersaDumps")

Write-Host "🔧 Instalando VersaDumps Visualizer..." -ForegroundColor Cyan
Write-Host "📍 Ruta de instalación: `$InstallPath" -ForegroundColor Yellow

# Crear directorio de instalación
if (-not (Test-Path `$InstallPath)) {
    New-Item -ItemType Directory -Path `$InstallPath -Force | Out-Null
    Write-Host "✅ Directorio de instalación creado" -ForegroundColor Green
}

# Copiar archivos
try {
    Copy-Item "VersaDumps.exe" "`$InstallPath\" -Force
    Write-Host "✅ Ejecutable instalado" -ForegroundColor Green

    if (Test-Path "config.yml") {
        Copy-Item "config.yml" "`$InstallPath\" -Force
        Write-Host "✅ Archivo de configuración instalado" -ForegroundColor Green
    }

    if (Test-Path "icon.ico") {
        Copy-Item "icon.ico" "`$InstallPath\" -Force
        Write-Host "✅ Icono instalado" -ForegroundColor Green
    }

    # Crear acceso directo en el escritorio
    `$WshShell = New-Object -comObject WScript.Shell
    `$Shortcut = `$WshShell.CreateShortcut("`$env:USERPROFILE\Desktop\VersaDumps Visualizer.lnk")
    `$Shortcut.TargetPath = "`$InstallPath\VersaDumps.exe"
    `$Shortcut.WorkingDirectory = `$InstallPath
    `$Shortcut.IconLocation = "`$InstallPath\icon.ico"
    `$Shortcut.Description = "VersaDumps Visualizer - A powerful debugging tool"
    `$Shortcut.Save()
    Write-Host "✅ Acceso directo creado en el escritorio" -ForegroundColor Green

    # Crear entrada en el menú inicio
    `$StartMenuPath = "`$env:ProgramData\Microsoft\Windows\Start Menu\Programs"
    `$StartShortcut = `$WshShell.CreateShortcut("`$StartMenuPath\VersaDumps Visualizer.lnk")
    `$StartShortcut.TargetPath = "`$InstallPath\VersaDumps.exe"
    `$StartShortcut.WorkingDirectory = `$InstallPath
    `$StartShortcut.IconLocation = "`$InstallPath\icon.ico"
    `$StartShortcut.Description = "VersaDumps Visualizer - A powerful debugging tool"
    `$StartShortcut.Save()
    Write-Host "✅ Entrada en menú inicio creada" -ForegroundColor Green

    Write-Host ""
    Write-Host "🎉 ¡Instalación completada exitosamente!" -ForegroundColor Magenta
    Write-Host "📍 VersaDumps instalado en: `$InstallPath" -ForegroundColor White
    Write-Host "🖥️  Acceso directo disponible en el escritorio" -ForegroundColor White
    Write-Host ""

    # Preguntar si ejecutar la aplicación
    `$response = Read-Host "¿Deseas ejecutar VersaDumps ahora? (s/n)"
    if (`$response -eq "s" -or `$response -eq "S") {
        Start-Process "`$InstallPath\VersaDumps.exe"
    }

} catch {
    Write-Error "❌ Error durante la instalación: `$(`$_.Exception.Message)"
    exit 1
}
"@

# Guardar script de instalación
$InstallScript | Out-File "$TempDir\install.ps1" -Encoding UTF8
Write-Host "✅ Script de instalación creado" -ForegroundColor Green

# Crear archivo de información
$InfoContent = @"
VersaDumps Visualizer v$Version
==============================

Instrucciones de instalación:
1. Ejecuta 'install.ps1' como administrador
2. Sigue las instrucciones en pantalla
3. ¡Disfruta de VersaDumps!

Archivos incluidos:
- VersaDumps.exe (aplicación principal)
- config.yml (configuración)
- icon.ico (icono de la aplicación)
- install.ps1 (script de instalación)

Para desinstalar:
- Elimina la carpeta de instalación (normalmente C:\Program Files\VersaDumps)
- Elimina los accesos directos del escritorio y menú inicio

Soporte: https://github.com/kriollo/versaDumps
"@

$InfoContent | Out-File "$TempDir\README.txt" -Encoding UTF8

# Crear archivo batch para facilitar la instalación
$BatchContent = @"
@echo off
echo Iniciando instalacion de VersaDumps Visualizer...
powershell -ExecutionPolicy Bypass -File install.ps1
pause
"@

$BatchContent | Out-File "$TempDir\instalar.bat" -Encoding ASCII

Write-Host "✅ Archivos auxiliares creados" -ForegroundColor Green

# Crear archivo ZIP como instalador
$ZipPath = (Get-Item .).FullName + "\$OutputPath".Replace(".exe", ".zip")
try {
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::CreateFromDirectory($TempDir, $ZipPath)
    Write-Host "✅ Instalador ZIP creado: $ZipPath" -ForegroundColor Green
} catch {
    Write-Warning "⚠️  No se pudo crear ZIP, usando Compress-Archive..."
    Compress-Archive -Path "$TempDir\*" -DestinationPath $ZipPath -Force
    Write-Host "✅ Instalador ZIP creado: $ZipPath" -ForegroundColor Green
}

# Limpiar directorio temporal
Remove-Item $TempDir -Recurse -Force

Write-Host ""
Write-Host "🎉 ¡Instalador creado exitosamente!" -ForegroundColor Magenta
Write-Host "📦 Archivo: $ZipPath" -ForegroundColor White
Write-Host ""
Write-Host "📋 Instrucciones para el usuario:" -ForegroundColor Cyan
Write-Host "1. Extrae el archivo ZIP" -ForegroundColor White
Write-Host "2. Ejecuta 'instalar.bat' como administrador" -ForegroundColor White
Write-Host "3. Sigue las instrucciones en pantalla" -ForegroundColor White
Write-Host ""
