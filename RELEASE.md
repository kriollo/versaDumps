# 📦 Release Process

Este documento describe el proceso para crear nuevas releases de VersaDumps.

## 🚀 Proceso Automático (Recomendado)

### Prerrequisitos

1. Asegúrate de que todos los cambios estén commiteados y pusheados a `main`
2. Verifica que el código compile correctamente localmente:
   ```powershell
   cd app
   wails build
   ```

### Crear una Release

#### Opción 1: Usando el script local (Windows)

```powershell
# Crear release v1.0.0 y subirla automáticamente
.\release.ps1 -Version 1.0.0 -Push

# O crear solo el tag localmente
.\release.ps1 -Version 1.0.0

# Con mensaje personalizado
.\release.ps1 -Version 1.0.0 -Message "Primera release estable" -Push
```

#### Opción 2: Manualmente con Git

```bash
# Crear el tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# Subir el tag a GitHub
git push origin v1.0.0
```

#### Opción 3: Desde GitHub Actions

1. Ve a la pestaña **Actions** en GitHub
2. Selecciona el workflow **Release**
3. Click en **Run workflow**
4. Ingresa la versión (ej: `v1.0.0`)
5. Click en **Run workflow**

### ¿Qué sucede después?

Una vez que el tag se sube a GitHub:

1. **GitHub Actions se activa automáticamente** y comienza a compilar
2. **Compila para 3 plataformas**:
   - Windows (amd64)
   - macOS (amd64)
   - Linux (amd64)
3. **Crea archivos comprimidos** con el ejecutable y `config.yml`
4. **Genera un changelog** automático con los commits desde la última release
5. **Crea la release en GitHub** con todos los binarios adjuntos

### Verificar la Release

1. Ve a https://github.com/[tu-usuario]/versaDumps/releases
2. Verifica que la nueva release aparezca con todos los binarios
3. Descarga y prueba al menos uno de los binarios

## 📋 Versionado

Seguimos [Semantic Versioning](https://semver.org/):

- **MAJOR** (v**1**.0.0): Cambios incompatibles con versiones anteriores
- **MINOR** (v1.**1**.0): Nueva funcionalidad compatible con versiones anteriores
- **PATCH** (v1.0.**1**): Correcciones de bugs compatibles con versiones anteriores

### Ejemplos:

- `v1.0.0` - Primera versión estable
- `v1.1.0` - Se agregó nueva funcionalidad (ej: badge de notificaciones)
- `v1.0.1` - Se corrigió un bug menor
- `v2.0.0` - Cambio mayor en la API o estructura

## 🛠️ Compilación Manual

Si necesitas compilar manualmente para una plataforma específica:

### Windows
```powershell
cd app
wails build -platform windows/amd64 -o versaDumps.exe
```

### macOS
```bash
cd app
wails build -platform darwin/amd64 -o versaDumps
```

### Linux
```bash
cd app
wails build -platform linux/amd64 -o versaDumps
```

### Crear archivo distribuible

#### Windows (ZIP)
```powershell
cd app/build/bin
Compress-Archive -Path versaDumps.exe, config.yml -DestinationPath versaDumps-windows-amd64.zip
```

#### macOS/Linux (TAR.GZ)
```bash
cd app/build/bin
tar -czf versaDumps-platform-amd64.tar.gz versaDumps config.yml
```

## 🔍 Troubleshooting

### El workflow de GitHub Actions falla

1. **Verifica los logs**: Click en el workflow fallido para ver los detalles
2. **Problemas comunes**:
   - Dependencias de frontend no instaladas: Asegúrate de que `package-lock.json` esté commiteado
   - Error de compilación Go: Verifica que el código compile localmente
   - Permisos: Asegúrate de que el workflow tenga permisos de escritura

### El tag ya existe

```bash
# Eliminar tag local
git tag -d v1.0.0

# Eliminar tag remoto
git push --delete origin v1.0.0
```

### Necesito hacer una release de emergencia

1. Crea un branch desde el último tag estable:
   ```bash
   git checkout -b hotfix/critical-bug v1.0.0
   ```
2. Aplica el fix
3. Crea el tag con versión patch:
   ```bash
   git tag -a v1.0.1 -m "Hotfix: critical bug"
   git push origin v1.0.1
   ```

## 📊 Checklist Pre-Release

- [ ] Todos los tests pasan (si existen)
- [ ] El código compila sin warnings
- [ ] La documentación está actualizada
- [ ] El CHANGELOG está actualizado (si se mantiene manualmente)
- [ ] Se probó la aplicación en al menos una plataforma
- [ ] No hay secretos o información sensible en el código

## 🔗 Enlaces Útiles

- [Releases en GitHub](https://github.com/[tu-usuario]/versaDumps/releases)
- [Actions/Workflows](https://github.com/[tu-usuario]/versaDumps/actions)
- [Semantic Versioning](https://semver.org/)
- [Wails Build Documentation](https://wails.io/docs/reference/cli#build)
