<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { membership, activities } from '@/api';
import type { Member, Activity } from '@/types';
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

onMounted(async () => {
  try {
    member.value = await membership.get(props.id);
    acts.value = await activities.list({ memberId: props.id }).catch(() => []);
  } catch (e) { error.value = (e as Error).message; }
  finally { loading.value = false; }
});
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

      <div v-else-if="tab === 'risks'" class="dcgg-card text-xs text-ink-500">
        Risk drilldown — to be wired to <code>/api/members/{{ member.id }}/risks</code>.
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

      <div v-else-if="tab === 'intelligence'" class="dcgg-card text-xs text-ink-500">
        Intelligence — to be wired to <code>/api/members/{{ member.id }}/intelligence</code>.
      </div>
    </template>
  </section>
</template>
