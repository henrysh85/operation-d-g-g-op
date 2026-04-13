<script setup lang="ts">
import { computed } from 'vue';
import type { Impact } from '@/types';

const props = defineProps<{ impact: Impact }>();

const cls = computed(() => {
  switch (props.impact) {
    case 'critical':      return 'bg-red-50 text-err border-red-200';
    case 'high':          return 'bg-amber-50 text-warn border-amber-200';
    case 'medium':        return 'bg-blue-50 text-brand-600 border-blue-200';
    case 'low':           return 'bg-emerald-50 text-ok border-emerald-200';
    case 'informational': return 'bg-ink-100 text-ink-500 border-ink-200';
    default:              return 'bg-ink-100 text-ink-500 border-ink-200';
  }
});
</script>

<template>
  <span class="inline-block text-xxs px-1.5 py-0.5 rounded border font-medium capitalize"
        :class="cls">
    {{ impact }}
  </span>
</template>
