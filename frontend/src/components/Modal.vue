<script setup lang="ts">
import { onMounted, onBeforeUnmount, watch } from 'vue';
const props = defineProps<{ open: boolean; title?: string; width?: string }>();
const emit = defineEmits<{ (e: 'close'): void }>();

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) emit('close');
}
onMounted(() => window.addEventListener('keydown', onKey));
onBeforeUnmount(() => window.removeEventListener('keydown', onKey));

// Prevent body scroll while open.
watch(() => props.open, (v) => {
  document.body.style.overflow = v ? 'hidden' : '';
});
</script>

<template>
  <transition name="fade">
    <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center" @click.self="emit('close')">
      <div class="absolute inset-0 bg-ink-900/40" />
      <div
        class="relative bg-white rounded-lg shadow-xl border border-ink-200 max-h-[90vh] overflow-hidden flex flex-col"
        :style="{ width: width ?? '520px' }"
      >
        <header v-if="title || $slots.header" class="px-4 py-3 border-b border-ink-200 flex items-center">
          <div class="text-sm font-semibold text-ink-900 flex-1">
            <slot name="header">{{ title }}</slot>
          </div>
          <button class="text-ink-400 hover:text-ink-700" aria-label="Close" @click="emit('close')">
            &times;
          </button>
        </header>
        <div class="p-4 overflow-y-auto">
          <slot />
        </div>
        <footer v-if="$slots.footer" class="px-4 py-3 border-t border-ink-200 flex justify-end gap-2">
          <slot name="footer" />
        </footer>
      </div>
    </div>
  </transition>
</template>
