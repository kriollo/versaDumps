# Changelog

Todos los cambios notables en VersaDumps Visualizer serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.2.0] - 2025-10-28

### ✨ Agregado
- **Soporte completo para versadumps-php 2.2.0**:
  - Integración con la nueva estructura de payload que incluye objeto `metadata`
  - Visualización de stack traces completos con información detallada de frames
  - Soporte para métodos semánticos (success, error, info, warning, important)
  - Compatibilidad con 10 colores personalizados (red, blue, green, yellow, orange, purple, pink, cyan, gray, white)
  - Procesamiento de etiquetas con emojis
  - Soporte para ejecución condicional (if/unless)
  - Manejo de método once() para prevenir duplicados en loops
  - Profundidad máxima configurable para serialización

### 🎨 Interfaz
- **Componente de Stack Trace**: Nueva sección expandible en cada log que muestra:
  - Clase y método donde se originó el dump
  - Ruta completa del archivo
  - Número de línea exacto
  - Jerarquía completa de llamadas (frames)
- **Colores semánticos**: Cada log muestra un borde de color según su tipo:
  - Verde para success
  - Rojo para error/important
  - Azul para info
  - Amarillo para warning
  - Y 6 colores adicionales personalizables
- **Normalización mejorada de payloads**: Procesamiento inteligente que soporta tanto el formato 2.2.0 como versiones anteriores

### 🔧 Mejorado
- **Procesamiento de metadata**: Extracción correcta de `metadata.trace` del payload PHP
- **Compatibilidad retroactiva**: El visualizador mantiene soporte para formatos antiguos de payload
- **Logging de debugging**: Logs detallados en consola para diagnóstico (`📦 Payload recibido`, `🔄 Datos normalizados`)
- **Preservación de metadata**: El objeto metadata ahora se mantiene completo para debugging y procesamiento posterior

### 🐛 Corregido
- **Bug crítico en extracción de trace**: Corregida la lectura de `metadata.trace` (anteriormente buscaba incorrectamente `metadata.includeTrace`)
- **Pérdida de metadata**: Se eliminó el código que borraba prematuramente el objeto `metadata` del payload
- **Mapeo de colores**: Corrección en el mapeo de colores semánticos a clases Tailwind CSS
- **Visualización de traces**: Los stack traces ahora se muestran correctamente en la interfaz

### 🔧 Técnico
- **Estructura de payload normalizada**: Sistema robusto que maneja:
  - `metadata.trace`: Array de frames con información de stack trace
  - `metadata.color`: Color personalizado del log
  - `metadata.max_depth`: Profundidad máxima de serialización
  - Fallbacks para compatibilidad con versiones anteriores
- **Computed properties optimizadas**:
  - `traceFrames`: Procesa y formatea frames de stack trace
  - `borderColor`: Determina color del borde basado en metadata o hash de archivo
  - `semanticColors`: Mapeo completo de colores a clases CSS
- **Mejor manejo de context.variables**: Soporte para la nueva estructura donde variables vienen dentro de `context.variables`

### 📝 Compatibilidad
- **versadumps-php 2.2.0**: Soporte completo para todas las características de la librería PHP actualizada
- **Builder Pattern**: Compatible con el nuevo patrón de construcción encadenado
- **Métodos semánticos**: Integración con success(), error(), info(), warning(), important()
- **Backward compatible**: Mantiene compatibilidad con payloads de versiones 2.1.0 y anteriores

### 🚀 Performance
- **Procesamiento optimizado**: Normalización de payloads sin impacto en rendimiento
- **Renderizado condicional**: Stack traces solo se procesan y muestran cuando están presentes
- **Carga eficiente**: Expansión/colapso de traces sin recargar componentes

## [2.1.0] - 2025-09-16

### 🔧 Corregido
- **Indicador de estado del servidor**: Se añadió un endpoint `/health` y un indicador en la interfaz que muestra estado 'online'/'offline'/'unknown' con sondeo cada 5s.
- **Reinicio del servidor al guardar configuración**: Ahora la aplicación reinicia el servidor HTTP internamente cuando se guardan cambios relevantes en `config.yml` (por ejemplo, cambio de puerto o host), aplicando los nuevos valores sin requerir reinicio manual.
- **Manejo de puerto y arranque**: Se corrigieron problemas con la configuración del puerto (anteriormente permisos y puerto erróneo). El servidor ahora se inicia correctamente en el arranque con la configuración cargada.
- **Corrección en comprobación de actualizaciones**: Evita falsos positivos cuando la API de GitHub responde con rate limiting; se añade verificación adicional en el frontend para asegurar que la versión reportada es realmente más nueva.

### 🎨 Interfaz
- **Recarga de configuración en caliente**: Al guardar la configuración desde el modal, la UI aplica los cambios de tema y idioma inmediatamente y reinicia el sondeo de salud del servidor si cambian host/puerto.

