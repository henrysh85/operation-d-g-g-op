<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue';
import { activities } from '@/api';
import type { Activity } from '@/types';
import FilterBar from '@/components/FilterBar.vue';
import DataTable from '@/components/DataTable.vue';
import Modal from '@/components/Modal.vue';
import { useFiltersStore } from '@/stores/filters';
import { format } from 'date-fns';

const rows = ref<Activity[]>([]);
const loading = ref(true);
const filters = useFiltersStore();
const showLog = ref(false);
const saving = ref(false);

const form = ref<Partial<Activity>>({
  type: 'meeting', title: '', summary: '',
  occurredAt: new Date().toISOString(),
  personIds: [], clientIds: [], jurisdictionIds: [],
});

async function load() {
  loading.value = true;
  try { rows.value = await activities.list(filters.asQuery); }
  finally { loading.value = false; }
}
onMounted(load);
watch(() => filters.asQuery, load, { deep: true });

const filtered = computed(() =>
  rows.value.filter((r) =>
    !filters.search || r.title.toLowerCase().includes(filters.search.toLowerCase()),
  ),
);

async function save() {
  saving.value = true;
  try {
    const created = await activities.create(form.value);
    rows.value.unshift(created);
    showLog.value = false;
  } finally { saving.value = false; }
}

const columns = [
  { key: 'title',      label: 'Activity', width: '40%' },
  { key: 'type',       label: 'Type' },
  { key: 'occurredAt', label: 'When', get: (r: Activity) => format(new Date(r.occurredAt), 'PPp') },
  { key: 'tags',       label: 'Tags', get: (r: Activity) => (r.tags ?? []).join(', ') },
];
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200 flex items-start">
    <div class="flex-1">
      <h1 class="text-base font-semibold text-ink-900">Activities</h1>
      <p class="text-xs text-ink-500 mt-0.5">Meetings, calls, filings and research.</p>
    </div>
    <button class="dcgg-btn-primary" @click="showLog = true">+ Log activity</button>
  </header>

  <FilterBar />

  <section class="flex-1 overflow-y-auto p-6">
    <DataTable :rows="filtered" :columns="columns" :loading="loading" empty="No activities logged." />

    <Modal :open="showLog" title="Log activity" width="520px" @close="showLog = false">
      <div class="space-y-3">
        <label class="block">
          <span class="text-xxs font-semibold text-ink-500 uppercase">Type</span>
          <select v-model="form.type" class="dcgg-input w-full mt-1">
            <option>meeting</option><option>call</option><option>email</option>
            <option>research</option><option>filing</option><option>publication</option><option>other</option>
          </select>
        </label>
        <label class="block">
          <span class="text-xxs font-semibold text-ink-500 uppercase">Title</span>
          <input v-model="form.title" class="dcgg-input w-full mt-1" />
        </label>
        <label class="block">
          <span class="text-xxs font-semibold text-ink-500 uppercase">Summary</span>
          <textarea v-model="form.summary" rows="4" class="dcgg-input w-full mt-1" />
        </label>
      </div>
      <template #footer>
        <button class="dcgg-btn" @click="showLog = false">Cancel</button>
        <button class="dcgg-btn-primary" :disabled="saving || !form.title" @click="save">
          {{ saving ? 'Saving…' : 'Save' }}
        </button>
      </template>
    </Modal>
  </section>
</template>
