<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue';
import { activities } from '@/api';
import type { Activity } from '@/types';
import type { ActivityOutput } from '@/api/activities';
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

const filesFor = ref<Activity | null>(null);
const fileList = ref<ActivityOutput[]>([]);
const uploading = ref(false);
const fileError = ref<string | null>(null);

async function openFiles(a: Activity) {
  filesFor.value = a;
  fileError.value = null;
  fileList.value = await activities.listOutputs(a.id).catch(() => []);
}

async function handleUpload(ev: Event) {
  if (!filesFor.value) return;
  const input = ev.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  uploading.value = true; fileError.value = null;
  try {
    await activities.uploadOutput(filesFor.value.id, file);
    fileList.value = await activities.listOutputs(filesFor.value.id);
    input.value = '';
  } catch (e) { fileError.value = (e as Error).message; }
  finally { uploading.value = false; }
}

const columns = [
  { key: 'title',      label: 'Activity', width: '40%' },
  { key: 'type',       label: 'Type' },
  { key: 'occurredAt', label: 'When', get: (r: Activity) => format(new Date(r.occurredAt), 'PPp') },
  { key: 'tags',       label: 'Tags', get: (r: Activity) => (r.tags ?? []).join(', ') },
  { key: 'files',      label: 'Files' },
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
    <DataTable :rows="filtered" :columns="columns" :loading="loading" empty="No activities logged.">
      <template #cell-files="{ row }">
        <button class="text-xxs text-brand-600 hover:underline" @click.stop="openFiles(row)">Files</button>
      </template>
    </DataTable>

    <Modal :open="!!filesFor" :title="filesFor ? 'Files — ' + filesFor.title : ''" width="520px" @close="filesFor = null">
      <div v-if="!fileList.length" class="text-xs text-ink-400 mb-3">No files attached yet.</div>
      <ul v-else class="divide-y divide-ink-100 mb-3">
        <li v-for="f in fileList" :key="f.id" class="py-2 flex items-center gap-2">
          <a :href="f.url" target="_blank" class="text-xs text-brand-600 hover:underline truncate flex-1">{{ f.label }}</a>
          <span class="text-xxs text-ink-400">{{ (f.size_bytes / 1024).toFixed(1) }} KB</span>
        </li>
      </ul>
      <label class="block">
        <span class="text-xxs font-semibold text-ink-500 uppercase">Attach file (max 25 MB)</span>
        <input type="file" :disabled="uploading" class="dcgg-input w-full mt-1" @change="handleUpload" />
      </label>
      <div v-if="uploading" class="text-xs text-ink-400 mt-2">Uploading…</div>
      <div v-if="fileError" class="text-xs text-err mt-2">{{ fileError }}</div>
    </Modal>

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
