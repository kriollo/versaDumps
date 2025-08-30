<template>
  <div class="bg-white dark:bg-slate-900 shadow-md rounded-lg p-3.5 border-l-4" :class="borderColor">
    <div class="flex items-center gap-3">
      <!-- Frame Info -->
      <div class="flex-grow font-mono text-xs overflow-hidden whitespace-nowrap text-ellipsis">
        <span class="font-semibold text-green-600 dark:text-green-500">{{ log.frame.function }}()</span>
        <span class="text-slate-500 dark:text-slate-400 ml-2">@ {{ log.frame.file }}:{{ log.frame.line }}</span>
      </div>
      <!-- Timestamp and Delete Button -->
      <span class="text-xs text-slate-400 dark:text-slate-500">{{ timestamp }}</span>
      <button class="p-1 rounded-full hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200" @click="$emit('delete')" title="Delete Log">
        <Icon name="delete" />
      </button>
    </div>
    <!-- Context Tree View -->
    <JsonTreeView v-if="parsedContext" :json-data="parsedContext" />
  </div>
</template>

<script setup>
import { computed } from 'vue';
import Icon from './Icon.vue';
import JsonTreeView from './JsonTreeView.vue';

const props = defineProps({
  log: {
    type: Object,
    required: true,
  },
});

defineEmits(['delete']);

const timestamp = computed(() => new Date(props.log.id).toLocaleTimeString());

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

// A simple hashing function to get a consistent color for a given file path
const stringToHslColor = (str, s, l) => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  let h = hash % 360;
  return 'hsl('+ h +', '+ s +'%, '+ l +'%)';
}

const borderColor = computed(() => {
    const color = stringToHslColor(props.log.frame.file, 50, 60);
    return `border-color: ${color}`;
});

</script>
