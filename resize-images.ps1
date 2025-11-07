# Script para redimensionar imágenes para README de GitHub
# Mantiene las proporciones originales

Add-Type -AssemblyName System.Drawing

function Resize-Image {
    param(
        [string]$InputPath,
        [string]$OutputPath,
        [int]$MaxWidth
    )

    # Cargar la imagen original
    $img = [System.Drawing.Image]::FromFile($InputPath)

    # Calcular nuevas dimensiones manteniendo proporción
    $ratio = $img.Height / $img.Width
    $newWidth = $MaxWidth
    $newHeight = [int]($newWidth * $ratio)

    Write-Host "Redimensionando $([System.IO.Path]::GetFileName($InputPath)): $($img.Width)x$($img.Height) -> ${newWidth}x${newHeight}"

    # Crear nueva imagen con las dimensiones calculadas
    $newImg = New-Object System.Drawing.Bitmap($newWidth, $newHeight)
    $graphics = [System.Drawing.Graphics]::FromImage($newImg)

    # Configurar calidad alta
    $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
    $graphics.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
    $graphics.CompositingQuality = [System.Drawing.Drawing2D.CompositingQuality]::HighQuality

    # Dibujar la imagen redimensionada
    $graphics.DrawImage($img, 0, 0, $newWidth, $newHeight)

    # Limpiar recursos primero
    $graphics.Dispose()
    $img.Dispose()

    # Usar archivo temporal
    $tempPath = "$OutputPath.temp"
    $newImg.Save($tempPath, [System.Drawing.Imaging.ImageFormat]::Png)
    $newImg.Dispose()

    # Reemplazar el archivo original
    Remove-Item $OutputPath -Force
    Move-Item $tempPath $OutputPath -Force
}

# Definir tamaños objetivo para cada imagen
$artPath = "c:\Users\jjara\Desktop\proyectos\versaDumps\art"

# Redimensionar cada imagen - tamaños más pequeños para GitHub
Resize-Image -InputPath "$artPath\versaDumpsVisualizer.png" -OutputPath "$artPath\versaDumpsVisualizer.png" -MaxWidth 350
Resize-Image -InputPath "$artPath\visualizerExample.png" -OutputPath "$artPath\visualizerExample.png" -MaxWidth 400
Resize-Image -InputPath "$artPath\visualizerExampleConfig1.png" -OutputPath "$artPath\visualizerExampleConfig1.png" -MaxWidth 350
Resize-Image -InputPath "$artPath\visualizerExampleConfig2.png" -OutputPath "$artPath\visualizerExampleConfig2.png" -MaxWidth 350
Resize-Image -InputPath "$artPath\visualizerExampleConfig3.png" -OutputPath "$artPath\visualizerExampleConfig3.png" -MaxWidth 350

Write-Host "`n✓ Todas las imágenes han sido redimensionadas exitosamente" -ForegroundColor Green
