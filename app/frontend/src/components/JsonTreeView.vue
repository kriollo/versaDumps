<template>
  <div class="mt-2.5 p-2.5 bg-slate-100 dark:bg-slate-800 rounded font-mono text-xs overflow-auto">
    <ul class="pl-5 break-words">
      <JsonTreeViewNode v-for="(value, key) in jsonData" :node-key="key" :node-value="value" :key="key" />
    </ul>
  </div>
</template>

<script setup>
import JsonTreeViewNode from './JsonTreeViewNode.vue';

defineProps({
  jsonData: {
    type: Object,
    required: true,
  },
});
</script>