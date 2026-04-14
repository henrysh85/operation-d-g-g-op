<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { membership, activities } from '@/api';
import type { Member, Activity } from '@/types';
import type { MemberIntel } from '@/api/membership';
import StatusBadge from '@/components/StatusBadge.vue';
import { format } from 'date-fns';

const props = defineProps<{ id: string }>();

type Tab = 'overview' | 'risks' | 'activity' | 'intelligence';
const tab = ref<Tab>('overview');
const tabs: Array<{ id: Tab; label: string }> = [
  { id: 'overview',     label: 'Overview' },
  { id: 'risks',        label: 'Risks' },
  { id: 'activity',     label: 'Activity' },
  { id: 'intelligence', label: 'Intelligence' },
];

const member = ref<Member | null>(null);
const acts = ref<Activity[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

const intel = ref<MemberIntel | null>(null);

onMounted(async () => {
  try {
    member.value = await membership.get(props.id);
    acts.value = await activities.list({ memberId: props.id }).catch(() => []);
  } catch (e) { error.value = (e as Error).message; }
  finally { loading.value = false; }
});

async function loadIntel() {
  if (!member.value || intel.value) return;
  intel.value = await membership.intel(member.value.id).catch(() => null);
}
watch(tab, (t) => { if (t === 'risks' || t === 'intelligence') loadIntel(); });
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200 flex gap-3 items-start">
    <router-link to="/members" class="text-xxs text-ink-500 hover:text-brand-600">&larr; Members</router-link>
    <div class="flex-1">
      <h1 class="text-base font-semibold text-ink-900">
        {{ member?.legalName ?? (loading ? 'Loading…' : 'Member') }}
      </h1>
      <p v-if="member" class="text-xs text-ink-500 mt-0.5 flex items-center gap-2">
        <StatusBadge :status="member.status" />
        <span v-if="member.tier">· {{ member.tier }} tier</span>
      </p>
    </div>
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

  <section class="flex-1 overflow-y-auto p-6">
    <div v-if="error" class="dcgg-card text-xs text-err">{{ error }}</div>
    <div v-else-if="loading" class="text-xs text-ink-400">Loading…</div>

    <template v-else-if="member">
      <div v-if="tab === 'overview'" class="dcgg-card">
        <dl class="grid grid-cols-2 gap-y-2 text-sm">
          <dt class="text-ink-500">Legal name</dt><dd>{{ member.legalName }}</dd>
          <dt class="text-ink-500">Status</dt><dd><StatusBadge :status="member.status" /></dd>
          <dt class="text-ink-500">Tier</dt><dd>{{ member.tier ?? '—' }}</dd>
          <dt class="text-ink-500">Joined</dt>
          <dd>{{ member.joinedAt ? format(new Date(member.joinedAt), 'PP') : '—' }}</dd>
          <dt class="text-ink-500">Risk score</dt><dd>{{ member.riskScore ?? '—' }}</dd>
        </dl>
      </div>

      <div v-else-if="tab === 'risks'" class="space-y-3">
        <div v-if="!intel" class="text-xs text-ink-400">Loading risks…</div>
        <template v-else>
          <div class="dcgg-card">
            <div class="text-sm font-semibold mb-3">Risk components</div>
            <div class="space-y-3">
              <div v-for="r in intel.risks" :key="r.key">
                <div class="flex items-baseline justify-between">
                  <span class="text-xs text-ink-700">{{ r.label }}</span>
                  <span class="text-xs font-semibold tabular-nums">{{ r.value }}</span>
                </div>
                <div class="mt-1 h-1.5 bg-ink-100 rounded">
                  <div class="h-full rounded" :style="{ width: Math.min(100, r.value) + '%' }"
                       :class="r.value > 60 ? 'bg-err' : r.value > 30 ? 'bg-warn' : 'bg-ok'" />
                </div>
                <div class="text-xxs text-ink-500 mt-0.5">{{ r.note }}</div>
              </div>
            </div>
          </div>
        </template>
      </div>

      <div v-else-if="tab === 'activity'">
        <ul v-if="acts.length" class="dcgg-card !p-0 divide-y divide-ink-100">
          <li v-for="a in acts" :key="a.id" class="px-4 py-2">
            <div class="text-sm text-ink-900">{{ a.title }}</div>
            <div class="text-xxs text-ink-500">{{ a.type }} · {{ format(new Date(a.occurredAt), 'PPp') }}</div>
          </li>
        </ul>
        <div v-else class="dcgg-card text-xs text-ink-400">No activity recorded.</div>
      </div>

      <div v-else-if="tab === 'intelligence'" class="space-y-3">
        <div v-if="!intel" class="text-xs text-ink-400">Loading intelligence…</div>
        <template v-else>
          <div class="dcgg-card">
            <div class="text-sm font-semibold mb-3">Open consultations in jurisdiction</div>
            <div v-if="!intel.openConsults.length" class="text-xs text-ink-400">None.</div>
            <ul v-else class="divide-y divide-ink-100">
              <li v-for="r in intel.openConsults" :key="r.id" class="py-1.5 flex items-baseline gap-2">
                <span class="text-sm text-ink-900 flex-1 truncate">{{ r.title }}</span>
                <span class="text-xxs text-ink-500">{{ r.tag }}</span>
                <span class="text-xxs text-ink-400">{{ r.when?.slice(0,10) || '—' }}</span>
              </li>
            </ul>
          </div>
          <div class="dcgg-card">
            <div class="text-sm font-semibold mb-3">Recent regulatory changes</div>
            <div v-if="!intel.recentRegChanges.length" class="text-xs text-ink-400">None.</div>
            <ul v-else class="divide-y divide-ink-100">
              <li v-for="r in intel.recentRegChanges" :key="r.id" class="py-1.5 flex items-baseline gap-2">
                <span class="text-sm text-ink-900 flex-1 truncate">{{ r.title }}</span>
                <span class="text-xxs text-ink-500">{{ r.tag }}</span>
                <span class="text-xxs text-ink-400">{{ r.when?.slice(0,10) || '—' }}</span>
              </li>
            </ul>
          </div>
          <div class="dcgg-card">
            <div class="text-sm font-semibold mb-3">Recent activities</div>
            <div v-if="!intel.activities.length" class="text-xs text-ink-400">None.</div>
            <ul v-else class="divide-y divide-ink-100">
              <li v-for="r in intel.activities" :key="r.id" class="py-1.5 flex items-baseline gap-2">
                <span class="text-sm text-ink-900 flex-1 truncate">{{ r.title }}</span>
                <span class="text-xxs text-ink-500">{{ r.tag }}</span>
                <span class="text-xxs text-ink-400">{{ r.when?.slice(0,10) || '—' }}</span>
              </li>
            </ul>
          </div>
        </template>
      </div>
    </template>
  </section>
</template>
