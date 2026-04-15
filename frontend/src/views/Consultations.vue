<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue';
import { consultations } from '@/api';
import type { Consultation } from '@/types';
import FilterBar from '@/components/FilterBar.vue';
import DataTable from '@/components/DataTable.vue';
import ImpactBadge from '@/components/ImpactBadge.vue';
import StatusBadge from '@/components/StatusBadge.vue';
import { useFiltersStore } from '@/stores/filters';
import { differenceInCalendarDays, format } from 'date-fns';

const rows = ref<Consultation[]>([]);
const loading = ref(true);
const filters = useFiltersStore();

const PAGE = 100;
const offset = ref(0);
const hasMore = ref(true);
const loadingMore = ref(false);

async function load() {
  loading.value = true;
  offset.value = 0;
  try {
    const page = await consultations.list({ ...filters.asQuery, limit: PAGE, offset: 0 });
    rows.value = page;
    hasMore.value = page.length === PAGE;
  } finally { loading.value = false; }
}
async function loadMore() {
  if (loadingMore.value || !hasMore.value) return;
  loadingMore.value = true;
  try {
    offset.value += PAGE;
    const page = await consultations.list({ ...filters.asQuery, limit: PAGE, offset: offset.value });
    rows.value.push(...page);
    hasMore.value = page.length === PAGE;
  } finally { loadingMore.value = false; }
}
onMounted(load);
watch(() => filters.asQuery, load, { deep: true });

const filtered = computed(() =>
  rows.value.filter((r) =>
    (filters.vertical === 'all' || r.vertical === filters.vertical) &&
    (!filters.search || r.title.toLowerCase().includes(filters.search.toLowerCase())),
  ),
);

function daysLeft(d: string) {
  return differenceInCalendarDays(new Date(d), new Date());
}

const columns = [
  { key: 'title',     label: 'Consultation', width: '38%' },
  { key: 'regulator', label: 'Regulator' },
  { key: 'deadline',  label: 'Deadline',  get: (r: Consultation) => r.deadlineAt },
  { key: 'impact',    label: 'Impact' },
  { key: 'status',    label: 'Status' },
];
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">Consultations</h1>
    <p class="text-xs text-ink-500 mt-0.5">Open regulatory consultations with deadline countdowns.</p>
  </header>

  <FilterBar />

  <section class="flex-1 overflow-y-auto p-6">
    <DataTable :rows="filtered" :columns="columns" :loading="loading" empty="No consultations found.">
      <template #cell-deadline="{ row }">
        <div class="flex items-center gap-2">
          <span>{{ format(new Date(row.deadlineAt), 'PP') }}</span>
          <span
            class="text-xxs font-semibold"
            :class="daysLeft(row.deadlineAt) < 0
              ? 'text-ink-400'
              : daysLeft(row.deadlineAt) <= 7
                ? 'text-err'
                : daysLeft(row.deadlineAt) <= 30
                  ? 'text-warn'
                  : 'text-ok'"
          >
            {{ daysLeft(row.deadlineAt) < 0 ? 'closed' : `${daysLeft(row.deadlineAt)}d left` }}
          </span>
        </div>
      </template>
      <template #cell-impact="{ row }"><ImpactBadge :impact="row.impact" /></template>
      <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
    </DataTable>
    <div v-if="hasMore && !loading" class="flex justify-center mt-4">
      <button class="dcgg-btn" :disabled="loadingMore" @click="loadMore">
        {{ loadingMore ? 'Loading…' : `Load ${PAGE} more` }}
      </button>
    </div>
  </section>
</template>
