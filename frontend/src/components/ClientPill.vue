<script setup lang="ts">
defineProps<{ name: string; color?: string }>();
</script>

<template>
  <span
    class="inline-flex items-center gap-1 text-xxs px-2 py-0.5 rounded-full border border-ink-200 bg-white text-ink-700"
  >
    <span
      class="w-1.5 h-1.5 rounded-full"
      :style="{ background: color ?? '#1d4ed8' }"
    />
    {{ name }}
  </span>
</template>
