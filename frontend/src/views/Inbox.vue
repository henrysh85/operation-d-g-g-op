<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { inbox } from '@/api';
import type { InboxItem } from '@/api/inbox';
import { formatDistanceToNow } from 'date-fns';

const router = useRouter();
const items = ref<InboxItem[]>([]);
const loading = ref(true);

async function load() {
  loading.value = true;
  try { items.value = (await inbox.tasks()).items; }
  finally { loading.value = false; }
}
onMounted(load);

const groups = computed(() => {
  const m = new Map<string, InboxItem[]>();
  for (const it of items.value) {
    const k = it.kind;
    if (!m.has(k)) m.set(k, []);
    m.get(k)!.push(it);
  }
  return Array.from(m.entries());
});

const headings: Record<string, string> = {
  holiday_pending:         'Holiday requests awaiting your decision',
  expense_pending:         'Expense claims awaiting your decision',
  consultation_deadline:   'Consultations assigned to you (≤ 60 days)',
  holiday_decided:         'Recent decisions on your holidays',
};

function go(it: InboxItem) {
  if (it.link) router.push(it.link);
}
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200 flex items-center">
    <div class="flex-1">
      <h1 class="text-base font-semibold text-ink-900">Inbox</h1>
      <p class="text-xs text-ink-500 mt-0.5">Approvals, assignments and decisions that concern you.</p>
    </div>
    <button class="dcgg-btn" @click="load">Refresh</button>
  </header>

  <section class="flex-1 overflow-y-auto p-6 space-y-6">
    <div v-if="loading" class="text-xs text-ink-400">Loading…</div>
    <div v-else-if="!items.length" class="dcgg-card text-xs text-ink-400">
      Inbox zero. Nothing is waiting on you.
    </div>

    <div v-for="[kind, list] in groups" :key="kind">
      <div class="text-xxs font-semibold text-ink-400 uppercase tracking-wider mb-2 pb-1 border-b border-ink-100">
        {{ headings[kind] ?? kind }} · {{ list.length }}
      </div>
      <ul class="dcgg-card !p-0 divide-y divide-ink-100">
        <li
          v-for="it in list"
          :key="it.kind + it.id"
          class="px-4 py-2 flex items-baseline gap-3 cursor-pointer hover:bg-ink-50"
          @click="go(it)"
        >
          <span class="text-sm text-ink-900 flex-1 truncate">{{ it.title }}</span>
          <span class="text-xxs text-ink-500 truncate">{{ it.detail }}</span>
          <span class="text-xxs text-ink-400">{{ formatDistanceToNow(new Date(it.when), { addSuffix: true }) }}</span>
        </li>
      </ul>
    </div>
  </section>
</template>
