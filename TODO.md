🚀 Roadmap para VersaDumps
Fase 1 – MVP sólido (lo mínimo que te diferencia ya)
-----

Objetivo: reemplazar Laradumps como opción rápida y usable.

# Pinned por label

Implementar ->label("...") con opción pin.

UI: sección fija arriba, mostrando el último mensaje por label.

Botón “historial” para ver todos los dumps con ese label.

Esto ya supera a Laradumps, porque añade persistencia por label.

# Type de mensajes (info, warning, error, success)

Cambiar color de fondo según type.

Si no existe, neutro.

Con esto ya tienes la semántica básica que Ray da (colores → contexto visual rápido).

# Persistencia en SQLite

Guardar todos los mensajes.

Botón “Activar permanencia” en la UI → confirmación.

Crear DB si no existe.



Fase 2 – Nivel Avanzado (profundidad y potencia)
-----

Objetivo: ponerte a la altura de Ray, pero con tu propio sabor.

# Debugging Distribuido (agrupación por servicio)

vd($user)->service("auth").

UI: tabs o filtros por servicio.

Útil para microservicios → nadie más lo hace bien hoy.

# Timers y Benchmarks Integrados

timeStart, timeEnd.

Mostrar resultados como barras, tiempo acumulado, memoria usada.

“Comparación” de benchmarks → varios procesos en paralelo.

Mini Blackfire incluido sin configuración.

# Control de Ejecución Avanzado

->checkpoint() pausa ejecución.

En la UI aparece “Continuar” o “Abortar”.

Técnica: socket bloqueante entre cliente PHP y tu app Go.

Esto es brutal: un pseudo-debugger sin Xdebug → feature premium.

-----

Fase 3 – Diferenciadores (killer features que solo VersaDumps tendría)
-----

Objetivo: convertirte en referente.

# Notificaciones Inteligentes

vd($balance)->warnIf(fn($b) => $b < 0).

Al dispararse → notificación nativa de sistema.

Útil en procesos largos, workers, colas.

# Snapshots de Estado (comparación)

Guardar dumps de objetos/arrays en un punto A y B.

UI para comparar cambios → ideal para debugging de mutaciones.

# Integración con Git

Mostrar commit, branch, archivo y línea del dump.

Botón “abrir en VSCode”.

# Panel Colaborativo (moonshot)

Compartir sesión de dumps en tiempo real con otro dev (via WebSocket relay o WebRTC).

“Mira este bug conmigo” → envías link, otro ve lo mismo.

Esto te convierte en la primera herramienta colaborativa de debugging visual.
