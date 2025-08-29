<template>
  <main>
    <div
      class="sticky top-0 z-10 flex items-center justify-between p-2.5 bg-slate-100/80 dark:bg-slate-900/80 backdrop-blur-sm shadow-sm"
    >
      <h1 class="text-lg font-semibold text-slate-800 dark:text-slate-200">
        VersaDumps Visualizer
      </h1>
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

    <div class="p-2.5 space-y-2.5">
      <div v-if="logs.length === 0" class="text-center py-10 text-slate-500">
        <p>{{ t('waiting_data') }}</p>
      </div>
      <LogItem
        v-for="log in sortedLogs"
        :key="log.id"
        :log="log"
        @delete="deleteLog(log.id)"
      />
    </div>

    <ConfigModal :is-open="isConfigModalOpen" @close="closeConfigModal" />
  </main>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import Icon from "./components/Icon.vue";
import LogItem from "./components/LogItem.vue";
import ConfigModal from "./components/ConfigModal.vue";
import { t, currentLanguage } from "./i18n";

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
onMounted(() => {
  const savedTheme = localStorage.getItem("theme") || "dark";
  theme.value = savedTheme;
  if (savedTheme === "dark") {
    document.documentElement.classList.add("dark");
  }
  
  // Initialize logs with some test data if none is received
  setTimeout(() => {
    if (logs.value.length === 0) {
      logs.value.push({
        id: Date.now(),
        frame: { file: "Example.js", line: 42, function: "testFunction" },
        context: "This is a test log to show how the visualizer works"
      });
    }
  }, 2000);
});

// CONFIG MODAL
const isConfigModalOpen = ref(false);
const openConfigModal = () => {
  isConfigModalOpen.value = true;
};
const closeConfigModal = () => {
  isConfigModalOpen.value = false;
};

// LOGS
const logs = ref([]);
onMounted(() => {
  EventsOn("newData", (data) => {
    try {
      const parsedData = JSON.parse(data);
      logs.value.push({ ...parsedData, id: Date.now() });
    } catch (e) {
      logs.value.push({
        id: Date.now(),
        frame: { file: "Error", line: 0, function: "Invalid Data" },
        context: data,
      });
    }
  });
});

const deleteLog = (id) => {
  logs.value = logs.value.filter((log) => log.id !== id);
};

const clearLogs = () => {
  logs.value = [];
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
</script>

<style>
.icon-button {
  @apply p-1 rounded-full hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200;
}
</style>
