# Changelog

Todos los cambios notables en VersaDumps Visualizer serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.0.2] - 2025-11-07

### 🔥 CRÍTICO - Solución de Permisos
- **Ubicación del archivo de configuración movida a AppData**: Solución definitiva al problema de permisos de escritura
  - El archivo `config.yml` ahora se guarda en `%APPDATA%\VersaDumps\` en lugar de `C:\Program Files\VersaDumps\`
  - Eliminados todos los errores relacionados con permisos de escritura en Windows
  - La aplicación ya no requiere permisos de administrador para funcionar correctamente
  - Cada usuario de Windows tiene su propia configuración independiente
  - Migración automática del archivo de configuración desde la ubicación antigua

### 🐛 Corregido
- **"Error adding folder"**: Solucionado el error que impedía agregar carpetas de logs
  - Problema causado por falta de permisos de escritura en `C:\Program Files`
  - Ahora la configuración se guarda en el directorio del usuario con permisos completos
  - Agregada validación de rutas antes de guardar configuración
  - Validación de que la ruta existe en el sistema
  - Validación de que la ruta es un directorio y no un archivo
  - Validación de permisos de acceso a la ruta especificada

### ✨ Nuevo
- **Mensajes de error mejorados**: Errores más descriptivos al agregar carpetas
  - "La ruta especificada no existe" cuando la carpeta no se encuentra
  - "La ruta especificada no es un directorio" cuando se intenta agregar un archivo
  - "No se puede acceder a la ruta especificada" cuando hay problemas de permisos
  - Mensajes traducidos tanto en español como en inglés

### 🔧 Mejorado
- **Sistema de configuración robusto**:
  - Función `getConfigPath()` para obtener la ubicación correcta del archivo de configuración
  - Creación automática del directorio de configuración si no existe
  - Migración automática y transparente de configuraciones existentes
  - Mejor manejo de errores en carga y guardado de configuración
  - Logs informativos mostrando la ubicación del archivo de configuración en uso

### 📝 Técnico
- **Cambios en `config.go`**:
  - Nueva función `getConfigPath()` que usa `os.UserConfigDir()`
  - Migración automática desde ubicación antigua (`config.yml` en directorio actual)
  - Creación automática de directorio `VersaDumps` en AppData
  - Actualización de `LoadConfig()` y `SaveConfig()` para usar nueva ubicación

- **Cambios en `app.go`**:
  - Validación de rutas en `AddLogFolder()` antes de guardar
  - Verificación de existencia con `os.Stat()`
  - Verificación de tipo de archivo con `info.IsDir()`
  - Mensajes de error más descriptivos y específicos
  - Logs mostrando la ubicación del archivo de configuración

- **Cambios en frontend**:
  - Detección inteligente de tipos de error en `LogFoldersManager.vue`
  - Mapeo de errores del backend a mensajes de traducción apropiados
  - Nuevas claves de traducción en `es.js` y `en.js`

### 📚 Documentación
- **Guía de migración**: Nuevo archivo `MIGRATION.md` con instrucciones detalladas
  - Explicación del cambio de ubicación del archivo de configuración
  - Instrucciones para verificar la migración
  - Solución de problemas comunes
  - Guía de respaldo y restauración de configuración

### 💡 Notas de Actualización
- **Acción requerida**: Ninguna, la migración es automática
- **Ubicación antigua**: `C:\Program Files\VersaDumps\config.yml` (solo lectura)
- **Ubicación nueva**: `%APPDATA%\VersaDumps\config.yml` (lectura/escritura)
- **Compatibilidad**: El archivo antiguo se mantiene intacto como respaldo

---

## [3.0.1] - 2025-11-06

### 🐛 Corregido
- **Sistema de monitoreo de archivos de log**: Corrección crítica en el manejo de archivos
  - Solucionado problema de archivos bloqueados en Windows que impedía la escritura por otras aplicaciones
  - Eliminado el mantenimiento de handles de archivos abiertos permanentemente
  - Implementado sistema de apertura temporal solo para lectura con cierre inmediato
  - Los archivos ahora se abren con acceso compartido de lectura (`os.O_RDONLY`)
  - Cierre automático de archivos después de cada lectura mediante `defer`
  - Solución completa al error "locked a portion of the file" en Windows
  - Mejora significativa en la gestión de recursos del sistema
  - Prevención de errores "file locked" y "access denied" en sistemas Windows

### 🔧 Mejorado
- **LogWatcher optimizado**: Optimización en la gestión de recursos
  - Cambio de arquitectura: `LogFile` ahora solo almacena metadata (Path, LastPosition, LastModTime, LastSize)
  - Eliminado el campo `File` de la estructura `LogFile`
  - Los archivos se abren solo cuando es necesario leer nuevas líneas
  - Mejor detección de rotación de logs comparando tamaño actual vs. última posición
  - Reinicio automático desde el principio del archivo cuando se detecta rotación
  - Logs informativos cuando se detecta rotación de archivos

- **Performance del LogWatcher**:
  - Reducción drástica del uso de memoria al no mantener archivos abiertos
  - Eliminación de posibles fugas de memoria por archivos no cerrados
  - Sistema más robusto de lectura con cierre garantizado mediante `defer`
  - Manejo robusto de archivos eliminados con verificación de existencia
  - Mensajes de log más descriptivos con nombres de archivo cortos
  - Eliminada la lógica innecesaria de detección de archivos bloqueados (ya no es necesaria)

### 🎨 Interfaz
- **Código formateado**: Reformateado de `App.vue` y `LogFileViewer.vue` para mejor legibilidad
  - Indentación consistente en toda la plantilla
  - Mejor organización de atributos en elementos Vue
  - Código más limpio y mantenible

### 💡 Compatibilidad
- **Windows**: Solución definitiva para problemas de bloqueo de archivos en sistemas Windows
  - Compatible con aplicaciones que escriben en logs simultáneamente
  - No más errores de "file being used by another process"
  - Acceso compartido correcto a archivos de log

### 📝 Técnico
- **Arquitectura mejorada**:
  - Cambio de modelo de "archivos abiertos permanentemente" a "apertura temporal bajo demanda"
  - `LogFile` ahora solo contiene metadata (Path, LastPosition, LastModTime, LastSize)
  - Método `registerFile()` reemplaza a `tailFile()` para registro sin apertura
  - Método `readNewLines()` ahora maneja todo el ciclo de vida del archivo (open/read/close)
  - Eliminación del campo `File *os.File` de la estructura `LogFile`

- **Gestión de recursos**:
  - Limpieza automática de recursos al detener el watcher
  - No se requiere cerrar archivos en el shutdown (no hay archivos abiertos)
  - Mejor compatibilidad con aplicaciones que escriben a los mismos archivos de log

## [3.0.0] - 2025-11-05

### ✨ Agregado
- **Monitoreo de archivos de log**: Nueva funcionalidad completa para monitorear carpetas de archivos de log en tiempo real
  - Gestión de carpetas con rutas personalizables
  - Filtrado por extensiones de archivo (.log, .txt, etc.)
  - Filtrado por patrones de nombres de archivo (errors_*, access_*, etc.)
  - Selección de formato de log (JSON o texto plano)
  - Edición completa de configuraciones de carpetas monitoreadas
- **Visualizador de archivos de log**: Componente dedicado para ver contenido de archivos
  - Vista en tiempo real con actualización automática
  - Detección automática de formato JSON en archivos
  - Pretty-printing de JSON con indentación de 2 espacios
  - Resaltado de sintaxis para archivos JSON con colores diferenciados:
    - Claves en azul (#0066cc light / #61afef dark)
    - Strings en verde (#067d17 light / #98c379 dark)
    - Números en rojo (#d73a49 light / #d19a66 dark)
    - Booleanos en azul negrita (#005cc5 light / #56b6c2 dark)
    - Valores null en morado cursiva (#6f42c1 light / #c678dd dark)
  - Filtrado de logs en tiempo real por texto
  - Contador de líneas totales y filtradas
  - Botón para limpiar todos los logs

### 🎨 Interfaz
- **Panel horizontal dividido**: Nueva distribución de pantalla
  - Panel superior (60% altura): Lista de logs de dumps HTTP
  - Panel inferior (40% altura): Visor de archivos de log monitoreados
  - Separador redimensionable con límites 30%-70%
  - Cursor row-resize para indicar área de ajuste
- **Gestión mejorada de carpetas**: Interface completa CRUD
  - Botón de editar con ícono de lápiz
  - Modal de edición con campos prellenados
  - Ruta no editable en modo edición (previene inconsistencias)
  - Badge visual que muestra el formato (TEXT/JSON)
  - Selector de formato en modales de agregar/editar
- **Nuevos iconos**: Agregados al sistema de iconos personalizado
  - `file`: Ícono de documento para abrir panel de archivos
  - `edit`: Ícono de lápiz para editar configuraciones
  - `plus`: Ícono + para agregar nuevas carpetas

### 🔧 Mejorado
- **Backend robusto para monitoreo de archivos**:
  - Sistema de FileWatcher con fsnotify para detección de cambios
  - Soporte para múltiples carpetas simultáneas
  - Reinicio automático del watcher al cambiar perfiles activos
  - Manejo eficiente de eventos de archivo (CREATE, WRITE, REMOVE)
  - Lectura incremental de archivos grandes
- **Gestión de configuración expandida**:
  - Nuevas funciones `AddLogFolder` y `UpdateLogFolder` en el backend
  - Persistencia automática en config.yml
  - Validación de rutas y parámetros
  - Campo `Format` añadido a la estructura `LogFolder`
- **Sistema de eventos mejorado**:
  - Evento `log:file:line` para transmitir líneas de log al frontend
  - Evento `log:file:clear` para limpiar logs del archivo actual
  - Sincronización en tiempo real entre backend y frontend

### 🔧 Técnico
- **Estructura de datos mejorada**:
  ```go
  type LogFolder struct {
      Path       string   `yaml:"path" json:"path"`
      Extensions []string `yaml:"extensions" json:"extensions"`
      Filters    []string `yaml:"filters,omitempty" json:"filters,omitempty"`
      Format     string   `yaml:"format,omitempty" json:"format,omitempty"` // "text" or "json"
  }
  ```
- **Funciones de backend con 5 parámetros**:
  - `AddLogFolder(profileName, path, extensions, filters, format string)`
  - `UpdateLogFolder(profileName, path, extensions, filters, format string)`
- **Algoritmo de detección JSON**:
  ```javascript
  const tryParseJson = (line) => {
    try {
      const parsed = JSON.parse(line);
      const formatted = JSON.stringify(parsed, null, 2);
      return { isJson: true, formattedLine: formatted, coloredJson: colorizeJson(formatted) };
    } catch (e) {
      return { isJson: false, formattedLine: line, coloredJson: '' };
    }
  };
  ```
- **Colorización de JSON con regex**:
  - Claves: `/(".*?")\s*:/g`
  - Strings: `/:\s*(".*?")/g`
  - Números: `/:\s*(\d+)/g`
  - Booleanos: `/:\s*(true|false)/g`
  - Null: `/:\s*(null)/g`
- **CSS con :deep() para v-html**: Penetración de estilos en contenido renderizado dinámicamente
- **Wails bindings regenerados**: TypeScript definitions actualizadas con firmas correctas

### 📝 Traducciones
- **Nuevas claves en i18n**:
  - `log_format`, `format_text`, `format_json`, `format_help`
  - `edit_log_folder`, `edit`, `path_cannot_be_changed`
  - `error_updating_folder`, `log_folders`, `add_log_folder`
  - `file_path`, `file_extensions`, `file_filters`
- **Soporte completo** en español e inglés para todas las nuevas funcionalidades

### 🐛 Corregido
- **Botón de archivo invisible**: Agregado ícono `file` faltante al componente Icon.vue
- **Layout vertical en lugar de horizontal**: Cambiado de split izquierda/derecha a arriba/abajo
- **Error "UpdateLogFolder is not a function"**: Bindings de Wails regenerados correctamente
- **Error de firma de función**: Parámetro `format` agregado y bindings actualizados (4 args → 5 args)
- **JSON sin formato**: Implementado sistema completo de detección, formateo y colorización

### 💡 Compatibilidad
- **Retrocompatibilidad**: Campo `format` con valor por defecto "text" para configuraciones existentes
- **Degradación elegante**: JSON inválido se muestra como texto plano sin errores
- **Multi-plataforma**: Monitoreo de archivos funciona en Windows, macOS y Linux
- **Temas adaptativos**: Colores de sintaxis JSON optimizados para modo claro y oscuro

### 🚀 Performance
- **Lectura eficiente de archivos**: Buffer de 4KB para archivos grandes
- **Procesamiento incremental**: Solo se procesan líneas nuevas
- **Regex optimizado**: Colorización sin impacto perceptible en rendimiento
- **Renderizado condicional**: JSON solo se procesa si es detectado como válido

### 🔧 Arquitectura
- **Separación de responsabilidades**:
  - `App.vue`: Layout principal con split panel horizontal
  - `LogFileViewer.vue`: Visualización y formateo de logs de archivo
  - `LogFoldersManager.vue`: Gestión CRUD de carpetas monitoreadas
  - `config.go`: Estructuras de datos y persistencia
  - `app.go`: Lógica de negocio y funciones exportadas a frontend
  - `server.go`: FileWatcher y eventos de archivo en tiempo real

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
