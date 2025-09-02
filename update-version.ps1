#!/usr/bin/env pwsh
# Script para actualizar la versión en todos los archivos del proyecto

param(
    [Parameter(Mandatory=$true)]
    [string]$Version,

    [Parameter(Mandatory=$false)]
    [switch]$DryRun = $false
)

# Validar formato de versión
if ($Version -notmatch '^\d+\.\d+\.\d+$') {
    Write-Host "Error: La versión debe tener el formato X.X.X (ej: 1.0.8)" -ForegroundColor Red
    exit 1
}

Write-Host "=== Actualizando versión a $Version ===" -ForegroundColor Cyan
if ($DryRun) {
    Write-Host "MODO DRY RUN - No se realizarán cambios" -ForegroundColor Yellow
}

# Archivos a actualizar
$files = @(
    @{
        Path = "app\updater.go"
        Pattern = 'const CurrentVersion = "\d+\.\d+\.\d+"'
        Replace = "const CurrentVersion = `"$Version`""
        Description = "updater.go (backend version)"
    },
    @{
        Path = "app\wails.json"
        Pattern = '"productVersion": "\d+\.\d+\.\d+"'
        Replace = "`"productVersion`": `"$Version`""
        Description = "wails.json (product info)"
    },
    @{
        Path = "app\build\windows\installer\project.nsi"
        Pattern = '!define INFO_PRODUCTVERSION "\d+\.\d+\.\d+"'
        Replace = "!define INFO_PRODUCTVERSION `"$Version`""
        Description = "project.nsi (installer version)"
    },
    @{
        Path = "app\frontend\package.json"
        Pattern = '"version": "\d+\.\d+\.\d+"'
        Replace = "`"version`": `"$Version`""
        Description = "package.json (frontend version)"
    },
    @{
        Path = "README.md"
        Pattern = 'versaDumps v\d+\.\d+\.\d+'
        Replace = "versaDumps v$Version"
        Description = "README.md (documentation version)"
    }
)

$updated = 0
$errors = 0

foreach ($file in $files) {
    $filePath = Join-Path $PSScriptRoot $file.Path

    if (-not (Test-Path $filePath)) {
        Write-Host "  ⚠ No encontrado: $($file.Description)" -ForegroundColor Yellow
        continue
    }

    try {
        $content = Get-Content $filePath -Raw
        $newContent = $content -replace $file.Pattern, $file.Replace

        if ($content -eq $newContent) {
            Write-Host "  ○ Sin cambios: $($file.Description)" -ForegroundColor Gray
        } else {
            # Mostrar qué línea se va a cambiar
            $oldLines = $content -split "`n"
            $newLines = $newContent -split "`n"
            for ($i = 0; $i -lt $oldLines.Length; $i++) {
                if ($oldLines[$i] -ne $newLines[$i]) {
                    if ($DryRun) {
                        Write-Host "  📝 $($file.Description):" -ForegroundColor Cyan
                        Write-Host "     Antes: $($oldLines[$i].Trim())" -ForegroundColor Red
                        Write-Host "     Después: $($newLines[$i].Trim())" -ForegroundColor Green
                    }
                    break
                }
            }

            if (-not $DryRun) {
                Set-Content -Path $filePath -Value $newContent -NoNewline
            }
            Write-Host "  ✓ Actualizado: $($file.Description)" -ForegroundColor Green
            $updated++
        }
    } catch {
        Write-Host "  ✗ Error en: $($file.Description) - $_" -ForegroundColor Red
        $errors++
    }
}

Write-Host ""
Write-Host "=== Resumen ===" -ForegroundColor Cyan
Write-Host "Archivos actualizados: $updated" -ForegroundColor $(if ($updated -gt 0) { "Green" } else { "Gray" })
if ($errors -gt 0) {
    Write-Host "Errores: $errors" -ForegroundColor Red
}

if ($DryRun) {
    Write-Host ""
    Write-Host "Para aplicar los cambios, ejecuta sin -DryRun:" -ForegroundColor Yellow
    Write-Host "  .\update-version.ps1 -Version $Version" -ForegroundColor White
} else {
    Write-Host ""
    Write-Host "✓ Versión actualizada a $Version" -ForegroundColor Green
    Write-Host ""
    Write-Host "Próximos pasos:" -ForegroundColor Yellow
    Write-Host "  1. Commitear los cambios: git add -A && git commit -m `"chore: bump version to $Version`"" -ForegroundColor White
    Write-Host "  2. Crear release: .\release.ps1 -Version `"$Version`" -Push" -ForegroundColor White
}
