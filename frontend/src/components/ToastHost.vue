<script setup lang="ts">
import { useToastStore } from '@/stores/toast';
const toasts = useToastStore();

const styleMap: Record<string, string> = {
  info:    'bg-white border-ink-200 text-ink-700',
  success: 'bg-emerald-50 border-emerald-200 text-ok',
  warn:    'bg-amber-50 border-amber-200 text-warn',
  error:   'bg-red-50 border-red-200 text-err',
};
</script>

<template>
  <div class="fixed top-3 right-3 z-[60] flex flex-col gap-2 max-w-sm pointer-events-none">
    <transition-group name="toast">
      <div
        v-for="t in toasts.items"
        :key="t.id"
        class="pointer-events-auto border rounded-md shadow-sm px-3 py-2 text-xs flex items-start gap-2"
        :class="styleMap[t.kind]"
      >
        <span class="flex-1 whitespace-pre-line">{{ t.text }}</span>
        <button class="text-ink-400 hover:text-ink-700" @click="toasts.dismiss(t.id)">&times;</button>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-enter-active, .toast-leave-active { transition: all .15s ease; }
.toast-enter-from { opacity: 0; transform: translateY(-4px); }
.toast-leave-to   { opacity: 0; transform: translateY(-4px); }
</style>
