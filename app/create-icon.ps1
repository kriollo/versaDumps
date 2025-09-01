# Create a simple icon for VersaDumps
Add-Type -AssemblyName System.Drawing

# Create a bitmap
$bitmap = New-Object System.Drawing.Bitmap 256, 256

# Create graphics object
$graphics = [System.Drawing.Graphics]::FromImage($bitmap)

# Set background gradient
$rect = New-Object System.Drawing.Rectangle 0, 0, 256, 256
$brush = New-Object System.Drawing.Drawing2D.LinearGradientBrush($rect, 
    [System.Drawing.Color]::FromArgb(102, 126, 234),  # #667eea
    [System.Drawing.Color]::FromArgb(118, 75, 162),   # #764ba2
    [System.Drawing.Drawing2D.LinearGradientMode]::ForwardDiagonal)

$graphics.FillRectangle($brush, 0, 0, 256, 256)

# Draw text "VD"
$font = New-Object System.Drawing.Font("Arial", 100, [System.Drawing.FontStyle]::Bold)
$textBrush = New-Object System.Drawing.SolidBrush([System.Drawing.Color]::White)
$stringFormat = New-Object System.Drawing.StringFormat
$stringFormat.Alignment = [System.Drawing.StringAlignment]::Center
$stringFormat.LineAlignment = [System.Drawing.StringAlignment]::Center

$graphics.DrawString("VD", $font, $textBrush, 128, 128, $stringFormat)

# Save as PNG first
$pngPath = "appicon.png"
$bitmap.Save($pngPath, [System.Drawing.Imaging.ImageFormat]::Png)

# Create ICO file (basic version)
$icon = [System.Drawing.Icon]::FromHandle($bitmap.GetHicon())
$fs = New-Object System.IO.FileStream("appicon.ico", [System.IO.FileMode]::Create)
$icon.Save($fs)
$fs.Close()

# Clean up
$graphics.Dispose()
$bitmap.Dispose()
$brush.Dispose()
$textBrush.Dispose()
$font.Dispose()

Write-Host "Icon created successfully: appicon.ico and appicon.png" -ForegroundColor Green
