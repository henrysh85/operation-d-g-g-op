<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{ status: string }>();

const map: Record<string, string> = {
  // generic
  live:      'bg-emerald-50 text-ok border-emerald-200',
  active:    'bg-amber-50 text-warn border-amber-200',
  open:      'bg-blue-50 text-brand-600 border-blue-200',
  responded: 'bg-emerald-50 text-ok border-emerald-200',
  draft:     'bg-ink-100 text-ink-500 border-ink-200',
  archived:  'bg-ink-100 text-ink-400 border-ink-200',
  closed:    'bg-ink-100 text-ink-500 border-ink-200',
  early:     'bg-ink-100 text-ink-500 border-ink-200',
  none:      'bg-ink-100 text-ink-400 border-ink-200',
  done:      'bg-emerald-50 text-ok border-emerald-200',
  // membership
  prospect:  'bg-blue-50 text-brand-600 border-blue-200',
  applied:   'bg-amber-50 text-warn border-amber-200',
  lapsed:    'bg-red-50 text-err border-red-200',
  // consultations / submissions
  submitted: 'bg-blue-50 text-brand-600 border-blue-200',
  withdrawn: 'bg-ink-100 text-ink-500 border-ink-200',
  // approval workflow (holidays / expenses)
  pending:   'bg-amber-50 text-warn border-amber-200',
  approved:  'bg-emerald-50 text-ok border-emerald-200',
  rejected:  'bg-red-50 text-err border-red-200',
  paid:      'bg-emerald-50 text-ok border-emerald-200',
};

const cls = computed(() => map[props.status] ?? 'bg-ink-100 text-ink-500 border-ink-200');
</script>

<template>
  <span class="inline-flex items-center gap-1 text-xxs px-1.5 py-0.5 rounded border font-medium capitalize"
        :class="cls">
    <span class="w-1.5 h-1.5 rounded-full bg-current opacity-70" />
    {{ status }}
  </span>
</template>
