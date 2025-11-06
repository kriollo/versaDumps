<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-slate-800 p-6 rounded-lg shadow-lg w-11/12 md:w-2/3 lg:w-3/5 xl:w-1/2 relative max-h-[90vh] overflow-y-auto">
      <button @click="closeModal" class="absolute top-3 right-3 text-slate-500 hover:text-slate-700 dark:hover:text-slate-300 z-10">
        <Icon name="delete" />
      </button>
      <h2 class="text-xl font-semibold mb-4 text-slate-800 dark:text-slate-200">{{ t('settings') }}</h2>

      <!-- Tabs Navigation -->
      <div class="flex border-b border-slate-200 dark:border-slate-700 mb-4">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          class="px-4 py-2 text-sm font-medium transition-colors"
          :class="activeTab === tab.id
            ? 'border-b-2 border-blue-600 text-blue-600 dark:text-blue-400'
            : 'text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200'"
        >
          {{ t(tab.label) }}
        </button>
      </div>

      <!-- Tab Content -->
      <div class="space-y-4">
        <!-- General Settings Tab -->
        <div v-show="activeTab === 'general'">
          <!-- Server configuration -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">{{ t('server') }}</label>
            <input
              v-model="selectedServer"
              type="text"
              :placeholder="t('server_placeholder')"
              class="w-full px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:text-white placeholder-slate-400 dark:placeholder-slate-500"
            />
          </div>

          <!-- Port configuration -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">{{ t('port') }}</label>
            <input
              v-model="selectedPort"
              type="number"
              :placeholder="t('port_placeholder')"
              min="1"
              max="65535"
              class="w-full px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:text-white placeholder-slate-400 dark:placeholder-slate-500"
            />
          </div>

          <!-- Selección de idioma -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">{{ t('language') }}</label>
            <select
              v-model="selectedLanguage"
              class="w-full px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:text-white"
            >
              <option value="en">{{ t('english') }}</option>
              <option value="es">{{ t('spanish') }}</option>
            </select>
          </div>

          <!-- Mostrar tipos de variable (toggle moderno) -->
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <label for="showTypes" class="text-sm text-slate-700 dark:text-slate-300">{{ t('show_variable_types') }}</label>
              <p class="text-xs text-slate-400">{{ selectedShowTypes ? 'On' : 'Off' }}</p>
            </div>
            <button
              id="showTypes"
              @click="selectedShowTypes = !selectedShowTypes"
              :aria-pressed="selectedShowTypes.toString()"
              class="relative inline-flex items-center h-6 rounded-full w-11 transition-colors focus:outline-none"
              :class="selectedShowTypes ? 'bg-blue-600' : 'bg-slate-300 dark:bg-slate-600'"
            >
              <span
                class="inline-block w-4 h-4 bg-white rounded-full transform transition-transform"
                :style="{ transform: selectedShowTypes ? 'translateX(20px)' : 'translateX(2px)' }"
              ></span>
            </button>
          </div>

          <!-- Sección de actualizaciones -->
          <div class="mt-4 pt-4 border-t border-slate-200 dark:border-slate-700">
            <button @click="checkUpdates" class="w-full px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md transition-colors flex items-center justify-center gap-2">
              <Icon name="download" class="w-4 h-4" />
              {{ t('check_for_updates') }}
            </button>
          </div>
        </div>

        <!-- Profiles Tab -->
        <div v-show="activeTab === 'profiles'">
          <!-- Profile Selector -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">{{ t('active_profile') }}</label>
            <select
              v-model="selectedProfile"
              @change="onProfileChange"
              class="w-full px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:text-white"
            >
              <option v-for="profile in profiles" :key="profile.name || profile.Name" :value="profile.name || profile.Name">
                {{ profile.name || profile.Name }}
              </option>
            </select>
          </div>

          <!-- Profile Actions -->
          <div class="flex gap-2 mt-4">
            <button
              @click="showCreateProfileModal = true"
              class="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors flex items-center justify-center gap-2"
            >
              <Icon name="add" class="w-4 h-4" />
              {{ t('create_profile') }}
            </button>
            <button
              @click="confirmDeleteProfile"
              :disabled="profiles.length <= 1"
              class="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors disabled:bg-slate-400 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              <Icon name="delete" class="w-4 h-4" />
              {{ t('delete_profile') }}
            </button>
          </div>

          <!-- Current Profile Info -->
          <div class="mt-4 p-4 bg-slate-100 dark:bg-slate-700 rounded-md">
            <h3 class="font-medium text-slate-800 dark:text-slate-200 mb-2">{{ t('profile_settings') }}</h3>
            <div class="space-y-2 text-sm text-slate-600 dark:text-slate-400">
              <p><span class="font-medium">{{ t('server') }}:</span> {{ currentProfile?.server || currentProfile?.Server || 'localhost' }}</p>
              <p><span class="font-medium">{{ t('port') }}:</span> {{ currentProfile?.port || currentProfile?.Port || 9191 }}</p>
              <p><span class="font-medium">{{ t('language') }}:</span> {{ currentProfile?.language || currentProfile?.Language || 'es' }}</p>
              <p><span class="font-medium">{{ t('log_folders') }}:</span> {{ (currentProfile?.log_folders || currentProfile?.LogFolders)?.length || 0 }}</p>
            </div>
          </div>
        </div>

        <!-- Log Folders Tab -->
        <div v-show="activeTab === 'logfolders'">
          <div v-if="!selectedProfile" class="text-center text-slate-500 dark:text-slate-400 py-10">
            {{ t('loading') }}...
          </div>
          <LogFoldersManager
            v-else
            :profile-name="selectedProfile"
            :folders="currentProfile?.log_folders || currentProfile?.LogFolders || []"
            @update="loadProfiles"
            @error="handleLogFolderError"
          />
        </div>

        <!-- Botones de acción (always visible) -->
        <div class="flex justify-end gap-2 pt-4 border-t border-slate-200 dark:border-slate-700">
          <button
            @click="closeModal"
            class="px-4 py-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 rounded-md text-slate-800 dark:text-slate-200 transition-colors"
          >
            {{ t('close') }}
          </button>
          <button
            v-if="activeTab === 'general'"
            @click="saveSettings"
            class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors"
          >
            {{ t('save') }}
          </button>
        </div>
      </div>

      <!-- Create Profile Modal -->
      <div v-if="showCreateProfileModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div class="bg-white dark:bg-slate-800 p-6 rounded-lg shadow-lg w-11/12 md:w-1/2 lg:w-1/3">
          <h3 class="text-lg font-semibold mb-4 text-slate-800 dark:text-slate-200">{{ t('create_profile') }}</h3>
          <div class="space-y-3">
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">{{ t('profile_name') }}</label>
              <input
                v-model="newProfileName"
                type="text"
                :placeholder="t('profile_name')"
                class="w-full px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:text-white"
              />
            </div>
            <div class="flex items-center gap-2">
              <input
                v-model="copyCurrentSettings"
                type="checkbox"
                id="copySettings"
                class="rounded"
              />
              <label for="copySettings" class="text-sm text-slate-700 dark:text-slate-300">{{ t('copy_current_settings') }}</label>
            </div>
          </div>
          <div class="flex justify-end gap-2 mt-4">
            <button
              @click="showCreateProfileModal = false; newProfileName = ''"
              class="px-4 py-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 rounded-md text-slate-800 dark:text-slate-200 transition-colors"
            >
              {{ t('cancel') }}
            </button>
            <button
              @click="createProfile"
              :disabled="!newProfileName.trim()"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors disabled:bg-slate-400 disabled:cursor-not-allowed"
            >
              {{ t('create') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, defineEmits, defineProps, onMounted, ref, watch } from 'vue';
import * as BackendApp from '../../wailsjs/go/main/App';
import { currentLanguage, setLanguage, t } from '../i18n';
import Icon from './Icon.vue';
import LogFoldersManager from './LogFoldersManager.vue';

const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(['close', 'check-updates']);

// Tabs
const activeTab = ref('general');
const tabs = [
  { id: 'general', label: 'general_settings' },
  { id: 'profiles', label: 'profiles' },
  { id: 'logfolders', label: 'log_folders' }
];

// Configuración del servidor (from active profile)
const selectedServer = ref('localhost');
const selectedPort = ref(9191);
// Idioma seleccionado (inicialmente el actual)
const selectedLanguage = ref(currentLanguage.value);
// Mostrar tipos (inicialmente false, se actualizará en mounted si backend devuelve valor)
const selectedShowTypes = ref(false);

// Profile Management
const profiles = ref([]);
const selectedProfile = ref('');
const currentProfile = computed(() => {
  return profiles.value.find(p => (p.name || p.Name) === selectedProfile.value);
});
const showCreateProfileModal = ref(false);
const newProfileName = ref('');
const copyCurrentSettings = ref(true);

// Load profiles from backend
const loadProfiles = async () => {
  try {
    console.log('Loading profiles...');
    const profileList = await BackendApp.ListProfiles();
    profiles.value = profileList || [];
    console.log('Profiles loaded:', profiles.value.length);
    if (profiles.value.length > 0) {
      console.log('First profile:', profiles.value[0]);
    }

    // Get active profile from config
    const config = await BackendApp.GetConfig();
    const activeProfileName = config?.active_profile || config?.ActiveProfile;

    if (activeProfileName) {
      selectedProfile.value = activeProfileName;
      console.log('Active profile:', activeProfileName);
    } else if (profiles.value.length > 0) {
      // If no active profile, select the first one
      selectedProfile.value = profiles.value[0].name || profiles.value[0].Name;
      console.log('No active profile, selecting first:', selectedProfile.value);
    }

    // Load active profile settings into form
    if (selectedProfile.value) {
      const active = profiles.value.find(p => (p.name || p.Name) === selectedProfile.value);
      if (active) {
        selectedServer.value = active.server || active.Server || 'localhost';
        selectedPort.value = active.port || active.Port || 9191;
        selectedLanguage.value = active.language || active.Language || 'es';
        selectedShowTypes.value = active.show_types !== undefined ? active.show_types : (active.ShowTypes || false);
        const logFoldersCount = (active.log_folders || active.LogFolders)?.length || 0;
        console.log('Active profile log folders:', logFoldersCount);
      }
    }
  } catch (e) {
    console.error('Error loading profiles:', e);
  }
};

// Create new profile
const createProfile = async () => {
  if (!newProfileName.value.trim()) return;

  try {
    let server = 'localhost';
    let port = 9191;
    let language = 'es';
    let showTypes = false;

    if (copyCurrentSettings.value && currentProfile.value) {
      server = currentProfile.value.server || currentProfile.value.Server || 'localhost';
      port = currentProfile.value.port || currentProfile.value.Port || 9191;
      language = currentProfile.value.language || currentProfile.value.Language || 'es';
      showTypes = currentProfile.value.show_types !== undefined ? currentProfile.value.show_types : (currentProfile.value.ShowTypes || false);
    }

    await BackendApp.CreateProfile(
      newProfileName.value.trim(),
      server,
      port,
      '',
      language,
      showTypes
    );

    await loadProfiles();
    showCreateProfileModal.value = false;
    newProfileName.value = '';
    copyCurrentSettings.value = true;
  } catch (e) {
    console.error('Error creating profile:', e);
    alert(t.value('error') + ': ' + e);
  }
};

// Delete current profile
const confirmDeleteProfile = async () => {
  if (profiles.value.length <= 1) {
    alert(t.value('cannot_delete_last_profile'));
    return;
  }

  if (!confirm(t.value('confirm_delete_profile').replace('{name}', selectedProfile.value))) {
    return;
  }

  try {
    await BackendApp.DeleteProfile(selectedProfile.value);
    await loadProfiles();

    // Select first available profile
    if (profiles.value.length > 0) {
      selectedProfile.value = profiles.value[0].name || profiles.value[0].Name;
      await onProfileChange();
    }
  } catch (e) {
    console.error('Error deleting profile:', e);
    alert(t.value('error') + ': ' + e);
  }
};

// Switch to different profile
const onProfileChange = async () => {
  try {
    await BackendApp.SwitchProfile(selectedProfile.value);
    await loadProfiles();
    const message = `${t.value('profile_switched')}: ${selectedProfile.value}`;
    alert(message);
  } catch (err) {
    console.error('Error switching profile:', err);
    const errorMsg = `${t.value('error')}: ${err}`;
    alert(errorMsg);
  }
};

// Handle log folder errors
const handleLogFolderError = (error) => {
  alert(t.value('error') + ': ' + error);
};

const closeModal = () => {
  // Restaurar el idioma seleccionado al actual si se cierra sin guardar
  selectedLanguage.value = currentLanguage.value;
  emit('close', { action: 'closed'});
};

const saveSettings = async () => {
  // Guardar el idioma seleccionado
  setLanguage(selectedLanguage.value);

  // Obtener configuración actual antes de guardar para detectar cambios
  let currentConfig = null;
  try {
    currentConfig = await BackendApp.GetConfig();
  } catch (e) {}

  // Preparar objeto de configuración
  const config = {
    server: selectedServer.value,
    port: parseInt(selectedPort.value),
    language: selectedLanguage.value,
    show_types: selectedShowTypes.value
  };

  // Persistir en config.yml via backend
  try {
    await BackendApp.SaveFrontendConfig(config);
    // Also persist locally for immediate UI use
    try {
      localStorage.setItem('show_types', selectedShowTypes.value ? 'true' : 'false');
    } catch (e) {}

    // Mostrar mensaje de reinicio requerido si cambió servidor o puerto
    const oldServer = currentConfig?.server || currentConfig?.Server;
    const oldPort = currentConfig?.port || currentConfig?.Port;
    if (currentConfig &&
        (oldServer !== selectedServer.value ||
         oldPort !== parseInt(selectedPort.value))) {
      alert(t.value('restart_required'));
    }
  } catch (e) {
    console.error('Error saving config:', e);
  }

  emit('close', { action: 'saved' });
};

const checkUpdates = () => {
  console.log('ConfigModal: checkUpdates called');
  emit('check-updates');
};

// Emitir evento global al cambiar el toggle para que otros componentes reaccionen sin recargar
watch(selectedShowTypes, (val) => {
  try { localStorage.setItem('show_types', val ? 'true' : 'false'); } catch (e) {}
  try {
    window.dispatchEvent(new CustomEvent('show_types_changed', { detail: { value: val } }));
  } catch (e) {}
});

// Watch for modal open to load profiles
watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    loadProfiles();
  }
});

// Obtener valores iniciales desde backend si está disponible
try {
  BackendApp.GetConfig().then((cfg) => {
    if (cfg) {
      if (cfg.server || cfg.Server) {
        selectedServer.value = cfg.server || cfg.Server;
      }
      if (cfg.port || cfg.Port) {
        selectedPort.value = cfg.port || cfg.Port;
      }
      if (cfg.show_types !== undefined) {
        selectedShowTypes.value = !!cfg.show_types;
      } else if (typeof cfg.ShowTypes !== 'undefined') {
        selectedShowTypes.value = !!cfg.ShowTypes;
      }
    }
  }).catch(()=>{});
} catch (e) {}

// Load profiles on mount
onMounted(async () => {
  await loadProfiles();
});
</script>
