<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { activities, publications } from '@/api';
import type { Activity, Publication } from '@/types';
import { format } from 'date-fns';

type Tab = 'events' | 'meetings' | 'publications' | 'consultations';
const tab = ref<Tab>('events');
const tabs: Array<{ id: Tab; label: string }> = [
  { id: 'events',        label: 'Events' },
  { id: 'meetings',      label: 'Meetings' },
  { id: 'publications',  label: 'Publications' },
  { id: 'consultations', label: 'Consultations' },
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

function groupByVertical<T extends Record<string, any>>(items: T[]): Array<[string, T[]]> {
  const m = new Map<string, T[]>();
  for (const i of items) {
    const v = (i.vertical as string | undefined) ?? 'general';
    if (!m.has(v)) m.set(v, []);
    m.get(v)!.push(i);
  }
  return Array.from(m.entries()).sort(([a], [b]) => a.localeCompare(b));
}

const events     = computed(() => groupByVertical(acts.value.filter((a) => (a.type as string) === 'event')));
const meetings   = computed(() => groupByVertical(acts.value.filter((a) => (a.type as string) === 'meeting')));
const consults   = computed(() => groupByVertical(acts.value.filter((a) => (a.type as string) === 'consultation')));
const pubsByVert = computed(() => groupByVertical(pubs.value));
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
        <div v-if="!pubs.length" class="dcgg-card text-xs text-ink-400">No publications.</div>
        <div v-for="[vert, list] in pubsByVert" :key="vert" class="space-y-2">
          <div class="text-xxs font-semibold text-ink-400 uppercase tracking-wider">{{ vert }}</div>
          <ul class="dcgg-card !p-0 divide-y divide-ink-100">
            <li v-for="p in list" :key="p.id" class="px-4 py-2">
              <div class="text-sm text-ink-900 truncate">{{ p.title }}</div>
              <div class="text-xxs text-ink-500">
                {{ p.publishedAt ? format(new Date(p.publishedAt), 'PP') : '—' }}
                <span v-if="(p as any).venue"> · {{ (p as any).venue }}</span>
              </div>
            </li>
          </ul>
        </div>
      </template>

      <template v-else>
        <div
          v-for="[vert, list] in (tab === 'events' ? events : tab === 'meetings' ? meetings : consults)"
          :key="vert"
          class="space-y-2"
        >
          <div class="text-xxs font-semibold text-ink-400 uppercase tracking-wider">{{ vert }}</div>
          <ul class="dcgg-card !p-0 divide-y divide-ink-100">
            <li v-for="a in list" :key="a.id" class="px-4 py-2">
              <div class="text-sm text-ink-900">{{ a.title }}</div>
              <div class="text-xxs text-ink-500">{{ format(new Date(a.occurredAt), 'PPp') }}</div>
            </li>
          </ul>
        </div>
        <div v-if="(tab === 'events' ? events : tab === 'meetings' ? meetings : consults).length === 0"
             class="dcgg-card text-xs text-ink-400">No items.</div>
      </template>
    </template>
  </section>
</template>
