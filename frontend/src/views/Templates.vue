<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { templates } from '@/api';
import type { Template } from '@/types';
import { format } from 'date-fns';

const rows = ref<Template[]>([]);
const loading = ref(true);
const search = ref('');

onMounted(async () => {
  try { rows.value = await templates.list(); }
  finally { loading.value = false; }
});

const byCategory = computed(() => {
  const q = search.value.toLowerCase();
  const list = rows.value.filter((t) => !q || t.name.toLowerCase().includes(q));
  const m = new Map<string, Template[]>();
  for (const t of list) {
    if (!m.has(t.category)) m.set(t.category, []);
    m.get(t.category)!.push(t);
  }
  return Array.from(m.entries());
});
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200 flex gap-3 items-start">
    <div class="flex-1">
      <h1 class="text-base font-semibold text-ink-900">Templates</h1>
      <p class="text-xs text-ink-500 mt-0.5">Document and response templates by category.</p>
    </div>
    <input v-model="search" type="search" placeholder="Search templates…" class="dcgg-input w-64" />
  </header>

  <section class="flex-1 overflow-y-auto p-6 space-y-6">
    <div v-if="loading" class="text-xs text-ink-400">Loading…</div>
    <div v-else-if="!rows.length" class="dcgg-card text-xs text-ink-400">No templates uploaded yet.</div>

    <div v-for="[cat, items] in byCategory" :key="cat">
      <div class="text-xxs font-semibold text-ink-400 uppercase tracking-wider mb-2 pb-1 border-b border-ink-100">
        {{ cat }}
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        <article v-for="t in items" :key="t.id" class="dcgg-card hover:border-brand-200 transition-colors">
          <div class="flex items-start gap-3">
            <div class="w-9 h-9 rounded-md bg-brand-50 text-brand-600 grid place-items-center text-xs font-bold">
              {{ t.name[0] }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm font-semibold text-ink-900 truncate">{{ t.name }}</div>
              <div class="text-xxs text-ink-500 mt-0.5">
                {{ t.owner ?? '—' }} · updated {{ format(new Date(t.updatedAt), 'PP') }}
              </div>
              <p v-if="t.description" class="text-xs text-ink-700 mt-2 line-clamp-2">{{ t.description }}</p>
              <a v-if="t.downloadUrl" :href="t.downloadUrl" class="text-xxs text-brand-600 mt-2 inline-block">
                Download →
              </a>
            </div>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
