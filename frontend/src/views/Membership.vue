<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { membership, regulatory } from '@/api';
import type { Jurisdiction } from '@/types';
import type { QuoteResult } from '@/api/membership';

const step = ref(1);

const entityTypes = [
  { id: 'exchange',  label: 'Exchange',          sub: 'Spot / derivatives' },
  { id: 'custodian', label: 'Custodian / VASP',  sub: 'Wallet & safekeeping' },
  { id: 'bank',      label: 'Bank',              sub: 'Deposit / credit' },
  { id: 'broker',    label: 'Broker / dealer',   sub: 'Securities trading' },
  { id: 'issuer',    label: 'Token issuer',      sub: 'Primary issuance' },
  { id: 'other',     label: 'Other',             sub: 'Describe in details' },
];

const form = ref({
  entityType: '',
  jurisdictionId: '',
  tier: 'standard' as 'standard' | 'premium' | 'enterprise',
  legalName: '',
  contactEmail: '',
  notes: '',
});

const jurisdictions = ref<Jurisdiction[]>([]);
const quote = ref<QuoteResult | null>(null);
const submitting = ref(false);
const resultPdf = ref<string | null>(null);

onMounted(async () => {
  try { jurisdictions.value = await regulatory.list(); } catch { /* ignore for scaffold */ }
});

const canNext = computed(() => {
  if (step.value === 1) return !!form.value.entityType;
  if (step.value === 2) return !!form.value.jurisdictionId;
  if (step.value === 3) return !!form.value.tier;
  if (step.value === 4) return !!form.value.legalName && !!form.value.contactEmail;
  return false;
});

async function next() {
  if (!canNext.value) return;
  if (step.value === 3) {
    try {
      quote.value = await membership.quote({
        entityType: form.value.entityType,
        jurisdictionId: form.value.jurisdictionId,
        tier: form.value.tier,
      });
    } catch { quote.value = null; }
  }
  step.value++;
}

async function submit() {
  submitting.value = true;
  try {
    const res = await membership.submitApplication(form.value);
    resultPdf.value = res.pdfUrl;
  } finally { submitting.value = false; }
}
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">Membership application</h1>
    <p class="text-xs text-ink-500 mt-0.5">Five-step onboarding wizard.</p>
  </header>

  <section class="flex-1 overflow-y-auto p-6">
    <!-- Stepper -->
    <ol class="flex items-center gap-2 mb-6 text-xs text-ink-500">
      <li v-for="n in 5" :key="n" class="flex items-center gap-2">
        <span
          class="w-6 h-6 rounded-full grid place-items-center text-xxs font-bold"
          :class="n < step ? 'bg-ok text-white' : n === step ? 'bg-brand-600 text-white' : 'bg-ink-100 text-ink-500'"
        >{{ n }}</span>
        <span class="font-medium" :class="n === step && 'text-ink-900'">
          {{ ['Entity', 'Jurisdiction', 'Pricing', 'Details', 'PDF'][n - 1] }}
        </span>
        <span v-if="n < 5" class="w-6 h-px bg-ink-200" />
      </li>
    </ol>

    <div class="dcgg-card max-w-3xl">
      <!-- 1: Entity type -->
      <div v-if="step === 1">
        <div class="text-sm font-semibold mb-3">Entity type</div>
        <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
          <button
            v-for="e in entityTypes"
            :key="e.id"
            class="text-center border-2 border-ink-200 rounded-lg py-4 hover:border-brand-200 hover:bg-brand-50/50"
            :class="form.entityType === e.id && '!border-brand-600 bg-brand-50'"
            @click="form.entityType = e.id"
          >
            <div class="text-sm font-semibold text-ink-900">{{ e.label }}</div>
            <div class="text-xxs text-ink-500 mt-0.5">{{ e.sub }}</div>
          </button>
        </div>
      </div>

      <!-- 2: Jurisdiction -->
      <div v-else-if="step === 2">
        <div class="text-sm font-semibold mb-3">Jurisdiction</div>
        <select v-model="form.jurisdictionId" class="dcgg-input w-full">
          <option value="">Select a jurisdiction…</option>
          <option v-for="j in jurisdictions" :key="j.id" :value="j.id">
            {{ j.name }} ({{ j.region }})
          </option>
        </select>
      </div>

      <!-- 3: Pricing -->
      <div v-else-if="step === 3">
        <div class="text-sm font-semibold mb-3">Pricing tier</div>
        <div class="grid grid-cols-3 gap-3">
          <button
            v-for="t in (['standard', 'premium', 'enterprise'] as const)"
            :key="t"
            class="border-2 border-ink-200 rounded-lg py-4 hover:border-brand-200"
            :class="form.tier === t && '!border-brand-600 bg-brand-50'"
            @click="form.tier = t"
          >
            <div class="text-sm font-semibold capitalize">{{ t }}</div>
          </button>
        </div>
      </div>

      <!-- 4: Details -->
      <div v-else-if="step === 4">
        <div class="text-sm font-semibold mb-3">Entity details</div>
        <div class="space-y-3">
          <label class="block">
            <span class="text-xxs font-semibold text-ink-500 uppercase">Legal name</span>
            <input v-model="form.legalName" class="dcgg-input w-full mt-1" />
          </label>
          <label class="block">
            <span class="text-xxs font-semibold text-ink-500 uppercase">Contact email</span>
            <input v-model="form.contactEmail" type="email" class="dcgg-input w-full mt-1" />
          </label>
          <label class="block">
            <span class="text-xxs font-semibold text-ink-500 uppercase">Notes</span>
            <textarea v-model="form.notes" rows="4" class="dcgg-input w-full mt-1" />
          </label>
          <div v-if="quote" class="mt-2 p-3 bg-ink-50 rounded-md text-xs">
            <div class="font-semibold mb-1">Quote</div>
            <div>Setup: {{ quote.currency }} {{ quote.setupFee }}</div>
            <div>Annual: {{ quote.currency }} {{ quote.annualFee }}</div>
          </div>
        </div>
      </div>

      <!-- 5: PDF -->
      <div v-else-if="step === 5">
        <div class="text-sm font-semibold mb-3">Generate PDF</div>
        <p class="text-xs text-ink-500 mb-3">Submit your application and download the generated PDF.</p>
        <button class="dcgg-btn-primary" :disabled="submitting" @click="submit">
          {{ submitting ? 'Generating…' : 'Submit & download' }}
        </button>
        <a v-if="resultPdf" :href="resultPdf" target="_blank" class="ml-3 text-brand-600 text-xs underline">
          Download PDF
        </a>
      </div>

      <!-- Nav -->
      <div class="flex justify-between mt-6">
        <button class="dcgg-btn" :disabled="step === 1" @click="step--">Back</button>
        <button v-if="step < 5" class="dcgg-btn-primary" :disabled="!canNext" @click="next">Next</button>
      </div>
    </div>
  </section>
</template>
