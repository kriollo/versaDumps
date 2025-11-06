<template>
  <div class="log-folders-manager">
    <div class="mb-4">
      <h4 class="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-2">
        {{ t('log_folders') }}
      </h4>

      <!-- Add folder button -->
      <button
        @click="showAddFolder = true"
        class="w-full px-3 py-2 text-sm bg-blue-500 hover:bg-blue-600 text-white rounded flex items-center justify-center gap-2"
      >
        <Icon name="plus" />
        {{ t('add_log_folder') }}
      </button>
    </div>

    <!-- Folder list -->
    <div class="space-y-2">
      <div
        v-for="(folder, index) in folders"
        :key="index"
        class="p-3 rounded border border-slate-300 dark:border-slate-600 bg-slate-50 dark:bg-slate-800"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-1">
              <input
                type="checkbox"
                :checked="folder.enabled !== undefined ? folder.enabled : folder.Enabled"
                @change="toggleFolder(index)"
                class="rounded"
              />
              <span class="text-xs font-mono text-slate-700 dark:text-slate-300 truncate" :title="folder.path || folder.Path">
                {{ folder.path || folder.Path || '[Sin ruta]' }}
              </span>
            </div>

            <div class="text-xs text-slate-500 dark:text-slate-400 space-y-0.5">
              <div>
                <span class="font-semibold">{{ t('extensions') }}:</span>
                {{ ((folder.extensions || folder.Extensions) && (folder.extensions || folder.Extensions).length > 0) ? (folder.extensions || folder.Extensions).join(', ') : t('all') }}
              </div>
              <div>
                <span class="font-semibold">{{ t('log_format') }}:</span>
                <span class="px-1.5 py-0.5 rounded bg-slate-200 dark:bg-slate-700 text-xs">
                  {{ (folder.format || folder.Format || 'text').toUpperCase() }}
                </span>
              </div>
              <div v-if="(folder.filters || folder.Filters) && (folder.filters || folder.Filters).length > 0">
                <span class="font-semibold">{{ t('filters') }}:</span>
                {{ (folder.filters || folder.Filters).join(', ') }}
              </div>
            </div>
          </div>

          <div class="flex gap-1">
            <button
              @click="editFolder(index)"
              class="p-1 text-blue-500 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
              :title="t('edit')"
            >
              <Icon name="edit" class="text-sm" />
            </button>
            <button
              @click="removeFolder(index)"
              class="p-1 text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
              :title="t('remove')"
            >
              <Icon name="trash" class="text-sm" />
            </button>
          </div>
        </div>
      </div>

      <div v-if="folders.length === 0" class="text-center text-sm text-slate-500 dark:text-slate-400 py-4">
        {{ t('no_log_folders') }}
      </div>
    </div>

    <!-- Add folder modal -->
    <div
      v-if="showAddFolder"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]"
      @click.self="showAddFolder = false"
    >
      <div class="bg-white dark:bg-slate-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4" @click.stop>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-200 mb-4">
          {{ t('add_log_folder') }}
        </h3>

        <div class="space-y-4">
          <!-- Folder path -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              {{ t('folder_path') }}
            </label>
            <div class="flex gap-2">
              <input
                v-model="newFolder.path"
                type="text"
                class="flex-1 px-3 py-2 text-sm border border-slate-300 dark:border-slate-600 rounded bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200"
                :placeholder="t('folder_path_placeholder')"
              />
              <button
                @click="selectFolder"
                class="px-3 py-2 bg-slate-500 hover:bg-slate-600 text-white rounded text-sm"
              >
                {{ t('browse') }}
              </button>
            </div>
          </div>

          <!-- Extensions -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              {{ t('file_extensions') }}
            </label>
            <input
              v-model="extensionsInput"
              type="text"
              class="w-full px-3 py-2 text-sm border border-slate-300 dark:border-slate-600 rounded bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200"
              placeholder="*.log, *.txt"
            />
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {{ t('extensions_help') }}
            </p>
          </div>

          <!-- Log Format -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              {{ t('log_format') }}
            </label>
            <select
              v-model="newFolder.format"
              class="w-full px-3 py-2 text-sm border border-slate-300 dark:border-slate-600 rounded bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200"
            >
              <option value="text">{{ t('format_text') }}</option>
              <option value="json">{{ t('format_json') }}</option>
            </select>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {{ t('format_help') }}
            </p>
          </div>

          <!-- Filters -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              {{ t('log_level_filters') }}
            </label>
            <div class="flex flex-wrap gap-2">
              <label v-for="level in ['error', 'warning', 'info', 'debug', 'success']" :key="level" class="flex items-center gap-1">
                <input
                  type="checkbox"
                  :value="level"
                  v-model="newFolder.filters"
                  class="rounded"
                />
                <span class="text-sm text-slate-700 dark:text-slate-300">{{ t(level) }}</span>
              </label>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {{ t('filters_help') }}
            </p>
          </div>
        </div>

        <div class="flex gap-2 mt-6">
          <button
            @click="addFolder"
            class="flex-1 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="!newFolder.path || newFolder.path.trim() === ''"
          >
            {{ t('add') }}
          </button>
          <button
            @click="showAddFolder = false"
            class="flex-1 px-4 py-2 bg-slate-300 hover:bg-slate-400 dark:bg-slate-600 dark:hover:bg-slate-500 text-slate-800 dark:text-slate-200 rounded"
          >
            {{ t('cancel') }}
          </button>
        </div>

        <!-- Debug info -->
        <div class="mt-2 text-xs text-slate-500" v-if="false">
          Path: "{{ newFolder.path }}" | Length: {{ newFolder.path?.length || 0 }} | Disabled: {{ !newFolder.path || newFolder.path.trim() === '' }}
        </div>
      </div>
    </div>

    <!-- Edit folder modal -->
    <div
      v-if="showEditFolder"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]"
      @click.self="showEditFolder = false"
    >
      <div class="bg-white dark:bg-slate-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4" @click.stop>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-200 mb-4">
          {{ t('edit_log_folder') }}
        </h3>

        <div class="space-y-4">
          <!-- Folder path (read-only) -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              {{ t('folder_path') }}
            </label>
            <input
              v-model="editingFolder.path"
              type="text"
              disabled
              class="w-full px-3 py-2 text-sm border border-slate-300 dark:border-slate-600 rounded bg-slate-100 dark:bg-slate-900 text-slate-600 dark:text-slate-400 cursor-not-allowed"
            />
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {{ t('path_cannot_be_changed') }}
            </p>
          </div>

          <!-- Extensions -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              {{ t('file_extensions') }}
            </label>
            <input
              v-model="editExtensionsInput"
              type="text"
              class="w-full px-3 py-2 text-sm border border-slate-300 dark:border-slate-600 rounded bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200"
              placeholder="*.log, *.txt"
            />
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {{ t('extensions_help') }}
            </p>
          </div>

          <!-- Log Format -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              {{ t('log_format') }}
            </label>
            <select
              v-model="editingFolder.format"
              class="w-full px-3 py-2 text-sm border border-slate-300 dark:border-slate-600 rounded bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200"
            >
              <option value="text">{{ t('format_text') }}</option>
              <option value="json">{{ t('format_json') }}</option>
            </select>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {{ t('format_help') }}
            </p>
          </div>

          <!-- Filters -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              {{ t('log_level_filters') }}
            </label>
            <div class="flex flex-wrap gap-2">
              <label v-for="level in ['error', 'warning', 'info', 'debug', 'success']" :key="level" class="flex items-center gap-1">
                <input
                  type="checkbox"
                  :value="level"
                  v-model="editingFolder.filters"
                  class="rounded"
                />
                <span class="text-sm text-slate-700 dark:text-slate-300">{{ t(level) }}</span>
              </label>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {{ t('filters_help') }}
            </p>
          </div>
        </div>

        <div class="flex gap-2 mt-6">
          <button
            @click="updateFolder"
            class="flex-1 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded"
          >
            {{ t('save') }}
          </button>
          <button
            @click="showEditFolder = false"
            class="flex-1 px-4 py-2 bg-slate-300 hover:bg-slate-400 dark:bg-slate-600 dark:hover:bg-slate-500 text-slate-800 dark:text-slate-200 rounded"
          >
            {{ t('cancel') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';
import * as BackendApp from '../../wailsjs/go/main/App';
import { t } from '../i18n';
import Icon from './Icon.vue';

// Props
const props = defineProps({
  profileName: {
    type: String,
    required: true
  },
  folders: {
    type: Array,
    default: () => []
  }
});

// Debug: Watch folders prop
watch(() => props.folders, (newFolders) => {
  console.log('LogFoldersManager - Folders updated:', newFolders);
  if (newFolders && newFolders.length > 0) {
    console.log('First folder:', newFolders[0]);
  }
}, { immediate: true });

// Emits
const emit = defineEmits(['update', 'error']);

// State
const showAddFolder = ref(false);
const showEditFolder = ref(false);
const editingFolderIndex = ref(-1);
const extensionsInput = ref('*.log, *.txt');
const editExtensionsInput = ref('');
const newFolder = ref({
  path: '',
  extensions: [],
  filters: [],
  enabled: true,
  format: 'text'
});
const editingFolder = ref({
  path: '',
  originalPath: '',
  extensions: [],
  filters: [],
  enabled: true,
  format: 'text'
});

// Methods
const selectFolder = async () => {
  try {
    const folder = await BackendApp.SelectFolder();
    if (folder) {
      newFolder.value.path = folder;
    }
  } catch (e) {
    console.error('Error selecting folder:', e);
    emit('error', t.value('error_selecting_folder'));
  }
};

const addFolder = async () => {
  if (!newFolder.value.path || !newFolder.value.path.trim()) {
    emit('error', t.value('folder_path_required'));
    return;
  }

  // Check if folder already exists
  const folderPath = newFolder.value.path.trim();
  const isDuplicate = props.folders.some(f => (f.path || f.Path) === folderPath);
  if (isDuplicate) {
    emit('error', t.value('folder_already_exists'));
    return;
  }

  try {
    // Parse extensions from input
    const extensions = extensionsInput.value
      .split(',')
      .map(ext => ext.trim())
      .filter(ext => ext.length > 0);

    if (!extensions.length) {
      extensions.push('*.log'); // Default
    }

    console.log('Adding folder:', {
      profileName: props.profileName,
      path: newFolder.value.path,
      extensions,
      filters: newFolder.value.filters,
      format: newFolder.value.format
    });

    // Add folder via backend
    await BackendApp.AddLogFolder(
      props.profileName,
      newFolder.value.path,
      extensions,
      newFolder.value.filters,
      newFolder.value.format || 'text'
    );

    console.log('Folder added successfully');

    // Reset form
    newFolder.value = {
      path: '',
      extensions: [],
      filters: [],
      enabled: true,
      format: 'text'
    };
    extensionsInput.value = '*.log, *.txt';
    showAddFolder.value = false;

    // Emit update to reload the profile data
    emit('update');
  } catch (e) {
    console.error('Error adding folder:', e);

    // Check for specific error messages
    let errorMessage = t.value('error_adding_folder');
    if (e && e.message) {
      if (e.message.includes('already exists')) {
        errorMessage = t.value('folder_already_exists');
      } else {
        errorMessage = e.message;
      }
    }

    emit('error', errorMessage);
  }
};

const removeFolder = async (index) => {
  const folder = props.folders[index];
  const folderPath = folder.path || folder.Path;

  if (!folderPath) {
    console.error('Folder has no path:', folder);
    emit('error', 'Error: carpeta sin ruta');
    return;
  }

  if (!confirm(t.value('confirm_remove_folder'))) {
    return;
  }

  try {
    await BackendApp.RemoveLogFolder(props.profileName, folderPath);
    emit('update');
  } catch (e) {
    console.error('Error removing folder:', e);
    emit('error', e.message || t.value('error_removing_folder'));
  }
};

const toggleFolder = async (index) => {
  const folder = props.folders[index];
  const folderPath = folder.path || folder.Path;
  const isEnabled = folder.enabled !== undefined ? folder.enabled : folder.Enabled;

  if (!folderPath) {
    console.error('Folder has no path:', folder);
    emit('error', 'Error: carpeta sin ruta');
    return;
  }

  try {
    await BackendApp.ToggleLogFolder(props.profileName, folderPath, !isEnabled);
    emit('update');
  } catch (e) {
    console.error('Error toggling folder:', e);
    emit('error', e.message || t.value('error_toggling_folder'));
  }
};

const editFolder = (index) => {
  const folder = props.folders[index];
  const folderPath = folder.path || folder.Path;
  const folderExtensions = folder.extensions || folder.Extensions || [];
  const folderFilters = folder.filters || folder.Filters || [];
  const folderFormat = folder.format || folder.Format || 'text';

  editingFolderIndex.value = index;
  editingFolder.value = {
    path: folderPath,
    originalPath: folderPath,
    extensions: [...folderExtensions],
    filters: [...folderFilters],
    format: folderFormat,
    enabled: folder.enabled !== undefined ? folder.enabled : folder.Enabled
  };

  // Convert extensions array to input string
  editExtensionsInput.value = folderExtensions.join(', ');

  showEditFolder.value = true;
};

const updateFolder = async () => {
  try {
    // Parse extensions from input
    const extensions = editExtensionsInput.value
      .split(',')
      .map(ext => ext.trim())
      .filter(ext => ext.length > 0);

    if (!extensions.length) {
      extensions.push('*.log'); // Default
    }

    console.log('Updating folder:', {
      profileName: props.profileName,
      path: editingFolder.value.originalPath,
      extensions,
      filters: editingFolder.value.filters,
      format: editingFolder.value.format
    });

    // Update folder via backend
    await BackendApp.UpdateLogFolder(
      props.profileName,
      editingFolder.value.originalPath,
      extensions,
      editingFolder.value.filters,
      editingFolder.value.format || 'text'
    );

    console.log('Folder updated successfully');

    // Close modal
    showEditFolder.value = false;
    editingFolderIndex.value = -1;

    // Emit update to reload the profile data
    emit('update');
  } catch (e) {
    console.error('Error updating folder:', e);
    emit('error', e.message || t.value('error_updating_folder'));
  }
};
</script>
