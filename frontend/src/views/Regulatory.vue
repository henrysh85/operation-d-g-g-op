<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue';
import { regulatory } from '@/api';
import type { Jurisdiction } from '@/types';
import FilterBar from '@/components/FilterBar.vue';
import DataTable from '@/components/DataTable.vue';
import StatusBadge from '@/components/StatusBadge.vue';
import Modal from '@/components/Modal.vue';
import { useFiltersStore } from '@/stores/filters';

const filters = useFiltersStore();
const rows = ref<Jurisdiction[]>([]);
const loading = ref(true);
const selected = ref<Jurisdiction | null>(null);

async function load() {
  loading.value = true;
  try { rows.value = await regulatory.list(filters.asQuery); }
  finally { loading.value = false; }
}
onMounted(load);
watch(() => filters.asQuery, load, { deep: true });

const filtered = computed(() =>
  rows.value.filter((r) =>
    (filters.vertical === 'all' || r.vertical === filters.vertical) &&
    (filters.region   === 'all' || r.region   === filters.region) &&
    (!filters.search || r.name.toLowerCase().includes(filters.search.toLowerCase())),
  ),
);

const columns = [
  { key: 'name',     label: 'Jurisdiction', width: '28%' },
  { key: 'region',   label: 'Region' },
  { key: 'vertical', label: 'Vertical' },
  { key: 'tier',     label: 'Tier' },
  { key: 'status',   label: 'Status' },
];
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">Regulatory</h1>
    <p class="text-xs text-ink-500 mt-0.5">All tracked jurisdictions across verticals.</p>
  </header>

  <FilterBar />

  <section class="flex-1 overflow-y-auto p-6">
    <DataTable
      :rows="filtered"
      :columns="columns"
      :loading="loading"
      empty="No jurisdictions match the current filters."
      @row-click="(r) => (selected = r)"
    >
      <template #cell-tier="{ value }">
        <span class="dcgg-tag">Tier {{ value }}</span>
      </template>
      <template #cell-status="{ value }">
        <StatusBadge :status="String(value)" />
      </template>
    </DataTable>

    <Modal :open="!!selected" :title="selected?.name" width="520px" @close="selected = null">
      <template v-if="selected">
        <dl class="grid grid-cols-2 gap-y-2 text-sm">
          <dt class="text-ink-500">Code</dt><dd>{{ selected.code }}</dd>
          <dt class="text-ink-500">Region</dt><dd>{{ selected.region }}</dd>
          <dt class="text-ink-500">Vertical</dt><dd>{{ selected.vertical }}</dd>
          <dt class="text-ink-500">Tier</dt><dd>{{ selected.tier }}</dd>
          <dt class="text-ink-500">Status</dt><dd><StatusBadge :status="selected.status" /></dd>
          <dt class="text-ink-500">Regulators</dt><dd>{{ selected.regulators?.join(', ') || '—' }}</dd>
        </dl>
      </template>
    </Modal>
  </section>
</template>
