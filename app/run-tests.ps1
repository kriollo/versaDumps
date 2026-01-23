# Script para configurar y ejecutar tests en versaDumps
# PowerShell version

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  VersaDumps - Setup de Tests" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Funciones para output con color
function Print-Success {
    param($Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Print-Warning {
    param($Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

function Print-Error {
    param($Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

# Verificar que estamos en el directorio correcto
if (-not (Test-Path "go.mod")) {
    Print-Error "Este script debe ejecutarse desde el directorio 'app'"
    exit 1
}

Print-Success "Directorio correcto detectado"

# 1. Verificar Go
Write-Host ""
Write-Host "Verificando Go..." -ForegroundColor Yellow

try {
    $goVersion = go version
    Print-Success "Go encontrado: $goVersion"
} catch {
    Print-Error "Go no está instalado"
    exit 1
}

# 2. Verificar Node.js
Write-Host ""
Write-Host "Verificando Node.js..." -ForegroundColor Yellow

try {
    $nodeVersion = node --version
    Print-Success "Node.js encontrado: $nodeVersion"
} catch {
    Print-Error "Node.js no está instalado"
    exit 1
}

# 3. Ejecutar tests de Go
Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Ejecutando Tests Go" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

go test ./... -v | Tee-Object -FilePath "test-results-go.log"
$goExitCode = $LASTEXITCODE

if ($goExitCode -eq 0) {
    Print-Success "Tests Go: PASARON"
} else {
    Print-Warning "Tests Go: ALGUNOS FALLARON (ver test-results-go.log)"
}

# 4. Instalar dependencias de Vue si es necesario
Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Configurando Tests Vue" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

Set-Location frontend

if (-not (Test-Path "node_modules")) {
    Print-Warning "node_modules no encontrado. Instalando dependencias..."
    npm install
    if ($LASTEXITCODE -eq 0) {
        Print-Success "Dependencias instaladas correctamente"
    } else {
        Print-Error "Error instalando dependencias"
        Set-Location ..
        exit 1
    }
} else {
    Print-Success "node_modules ya existe"
    
    # Verificar si necesitamos instalar las nuevas dependencias de testing
    $hasVitest = npm list vitest 2>$null
    if ($LASTEXITCODE -ne 0) {
        Print-Warning "Dependencias de testing no encontradas. Instalando..."
        npm install
        if ($LASTEXITCODE -eq 0) {
            Print-Success "Dependencias de testing instaladas"
        } else {
            Print-Error "Error instalando dependencias de testing"
            Set-Location ..
            exit 1
        }
    } else {
        Print-Success "Dependencias de testing ya instaladas"
    }
}

# 5. Ejecutar tests de Vue
Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Ejecutando Tests Vue" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

npm run test run | Tee-Object -FilePath "../test-results-vue.log"
$vueExitCode = $LASTEXITCODE

if ($vueExitCode -eq 0) {
    Print-Success "Tests Vue: PASARON"
} else {
    Print-Warning "Tests Vue: ALGUNOS FALLARON (ver test-results-vue.log)"
}

# Volver al directorio app
Set-Location ..

# 6. Resumen final
Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Resumen de Tests" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

if ($goExitCode -eq 0) {
    Print-Success "Tests Go: PASARON"
} else {
    Print-Warning "Tests Go: FALLARON"
}

if ($vueExitCode -eq 0) {
    Print-Success "Tests Vue: PASARON"
} else {
    Print-Warning "Tests Vue: FALLARON"
}

Write-Host ""
Write-Host "Logs guardados en:"
Write-Host "  - test-results-go.log"
Write-Host "  - test-results-vue.log"
Write-Host ""

# 7. Generar reporte de coverage
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Generando Reportes de Coverage" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Generando coverage Go..." -ForegroundColor Yellow
go test ./... -coverprofile=coverage.out -cover
if ($LASTEXITCODE -eq 0) {
    go tool cover -html=coverage.out -o coverage-go.html
    Print-Success "Coverage Go generado: coverage-go.html"
}

Write-Host ""
Write-Host "Generando coverage Vue..." -ForegroundColor Yellow
Set-Location frontend
npm run test:coverage | Out-Null
if ($LASTEXITCODE -eq 0) {
    Print-Success "Coverage Vue generado en: frontend/coverage/"
}
Set-Location ..

Write-Host ""
Print-Success "Setup y tests completados!"
Write-Host ""
Write-Host "Para ver los reportes HTML de coverage:"
Write-Host "  Go:  abrir coverage-go.html en navegador"
Write-Host "  Vue: abrir frontend/coverage/index.html en navegador"
Write-Host ""
