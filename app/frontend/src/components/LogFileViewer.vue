<template>
    <div class="log-file-viewer h-full flex flex-col bg-slate-50 dark:bg-slate-900">
        <!-- Header con filtros -->
        <div class="p-3 bg-slate-100 dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700">
            <div class="flex items-center justify-between mb-2">
                <h3 class="text-sm font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-2">
                    <Icon name="file" class="text-sm" />
                    {{ t("log_files") }}
                    <span class="text-xs text-slate-500 dark:text-slate-400" v-if="logLines.length > 0">
                        ({{ logLines.length }} {{ t("lines") }})
                    </span>
                </h3>
                <div class="flex items-center gap-2">
                    <!-- Filter dropdown -->
                    <select
                        v-model="levelFilter"
                        class="text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200"
                    >
                        <option value="all">{{ t("all_levels") }}</option>
                        <option value="error">{{ t("error") }}</option>
                        <option value="warning">{{ t("warning") }}</option>
                        <option value="info">{{ t("info") }}</option>
                        <option value="debug">{{ t("debug") }}</option>
                        <option value="success">{{ t("success") }}</option>
                    </select>

                    <!-- Auto-scroll toggle -->
                    <button
                        @click="autoScroll = !autoScroll"
                        :class="[
                            'text-xs px-2 py-1 rounded border',
                            autoScroll
                                ? 'bg-green-500 text-white border-green-600'
                                : 'bg-white dark:bg-slate-700 border-slate-300 dark:border-slate-600 text-slate-800 dark:text-slate-200',
                        ]"
                        :title="autoScroll ? t('auto_scroll_on') : t('auto_scroll_off')"
                    >
                        <Icon :name="autoScroll ? 'check' : 'x'" class="inline" />
                        {{ t("auto_scroll") }}
                    </button>

                    <!-- Clear button -->
                    <button
                        @click="clearLogs"
                        class="text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-600"
                        :title="t('clear_logs')"
                    >
                        <Icon name="trash" class="inline" />
                    </button>
                </div>
            </div>

            <!-- File filter chips -->
            <div class="flex flex-wrap gap-1" v-if="activeFiles.length > 0">
                <button
                    v-for="file in activeFiles"
                    :key="file"
                    @click="toggleFileFilter(file)"
                    :class="[
                        'text-xs px-2 py-1 rounded',
                        fileFilters.includes(file)
                            ? 'bg-blue-500 text-white'
                            : 'bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-300',
                    ]"
                >
                    {{ getFileName(file) }}
                </button>
            </div>
        </div>

        <!-- Log lines container with virtual scrolling -->
        <div
            ref="logContainer"
            class="flex-1 overflow-y-auto overflow-x-auto font-mono text-xs p-2 space-y-0.5"
            @scroll="handleScroll"
        >
            <div v-if="filteredLogs.length === 0" class="text-center text-slate-500 dark:text-slate-400 py-10">
                {{ t("no_logs_yet") }}
            </div>

            <div
                v-for="(log, index) in filteredLogs"
                :key="`${log.filePath}-${log.lineNum}-${index}`"
                :class="['p-1.5 rounded border-l-4 min-w-max', getLogLevelClass(log.level)]"
            >
                <div class="flex items-start gap-2">
                    <span class="text-slate-400 dark:text-slate-500 text-[10px] shrink-0">
                        {{ formatTime(log.timestamp) }}
                    </span>
                    <span :class="['font-semibold text-[10px] shrink-0 uppercase', getLogLevelTextClass(log.level)]">
                        {{ log.level }}
                    </span>
                    <span
                        class="text-slate-500 dark:text-slate-400 text-[10px] shrink-0 truncate max-w-[150px]"
                        :title="log.fileName"
                    >
                        {{ log.fileName }}
                    </span>
                    <div class="text-slate-800 dark:text-slate-200 flex-1 min-w-0">
                        <!-- Format JSON if applicable -->
                        <pre
                            v-if="log.isJson"
                            class="json-content whitespace-pre text-[11px] leading-relaxed overflow-x-auto"
                            v-html="log.coloredJson"
                        ></pre>
                        <span v-else class="whitespace-nowrap">{{ log.line }}</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import { t } from "../i18n";
import Icon from "./Icon.vue";

// State
const logLines = ref([]);
const levelFilter = ref("all");
const autoScroll = ref(true);
const fileFilters = ref([]);
const logContainer = ref(null);
const maxLines = 1000; // Limit to prevent memory issues

// Computed
const activeFiles = computed(() => {
    const files = new Set();
    logLines.value.forEach((log) => files.add(log.filePath));
    return Array.from(files);
});

const filteredLogs = computed(() => {
    let filtered = logLines.value;

    // Filter by level
    if (levelFilter.value !== "all") {
        filtered = filtered.filter((log) => log.level === levelFilter.value);
    }

    // Filter by files (if any selected)
    if (fileFilters.value.length > 0) {
        filtered = filtered.filter((log) => fileFilters.value.includes(log.filePath));
    }

    return filtered;
});

// Methods
const getFileName = (filePath) => {
    return filePath.split(/[\\/]/).pop();
};

