<template>
  <li class="relative">
    <div class="inline-flex items-baseline cursor-pointer" @click="toggle">
      <span v-if="isObject" class="absolute -left-4 top-0.5 text-slate-500 transition-transform duration-100" :class="{ 'rotate-90': isExpanded }">►</span>
      <span class="text-slate-800 dark:text-slate-400">"{{ nodeKey }}":&nbsp;</span>
      <span v-if="isObject" class="text-slate-500 italic">{{ objectType }}({{ entries }})</span>
  <span v-else :class="valueClass">{{ formattedValue }}<span v-if="showTypes && !isObject" class="ml-2 text-xs text-slate-400">({{ valueType }})</span></span>
    </div>
    <JsonTreeView v-if="isObject && isExpanded" :json-data="nodeValue" />
  </li>
</template>

<script setup>
import { computed, ref } from 'vue';
import JsonTreeView from './JsonTreeView.vue';

const props = defineProps({
  nodeKey: {
    type: [String, Number],
    required: true,
  },
  nodeValue: {
    type: [Object, String, Number, Boolean, Array],
    default: null,
  },
});

const isExpanded = ref(false);

const isObject = computed(() => typeof props.nodeValue === 'object' && props.nodeValue !== null);
const objectType = computed(() => (Array.isArray(props.nodeValue) ? 'Array' : 'Object'));
const entries = computed(() => (isObject.value ? Object.keys(props.nodeValue).length : 0));
const formattedValue = computed(() => JSON.stringify(props.nodeValue));

// show types toggle (leer desde localStorage para rendimiento)
const showTypes = ref(false);
try {
  const v = localStorage.getItem('show_types');
  if (v !== null) {
    showTypes.value = v === 'true';
  }
} catch (e) {}

const valueType = computed(() => {
  if (props.nodeValue === null) return 'null';
  if (Array.isArray(props.nodeValue)) return 'Array';
  return typeof props.nodeValue;
});

const valueClass = computed(() => {
    const type = typeof props.nodeValue;
    switch(type) {
        case 'string': return 'text-green-600 dark:text-green-400';
        case 'number': return 'text-amber-600 dark:text-amber-400';
        case 'boolean': return 'text-red-600 dark:text-red-500';
        default: return 'text-slate-500'; // for null
    }
});

function toggle() {
  if (isObject.value) {
    isExpanded.value = !isExpanded.value;
  }
}
</script>