### 🔧 Técnico
- **Mejor sincronización frontend/backend**: Guardado de configuración desde frontend ahora persiste y notifica al backend para aplicar los cambios sin necesidad de reiniciar la aplicación.
- **Logging mejorado**: Mensajes informativos añadidos para el proceso de guardado y reinicio del servidor para facilitar debugging.

## [2.0.1] - 2025-09-08

### ✨ Agregado
- **Sistema de etiquetas personalizadas (Labels)**: Nueva propiedad `label` en dumps de PHP que reemplaza automáticamente la primera clave del contexto
- **Función de copia al portapapeles**: Botón de copia en cada log con notificación toast de confirmación
- **Notificaciones toast**: Sistema de feedback visual para acciones del usuario con animaciones suaves
- **Soporte completo para i18n**: Textos de notificaciones traducidos en español e inglés

### 🔧 Mejorado
- **Sincronización mejorada de badges**: Corrección completa del sistema de contadores entre frontend y backend
- **Logs de debugging avanzados**: Sistema completo de logs para diagnóstico de problemas de sincronización
- **Procesamiento inteligente de datos**: Manejo automático de arrays y objetos en el sistema de labels
- **Gestión robusta de contadores**: Verificación y corrección automática de desincronizaciones entre UI y sistema operativo
- **Experiencia de usuario mejorada**: Feedback inmediato para todas las acciones principales

### 🎨 Interfaz
- **Toast notifications**: Notificaciones elegantes en la esquina superior derecha con animación slide-in
- **Iconos mejorados**: Nuevo botón de copia con icono dedicado en cada log
- **Animaciones fluidas**: Transiciones CSS para todas las notificaciones y estados de la UI
- **Mejor feedback visual**: Confirmaciones claras para acciones como copiar y limpiar logs

### 🔧 Técnico
- **Arquitectura de labels**: Sistema robusto para procesar etiquetas personalizadas desde PHP
  - Soporte para contextos de tipo array y object
  - Preservación del orden y estructura de datos
  - Eliminación automática de propiedades procesadas
- **Sistema de logs mejorado**:
  - Logs detallados en frontend (JavaScript console)
  - Logs del backend (Go runtime logs)
  - Tracking completo del flujo de datos
- **Sincronización de estado**:
  - Verificación automática de contadores al inicio
  - Corrección proactiva de desincronizaciones
  - Manejo robusto de errores en actualización de badges
- **API de clipboard moderna**: Uso de `navigator.clipboard` con manejo de errores completo

### 🐛 Corregido
- **Desincronización de badges**: Problema crítico donde el contador del título de ventana no se sincronizaba correctamente
- **Cache de Windows**: Issues con el cache del título de ventana en Windows resueltos con actualizaciones forzadas
- **Condiciones de carrera**: Eliminación de race conditions en la actualización de contadores
- **Persistencia incorrecta**: Corrección del problema donde contadores persistían incorrectamente entre sesiones

### 📝 Documentación
- Logs de debugging documentados para troubleshooting
- Especificación completa del sistema de labels
- Guía de uso del sistema de notificaciones

### 🚀 Performance
- **Optimización de contadores**: Reducción de llamadas redundantes al backend
- **Gestión eficiente de memoria**: Limpieza automática de referencias temporales
- **Lazy loading de iconos**: Carga optimizada de recursos de interfaz

### 💡 Compatibilidad
- **Windows**: Mejoras específicas para el sistema de badges en taskbar
- **Multiplataforma**: Mantenimiento de compatibilidad con macOS y Linux
- **Navegadores modernos**: Uso de APIs modernas con fallbacks apropiados

## [1.0.14] - 2025-09-01

### ✨ Agregado
- **Función de verificación manual de actualizaciones**: Nuevo botón "Revisar actualización" en el panel de configuración
- **Modal de confirmación mejorado**: Información detallada de versiones con estados diferenciados
- **Sistema de verificación dual**: Verificaciones automáticas (silenciosas) y manuales (con modal)

### 🔧 Mejorado
- **Notificaciones inteligentes**: Las verificaciones automáticas solo notifican cuando hay actualizaciones reales disponibles
- **Mejor experiencia de usuario**: Modal de confirmación con botones habilitados/deshabilitados según disponibilidad de actualizaciones
- **Interfaz refinada**: Estilos mejorados para el sistema de actualizaciones con soporte completo para modo oscuro
- **Manejo mejorado de errores**: Mejor gestión de rate limiting de la API de GitHub y errores de conexión

### 🎨 Interfaz
- Nuevo diseño del modal de actualizaciones con información clara de versiones
- Botones adaptativos que se deshabilitan cuando no hay actualizaciones disponibles
- Indicadores visuales mejorados para diferentes estados de actualización

### 🔧 Técnico
- Logs de debugging mejorados para diagnóstico del sistema de actualizaciones
- Optimización del rendimiento en verificaciones automáticas
- Compatibilidad mejorada con Windows para el sistema de actualizaciones

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
