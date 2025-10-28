<template>
  <div class="bg-white dark:bg-slate-900 shadow-md rounded-lg p-3.5 border-l-4" :class="borderColor">
    <div class="flex flex-col sm:flex-row sm:justify-between gap-2">
      <!-- Frame Info -->
      <div class="flex flex-col font-mono text-xs flex-1 min-w-0">
        <div class="group relative">
          <a
            href="#"
            @click.prevent="openInEditor"
            class="text-slate-500 dark:text-slate-400 hover:underline hover:text-blue-600 block transition-all duration-200 cursor-pointer"
            :title="`${log.frame.file}:${log.frame.line} - Click to open in editor`"
          >
            <!-- Mobile view: Show only filename -->
            <span class="sm:hidden">
              <span class="font-medium">{{ fileName }}</span>:<span class="font-bold text-blue-600 dark:text-blue-400">{{ log.frame.line }}</span>
            </span>
            <!-- Desktop view: Show full path with smart truncation -->
            <span class="hidden sm:inline-block w-full">
              <span class="inline-block truncate hover:text-clip hover:break-all" style="max-width: calc(100% - 3rem);">
                <span class="text-slate-400 dark:text-slate-500">{{ dirPath }}</span><span class="font-medium">{{ fileName }}</span>
              </span>:<span class="font-bold text-blue-600 dark:text-blue-400">{{ log.frame.line }}</span>
            </span>
          </a>
        </div>
        <div class="mt-1">
          <span class="font-semibold text-green-600 dark:text-green-500 break-words sm:break-normal">{{ log.frame.function }}()</span>
        </div>
      </div>
      <!-- Timestamp and Delete Button -->
      <div class="flex flex-row sm:flex-col items-center sm:items-end gap-2 sm:gap-0">
        <div>
          <button
            class="p-1 rounded-full hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 order-2 sm:order-1"
            @click="$emit('delete')"
            title="Delete Log"
          >
            <Icon name="delete" />
          </button>
          <button
            class="p-1 rounded-full hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 order-2 sm:order-1"
            @click="$emit('copy', log)"
            title="Copy Log"
          >
            <Icon name="copy" />
          </button>

        </div>
        <span class="text-xs text-slate-400 dark:text-slate-500 font-mono order-1 sm:order-2 sm:mt-1">{{ timestamp }}</span>
      </div>
    </div>
    <!-- Context Tree View -->
    <JsonTreeView v-if="parsedContext" :json-data="parsedContext" />

    <!-- Stack Trace (si está disponible) -->
    <div v-if="log.trace || log.stack || log.backtrace" class="mt-3 border-t border-slate-200 dark:border-slate-700 pt-3">
      <div class="text-xs font-semibold text-slate-600 dark:text-slate-400 mb-2 flex items-center gap-2">
        <Icon name="list" />
        Stack Trace
      </div>
      <div class="space-y-1 max-h-64 overflow-y-auto">
        <div
          v-for="(frame, index) in traceFrames"
          :key="index"
          class="text-xs font-mono bg-slate-50 dark:bg-slate-800 p-2 rounded border-l-2 border-blue-400"
        >
          <div class="text-slate-700 dark:text-slate-300">
            <span class="text-blue-600 dark:text-blue-400">{{ index + 1 }}.</span>
            <span v-if="frame.class" class="text-green-600 dark:text-green-400">{{ frame.class }}{{ frame.type || '::' }}</span>
            <span class="font-semibold">{{ frame.function || 'anonymous' }}()</span>
          </div>
          <div v-if="frame.file" class="text-slate-500 dark:text-slate-400 truncate" :title="frame.file">
            {{ frame.file }}<span v-if="frame.line" class="text-blue-600 dark:text-blue-400">:{{ frame.line }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import * as BackendApp from '../../wailsjs/go/main/App';
import Icon from './Icon.vue';
import JsonTreeView from './JsonTreeView.vue';

const props = defineProps({
  log: {
    type: Object,
    required: true,
  },
});

defineEmits(['delete', 'copy']);

const timestamp = computed(() => new Date(props.log.id).toLocaleTimeString());

// Extract filename from full path
const fileName = computed(() => {
  const path = props.log.frame.file;
  // Handle both Windows and Unix paths
  const parts = path.split(/[\\\/]/);
  return parts[parts.length - 1] || path;
});

// Get directory path (everything except filename)
const dirPath = computed(() => {
  const path = props.log.frame.file;
  const lastSlash = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'));
  if (lastSlash === -1) return '';
  return path.substring(0, lastSlash + 1);
});

const parseContext = (ctx) => {
  // If already an object/array, return as-is
  if (ctx !== null && typeof ctx === 'object') return ctx

  // If it's a string, try to parse JSON, otherwise return original string
  if (typeof ctx === 'string') {
    const trimmed = ctx.trim()
    // Fast-path: if looks like JSON object/array, attempt parse
    if ((trimmed.startsWith('{') && trimmed.endsWith('}')) || (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
      try {
        return JSON.parse(trimmed)
      } catch (e) {
        return { error: 'Malformed JSON', original_context: ctx }
      }
    }
    // Not JSON - return the raw string so UI can show it plainly
    return ctx
  }

  // Unknown type - return as-is wrapped for safety
  return { error: 'Unsupported context type', original_context: String(ctx) }
}

const parsedContext = computed(() => {
  if (!props.log.context) return null
  return parseContext(props.log.context)
})

// Procesar stack trace si está disponible
const traceFrames = computed(() => {
  const trace = props.log.trace || props.log.stack || props.log.backtrace;

  if (!trace) return [];

  // Si trace es un array, usarlo directamente
  if (Array.isArray(trace)) {
    return trace;
  }

  // Si trace es un número (niveles de trace solicitados), no hay frames
  if (typeof trace === 'number') {
    return [];
  }

  // Si trace es un objeto, intentar convertirlo a array
  if (typeof trace === 'object') {
    return Object.values(trace);
  }

  return [];
});

// A simple hashing function to get a consistent color for a given file path
const stringToHslColor = (str, s, l) => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  let h = hash % 360;
  return 'hsl('+ h +', '+ s +'%, '+ l +'%)';
}

// Mapa de colores semánticos de versadumps-php a Tailwind
const semanticColors = {
  'red': 'border-red-500',
  'green': 'border-green-500',
  'blue': 'border-blue-500',
  'yellow': 'border-yellow-500',
  'purple': 'border-purple-500',
  'orange': 'border-orange-500',
  'pink': 'border-pink-500',
  'cyan': 'border-cyan-500',
  'gray': 'border-gray-500',
  'white': 'border-slate-300 dark:border-slate-600',
};

const borderColor = computed(() => {
    // Si hay color semántico del metadata, usarlo
    if (props.log.color && semanticColors[props.log.color.toLowerCase()]) {
      return semanticColors[props.log.color.toLowerCase()];
    }

    // Fallback: color basado en el path del archivo
    const color = stringToHslColor(props.log.frame.file, 50, 60);
    return `border-color: ${color}`;
});

// Attempt to open the file in the user's editor. Use generated BackendApp if present,
// otherwise fall back to calling the wails bridge directly.
const openInEditor = async () => {
  const path = props.log.frame.file;
  const line = props.log.frame.line || 1;
  try {
    if (BackendApp && typeof BackendApp.OpenInEditor === 'function') {
      await BackendApp.OpenInEditor(path, line);
    } else if (window && window['go'] && window['go']['main'] && window['go']['main']['App'] && typeof window['go']['main']['App']['OpenInEditor'] === 'function') {
      window['go']['main']['App']['OpenInEditor'](path, line);
    } else {
      // last resort: OpenInEditor not available in this context
    }
  } catch (e) {
    // Silently fail in production
  }
};

</script>
