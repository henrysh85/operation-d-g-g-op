<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { activities, publications } from '@/api';
import type { Activity, Publication } from '@/types';
import { format } from 'date-fns';

type Tab = 'roundtables' | 'meetings' | 'publications' | 'governance';
const tab = ref<Tab>('roundtables');
const tabs: Array<{ id: Tab; label: string }> = [
  { id: 'roundtables',  label: 'Roundtables' },
  { id: 'meetings',     label: 'Meetings' },
  { id: 'publications', label: 'Publications' },
  { id: 'governance',   label: 'Governance' },
];

const acts = ref<Activity[]>([]);
const pubs = ref<Publication[]>([]);
const loading = ref(true);

onMounted(async () => {
  try {
    const [a, p] = await Promise.all([
      activities.list().catch(() => []),
      publications.list().catch(() => []),
    ]);
    acts.value = a; pubs.value = p;
  } finally { loading.value = false; }
});

function groupByRegion<T extends { region?: string; tags?: string[] }>(items: T[]): Array<[string, T[]]> {
  const m = new Map<string, T[]>();
  for (const i of items) {
    const r = i.region ?? (i.tags?.find((t) => /^(EMEA|Americas|APAC|MENA|Africa|Global)$/.test(t))) ?? 'Unassigned';
    if (!m.has(r)) m.set(r, []);
    m.get(r)!.push(i);
  }
  return Array.from(m.entries());
}

const roundtables = computed(() => groupByRegion(acts.value.filter((a) => a.tags?.includes('roundtable'))));
const meetings    = computed(() => groupByRegion(acts.value.filter((a) => a.type === 'meeting')));
const pubsByRegion = computed(() => groupByRegion(pubs.value));
const governance  = computed(() => groupByRegion(acts.value.filter((a) => a.tags?.includes('governance'))));
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">Engagement</h1>
    <p class="text-xs text-ink-500 mt-0.5">Roundtables, meetings, publications and governance by region.</p>
  </header>

  <div class="px-6 py-2 bg-white border-b border-ink-200 flex gap-2">
    <button
      v-for="t in tabs"
      :key="t.id"
      class="text-xs px-3 py-1.5 rounded-md border border-ink-200"
      :class="tab === t.id ? 'bg-brand-600 text-white border-brand-600' : 'bg-white text-ink-500 hover:bg-ink-100'"
      @click="tab = t.id"
    >{{ t.label }}</button>
  </div>

  <section class="flex-1 overflow-y-auto p-6 space-y-4">
    <div v-if="loading" class="text-xs text-ink-400">Loading…</div>

    <template v-else>
      <template v-if="tab === 'publications'">
        <div v-for="[region, list] in pubsByRegion" :key="region" class="space-y-2">
          <div class="text-xxs font-semibold text-ink-400 uppercase tracking-wider">{{ region }}</div>
          <ul class="dcgg-card !p-0 divide-y divide-ink-100">
            <li v-for="p in list" :key="p.id" class="px-4 py-2 flex items-center gap-3">
              <span class="dcgg-tag capitalize">{{ p.kind }}</span>
              <div class="flex-1 min-w-0">
                <div class="text-sm text-ink-900 truncate">{{ p.title }}</div>
                <div class="text-xxs text-ink-500">{{ format(new Date(p.publishedAt), 'PP') }}</div>
              </div>
            </li>
          </ul>
        </div>
      </template>

      <template v-else>
        <div
          v-for="[region, list] in (tab === 'roundtables' ? roundtables : tab === 'meetings' ? meetings : governance)"
          :key="region"
          class="space-y-2"
        >
          <div class="text-xxs font-semibold text-ink-400 uppercase tracking-wider">{{ region }}</div>
          <ul class="dcgg-card !p-0 divide-y divide-ink-100">
            <li v-for="a in list" :key="a.id" class="px-4 py-2">
              <div class="text-sm text-ink-900">{{ a.title }}</div>
              <div class="text-xxs text-ink-500">{{ format(new Date(a.occurredAt), 'PPp') }}</div>
            </li>
          </ul>
        </div>
      </template>
    </template>
  </section>
</template>
