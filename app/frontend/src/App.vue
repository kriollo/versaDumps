<template>
  <main>
    <div
      class="sticky top-0 z-10 flex items-center justify-between p-2.5 bg-slate-100/80 dark:bg-slate-900/80 backdrop-blur-sm shadow-sm"
    >
      <h1 class="text-lg font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-3">
        <span>VersaDumps Visualizer <small v-if="currentVersion">v{{ currentVersion }}</small></span>
        <span v-if="logs.length > 0" class="inline-flex items-center justify-center bg-red-600 text-white text-xs font-semibold rounded-full w-6 h-6">{{ logs.length }}</span>

      </h1>
      <!--Buttons-->
      <div class="flex items-center gap-2">
        <!-- Server status indicator -->
        <div class="flex items-center gap-2 mr-2">
          <span :title="serverStatusTitle" class="w-3 h-3 rounded-full"
            :class="serverStatus === 'online' ? 'bg-emerald-500' : serverStatus === 'offline' ? 'bg-gray-400' : 'bg-yellow-400'">
          </span>
          <span class="text-xs text-slate-600 dark:text-slate-300" v-if="showServerText">{{ serverStatusText }}</span>
        </div>
        <button
          class="icon-button"
          @click="toggleSortOrder"
          :title="sortButtonTitle"
        >
          <Icon name="sort" />
        </button>
        <button
          class="icon-button"
          @click="clearLogs"
          :title="t('clear_all_logs')"
        >
          <Icon name="trash" />
        </button>
        <button
          class="icon-button"
          @click="toggleTheme"
          :title="t('toggle_theme')"
        >
          <Icon :name="theme === 'dark' ? 'sun' : 'moon'" />
        </button>
        <button
          class="icon-button"
          @click="openConfigModal"
          :title="t('settings')"
        >
          <Icon name="gear" />
        </button>
        <!-- View toggle button -->
        <button
          class="icon-button"
          @click="toggleLogFilesPanel"
          :title="showLogFiles ? t('hide_logs') : t('show_logs')"
        >
          <Icon name="file" />
        </button>
      </div>
    </div>
    <LineHr />

    <!-- Split View Container -->
    <div class="split-container" :class="{ 'single-panel': !showLogFiles }">
      <!-- Top Panel: Dumps -->
      <div
        class="panel dumps-panel"
        :style="{ height: showLogFiles ? `${dumpsHeight}%` : '100%' }"
      >
        <div class="p-2.5 space-y-2.5">
          <div v-if="logs.length === 0" class="text-center py-10 text-slate-500">
            <p>{{ t('waiting_data') }}</p>
            <div class="mt-4">
                <div class="mx-auto w-full max-w-2xl h-64 md:h-80 lg:h-96 relative">
                  <img src="./assets/images/versaDumpsVisualizer.webp" alt="versaDumpVisualizer" class="block w-full h-full object-contain" />
                  <img src="./assets/images/texture.svg" alt="" class="pointer-events-none absolute inset-0 w-full h-full object-cover mix-blend-multiply opacity-80" />
                </div>
            </div>
          </div>
          <LogItem
            v-for="log in sortedLogs"
            :key="log.id"
            :log="log"
            @delete="deleteLog(log.id)"
            @copy="copyLogToClipboard"
          />
        </div>
      </div>

      <!-- Resizer -->
      <div
        v-if="showLogFiles"
        class="resizer-horizontal"
        @mousedown="startResize"
      ></div>

      <!-- Bottom Panel: Log Files -->
      <div
        v-if="showLogFiles"
        class="panel logfiles-panel"
        :style="{ height: `${100 - dumpsHeight}%` }"
      >
        <LogFileViewer />
      </div>
    </div>

    <ConfigModal :is-open="isConfigModalOpen" @close="closeConfigModal" @check-updates="handleCheckUpdates" />
    <UpdateNotification ref="updateNotificationRef" />

    <!-- Toast Notification -->
    <Transition name="toast">
      <div
        v-if="showToast"
        class="fixed top-4 right-4 bg-green-600 text-white px-4 py-2 rounded-lg shadow-lg z-50 flex items-center gap-2"
      >
        <Icon name="check" class="text-sm" />
        <span>{{ toastMessage }}</span>
      </div>
    </Transition>

    <!-- Version indicator -->
    <div class="version-indicator">
      <span class="version-text">v{{ currentVersion }}</span>
      <button
        v-if="hasUpdate"
        @click="showUpdateNotification"
        class="update-badge"
        :title="t('update_available')"
      >
        <Icon name="download" class="update-icon-small" />
        <span>{{ t('new_version') }}: v{{ newVersion }}</span>
      </button>
    </div>
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import * as BackendApp from "../wailsjs/go/main/App";
import { EventsOn, WindowIsMinimised } from "../wailsjs/runtime/runtime";
import ConfigModal from "./components/ConfigModal.vue";
import Icon from "./components/Icon.vue";
import LineHr from "./components/LineHr.vue";
import LogFileViewer from "./components/LogFileViewer.vue";
import LogItem from "./components/LogItem.vue";
import UpdateNotification from "./components/UpdateNotification.vue";
import { setLanguage, t } from "./i18n";

