<script setup lang="ts">
import { onMounted, ref, computed, reactive } from 'vue';
import { templates } from '@/api';
import type { TemplateFull } from '@/api/templates';
import Modal from '@/components/Modal.vue';
import { format } from 'date-fns';

const rows = ref<TemplateFull[]>([]);
const loading = ref(true);
const search = ref('');
const selected = ref<TemplateFull | null>(null);
const params = reactive<Record<string, string>>({});
const rendered = ref<string | null>(null);
const renderError = ref<string | null>(null);
const rendering = ref(false);

onMounted(async () => {
  try { rows.value = await templates.list(); }
  finally { loading.value = false; }
});

const byCategory = computed(() => {
  const q = search.value.toLowerCase();
  const list = rows.value.filter((t) => !q || t.name.toLowerCase().includes(q));
  const m = new Map<string, TemplateFull[]>();
  for (const t of list) {
    if (!m.has(t.category)) m.set(t.category, []);
    m.get(t.category)!.push(t);
  }
  return Array.from(m.entries());
});

function open(t: TemplateFull) {
  selected.value = t;
  rendered.value = null;
  renderError.value = null;
  for (const k of Object.keys(params)) delete params[k];
  for (const v of t.variables) params[v.name] = '';
}

async function preview() {
  if (!selected.value) return;
  rendering.value = true;
  renderError.value = null;
  try {
    rendered.value = await templates.render(selected.value.id, params);
  } catch (e) { renderError.value = (e as Error).message; }
  finally { rendering.value = false; }
}

function closeModal() {
  selected.value = null;
  rendered.value = null;
}
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200 flex gap-3 items-start">
    <div class="flex-1">
      <h1 class="text-base font-semibold text-ink-900">Templates</h1>
      <p class="text-xs text-ink-500 mt-0.5">Document and response templates by category. Click to preview.</p>
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
        <button
          v-for="t in items"
          :key="t.id"
          class="dcgg-card hover:border-brand-200 transition-colors text-left"
          @click="open(t)"
        >
          <div class="flex items-start gap-3">
            <div class="w-9 h-9 rounded-md bg-brand-50 text-brand-600 grid place-items-center text-xs font-bold">
              {{ t.name[0] }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm font-semibold text-ink-900 truncate">{{ t.name }}</div>
              <div class="text-xxs text-ink-500 mt-0.5">
                {{ t.variables.length }} variable{{ t.variables.length === 1 ? '' : 's' }} · updated {{ format(new Date(t.updatedAt), 'PP') }}
              </div>
              <div v-if="t.tags.length" class="mt-2 flex flex-wrap gap-1">
                <span v-for="tag in t.tags" :key="tag" class="dcgg-tag">{{ tag }}</span>
              </div>
            </div>
          </div>
        </button>
      </div>
    </div>

    <Modal :open="!!selected" :title="selected?.name ?? ''" width="640px" @close="closeModal">
      <template v-if="selected">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <div class="text-xxs font-semibold text-ink-500 uppercase mb-2">Variables</div>
            <div v-if="!selected.variables.length" class="text-xs text-ink-400">No variables defined.</div>
            <div v-else class="space-y-2">
              <label v-for="v in selected.variables" :key="v.name" class="block">
                <span class="text-xxs font-semibold text-ink-700">
                  {{ v.name }}<span v-if="v.required" class="text-err">*</span>
                </span>
                <input
                  v-model="params[v.name]"
                  :type="v.type === 'date' ? 'date' : 'text'"
                  class="dcgg-input w-full mt-0.5"
                />
              </label>
            </div>
            <button class="dcgg-btn-primary mt-3" :disabled="rendering" @click="preview">
              {{ rendering ? 'Rendering…' : 'Preview' }}
            </button>
            <div v-if="renderError" class="text-xs text-err mt-2">{{ renderError }}</div>
          </div>
          <div>
            <div class="text-xxs font-semibold text-ink-500 uppercase mb-2">Output</div>
            <pre v-if="rendered" class="text-xs whitespace-pre-wrap bg-ink-50 p-3 rounded-md max-h-[420px] overflow-auto">{{ rendered }}</pre>
            <pre v-else class="text-xs whitespace-pre-wrap bg-ink-50 p-3 rounded-md max-h-[420px] overflow-auto text-ink-500">{{ selected.body }}</pre>
          </div>
        </div>
      </template>
    </Modal>
  </section>
</template>
