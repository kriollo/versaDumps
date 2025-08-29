<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-slate-800 p-6 rounded-lg shadow-lg w-11/12 md:w-1/2 lg:w-1/3 relative">
      <button @click="closeModal" class="absolute top-3 right-3 text-slate-500 hover:text-slate-700 dark:hover:text-slate-300">
        <Icon name="delete" />
      </button>
      <h2 class="text-xl font-semibold mb-4 text-slate-800 dark:text-slate-200">Settings</h2>

      <!-- Configuration options will go here -->

    </div>
  </div>
</template>

<script setup>
import { defineProps, defineEmits } from 'vue';
import Icon from './Icon.vue';

const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(['close']);

const closeModal = () => {
  emit('close');
};
</script>