const toggleFileFilter = (file) => {
    const index = fileFilters.value.indexOf(file);
    if (index === -1) {
        fileFilters.value.push(file);
    } else {
        fileFilters.value.splice(index, 1);
    }
};

const getLogLevelClass = (level) => {
    const classes = {
        error: "bg-red-50 dark:bg-red-900/20 border-red-500",
        warning: "bg-yellow-50 dark:bg-yellow-900/20 border-yellow-500",
        info: "bg-blue-50 dark:bg-blue-900/20 border-blue-500",
        debug: "bg-slate-100 dark:bg-slate-800 border-slate-400",
        success: "bg-green-50 dark:bg-green-900/20 border-green-500",
    };
    return classes[level] || "bg-slate-50 dark:bg-slate-800 border-slate-300";
};

const getLogLevelTextClass = (level) => {
    const classes = {
        error: "text-red-600 dark:text-red-400",
        warning: "text-yellow-600 dark:text-yellow-400",
        info: "text-blue-600 dark:text-blue-400",
        debug: "text-slate-600 dark:text-slate-400",
        success: "text-green-600 dark:text-green-400",
    };
    return classes[level] || "text-slate-600 dark:text-slate-400";
};

const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString("en-US", {
        hour12: false,
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
    });
};

const clearLogs = () => {
    logLines.value = [];
    fileFilters.value = [];
};

const handleScroll = () => {
    if (!logContainer.value) return;

    const { scrollTop, scrollHeight, clientHeight } = logContainer.value;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;

    // Disable auto-scroll if user scrolls up
    if (!isAtBottom && autoScroll.value) {
        autoScroll.value = false;
    }
};

const scrollToBottom = () => {
    if (!autoScroll.value || !logContainer.value) return;

    nextTick(() => {
        if (logContainer.value) {
            logContainer.value.scrollTop = logContainer.value.scrollHeight;
        }
    });
};

const colorizeJson = (jsonString) => {
    return jsonString
        .replace(/(".*?")\s*:/g, '<span class="json-key">$1</span>:')
        .replace(/:\s*(".*?")/g, ': <span class="json-string">$1</span>')
        .replace(/:\s*(\d+)/g, ': <span class="json-number">$1</span>')
        .replace(/:\s*(true|false)/g, ': <span class="json-boolean">$1</span>')
        .replace(/:\s*(null)/g, ': <span class="json-null">$1</span>');
};

const tryParseJson = (line) => {
    try {
        const parsed = JSON.parse(line);
        const formatted = JSON.stringify(parsed, null, 2);
        return {
            isJson: true,
            formattedLine: formatted,
            coloredJson: colorizeJson(formatted),
        };
    } catch (e) {
        return {
            isJson: false,
            formattedLine: line,
            coloredJson: "",
        };
    }
};

const addLogLine = (logEntry) => {
    // Check if line is JSON and format it
    const { isJson, formattedLine, coloredJson } = tryParseJson(logEntry.line);

    // Add new log line with formatted version
    logLines.value.push({
        ...logEntry,
        isJson,
        formattedLine,
        coloredJson,
    });

    // Limit lines to prevent memory issues
    if (logLines.value.length > maxLines) {
        logLines.value.shift();
    }

    // Auto-scroll if enabled
    scrollToBottom();
};

// Lifecycle
onMounted(() => {
    // Listen for new log lines from backend
    EventsOn("logLine", addLogLine);
});

onUnmounted(() => {
    // Clean up event listener if needed
    // Note: Wails runtime doesn't have EventsOff, but it auto-cleans on component unmount
});
</script>

<style scoped>
/* Custom scrollbar */
.log-file-viewer ::-webkit-scrollbar {
    width: 8px;
}

.log-file-viewer ::-webkit-scrollbar-track {
    @apply bg-slate-200 dark:bg-slate-800;
}

.log-file-viewer ::-webkit-scrollbar-thumb {
    @apply bg-slate-400 dark:bg-slate-600 rounded;
}

.log-file-viewer ::-webkit-scrollbar-thumb:hover {
    @apply bg-slate-500 dark:bg-slate-500;
}

/* JSON formatting styles */
pre {
    margin: 0;
    font-family: "Courier New", monospace;
    line-height: 1.4;
}

/* JSON syntax highlighting */
.json-content :deep(.json-key) {
    color: #0066cc;
}

.dark .json-content :deep(.json-key) {
    color: #61afef;
}

.json-content :deep(.json-string) {
    color: #067d17;
}

.dark .json-content :deep(.json-string) {
    color: #98c379;
}

.json-content :deep(.json-number) {
    color: #d73a49;
}

.dark .json-content :deep(.json-number) {
    color: #d19a66;
}

.json-content :deep(.json-boolean) {
    color: #005cc5;
    font-weight: 600;
}

.dark .json-content :deep(.json-boolean) {
    color: #56b6c2;
    font-weight: 600;
}

.json-content :deep(.json-null) {
    color: #6f42c1;
    font-style: italic;
}

.dark .json-content :deep(.json-null) {
    color: #c678dd;
    font-style: italic;
}
</style>
