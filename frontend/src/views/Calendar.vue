<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { consultations, activities } from '@/api';
import {
  addMonths, endOfMonth, endOfWeek, format, isSameDay, isSameMonth,
  startOfMonth, startOfWeek, eachDayOfInterval,
} from 'date-fns';

const cursor = ref(new Date());
const events = ref<Array<{ id: string; title: string; at: Date; kind: 'consultation' | 'activity' }>>([]);
const loading = ref(true);

async function load() {
  loading.value = true;
  try {
    const [cs, as] = await Promise.all([
      consultations.list().catch(() => []),
      activities.list().catch(() => []),
    ]);
    events.value = [
      ...cs.map((c) => ({ id: c.id, title: c.title, at: new Date(c.deadlineAt), kind: 'consultation' as const })),
      ...as.map((a) => ({ id: a.id, title: a.title, at: new Date(a.occurredAt), kind: 'activity' as const })),
    ];
  } finally { loading.value = false; }
}
onMounted(load);

const days = computed(() => {
  const start = startOfWeek(startOfMonth(cursor.value), { weekStartsOn: 1 });
  const end   = endOfWeek(endOfMonth(cursor.value),   { weekStartsOn: 1 });
  return eachDayOfInterval({ start, end });
});

function eventsOn(d: Date) {
  return events.value.filter((e) => isSameDay(e.at, d));
}
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200 flex items-center gap-3">
    <div class="flex-1">
      <h1 class="text-base font-semibold text-ink-900">Calendar</h1>
      <p class="text-xs text-ink-500 mt-0.5">Consultations & activities overlaid on a month grid.</p>
    </div>
    <button class="dcgg-btn" @click="cursor = addMonths(cursor, -1)">‹</button>
    <div class="text-sm font-semibold min-w-[140px] text-center">{{ format(cursor, 'LLLL yyyy') }}</div>
    <button class="dcgg-btn" @click="cursor = addMonths(cursor, 1)">›</button>
    <button class="dcgg-btn" @click="cursor = new Date()">Today</button>
  </header>

  <section class="flex-1 overflow-y-auto p-6">
    <div class="grid grid-cols-7 gap-px bg-ink-200 border border-ink-200 rounded-lg overflow-hidden">
      <div
        v-for="d in ['Mon','Tue','Wed','Thu','Fri','Sat','Sun']"
        :key="d"
        class="bg-ink-50 text-xxs font-semibold text-ink-500 uppercase tracking-wider px-2 py-1.5"
      >{{ d }}</div>

      <div
        v-for="d in days"
        :key="d.toISOString()"
        class="bg-white min-h-[96px] p-1.5 text-xs"
        :class="!isSameMonth(d, cursor) && 'bg-ink-50 text-ink-400'"
      >
        <div class="flex items-center justify-between">
          <span :class="isSameDay(d, new Date()) && 'bg-brand-600 text-white rounded-full w-5 h-5 grid place-items-center text-xxs font-semibold'">
            {{ format(d, 'd') }}
          </span>
        </div>
        <ul class="mt-1 space-y-0.5">
          <li
            v-for="e in eventsOn(d).slice(0, 3)"
            :key="e.id"
            class="truncate text-xxs px-1 py-0.5 rounded"
            :class="e.kind === 'consultation' ? 'bg-amber-50 text-warn' : 'bg-brand-50 text-brand-600'"
            :title="e.title"
          >{{ e.title }}</li>
          <li v-if="eventsOn(d).length > 3" class="text-xxs text-ink-400">
            +{{ eventsOn(d).length - 3 }} more
          </li>
        </ul>
      </div>
    </div>
    <div v-if="loading" class="text-xs text-ink-400 mt-3">Loading events…</div>
  </section>
</template>
