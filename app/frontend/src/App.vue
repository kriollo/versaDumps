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
      </div>
    </div>
    <LineHr />

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
import LogItem from "./components/LogItem.vue";
import UpdateNotification from "./components/UpdateNotification.vue";
import { setLanguage, t } from "./i18n";

// VERSION AND UPDATES
const currentVersion = ref(''); // Se obtiene del backend
const hasUpdate = ref(false);
const newVersion = ref('');
const updateNotificationRef = ref(null);

// TOAST NOTIFICATIONS
const toastMessage = ref('');
const showToast = ref(false);

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
    try {
      const updateInfo = await BackendApp.CheckForUpdates();
      if (updateInfo && updateInfo.available) {
        hasUpdate.value = true;
        newVersion.value = updateInfo.version;
      }
    } catch (e) {
      console.error('Error checking updates:', e);
    }
  };

  // Check after 10 seconds and then every 30 minutes
  setTimeout(checkForUpdates, 10000);
  setInterval(checkForUpdates, 30 * 60 * 1000);

  // Try to get config from backend if available
  try {
    BackendApp.GetConfig().then((cfg) => {
      if (cfg) {
        if (cfg.Theme) {
          theme.value = cfg.Theme
          if (cfg.Theme === "dark") document.documentElement.classList.add("dark");
          else document.documentElement.classList.remove("dark");
          localStorage.setItem("theme", cfg.Theme);
        }
        if (cfg.Lang) {
          // set language in frontend i18n
          setLanguage(cfg.Lang);
        }
      }
    }).catch(()=>{})
  } catch (e) {}

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
  //recargar layout
  if (params.action === 'saved') {
    location.reload();
  }
};

// LOGS
const logs = ref([]);
onMounted(() => {
  EventsOn("newData", (data) => {
    try {
      const parsedData = JSON.parse(data);

      // Procesar la propiedad 'label' si existe
      if (parsedData.label && parsedData.context) {

        if (Array.isArray(parsedData.context)) {
          // Si context es un array, reemplazar el primer elemento
          if (parsedData.context.length > 0) {
            const firstValue = parsedData.context[0];

            // Crear un nuevo objeto donde la clave es el label y el valor es el primer elemento del array
            const newContext = { [parsedData.label]: firstValue };

            // Agregar el resto de elementos del array como claves numéricas empezando desde 1
            for (let i = 1; i < parsedData.context.length; i++) {
              newContext[i.toString()] = parsedData.context[i];
            }

            parsedData.context = newContext;
          }
        } else if (typeof parsedData.context === 'object') {
          const keys = Object.keys(parsedData.context);

          if (keys.length > 0) {
            const firstKey = keys[0];
            const firstValue = parsedData.context[firstKey];



            // Crear un nuevo objeto reemplazando la primera clave por el label
            const newContext = { [parsedData.label]: firstValue };

            // Agregar el resto de propiedades manteniendo sus claves originales
            for (let i = 1; i < keys.length; i++) {
              newContext[keys[i]] = parsedData.context[keys[i]];
            }
            parsedData.context = newContext;

          }
        }

        // Eliminar la propiedad label ya que fue procesada
        delete parsedData.label;
      }

      logs.value.push({ ...parsedData, id: Date.now() });
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
  EventsOn("configLoaded", (cfgJson) => {
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
  if (updateNotificationRef.value) {
    updateNotificationRef.value.checkForUpdates();
  }
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
</style>
