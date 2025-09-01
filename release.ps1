#!/usr/bin/env pwsh
# Script para crear una nueva release de VersaDumps

param(
    [Parameter(Mandatory=$true)]
    [string]$Version,
    
    [Parameter(Mandatory=$false)]
    [string]$Message = "Release version $Version",
    
    [Parameter(Mandatory=$false)]
    [switch]$Push = $false
)

# Validar formato de versión
if ($Version -notmatch '^v?\d+\.\d+\.\d+$') {
    Write-Host "Error: La versión debe tener el formato v1.0.0 o 1.0.0" -ForegroundColor Red
    exit 1
}

# Asegurar que la versión empiece con 'v'
if ($Version -notmatch '^v') {
    $Version = "v$Version"
}

Write-Host "=== VersaDumps Release Script ===" -ForegroundColor Cyan
Write-Host "Version: $Version" -ForegroundColor Yellow
Write-Host "Message: $Message" -ForegroundColor Yellow
Write-Host ""

# Verificar que estamos en el directorio correcto
if (-not (Test-Path "app\main.go")) {
    Write-Host "Error: Este script debe ejecutarse desde la raíz del proyecto" -ForegroundColor Red
    exit 1
}

# Verificar que no hay cambios sin commitear
$gitStatus = git status --porcelain
if ($gitStatus) {
    Write-Host "Error: Hay cambios sin commitear. Por favor, commitea o descarta los cambios primero." -ForegroundColor Red
    Write-Host $gitStatus
    exit 1
}

# Actualizar la versión en el README si existe
$readmePath = "README.md"
if (Test-Path $readmePath) {
    Write-Host "Actualizando versión en README.md..." -ForegroundColor Green
    $readme = Get-Content $readmePath -Raw
    $readme = $readme -replace 'VersaDumps Visualizer v[\d\.]+', "VersaDumps Visualizer $Version"
    $readme = $readme -replace 'version-v[\d\.]+-blue', "version-$Version-blue"
    Set-Content -Path $readmePath -Value $readme -NoNewline
    
    git add $readmePath
    git commit -m "chore: update version to $Version in README"
}

# Crear el tag
Write-Host "Creando tag $Version..." -ForegroundColor Green
git tag -a $Version -m "$Message"

if ($?) {
    Write-Host "✓ Tag $Version creado exitosamente" -ForegroundColor Green
    
    if ($Push) {
        Write-Host "Subiendo cambios y tag a GitHub..." -ForegroundColor Yellow
        git push origin main
        git push origin $Version
        
        if ($?) {
            Write-Host "✓ Tag subido exitosamente" -ForegroundColor Green
            Write-Host ""
            Write-Host "La GitHub Action se ejecutará automáticamente y creará la release con los binarios." -ForegroundColor Cyan
            Write-Host "Puedes ver el progreso en: https://github.com/tu-usuario/versaDumps/actions" -ForegroundColor Cyan
        } else {
            Write-Host "✗ Error al subir el tag" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host ""
        Write-Host "Tag creado localmente. Para subirlo a GitHub y crear la release, ejecuta:" -ForegroundColor Yellow
        Write-Host "  git push origin main" -ForegroundColor White
        Write-Host "  git push origin $Version" -ForegroundColor White
        Write-Host ""
        Write-Host "O ejecuta este script con -Push:" -ForegroundColor Yellow
        Write-Host "  .\release.ps1 -Version $Version -Push" -ForegroundColor White
    }
} else {
    Write-Host "✗ Error al crear el tag" -ForegroundColor Red
    exit 1
}
