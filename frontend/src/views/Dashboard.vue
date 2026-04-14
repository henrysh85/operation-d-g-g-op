<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { dashboard } from '@/api';
import type { DashboardSummary } from '@/api/dashboard';
import FilterBar from '@/components/FilterBar.vue';
import ImpactBadge from '@/components/ImpactBadge.vue';
import { format } from 'date-fns';

const loading = ref(true);
const error = ref<string | null>(null);
const data = ref<DashboardSummary | null>(null);

onMounted(async () => {
  try { data.value = await dashboard.summary(); }
  catch (e) { error.value = (e as Error).message; }
  finally   { loading.value = false; }
});
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">Dashboard</h1>
    <p class="text-xs text-ink-500 mt-0.5">Intelligence overview across all clients and jurisdictions.</p>
  </header>

  <FilterBar />

  <section class="flex-1 overflow-y-auto p-6 space-y-6">
    <div v-if="error" class="dcgg-card text-xs text-err">{{ error }}</div>

    <!-- KPI grid -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <div v-if="loading" v-for="i in 4" :key="i" class="dcgg-card h-20 animate-pulse" />
      <div v-else v-for="k in (data?.kpis ?? [])" :key="k.key" class="dcgg-card">
        <div class="text-xxs uppercase tracking-wider text-ink-500 font-semibold">{{ k.label }}</div>
        <div class="text-xl font-bold text-ink-900 mt-1">{{ k.value }}<span v-if="k.unit" class="text-xs text-ink-500 ml-1">{{ k.unit }}</span></div>
        <div v-if="k.delta" class="text-xxs mt-1" :class="k.delta >= 0 ? 'text-ok' : 'text-err'">
          {{ k.delta >= 0 ? '▲' : '▼' }} {{ Math.abs(k.delta) }} vs prev period
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div class="dcgg-card">
        <div class="text-sm font-semibold mb-3">Upcoming deadlines</div>
        <div v-if="loading" class="text-xs text-ink-400">Loading…</div>
        <div v-else-if="!data?.upcomingDeadlines.length" class="text-xs text-ink-400">No upcoming deadlines.</div>
        <ul v-else class="divide-y divide-ink-100">
          <li v-for="c in data.upcomingDeadlines" :key="c.id" class="py-2 flex items-center gap-3">
            <div class="flex-1 min-w-0">
              <div class="text-sm text-ink-900 truncate">{{ c.title }}</div>
              <div class="text-xxs text-ink-500">{{ c.regulator }} · {{ format(new Date(c.deadlineAt), 'PP') }}</div>
            </div>
            <ImpactBadge :impact="c.impact" />
          </li>
        </ul>
      </div>

      <div class="dcgg-card">
        <div class="text-sm font-semibold mb-3">Recent activity</div>
        <div v-if="loading" class="text-xs text-ink-400">Loading…</div>
        <div v-else-if="!data?.recentActivity.length" class="text-xs text-ink-400">No recent activity.</div>
        <ul v-else class="divide-y divide-ink-100">
          <li v-for="a in data.recentActivity" :key="a.id" class="py-2">
            <div class="text-sm text-ink-900">{{ a.title }}</div>
            <div class="text-xxs text-ink-500">
              {{ a.type }} · {{ format(new Date(a.occurredAt), 'PPp') }}
            </div>
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>
