<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { people } from '@/api';
import type { Person } from '@/types';
import OrgChart from '@/components/OrgChart.vue';
import DataTable from '@/components/DataTable.vue';
import { useAuthStore } from '@/stores/auth';

type Tab = 'org' | 'directory' | 'hr' | 'holidays' | 'performance' | 'expenses';
const tab = ref<Tab>('org');
const tabs: Array<{ id: Tab; label: string; gated?: boolean }> = [
  { id: 'org',         label: 'Org chart' },
  { id: 'directory',   label: 'Directory' },
  { id: 'hr',          label: 'HR', gated: true },
  { id: 'holidays',    label: 'Holidays' },
  { id: 'performance', label: 'Performance' },
  { id: 'expenses',    label: 'Expenses' },
];

const rows = ref<Person[]>([]);
const loading = ref(true);
onMounted(async () => {
  try { rows.value = await people.list(); }
  finally { loading.value = false; }
});

const auth = useAuthStore();
const pin = ref('');
const pinError = ref<string | null>(null);

async function unlockHR() {
  pinError.value = null;
  const ok = await auth.verifyPin(pin.value).catch(() => false);
  if (!ok) pinError.value = 'Invalid PIN.';
}

const columns = [
  { key: 'fullName', label: 'Name', width: '30%' },
  { key: 'role',     label: 'Role' },
  { key: 'team',     label: 'Team' },
  { key: 'region',   label: 'Region' },
  { key: 'email',    label: 'Email' },
];
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">People</h1>
    <p class="text-xs text-ink-500 mt-0.5">Team directory, org chart, and HR.</p>
  </header>

  <div class="px-6 py-2 bg-white border-b border-ink-200 flex gap-2 overflow-x-auto">
    <button
      v-for="t in tabs"
      :key="t.id"
      class="text-xs px-3 py-1.5 rounded-md border border-ink-200 whitespace-nowrap"
      :class="tab === t.id ? 'bg-brand-600 text-white border-brand-600' : 'bg-white text-ink-500 hover:bg-ink-100'"
      @click="tab = t.id"
    >
      {{ t.label }}<span v-if="t.gated" class="ml-1">🔒</span>
    </button>
  </div>

  <section class="flex-1 overflow-y-auto p-6">
    <div v-if="tab === 'org'" class="h-[70vh]">
      <OrgChart :people="rows" />
    </div>

    <div v-else-if="tab === 'directory'">
      <DataTable :rows="rows" :columns="columns" :loading="loading" empty="No people." />
    </div>

    <div v-else-if="tab === 'hr'">
      <div v-if="!auth.pinVerified" class="dcgg-card max-w-sm">
        <div class="text-sm font-semibold mb-2">HR area is PIN-protected</div>
        <p class="text-xs text-ink-500 mb-3">Enter your HR PIN to access salary and contract data.</p>
        <input v-model="pin" type="password" maxlength="8" placeholder="PIN" class="dcgg-input w-full" />
        <div v-if="pinError" class="text-xs text-err mt-2">{{ pinError }}</div>
        <button class="dcgg-btn-primary mt-3" @click="unlockHR">Unlock</button>
      </div>
      <div v-else class="dcgg-card">
        <div class="text-sm font-semibold mb-2">HR unlocked</div>
        <p class="text-xs text-ink-500">Contracts, salaries and benefits data would render here.</p>
      </div>
    </div>

    <div v-else-if="tab === 'holidays'" class="dcgg-card text-xs text-ink-500">
      Holiday calendar — to be wired to <code>/api/hr/holidays</code>.
    </div>
    <div v-else-if="tab === 'performance'" class="dcgg-card text-xs text-ink-500">
      Performance reviews — to be wired to <code>/api/hr/reviews</code>.
    </div>
    <div v-else-if="tab === 'expenses'" class="dcgg-card text-xs text-ink-500">
      Expenses — to be wired to <code>/api/hr/expenses</code>.
    </div>
  </section>
</template>
