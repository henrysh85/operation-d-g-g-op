<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { audit } from '@/api';
import type { AuditEntry } from '@/api/audit';
import { format } from 'date-fns';

const rows = ref<AuditEntry[]>([]);
const loading = ref(true);
const filterEntity = ref('');
const filterActor = ref('');

async function load() {
  loading.value = true;
  try {
    rows.value = await audit.list({
      entity: filterEntity.value || undefined,
      actor: filterActor.value || undefined,
    });
  } finally { loading.value = false; }
}
onMounted(load);
let t: number | undefined;
watch([filterEntity, filterActor], () => {
  if (t) clearTimeout(t);
  t = window.setTimeout(load, 250);
});
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">Audit log</h1>
    <p class="text-xs text-ink-500 mt-0.5">Every successful mutation under /api/v1, newest first.</p>
  </header>

  <div class="px-6 py-3 bg-white border-b border-ink-200 flex gap-2">
    <input v-model="filterEntity" placeholder="entity (e.g. activities)" class="dcgg-input w-56" />
    <input v-model="filterActor"  placeholder="actor email contains…" class="dcgg-input w-64" />
    <button class="dcgg-btn ml-auto" @click="load">Refresh</button>
  </div>

  <section class="flex-1 overflow-y-auto p-6">
    <div v-if="loading" class="text-xs text-ink-400">Loading…</div>
    <div v-else-if="!rows.length" class="dcgg-card text-xs text-ink-400">No entries match.</div>
    <table v-else class="dcgg-card !p-0 w-full text-xs">
      <thead class="text-xxs uppercase text-ink-500 bg-ink-50">
        <tr>
          <th class="text-left px-3 py-2">When</th>
          <th class="text-left px-3 py-2">Actor</th>
          <th class="text-left px-3 py-2">Action</th>
          <th class="text-left px-3 py-2">Entity</th>
          <th class="text-left px-3 py-2">Path</th>
          <th class="text-left px-3 py-2">Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.id" class="border-t border-ink-100">
          <td class="px-3 py-1.5 whitespace-nowrap">{{ format(new Date(r.createdAt), 'PPp') }}</td>
          <td class="px-3 py-1.5">{{ r.actorEmail ?? '—' }}</td>
          <td class="px-3 py-1.5 font-mono text-xxs">{{ r.action }}</td>
          <td class="px-3 py-1.5">{{ r.entity }}</td>
          <td class="px-3 py-1.5 font-mono text-xxs text-ink-600">{{ (r.metadata as any).path ?? '—' }}</td>
          <td class="px-3 py-1.5">{{ (r.metadata as any).status ?? '—' }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>
