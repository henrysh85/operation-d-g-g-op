<script setup lang="ts" generic="T extends Record<string, unknown>">
import { computed, ref } from 'vue';

interface Column<R> {
  key: string;
  label: string;
  get?: (row: R) => unknown;
  width?: string;
  class?: string;
}

const props = defineProps<{
  rows: T[];
  columns: Column<T>[];
  loading?: boolean;
  empty?: string;
  groupBy?: (row: T) => string;
  rowKey?: (row: T) => string | number;
}>();

const emit = defineEmits<{ (e: 'row-click', row: T): void }>();

const openGroups = ref<Record<string, boolean>>({});

const grouped = computed(() => {
  if (!props.groupBy) return null;
  const g = new Map<string, T[]>();
  for (const r of props.rows) {
    const k = props.groupBy!(r);
    if (!g.has(k)) g.set(k, []);
    g.get(k)!.push(r);
  }
  return Array.from(g.entries()).map(([name, items]) => ({ name, items }));
});

function key(r: T, i: number) {
  return props.rowKey ? props.rowKey(r) : ((r as { id?: string | number }).id ?? i);
}
function val(r: T, c: Column<T>) {
  return c.get ? c.get(r) : (r as Record<string, unknown>)[c.key];
}
function toggle(g: string) {
  openGroups.value[g] = openGroups.value[g] === false ? true : false;
}
</script>

<template>
  <div class="overflow-auto border border-ink-200 rounded-lg bg-white">
    <table class="dcgg-table">
      <thead>
        <tr>
          <th v-for="c in columns" :key="c.key" :style="c.width ? { width: c.width } : undefined" :class="c.class">
            {{ c.label }}
          </th>
        </tr>
      </thead>

      <tbody v-if="loading">
        <tr v-for="i in 6" :key="i">
          <td v-for="c in columns" :key="c.key">
            <div class="h-3 bg-ink-100 rounded animate-pulse w-3/4" />
          </td>
        </tr>
      </tbody>

      <tbody v-else-if="grouped">
        <template v-for="g in grouped" :key="g.name">
          <tr class="bg-ink-50 cursor-pointer" @click="toggle(g.name)">
            <td :colspan="columns.length" class="!py-1.5 !text-xxs !uppercase !font-semibold !text-ink-500">
              <span class="inline-block w-3">{{ openGroups[g.name] === false ? '▸' : '▾' }}</span>
              {{ g.name }} <span class="text-ink-400 normal-case">({{ g.items.length }})</span>
            </td>
          </tr>
          <tr
            v-for="(r, i) in (openGroups[g.name] === false ? [] : g.items)"
            :key="key(r, i)"
            class="cursor-pointer"
            @click="emit('row-click', r)"
          >
            <td v-for="c in columns" :key="c.key" :class="c.class">
              <slot :name="`cell-${c.key}`" :row="r" :value="val(r, c)">{{ val(r, c) }}</slot>
            </td>
          </tr>
        </template>
      </tbody>

      <tbody v-else-if="rows.length">
        <tr v-for="(r, i) in rows" :key="key(r, i)" class="cursor-pointer" @click="emit('row-click', r)">
          <td v-for="c in columns" :key="c.key" :class="c.class">
            <slot :name="`cell-${c.key}`" :row="r" :value="val(r, c)">{{ val(r, c) }}</slot>
          </td>
        </tr>
      </tbody>

      <tbody v-else>
        <tr>
          <td :colspan="columns.length" class="text-center text-ink-400 py-10">
            {{ empty ?? 'No records' }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
