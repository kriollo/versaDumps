<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-slate-800 p-6 rounded-lg shadow-lg w-11/12 md:w-1/2 lg:w-1/3 relative">
      <button @click="closeModal" class="absolute top-3 right-3 text-slate-500 hover:text-slate-700 dark:hover:text-slate-300">
        <Icon name="delete" />
      </button>
      <h2 class="text-xl font-semibold mb-4 text-slate-800 dark:text-slate-200">{{ t('settings') }}</h2>

      <div class="space-y-4">
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
import { defineProps, defineEmits, ref } from 'vue';
import Icon from './Icon.vue';
import { currentLanguage, setLanguage, t } from '../i18n';

const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(['close']);

// Idioma seleccionado (inicialmente el actual)
const selectedLanguage = ref(currentLanguage.value);

const closeModal = () => {
  // Restaurar el idioma seleccionado al actual si se cierra sin guardar
  selectedLanguage.value = currentLanguage.value;
  emit('close');
};

const saveSettings = () => {
  // Guardar el idioma seleccionado
  setLanguage(selectedLanguage.value);
  emit('close');
};
</script>
