<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { membership } from '@/api';
import type { Member } from '@/types';
import FilterBar from '@/components/FilterBar.vue';
import DataTable from '@/components/DataTable.vue';
import StatusBadge from '@/components/StatusBadge.vue';
import { useRouter } from 'vue-router';
import { useFiltersStore } from '@/stores/filters';

const rows = ref<Member[]>([]);
const loading = ref(true);
const router = useRouter();
const filters = useFiltersStore();

onMounted(async () => {
  try { rows.value = await membership.list(filters.asQuery); }
  finally { loading.value = false; }
});

const filtered = computed(() =>
  rows.value.filter((m) =>
    !filters.search || m.legalName.toLowerCase().includes(filters.search.toLowerCase()),
  ),
);

const columns = [
  { key: 'legalName',  label: 'Member', width: '40%' },
  { key: 'tier',       label: 'Tier' },
  { key: 'status',     label: 'Status' },
  { key: 'riskScore',  label: 'Risk' },
];
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">Members</h1>
    <p class="text-xs text-ink-500 mt-0.5">All members, prospects and applicants.</p>
  </header>

  <FilterBar />

  <section class="flex-1 overflow-y-auto p-6">
    <DataTable
      :rows="filtered"
      :columns="columns"
      :loading="loading"
      empty="No members yet."
      @row-click="(m) => router.push(`/members/${m.id}`)"
    >
      <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
      <template #cell-riskScore="{ row }">
        <span
          v-if="row.riskScore != null"
          class="dcgg-tag"
          :class="row.riskScore >= 75 ? 'bg-red-50 text-err border-red-200'
                : row.riskScore >= 40 ? 'bg-amber-50 text-warn border-amber-200'
                : 'bg-emerald-50 text-ok border-emerald-200'"
        >{{ row.riskScore }}</span>
        <span v-else class="text-ink-400">—</span>
      </template>
    </DataTable>
  </section>
</template>
