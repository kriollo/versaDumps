<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-slate-800 p-6 rounded-lg shadow-lg w-11/12 md:w-1/2 lg:w-1/3 relative">
      <button @click="closeModal" class="absolute top-3 right-3 text-slate-500 hover:text-slate-700 dark:hover:text-slate-300">
        <Icon name="delete" />
      </button>
      <h2 class="text-xl font-semibold mb-4 text-slate-800 dark:text-slate-200">{{ t('settings') }}</h2>

      <div class="space-y-4">
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

        <!-- Botones de acción -->
        <div class="flex justify-end gap-2 pt-4">
          <button
            @click="closeModal"
            class="px-4 py-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 rounded-md text-slate-800 dark:text-slate-200 transition-colors"
          >
            {{ t('close') }}
          </button>
          <button
            @click="saveSettings"
            class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors"
          >
            {{ t('save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { defineEmits, defineProps, ref, watch } from 'vue';
import * as BackendApp from '../../wailsjs/go/main/App';
import { currentLanguage, setLanguage, t } from '../i18n';
import Icon from './Icon.vue';

const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(['close']);

// Configuración del servidor
const selectedServer = ref('localhost');
const selectedPort = ref(9191);
// Idioma seleccionado (inicialmente el actual)
const selectedLanguage = ref(currentLanguage.value);
// Mostrar tipos (inicialmente false, se actualizará en mounted si backend devuelve valor)
const selectedShowTypes = ref(false);

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
    if (currentConfig && 
        (currentConfig.Server !== selectedServer.value || 
         currentConfig.Port !== parseInt(selectedPort.value))) {
      alert(t('restart_required'));
    }
  } catch (e) {
    console.error('Error saving config:', e);
  }
  
  emit('close', { action: 'saved' });
};

// Emitir evento global al cambiar el toggle para que otros componentes reaccionen sin recargar
watch(selectedShowTypes, (val) => {
  try { localStorage.setItem('show_types', val ? 'true' : 'false'); } catch (e) {}
  try {
    window.dispatchEvent(new CustomEvent('show_types_changed', { detail: { value: val } }));
  } catch (e) {}
});

// Obtener valores iniciales desde backend si está disponible
try {
  BackendApp.GetConfig().then((cfg) => {
    if (cfg) {
      if (cfg.Server) {
        selectedServer.value = cfg.Server;
      }
      if (cfg.Port) {
        selectedPort.value = cfg.Port;
      }
      if (typeof cfg.ShowTypes !== 'undefined') {
        selectedShowTypes.value = !!cfg.ShowTypes;
      }
    }
  }).catch(()=>{});
} catch (e) {}
</script>