// VERSION AND UPDATES
const currentVersion = ref(''); // Se obtiene del backend
const hasUpdate = ref(false);
const newVersion = ref('');
const updateNotificationRef = ref(null);

// SPLIT VIEW STATE
const showLogFiles = ref(false);
const dumpsHeight = ref(60); // Percentage - panel superior ocupará 60% por defecto
const isResizing = ref(false);

// Helper function to compare versions properly
const isNewerVersion = (newVer, currentVer) => {
  if (!newVer || !currentVer) return false;

  // Remove 'v' prefix if present
  const cleanNew = newVer.replace(/^v/, '');
  const cleanCurrent = currentVer.replace(/^v/, '');

  console.log(`Version comparison: new="${cleanNew}" vs current="${cleanCurrent}"`);

  const newParts = cleanNew.split('.').map(n => parseInt(n, 10));
  const currentParts = cleanCurrent.split('.').map(n => parseInt(n, 10));

  // Ensure arrays have same length
  while (newParts.length < 3) newParts.push(0);
  while (currentParts.length < 3) currentParts.push(0);

  for (let i = 0; i < 3; i++) {
    if (newParts[i] > currentParts[i]) {
      console.log(`Version check: ${cleanNew} is newer than ${cleanCurrent}`);
      return true;
    }
    if (newParts[i] < currentParts[i]) {
      console.log(`Version check: ${cleanNew} is older than ${cleanCurrent}`);
      return false;
    }
  }

  console.log(`Version check: ${cleanNew} is same as ${cleanCurrent}`);
  return false;
};

// TOAST NOTIFICATIONS
const toastMessage = ref('');
const showToast = ref(false);

// SERVER STATUS
const serverStatus = ref('unknown'); // 'online' | 'offline' | 'unknown'
const serverHost = ref('');
const serverPort = ref(0);
const showServerText = ref(true);
const configLoaded = ref(false); // Track if config has been loaded

const serverStatusText = computed(() => {
  if (serverStatus.value === 'online') return t.value ? (t.value('server_online') || 'En línea') : 'En línea';
  if (serverStatus.value === 'offline') return t.value ? (t.value('server_offline') || 'Desconectado') : 'Desconectado';
  return t.value ? (t.value('server_unknown') || 'Desconocido') : 'Desconocido';
});

const serverStatusTitle = computed(() => `Server: ${serverHost.value}:${serverPort.value}`);

let healthInterval = null;

const checkServerHealth = async () => {
  if (serverHost.value === '' || serverPort.value === 0) {
    serverStatus.value = 'unknown';
    return;
  }
  const url = `http://${serverHost.value}:${serverPort.value}/health`;
  try {
    const resp = await fetch(url, { method: 'GET', cache: 'no-cache' });
    if (resp.ok) {
      // Optionally verify JSON
      serverStatus.value = 'online';
    } else {
      serverStatus.value = 'offline';
    }
  } catch (e) {
    serverStatus.value = 'offline';
  }
};

