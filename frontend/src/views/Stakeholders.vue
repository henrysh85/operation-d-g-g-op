<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { stakeholders } from '@/api';

type Tree = Awaited<ReturnType<typeof stakeholders.tree>>;

const tree = ref<Tree>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const openRegion  = ref<Record<string, boolean>>({});
const openCountry = ref<Record<string, boolean>>({});

onMounted(async () => {
  try { tree.value = await stakeholders.tree(); }
  catch (e) { error.value = (e as Error).message; }
  finally   { loading.value = false; }
});
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">Stakeholders</h1>
    <p class="text-xs text-ink-500 mt-0.5">Regulators, trade bodies and key contacts by region.</p>
  </header>

  <section class="flex-1 overflow-y-auto p-6">
    <div v-if="loading" class="text-xs text-ink-400">Loading stakeholders…</div>
    <div v-else-if="error" class="dcgg-card text-xs text-err">{{ error }}</div>
    <div v-else-if="!tree.length" class="dcgg-card text-xs text-ink-400">No stakeholders yet.</div>

    <div v-else class="space-y-3">
      <div v-for="region in tree" :key="region.region" class="dcgg-card !p-0 overflow-hidden">
        <button
          class="w-full flex items-center gap-2 px-4 py-3 hover:bg-ink-50 text-left"
          @click="openRegion[region.region] = !openRegion[region.region]"
        >
          <span class="text-xxs text-ink-400">{{ openRegion[region.region] === false ? '▸' : '▾' }}</span>
          <span class="text-sm font-semibold text-ink-900 flex-1">{{ region.region }}</span>
          <span class="text-xxs text-ink-400">{{ region.countries.length }} countries</span>
        </button>

        <div v-if="openRegion[region.region] !== false" class="border-t border-ink-100 divide-y divide-ink-100">
          <div v-for="c in region.countries" :key="c.countryCode" class="px-4 py-2">
            <button
              class="w-full flex items-center gap-2 text-left"
              @click="openCountry[region.region + c.countryCode] = !openCountry[region.region + c.countryCode]"
            >
              <span class="text-xxs text-ink-400">
                {{ openCountry[region.region + c.countryCode] === false ? '▸' : '▾' }}
              </span>
              <span class="text-sm font-medium text-ink-900 flex-1">{{ c.countryCode }}</span>
              <span class="text-xxs text-ink-400">{{ c.institutions.length }} institutions</span>
            </button>

            <ul
              v-if="openCountry[region.region + c.countryCode] !== false"
              class="mt-2 ml-4 space-y-2"
            >
              <li v-for="inst in c.institutions" :key="inst.id" class="border-l-2 border-ink-100 pl-3">
                <div class="text-sm text-ink-900">{{ inst.name }}
                  <span class="dcgg-tag ml-2">{{ inst.type }}</span>
                </div>
                <ul v-if="inst.contacts?.length" class="mt-1 ml-3 space-y-0.5">
                  <li v-for="p in inst.contacts" :key="p.id" class="text-xxs text-ink-500">
                    · {{ p.name }}<span v-if="p.title"> — {{ p.title }}</span>
                  </li>
                </ul>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
