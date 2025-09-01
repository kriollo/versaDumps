# Changelog

Todos los cambios notables en VersaDumps Visualizer serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.9] - 2025-09-01

### ✨ Agregado
- Sistema de gestión de versiones dinámico desde el backend
- Script `update-version.ps1` para actualizar la versión en todos los archivos automáticamente
- La versión ahora se obtiene completamente desde `updater.go` sin valores hardcodeados en el frontend

### 🔧 Mejorado
- El frontend ahora muestra la versión real desde el backend en todos los lugares
- Simplificación del mantenimiento de versiones

## [1.0.8] - 2025-09-01

### ✨ Agregado
- **Sistema completo de auto-actualización**
  - Verificación automática de nuevas versiones desde GitHub Releases
  - Descarga con barra de progreso
  - Instalación automática con elevación de privilegios (UAC)
  - Notificaciones del sistema cuando hay actualizaciones
- **Indicador de versión** en la esquina inferior izquierda
- **Badge de actualización** que aparece cuando hay nueva versión disponible
- Componente `UpdateNotification.vue` para gestionar actualizaciones
- Soporte multi-idioma para el sistema de actualizaciones

### 🎨 Mejorado
- Icono corporativo actualizado
- Configuración correcta del nombre de la aplicación (VersaDumps en lugar de app)
- Ruta de instalación mejorada: `C:\Program Files\VersaDumps\`
- Información del producto en el instalador

### 🔧 Técnico
- Separación del código de actualización por plataforma (`updater_windows.go`, `updater_unix.go`)
- Integración con GitHub API para verificar releases

## [1.0.7] - 2025-08-29

### ✨ Agregado
- **Instalador NSIS para Windows**
  - Instalador profesional con interfaz gráfica
  - Versión portable en ZIP
  - Soporte para español e inglés
  - Desinstalación limpia desde Panel de Control

### 🔧 Corregido
- Configuración del instalador NSIS con rutas correctas
- Nombres de archivos de salida en el workflow

## [1.0.6] - 2025-08-29

### 🔧 Corregido
- Detección dinámica de la versión de webkit2gtk disponible en Ubuntu
- Compatibilidad mejorada con diferentes versiones de Ubuntu en GitHub Actions

## [1.0.5] - 2025-08-29

### 🔧 Corregido
- Dependencias de Ubuntu para webkit2gtk-4.0-dev
- Instalación correcta de paquetes en el workflow de GitHub Actions

## [1.0.4] - 2025-08-29

### 🐛 Corregido
- Error de case-sensitivity en Linux: `lineHr.vue` → `LineHr.vue`
- Compilación exitosa en sistemas Linux

## [1.0.3] - 2025-08-29

### 🔧 Corregido
- Nombres de paquetes webkit correctos para Ubuntu
- Manejo de nombres de salida por defecto de Wails
- Renombrado correcto de ejecutables después de la compilación

## [1.0.2] - 2025-08-29

### 🔧 Corregido
- Rutas de caché para `go.sum` y `package-lock.json` en GitHub Actions
- Copia correcta de `config.yml` al directorio de build
- Problemas de compilación en el workflow

## [1.0.1] - 2025-08-29

### ✨ Agregado
- Workflow simplificado de GitHub Actions con jobs separados por OS
- Workflow de prueba para Windows

### 🔧 Corregido
- Versión de Wails y parámetros de build
- Eliminación de flags problemáticos en el workflow
- Mejor compatibilidad con diferentes sistemas operativos

## [1.0.0] - 2025-08-29

### 🎉 Release Inicial

### ✨ Características Principales

#### **Interfaz de Usuario**
- Visualizador de dumps/logs en tiempo real
- Tema oscuro/claro con persistencia
- Soporte multi-idioma (Español/Inglés)
- Vista de árbol JSON expandible/colapsible
- Resaltado de sintaxis para código
- Ordenamiento de logs (más recientes/más antiguos)
- Badge contador de logs en la barra de título

#### **Funcionalidades Core**
- Servidor HTTP integrado para recibir dumps (puerto configurable)
- Procesamiento en tiempo real de datos JSON
- Vista detallada de stack traces con información de archivo y línea
- Capacidad de abrir archivos directamente en el editor (VS Code preferido)
- Notificaciones del sistema cuando la ventana está minimizada
- Limpieza de todos los logs con un clic

#### **Configuración**
- Modal de configuración con:
  - Servidor y puerto personalizables
  - Selección de idioma
  - Cambio de tema
  - Opción para mostrar/ocultar tipos de variables
- Persistencia de configuración en `config.yml`
- Carga automática de configuración al iniciar

#### **Integración con Sistema**
- **Windows**: Actualización del ícono en la barra de tareas con contador
- **macOS**: Badge en el Dock con contador
- **Linux**: Soporte básico de notificaciones
- Compilación multiplataforma con Wails

#### **Automatización y CI/CD**
- GitHub Actions workflow para builds automáticos
- Generación de releases para Windows, macOS y Linux
- Script de release (`release.ps1`) para facilitar versionado
- Documentación completa de instalación y uso

### 🛠️ Stack Tecnológico
- **Backend**: Go 1.23
- **Frontend**: Vue 3 + Vite
- **Framework**: Wails v2.10.2
- **Estilos**: Tailwind CSS
- **Iconos**: Sistema de iconos personalizado
- **Build**: GitHub Actions para CI/CD

### 📦 Formatos de Distribución
- **Windows**: Ejecutable portable (.exe)
- **macOS**: Archivo tar.gz
- **Linux**: Archivo tar.gz

### 📝 Documentación
- README completo en español
- Guía de instalación paso a paso
- Ejemplos de uso con Laravel
- Documentación de la API HTTP

---

## Convenciones

### Tipos de Cambios
- ✨ **Agregado**: Nueva funcionalidad
- 🔧 **Corregido**: Corrección de errores
- 🎨 **Mejorado**: Mejoras en funcionalidad existente
- 📝 **Documentación**: Cambios en documentación
- 🐛 **Bug Fix**: Corrección de bugs específicos
- ♻️ **Refactor**: Cambios de código sin afectar funcionalidad
- 🚀 **Performance**: Mejoras de rendimiento
- 🔒 **Seguridad**: Correcciones de seguridad

### Enlaces
- [Repositorio](https://github.com/kriollo/versaDumps)
- [Releases](https://github.com/kriollo/versaDumps/releases)
- [Issues](https://github.com/kriollo/versaDumps/issues)
