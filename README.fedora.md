# VersaDumps en Fedora

Guía para ejecutar VersaDumps en Fedora Linux.

## Dependencias del Sistema

```bash
# Herramientas de desarrollo y bibliotecas GTK/WebKit
sudo dnf install -y gcc pkg-config gtk3-devel webkit2gtk4.1-devel

# Go (si no está instalado)
sudo dnf install -y golang

# Node.js (si no está instalado)
sudo dnf install -y nodejs npm
```

## Configuración Inicial

### 1. Instalar Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0
```

### 2. Configurar WebKit Compatibility

Fedora usa `webkit2gtk-4.1` pero Wails busca `webkit2gtk-4.0`. El script `run-fedora.sh` configura esto automáticamente, pero si necesitas hacerlo manualmente:

```bash
mkdir -p ~/.local/lib/pkgconfig
ln -sf /usr/lib64/pkgconfig/webkit2gtk-4.1.pc ~/.local/lib/pkgconfig/webkit2gtk-4.0.pc
```

### 3. Instalar Dependencias del Frontend

```bash
cd app/frontend
npm install
```

## Ejecutar la Aplicación

### Modo Desarrollo (Recomendado)

Usando el script helper:

```bash
./run-fedora.sh
```

O manualmente:

```bash
cd app
export PKG_CONFIG_PATH="$HOME/.local/lib/pkgconfig:$PKG_CONFIG_PATH"
~/go/bin/wails dev
```

La aplicación estará disponible en:
- **DevServer**: http://localhost:34115
- **Frontend**: http://localhost:5173/ (o 5174, 5175 si el puerto está ocupado)

### Build de Producción

```bash
cd app
export PKG_CONFIG_PATH="$HOME/.local/lib/pkgconfig:$PKG_CONFIG_PATH"
~/go/bin/wails build -platform linux/amd64
```

El binario estará en `app/build/bin/VersaDumps`.

## Empaquetado para Distribución

### Crear Tarball

```bash
./packaging/build.sh <version>
# Ejemplo: ./packaging/build.sh 3.1.0
```

Esto genera:
- `dist/versaDumps-<version>-linux-amd64.tar.gz`

### Crear RPM (requiere rpmbuild)

```bash
sudo dnf install -y rpm-build rpmdevtools

# Crear estructura rpmbuild
rpmdev-setuptree

# Copiar tarball a SOURCES
cp dist/versaDumps-<version>-linux-amd64.tar.gz ~/rpmbuild/SOURCES/

# Construir RPM
rpmbuild -ba packaging/fedora/versaDumps.spec --define "version <version>"
```

El RPM estará en `~/rpmbuild/RPMS/x86_64/`.

## Notas

- **Compatibilidad Windows**: Todos los cambios son específicos de Linux. El soporte Windows permanece 100% intacto.
- **Config.yml**: Si el archivo de configuración tiene rutas Windows (C:\\Users\\...), estas no funcionarán en Linux. Ajusta las rutas en `~/.config/VersaDumps/config.yml` según sea necesario.
- **Auto-actualización**: El instalador Unix (`app/updater_unix.go`) soporta paquetes `.tar.gz` y binarios directos. Si la app no puede sobrescribir el binario por permisos, mostrará instrucciones.

## Troubleshooting

### Error: "Package webkit2gtk-4.0 was not found"

Ejecuta el script `run-fedora.sh` o configura manualmente PKG_CONFIG_PATH:

```bash
export PKG_CONFIG_PATH="$HOME/.local/lib/pkgconfig:$PKG_CONFIG_PATH"
```

### Error: "gcc not found"

Instala herramientas de desarrollo:

```bash
sudo dnf install -y gcc
```

### Error: "Cannot find module @rollup/rollup-linux-x64-gnu"

Reinstala dependencias del frontend:

```bash
cd app/frontend
rm -rf node_modules package-lock.json
npm install
```

## Estructura de Archivos Clave

- `app/updater_unix.go` - Actualizador para sistemas Unix/Linux
- `app/badge_unix.go` - Badge/tray para Unix (stub, puede mejorarse)
- `packaging/fedora/versaDumps.spec` - Spec RPM para Fedora
- `packaging/fedora/versaDumps.desktop` - Archivo desktop entry
- `packaging/build.sh` - Script de empaquetado
- `.github/workflows/build-linux.yml` - CI para builds Linux
- `run-fedora.sh` - Script helper para ejecutar en Fedora