// Centralized function to start health polling
const startHealthPolling = () => {
  if (serverHost.value === '' || serverPort.value === 0) {
    clearInterval(healthInterval);
    healthInterval = null;
  }
  console.log(`Starting health polling for ${serverHost.value}:${serverPort.value}`);
  // Clear any existing interval
  if (healthInterval) {
    clearInterval(healthInterval);
    healthInterval = null;
  }

  // Start new polling
  checkServerHealth();
  healthInterval = setInterval(checkServerHealth, 5000);
};


const showToastMessage = (message) => {
  toastMessage.value = message;
  showToast.value = true;

  // Auto-hide after 3 seconds
  setTimeout(() => {
    showToast.value = false;
  }, 3000);
};

// THEME
const theme = ref("dark");
const toggleTheme = () => {
  theme.value = theme.value === "dark" ? "light" : "dark";
  if (theme.value === "dark") {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
  localStorage.setItem("theme", theme.value);
};
onMounted(async () => {
  const savedTheme = localStorage.getItem("theme") || "dark";
  theme.value = savedTheme;
  if (savedTheme === "dark") {
    document.documentElement.classList.add("dark");
  }

  // Get current version from backend
  try {
    const version = await BackendApp.GetCurrentVersion();
    currentVersion.value = version;
  } catch (e) {
    console.error('Error getting version:', e);
  }

  // Check for updates periodically
  const checkForUpdates = async () => {
    console.log('=== CHECKING FOR UPDATES ===');
    console.log('Current version:', currentVersion.value);

    try {
      const updateInfo = await BackendApp.CheckForUpdates();
      console.log('Update info received:', updateInfo);

      if (updateInfo && updateInfo.available) {
        // Double-check with our own version comparison
        const isActuallyNewer = isNewerVersion(updateInfo.version, currentVersion.value);
        console.log('Backend says update available, our check says:', isActuallyNewer);

        if (isActuallyNewer) {
          hasUpdate.value = true;
          newVersion.value = updateInfo.version;
          console.log('✅ Update confirmed: Setting hasUpdate=true, newVersion=', updateInfo.version);
        } else {
          // Backend says there's an update, but version comparison disagrees
          console.log('⚠️ Backend reported update available, but version check disagrees. Ignoring.');
          hasUpdate.value = false;
          newVersion.value = '';
        }
      } else {
        // Clear update state when no update is available
        console.log('❌ No update available according to backend');
        hasUpdate.value = false;
        newVersion.value = '';
      }
    } catch (e) {
      console.error('Error checking updates:', e);
      // Clear update state on error
      hasUpdate.value = false;
      newVersion.value = '';
    }

    console.log('Final state: hasUpdate=', hasUpdate.value, 'newVersion=', newVersion.value);
  };

  // Check after 10 seconds and then every 30 minutes
  setTimeout(checkForUpdates, 10000);
  setInterval(checkForUpdates, 30 * 60 * 1000);

  // Try to get config from backend if available. Start polling only after config is confirmed.
  try {
    BackendApp.GetConfig().then((cfg) => {
      if (cfg) {
        // If config includes host/port for the server, pick them up
        // Support both capitalized (old) and lowercase (new with JSON tags) keys
        if (cfg.server || cfg.Server) serverHost.value = cfg.server || cfg.Server;
        if (cfg.port || cfg.Port) serverPort.value = cfg.port || cfg.Port;

        console.log('GetConfig result:', { host: serverHost.value, port: serverPort.value });

        if (cfg.theme || cfg.Theme) {
          theme.value = cfg.theme || cfg.Theme
          if ((cfg.theme || cfg.Theme) === "dark") document.documentElement.classList.add("dark");
          else document.documentElement.classList.remove("dark");
          localStorage.setItem("theme", cfg.theme || cfg.Theme);
        }
        if (cfg.language || cfg.lang || cfg.Lang) {
          // set language in frontend i18n
          setLanguage(cfg.language || cfg.lang || cfg.Lang);
        }

        // Mark config as loaded and start polling
        configLoaded.value = true;
        startHealthPolling();
      }
    }).catch(()=>{
      // If GetConfig fails, use defaults and start polling anyway
      console.log('GetConfig failed, using default values and starting polling');
      configLoaded.value = true;
    })
  } catch (e) {
    // If GetConfig is not available, use defaults and start polling
    console.log('GetConfig not available, using default values and starting polling');
    configLoaded.value = true;
  }

    // Ask backend for current visible count and sync UI/title
  try {
      if (BackendApp.GetVisibleCount) {
        BackendApp.GetVisibleCount().then((cnt) => {
          console.log('Backend counter at startup:', cnt);
          // If backend has a count but frontend is empty, reset backend to match frontend
          if (cnt > 0 && logs.value.length === 0) {
            console.log('Resetting backend counter from', cnt, 'to 0 to match empty frontend');
            try { BackendApp.UpdateVisibleCount(0); } catch (e) {}
          } else if (cnt !== logs.value.length) {
            console.log('Syncing backend counter from', cnt, 'to', logs.value.length);
            try { BackendApp.UpdateVisibleCount(logs.value.length); } catch (e) {}
          }
        }).catch(()=>{});
      }
  } catch (e) {}
});

// CONFIG MODAL
const isConfigModalOpen = ref(false);
const openConfigModal = () => {
  isConfigModalOpen.value = true;
};
const closeConfigModal = (params) => {
  isConfigModalOpen.value = false;

  if (params.action === 'saved') {
    console.log('=== CONFIG SAVED - RELOADING INTERFACE ===');

    // Force reload the entire configuration and apply all changes
    try {
      BackendApp.GetConfig().then((cfg) => {
        if (cfg) {
          console.log('New config loaded:', cfg);

          // Update server configuration
          const oldHost = serverHost.value;
          const oldPort = serverPort.value;

          if (cfg.Server) serverHost.value = cfg.Server;
          if (cfg.Port) serverPort.value = cfg.Port;

          // Update theme
          if (cfg.Theme) {
            theme.value = cfg.Theme;
            if (cfg.Theme === "dark") {
              document.documentElement.classList.add("dark");
            } else {
              document.documentElement.classList.remove("dark");
            }
            localStorage.setItem("theme", cfg.Theme);
            console.log('Theme updated to:', cfg.Theme);
          }

          // Update language
          if (cfg.Lang) {
            setLanguage(cfg.Lang);
            console.log('Language updated to:', cfg.Lang);
          }

          // Restart server polling if host/port changed
          if (oldHost !== serverHost.value || oldPort !== serverPort.value) {
            console.log(`Server config changed: ${oldHost}:${oldPort} -> ${serverHost.value}:${serverPort.value}`);
            startHealthPolling();
          }

          // Show success message
          showToastMessage(t.value('settings_saved') || 'Configuración guardada');

          console.log('✅ Interface reloaded successfully');
        }
      }).catch((e) => {
        console.error('Error reloading config:', e);
        showToastMessage('Error al cargar la configuración');
      });
    } catch (e) {
      console.error('Error in closeConfigModal:', e);
    }
  }
};

// LOGS
const logs = ref([]);
onMounted(() => {
  EventsOn("newData", (data) => {
    try {
      const parsedData = JSON.parse(data);

      // Normalizar estructura para versadumps-php 2.2.0
      // La nueva versión envía: { context: { variables, line, file }, frame: {...}, label, metadata: { color, includeTrace, max_depth } }
      let normalizedData = { ...parsedData };

      // Si tiene metadata (versión 2.2.0+), procesar
      if (parsedData.metadata) {
        // Extraer color del metadata
        if (parsedData.metadata.color) {
          normalizedData.color = parsedData.metadata.color;
        }

        // Extraer trace del metadata (la librería PHP envía metadata.trace)
        if (parsedData.metadata.trace) {
          normalizedData.trace = parsedData.metadata.trace;
        }

        // Extraer max_depth del metadata
        if (parsedData.metadata.max_depth) {
          normalizedData.max_depth = parsedData.metadata.max_depth;
        }
      }

      // Si color viene directamente (sin metadata), usarlo
      if (!normalizedData.color && parsedData.color) {
        normalizedData.color = parsedData.color;
      }

      // Si trace viene directamente (sin metadata), usarlo
      if (!normalizedData.trace && parsedData.trace) {
        normalizedData.trace = parsedData.trace;
      }

      // Buscar trace/stack en frame o en el payload raíz
      if (!normalizedData.trace) {
        if (parsedData.frame && parsedData.frame.trace) {
          normalizedData.trace = parsedData.frame.trace;
        } else if (parsedData.stack) {
          normalizedData.trace = parsedData.stack;
        } else if (parsedData.backtrace) {
          normalizedData.trace = parsedData.backtrace;
        }
      }

      console.log('📦 Payload recibido:', parsedData);
      console.log('🔄 Datos normalizados:', normalizedData);      // Procesar context.variables si existe (versión 2.2.0+)
      if (parsedData.context && parsedData.context.variables) {
        // La nueva estructura tiene variables dentro de context
        normalizedData.context = parsedData.context.variables;

        // Preservar file y line del context
        if (parsedData.context.file && !normalizedData.frame) {
          normalizedData.frame = {
            file: parsedData.context.file,
            line: parsedData.context.line || 0,
            function: parsedData.frame?.caller || parsedData.frame?.function || ''
          };
        }
      }

      // Procesar la propiedad 'label' si existe
      if (normalizedData.label && normalizedData.context) {

        if (Array.isArray(normalizedData.context)) {
          // Si context es un array, reemplazar el primer elemento
          if (normalizedData.context.length > 0) {
            const firstValue = normalizedData.context[0];

            // Crear un nuevo objeto donde la clave es el label y el valor es el primer elemento del array
            const newContext = { [normalizedData.label]: firstValue };

            // Agregar el resto de elementos del array como claves numéricas empezando desde 1
            for (let i = 1; i < normalizedData.context.length; i++) {
              newContext[i.toString()] = normalizedData.context[i];
            }

            normalizedData.context = newContext;
          }
        } else if (typeof normalizedData.context === 'object') {
          const keys = Object.keys(normalizedData.context);

          if (keys.length > 0) {
            const firstKey = keys[0];
            const firstValue = normalizedData.context[firstKey];

            // Crear un nuevo objeto reemplazando la primera clave por el label
            const newContext = { [normalizedData.label]: firstValue };

            // Agregar el resto de propiedades manteniendo sus claves originales
            for (let i = 1; i < keys.length; i++) {
              newContext[keys[i]] = normalizedData.context[keys[i]];
            }
            normalizedData.context = newContext;

          }
        }

        // Eliminar la propiedad label ya que fue procesada
        delete normalizedData.label;
      }

      // NO eliminar metadata - dejarlo para debugging y procesamiento posterior
      // El trace está dentro de metadata.trace y necesitamos preservarlo
      /*
      if (normalizedData.metadata) {
        delete normalizedData.metadata;
      }
      */

      logs.value.push({ ...normalizedData, id: Date.now() });
      console.log('=== NEW LOG ADDED ===');
      console.log('Frontend count after add:', logs.value.length);

      try {
        BackendApp.UpdateVisibleCount(logs.value.length);
        console.log('Sent to backend counter:', logs.value.length);

        // Verify it was set correctly
        BackendApp.GetVisibleCount().then((cnt) => {
          console.log('Verified backend counter after add:', cnt);
          if (cnt !== logs.value.length) {
            console.error('❌ MISMATCH! Frontend:', logs.value.length, 'Backend:', cnt);
          } else {
            console.log('✅ Counters in sync!');
          }
        }).catch((e) => {
          console.error('Error verifying backend counter:', e);
        });
      } catch (e) {
        console.error('Error updating backend counter:', e);
      }
      // Mostrar notificación si la ventana está minimizada
      try {
        const isMin = WindowIsMinimised();
        if (isMin && typeof Notification !== 'undefined' && Notification.permission === 'granted') {
          const preview = typeof parsedData.context === 'string' ? parsedData.context : JSON.stringify(parsedData.context);
          new Notification('VersaDumps', { body: preview.slice ? preview.slice(0, 200) : String(preview) });
        }
      } catch (e) {}
    } catch (e) {
      logs.value.push({
        id: Date.now(),
        frame: { file: "Error", line: 0, function: "Invalid Data" },
        context: data,
      });
      try { BackendApp.UpdateVisibleCount(logs.value.length); } catch (e) {}
    }
  });

  // Listen for config sent on startup
  EventsOn("configLoaded", async (cfgJson) => {
    try {
      const cfg = JSON.parse(cfgJson);
      if (cfg.theme) {
        theme.value = cfg.theme;
        if (cfg.theme === 'dark') document.documentElement.classList.add('dark');
        else document.documentElement.classList.remove('dark');
        localStorage.setItem('theme', cfg.theme);
      }

      if (cfg.language || cfg.lang) {
        const lang = cfg.language || cfg.lang;
        setLanguage(lang);
      }

      // Support both capitalized and lowercase keys from Go config
      if (cfg.server || cfg.Server) serverHost.value = cfg.server || cfg.Server;
      if (cfg.port || cfg.Port) serverPort.value = cfg.port || cfg.Port;

      console.log('Config loaded:', { host: serverHost.value, port: serverPort.value });

      // Mark config as loaded and restart polling with updated config
      configLoaded.value = true;
      startHealthPolling();

    } catch (e) {
      // ignore
    }
  })
});

const deleteLog = (id) => {
  logs.value = logs.value.filter((log) => log.id !== id);
  console.log('Deleted log, new count:', logs.value.length);
  try {
    BackendApp.UpdateVisibleCount(logs.value.length);
    console.log('Updated backend counter to:', logs.value.length);
  } catch (e) {
    console.error('Error updating backend counter:', e);
  }
};

const clearLogs = () => {
  console.log('=== CLEARING LOGS ===');
  console.log('Before clear - logs.length:', logs.value.length);

  logs.value = [];
  console.log('After clear - logs.length:', logs.value.length);

  // Ensure backend counter is properly reset
  try {
    BackendApp.UpdateVisibleCount(0);
    console.log('Backend counter reset to 0');

    // Double-check by getting the count back
    BackendApp.GetVisibleCount().then((cnt) => {
      console.log('Verified backend counter after reset:', cnt);
    }).catch((e) => {
      console.error('Error verifying backend counter:', e);
    });
  } catch (e) {
    console.error('Error resetting backend counter:', e);
  }
};

// COPY TO CLIPBOARD
const copyLogToClipboard = async (log) => {
  try {
    // Remove the internal id before copying
    const logCopy = { ...log };
    delete logCopy.id;

    await navigator.clipboard.writeText(JSON.stringify(logCopy, null, 2));
    showToastMessage(t.value('copied_to_clipboard') || 'Copiado al portapapeles');
  } catch (error) {
    console.error('Error copying to clipboard:', error);
    showToastMessage('Error al copiar al portapapeles');
  }
};

// SORTING
const sortOrder = ref("desc"); // 'desc' for newest first, 'asc' for oldest first
const toggleSortOrder = () => {
  sortOrder.value = sortOrder.value === "desc" ? "asc" : "desc";
};
const sortButtonTitle = computed(() => {
  return sortOrder.value === "desc" ? t.value('sort_newest') : t.value('sort_oldest');
});

const sortedLogs = computed(() => {
  return [...logs.value].sort((a, b) => {
    if (sortOrder.value === "desc") {
      return b.id - a.id;
    } else {
      return a.id - b.id;
    }
  });
});

// Show update notification manually
const showUpdateNotification = () => {
  if (updateNotificationRef.value) {
    updateNotificationRef.value.checkForUpdates();
  }
};

// Handle check updates from ConfigModal
const handleCheckUpdates = () => {
  console.log('Manual update check triggered from ConfigModal');
  if (updateNotificationRef.value) {
    updateNotificationRef.value.checkForUpdates();
  }
};

// Emergency function to clear update state (for debugging)
const clearUpdateState = () => {
  console.log('🚨 EMERGENCY: Clearing update state manually');
  hasUpdate.value = false;
  newVersion.value = '';
  localStorage.removeItem('versadumps_update_state'); // Clear any potential stored state
  console.log('Update state cleared');
};

// ========================================
// SPLIT VIEW FUNCTIONS
// ========================================

const toggleLogFilesPanel = () => {
  showLogFiles.value = !showLogFiles.value;

  // Load log folders if opening for the first time
  if (showLogFiles.value) {
    loadLogFolders();
  }
};

const loadLogFolders = async () => {
  try {
    const folders = await BackendApp.GetLogFolders();
    console.log('Log folders loaded:', folders);
  } catch (e) {
    console.error('Error loading log folders:', e);
  }
};

const startResize = (e) => {
  isResizing.value = true;
  document.addEventListener('mousemove', handleResize);
  document.addEventListener('mouseup', stopResize);
  e.preventDefault();
};

const handleResize = (e) => {
  if (!isResizing.value) return;

  const container = document.querySelector('.split-container');
  if (!container) return;

  const containerRect = container.getBoundingClientRect();
  const newHeight = ((e.clientY - containerRect.top) / containerRect.height) * 100;

  // Limit between 30% and 70%
  if (newHeight >= 30 && newHeight <= 70) {
    dumpsHeight.value = newHeight;
  }
};

const stopResize = () => {
  isResizing.value = false;
  document.removeEventListener('mousemove', handleResize);
  document.removeEventListener('mouseup', stopResize);
};
</script>

<style>
.icon-button {
  @apply p-1 rounded-full hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200;
}

.version-indicator {
  position: fixed;
  bottom: 10px;
  left: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  z-index: 100;
}

.version-text {
  font-size: 12px;
  color: #9ca3af;
  font-family: 'Courier New', monospace;
  padding: 4px 8px;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.dark .version-text {
  color: #6b7280;
  background: rgba(255, 255, 255, 0.05);
}

.update-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: linear-gradient(135deg, #397111 0%, #154bb1 100%);
  color: white;
  border: none;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  animation: pulse 2s infinite;
}

.update-badge:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.update-icon-small {
  font-size: 14px;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0.7);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(102, 126, 234, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0);
  }
}

/* Toast Animation */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100px);
}

/* Split View Styles - Horizontal Layout */
.split-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 60px); /* Adjust based on header height */
  overflow: hidden;
}

.split-container.single-panel {
  display: block;
}

.panel {
  overflow-y: auto;
  overflow-x: hidden;
  transition: height 0.3s ease;
}

.dumps-panel {
  min-height: 200px;
}

.logfiles-panel {
  min-height: 200px;
  border-top: 1px solid #cbd5e1;
}

.dark .logfiles-panel {
  border-top-color: #475569;
}

.resizer-horizontal {
  height: 8px;
  background: #e2e8f0;
  cursor: row-resize;
  flex-shrink: 0;
  position: relative;
  transition: background 0.2s;
}

.resizer-horizontal:hover {
  background: #94a3b8;
}

.dark .resizer-horizontal {
  background: #475569;
}

.dark .resizer-horizontal:hover {
  background: #64748b;
}

.resizer-horizontal::before {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 2px;
  background: #cbd5e1;
  border-radius: 1px;
}

.dark .resizer-horizontal::before {
  background: #64748b;
}
</style>
