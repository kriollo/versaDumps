<template>
  <Transition name="slide">
    <div v-if="updateInfo && updateInfo.available" class="update-notification">
      <div class="update-content">
        <div class="update-header">
          <Icon name="download" class="update-icon" />
          <div class="update-text">
            <h3>{{ t('update_available') }}</h3>
            <p class="update-version">
              {{ t('new_version') }}: v{{ updateInfo.version }}
              <span class="current-version">({{ t('current') }}: v{{ updateInfo.currentVersion }})</span>
            </p>
          </div>
        </div>

        <div class="update-actions">
          <button @click="downloadUpdate" class="btn-update" :disabled="downloading">
            <span v-if="!downloading">{{ t('update_now') }}</span>
            <span v-else>{{ t('downloading') }}... {{ Math.round(downloadProgress) }}%</span>
          </button>
          <button @click="dismiss" class="btn-dismiss">{{ t('later') }}</button>
        </div>
      </div>

      <!-- Progress bar -->
      <div v-if="downloading" class="progress-bar">
        <div class="progress-fill" :style="{ width: downloadProgress + '%' }"></div>
      </div>
    </div>
  </Transition>

  <!-- Modal de confirmación -->
  <div v-if="showConfirmModal" class="modal-overlay" @click.self="closeModal">
    <div class="modal-content">
      <!-- Si hay actualización disponible -->
      <div v-if="updateInfo.available">
        <h2>{{ t('update_available') }}</h2>
        <p class="update-version-info">
          {{ t('new_version') }}: <strong>v{{ updateInfo.version }}</strong><br>
          {{ t('current') }}: v{{ updateInfo.currentVersion }}
        </p>
        <p>{{ t('update_description') }}</p>

        <div class="update-changelog" v-if="updateInfo.description">
          <h4>{{ t('whats_new') }}:</h4>
          <div class="changelog-content" v-html="formatChangelog(updateInfo.description)"></div>
        </div>

        <div class="modal-actions">
          <button @click="confirmUpdate" class="btn-confirm">{{ t('install_restart') }}</button>
          <button @click="closeModal" class="btn-cancel">{{ t('cancel') }}</button>
        </div>
      </div>

      <!-- Si no hay actualización disponible -->
      <div v-else>
        <h2>{{ t('no_updates_available') }}</h2>
        <p class="update-version-info">
          {{ t('current') }}: <strong>v{{ updateInfo.currentVersion }}</strong><br>
          {{ t('latest_version') }}: v{{ updateInfo.version }}
        </p>
        <p>{{ t('up_to_date') }}</p>

        <div class="modal-actions">
          <button @click="confirmUpdate" class="btn-confirm btn-disabled" disabled>{{ t('install_restart') }}</button>
          <button @click="closeModal" class="btn-cancel">{{ t('close') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import * as BackendApp from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { t } from '../i18n';
import Icon from './Icon.vue';

const updateInfo = ref(null);
const downloading = ref(false);
const downloadProgress = ref(0);
const showConfirmModal = ref(false);
const downloadedFile = ref('');

// Verificar actualizaciones al montar
onMounted(async () => {
  // Verificar actualizaciones después de 5 segundos (automático, sin modal)
  setTimeout(checkForUpdatesAutomatic, 5000);

  // Verificar cada 30 minutos (automático, sin modal)
  setInterval(checkForUpdatesAutomatic, 30 * 60 * 1000);

  // Escuchar eventos de progreso de descarga
  EventsOn('updateDownloadProgress', (data) => {
    if (data.status === 'downloading') {
      downloadProgress.value = data.progress;
    } else if (data.status === 'complete') {
      downloading.value = false;
      showConfirmModal.value = true;
    } else if (data.status === 'error') {
      downloading.value = false;
      alert(t.value('update_error') + ': ' + data.error);
    }
  });
});

// Verificación automática: solo muestra notificación si hay actualización
async function checkForUpdatesAutomatic() {
  console.log('UpdateNotification: checkForUpdatesAutomatic called');
  try {
    const info = await BackendApp.CheckForUpdates();
    console.log('UpdateNotification: received info (automatic):', info);

    // Solo procesar si hay actualización disponible
    if (info && info.available) {
      updateInfo.value = info;

      // Mostrar notificación del sistema si está disponible
      if (Notification.permission === 'granted') {
        new Notification('VersaDumps', {
          body: `${t.value('update_available')}: v${info.version}`,
          icon: '/icon.png'
        });
      }
    }
  } catch (error) {
    console.error('Error checking for updates (automatic):', error);
  }
}

// Verificación manual: siempre muestra modal
async function checkForUpdates() {
  console.log('UpdateNotification: checkForUpdates called (manual)');
  try {
    console.log('UpdateNotification: calling BackendApp.CheckForUpdates (real API)');
    const info = await BackendApp.CheckForUpdates();
    console.log('UpdateNotification: received info:', info);

    // Siempre almacenar la información y abrir el modal
    updateInfo.value = info;
    showConfirmModal.value = true;

    // Solo mostrar notificación del sistema si hay actualización disponible
    if (info && info.available && Notification.permission === 'granted') {
      new Notification('VersaDumps', {
        body: `${t.value('update_available')}: v${info.version}`,
        icon: '/icon.png'
      });
    }
  } catch (error) {
    console.error('Error checking for updates:', error);
    // Si hay error, mostrar modal con información de "sin actualizaciones"
    updateInfo.value = {
      available: false,
      version: "1.0.13",
      currentVersion: "1.0.13",
      description: "Error al verificar actualizaciones: " + error
    };
    showConfirmModal.value = true;
  }
}

async function downloadUpdate() {
  if (!updateInfo.value || !updateInfo.value.downloadUrl) return;

  showConfirmModal.value = true;
}

async function confirmUpdate() {
  downloading.value = true;
  downloadProgress.value = 0;
  closeModal();

  try {
    await BackendApp.DownloadAndInstallUpdate(updateInfo.value.downloadUrl);
    // El instalador se ejecutará automáticamente con permisos de administrador
    // Mostrar mensaje al usuario
    setTimeout(() => {
      if (downloading.value) {
        downloading.value = false;
        alert(t.value('installer_launched'));
      }
    }, 2000);
  } catch (error) {
    downloading.value = false;
    // Si el error es por permisos, mostrar mensaje específico
    if (error.toString().includes('elevation') || error.toString().includes('admin')) {
      alert(t.value('admin_required'));
    } else {
      alert(t.value('update_error') + ': ' + error);
    }
  }
}

function dismiss() {
  updateInfo.value = null;
}

function closeModal() {
  showConfirmModal.value = false;
}

function formatChangelog(text) {
  // Convertir markdown simple a HTML
  return text
    .replace(/\n/g, '<br>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/- (.*?)(<br>|$)/g, '<li>$1</li>')
    .replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>');
}

// Expose methods for parent component
defineExpose({
  checkForUpdates
});
</script>

<style scoped>
.update-notification {
  position: fixed;
  top: 60px;
  right: 20px;
  background: linear-gradient(135deg, #397111 0%, #154bb1 100%);
  color: white;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  z-index: 1000;
  max-width: 400px;
  overflow: hidden;
}

.update-content {
  padding: 20px;
}

.update-header {
  display: flex;
  align-items: flex-start;
  gap: 15px;
  margin-bottom: 15px;
}

.update-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.update-text h3 {
  margin: 0 0 5px 0;
  font-size: 18px;
  font-weight: 600;
}

.update-version {
  margin: 0;
  font-size: 14px;
  opacity: 0.95;
}

.current-version {
  opacity: 0.8;
  font-size: 12px;
}

.update-actions {
  display: flex;
  gap: 10px;
}

.btn-update, .btn-dismiss {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-update {
  background: white;
  color: #397111;
  flex: 1;
}

.btn-update:hover:not(:disabled) {
  background: #f0f0f0;
}

.btn-update:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-dismiss {
  background: transparent;
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.btn-dismiss:hover {
  background: rgba(255, 255, 255, 0.1);
}

.progress-bar {
  height: 4px;
  background: rgba(255, 255, 255, 0.2);
  position: relative;
}

.progress-fill {
  height: 100%;
  background: white;
  transition: width 0.3s ease;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 30px;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
}

.dark .modal-content {
  background: #1f2937;
  color: #f3f4f6;
}

.modal-content h2 {
  margin: 0 0 15px 0;
  color: #154bb1;
}

.update-version-info {
  margin: 10px 0 15px 0;
  padding: 12px;
  background: #f0f9ff;
  border-left: 4px solid #3b82f6;
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.6;
}

.dark .update-version-info {
  background: #1e3a8a;
  border-left-color: #60a5fa;
  color: #dbeafe;
}

.update-version-info strong {
  color: #1d4ed8;
  font-weight: 600;
}

.dark .update-version-info strong {
  color: #93c5fd;
}

.update-changelog {
  margin: 20px 0;
  padding: 15px;
  background: #f5f5f5;
  border-radius: 8px;
}

.dark .update-changelog {
  background: #374151;
}

.changelog-content {
  margin-top: 10px;
  font-size: 14px;
  line-height: 1.6;
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.btn-confirm, .btn-cancel {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-confirm {
  background: #154bb1;
  color: white;
  flex: 1;
}

.btn-confirm:hover {
  background: #154bb1;
}

.btn-disabled {
  background: #9ca3af !important;
  color: #6b7280 !important;
  cursor: not-allowed !important;
  opacity: 0.6;
}

.btn-disabled:hover {
  background: #9ca3af !important;
}

.dark .btn-disabled {
  background: #4b5563 !important;
  color: #9ca3af !important;
}

.btn-cancel {
  background: #e5e7eb;
  color: #374151;
}

.dark .btn-cancel {
  background: #4b5563;
  color: #e5e7eb;
}

/* Animaciones */
.slide-enter-active, .slide-leave-active {
  transition: all 0.3s ease;
}

.slide-enter-from {
  transform: translateX(100%);
  opacity: 0;
}

.slide-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>
