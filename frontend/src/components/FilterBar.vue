<script setup lang="ts">
import { watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { useFiltersStore } from '@/stores/filters';

const filters = useFiltersStore();
const { vertical, region, clientId, month, search } = storeToRefs(filters);
const route = useRoute();
const router = useRouter();

// Hydrate on mount from URL.
filters.hydrateFromQuery(route.query as Record<string, string | undefined>);

// Sync to URL (shallow).
watch(
  () => filters.asQuery,
  (q) => {
    router.replace({ query: { ...route.query, ...q } }).catch(() => void 0);
  },
  { deep: true },
);
</script>

<template>
  <div class="flex flex-wrap items-center gap-2 px-6 py-3 bg-white border-b border-ink-200">
    <select v-model="vertical" class="dcgg-input">
      <option value="all">All verticals</option>
      <option value="crypto">Crypto</option>
      <option value="ai">AI</option>
      <option value="privacy">Privacy</option>
      <option value="market">Market infra</option>
    </select>

    <select v-model="region" class="dcgg-input">
      <option value="all">All regions</option>
      <option value="eu">Europe</option>
      <option value="na">North America</option>
      <option value="latam">Latin America</option>
      <option value="mena">MENA</option>
      <option value="apac">APAC</option>
      <option value="africa">Africa</option>
    </select>

    <input v-model="clientId" placeholder="Client" class="dcgg-input w-32" />

    <input v-model="month" type="month" class="dcgg-input" />

    <input
      v-model="search"
      type="search"
      placeholder="Search…"
      class="dcgg-input flex-1 min-w-[180px] max-w-xs"
    />

    <button class="dcgg-btn" @click="filters.reset()">Reset</button>
  </div>
</template>
