# Configuración de Iconos para VersaDumps

Este proyecto está configurado para usar el archivo `app/build/appicon.png` como icono base para todas las plataformas.

## Archivos de Configuración

### 1. Estructura de Iconos
```
app/build/
├── appicon.png          # Icono base (fuente)
├── windows/
│   └── icon.ico        # Icono para Windows (generado)
├── darwin/
│   └── icon.png        # Icono para macOS (copia del base)
└── linux/
    └── icon.png        # Icono para Linux (copia del base)
```

### 2. Scripts de Configuración

#### Windows (PowerShell)
```powershell
.\setup-icons.ps1
```

#### Unix/macOS/Linux (Bash)
```bash
./setup-icons.sh
```

### 3. Herramienta de Conversión Go
En `app/tools/convert-icon.go` se encuentra una herramienta personalizada que convierte PNG a ICO válido usando la biblioteca `golang-image-ico`.

## Configuración Automática

### GitHub Actions
El workflow `release-simple.yml` está configurado para generar automáticamente los iconos durante el build:

- **Windows**: Genera ICO desde PNG usando PowerShell
- **macOS**: Copia el PNG como icon.png
- **Linux**: Copia el PNG como icon.png

### Wails
El archivo `wails.json` está configurado para usar los iconos automáticamente.

### Instalador NSIS
El archivo `app/build/windows/installer/project.nsi` está configurado para usar el icono ICO generado.

## Uso

1. **Reemplazar el icono**: Simplemente reemplaza `app/build/appicon.png` con tu nuevo icono
2. **Ejecutar configuración**: Corre el script `setup-icons.ps1` (Windows) o `setup-icons.sh` (Unix)
3. **Compilar**: Ejecuta `wails build` para generar la aplicación con los nuevos iconos

## Requisitos del Icono Base

- **Formato**: PNG
- **Tamaño recomendado**: 256x256 píxeles o superior
- **Transparencia**: Soportada
- **Ubicación**: `app/build/appicon.png`

## Verificación

Para verificar que los iconos se han configurado correctamente:

```bash
# Verificar que existen los archivos
ls -la app/build/windows/icon.ico
ls -la app/build/darwin/icon.png
ls -la app/build/linux/icon.png

# Compilar aplicación con iconos
cd app && wails build -platform windows/amd64
```

## Troubleshooting

### Error "couldn't load icon from icon.ico: unexpected EOF"
- El archivo ICO está corrupto
- Ejecuta `setup-icons.ps1` nuevamente
- O usa la herramienta Go: `cd app/tools && go run convert-icon.go`

### Icono no aparece en la aplicación
- Verifica que el archivo `appicon.png` existe y es válido
- Asegúrate de que los iconos específicos de plataforma se generaron correctamente
- Recompila la aplicación después de configurar los iconos
