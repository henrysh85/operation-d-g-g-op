<script setup lang="ts">
/**
 * Minimal pan/zoomable SVG org chart skeleton.
 * Accepts a flat list of Persons and lays them out by manager chain.
 * Full hierarchical layout is deferred — this gives a functional baseline.
 */
import { computed, ref } from 'vue';
import type { Person } from '@/types';

const props = defineProps<{ people: Person[] }>();

// Simple level assignment: roots (no manager) = 0; children stack by depth.
const nodes = computed(() => {
  const byId = new Map(props.people.map((p) => [p.id, p]));
  const depth = new Map<string, number>();
  const level = (id: string): number => {
    if (depth.has(id)) return depth.get(id)!;
    const p = byId.get(id);
    if (!p || !p.managerId) { depth.set(id, 0); return 0; }
    const d = level(p.managerId) + 1;
    depth.set(id, d);
    return d;
  };
  const byLevel = new Map<number, Person[]>();
  for (const p of props.people) {
    const d = level(p.id);
    if (!byLevel.has(d)) byLevel.set(d, []);
    byLevel.get(d)!.push(p);
  }
  const NODE_W = 160, NODE_H = 56, GAP_X = 24, GAP_Y = 40;
  const placed: Array<Person & { x: number; y: number }> = [];
  for (const [d, list] of byLevel) {
    list.forEach((p, i) => {
      placed.push({ ...p, x: i * (NODE_W + GAP_X), y: d * (NODE_H + GAP_Y) });
    });
  }
  return placed;
});

const edges = computed(() =>
  nodes.value
    .filter((n) => n.managerId)
    .map((n) => {
      const m = nodes.value.find((x) => x.id === n.managerId);
      if (!m) return null;
      return {
        id: `${m.id}->${n.id}`,
        x1: m.x + 80, y1: m.y + 56,
        x2: n.x + 80, y2: n.y,
      };
    })
    .filter((e): e is NonNullable<typeof e> => !!e),
);

// Pan/zoom
const tx = ref(40);
const ty = ref(20);
const scale = ref(1);
const dragging = ref(false);
let startX = 0, startY = 0, startTx = 0, startTy = 0;

function down(e: PointerEvent) {
  dragging.value = true;
  startX = e.clientX; startY = e.clientY;
  startTx = tx.value; startTy = ty.value;
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
}
function move(e: PointerEvent) {
  if (!dragging.value) return;
  tx.value = startTx + (e.clientX - startX);
  ty.value = startTy + (e.clientY - startY);
}
function up() { dragging.value = false; }
function wheel(e: WheelEvent) {
  e.preventDefault();
  const factor = e.deltaY > 0 ? 0.9 : 1.1;
  scale.value = Math.min(2.5, Math.max(0.4, scale.value * factor));
}
</script>

<template>
  <div
    class="relative w-full h-full bg-ink-50 border border-ink-200 rounded-lg overflow-hidden select-none"
    @pointerdown="down"
    @pointermove="move"
    @pointerup="up"
    @wheel="wheel"
  >
    <svg class="absolute inset-0 w-full h-full" :style="{ cursor: dragging ? 'grabbing' : 'grab' }">
      <g :transform="`translate(${tx} ${ty}) scale(${scale})`">
        <line
          v-for="e in edges"
          :key="e.id"
          :x1="e.x1" :y1="e.y1" :x2="e.x2" :y2="e.y2"
          stroke="#d1d5db" stroke-width="1"
        />
        <g v-for="n in nodes" :key="n.id" :transform="`translate(${n.x} ${n.y})`">
          <rect width="160" height="56" rx="8" fill="#fff" stroke="#e5e7eb" />
          <text x="12" y="22" font-size="12" font-weight="600" fill="#111827">
            {{ n.fullName.slice(0, 22) }}
          </text>
          <text x="12" y="40" font-size="11" fill="#6b7280">
            {{ (n.role ?? '').slice(0, 26) }}
          </text>
        </g>
      </g>
    </svg>
    <div class="absolute bottom-3 right-3 flex gap-1 bg-white border border-ink-200 rounded-md p-1 shadow-card">
      <button class="dcgg-btn !py-0.5 !px-2" @click="scale = Math.max(0.4, scale * 0.9)">−</button>
      <button class="dcgg-btn !py-0.5 !px-2" @click="scale = Math.min(2.5, scale * 1.1)">+</button>
      <button class="dcgg-btn !py-0.5 !px-2" @click="tx = 40; ty = 20; scale = 1">Reset</button>
    </div>
  </div>
</template>
