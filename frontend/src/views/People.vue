<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { people, hr } from '@/api';
import type { Person } from '@/types';
import type { Holiday, HolidayBalance, Review, Expense } from '@/api/hr';
import OrgChart from '@/components/OrgChart.vue';
import DataTable from '@/components/DataTable.vue';
import { useAuthStore } from '@/stores/auth';
import { format } from 'date-fns';

type Tab = 'org' | 'directory' | 'hr' | 'holidays' | 'performance' | 'expenses';
const tab = ref<Tab>('org');
const tabs: Array<{ id: Tab; label: string; gated?: boolean }> = [
  { id: 'org',         label: 'Org chart' },
  { id: 'directory',   label: 'Directory' },
  { id: 'hr',          label: 'HR', gated: true },
  { id: 'holidays',    label: 'Holidays' },
  { id: 'performance', label: 'Performance' },
  { id: 'expenses',    label: 'Expenses' },
];

const rows = ref<Person[]>([]);
const loading = ref(true);
onMounted(async () => {
  try { rows.value = await people.list(); }
  finally { loading.value = false; }
});

const auth = useAuthStore();
const pin = ref('');
const pinError = ref<string | null>(null);

async function unlockHR() {
  pinError.value = null;
  const ok = await auth.verifyPin(pin.value).catch(() => false);
  if (!ok) pinError.value = 'Invalid PIN.';
}

// --- Holidays ---
const holidays = ref<Holiday[]>([]);
const balances = ref<HolidayBalance[]>([]);
const loadingHol = ref(false);
const newHoliday = ref({ personId: '', startDate: '', endDate: '', note: '' });
const holidayError = ref<string | null>(null);

async function loadHolidays() {
  if (!auth.pinVerified) return;
  loadingHol.value = true;
  try {
    holidays.value = await hr.listHolidays();
    balances.value = (await hr.balances()).data;
  } catch (e) { holidayError.value = (e as Error).message; }
  finally { loadingHol.value = false; }
}

watch(() => [tab.value, auth.pinVerified] as const, ([t, verified]) => {
  if ((t === 'holidays' || t === 'hr') && verified) loadHolidays();
});

async function submitHoliday() {
  holidayError.value = null;
  if (!newHoliday.value.personId || !newHoliday.value.startDate || !newHoliday.value.endDate) {
    holidayError.value = 'Person, start and end are required.';
    return;
  }
  try {
    await hr.createHoliday(newHoliday.value);
    newHoliday.value = { personId: '', startDate: '', endDate: '', note: '' };
    await loadHolidays();
  } catch (e) { holidayError.value = (e as Error).message; }
}

async function decideHoliday(id: string, status: 'approved' | 'rejected') {
  await hr.setHolidayStatus(id, status).catch(() => null);
  await loadHolidays();
}
async function deleteHoliday(id: string) {
  if (!confirm('Withdraw this holiday request?')) return;
  await hr.deleteHoliday(id).catch(() => null);
  await loadHolidays();
}

const selectedHolidays = ref<Set<string>>(new Set());
function toggleHoliday(id: string) {
  const s = new Set(selectedHolidays.value);
  if (s.has(id)) s.delete(id); else s.add(id);
  selectedHolidays.value = s;
}
function toggleAllPendingHolidays() {
  const pendingIds = holidays.value.filter((h) => h.status === 'pending').map((h) => h.id);
  const allSelected = pendingIds.every((id) => selectedHolidays.value.has(id));
  selectedHolidays.value = allSelected ? new Set() : new Set(pendingIds);
}
async function bulkDecideHolidays(status: 'approved' | 'rejected') {
  const ids = Array.from(selectedHolidays.value);
  if (!ids.length) return;
  await hr.bulkHolidayDecision(ids, status).catch(() => null);
  selectedHolidays.value = new Set();
  await loadHolidays();
}

// --- Reviews ---
const reviews = ref<Review[]>([]);
const newReview = ref({ personId: '', period: '', rating: 3, summary: '' });
async function loadReviews() {
  if (!auth.pinVerified) return;
  reviews.value = await hr.listReviews().catch(() => []);
}
async function submitReview() {
  if (!newReview.value.personId || !newReview.value.period) return;
  await hr.createReview(newReview.value).catch(() => null);
  newReview.value = { personId: '', period: '', rating: 3, summary: '' };
  await loadReviews();
}

// --- Expenses ---
const expenses = ref<Expense[]>([]);
const newExpense = ref({ personId: '', amount: 0, currency: 'USD', category: '', incurredOn: '', memo: '' });
async function loadExpenses() {
  if (!auth.pinVerified) return;
  expenses.value = await hr.listExpenses().catch(() => []);
}
async function submitExpense() {
  if (!newExpense.value.personId || !newExpense.value.amount || !newExpense.value.incurredOn) return;
  await hr.createExpense(newExpense.value).catch(() => null);
  newExpense.value = { personId: '', amount: 0, currency: 'USD', category: '', incurredOn: '', memo: '' };
  await loadExpenses();
}
async function decideExpense(id: string, status: 'approved' | 'rejected' | 'paid') {
  await hr.setExpenseStatus(id, status).catch(() => null);
  await loadExpenses();
}

watch(() => [tab.value, auth.pinVerified] as const, ([t, verified]) => {
  if (!verified) return;
  if (t === 'performance') loadReviews();
  if (t === 'expenses') loadExpenses();
});

const columns = [
  { key: 'fullName', label: 'Name', width: '30%' },
  { key: 'role',     label: 'Role' },
  { key: 'team',     label: 'Team' },
  { key: 'region',   label: 'Region' },
  { key: 'email',    label: 'Email' },
];
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200">
    <h1 class="text-base font-semibold text-ink-900">People</h1>
    <p class="text-xs text-ink-500 mt-0.5">Team directory, org chart, and HR.</p>
  </header>

  <div class="px-6 py-2 bg-white border-b border-ink-200 flex gap-2 overflow-x-auto">
    <button
      v-for="t in tabs"
      :key="t.id"
      class="text-xs px-3 py-1.5 rounded-md border border-ink-200 whitespace-nowrap"
      :class="tab === t.id ? 'bg-brand-600 text-white border-brand-600' : 'bg-white text-ink-500 hover:bg-ink-100'"
      @click="tab = t.id"
    >
      {{ t.label }}<span v-if="t.gated" class="ml-1">🔒</span>
    </button>
  </div>

  <section class="flex-1 overflow-y-auto p-6">
    <div v-if="tab === 'org'" class="h-[70vh]">
      <OrgChart :people="rows" />
    </div>

    <div v-else-if="tab === 'directory'">
      <DataTable :rows="rows" :columns="columns" :loading="loading" empty="No people." />
    </div>

    <div v-else-if="tab === 'hr'">
      <div v-if="!auth.pinVerified" class="dcgg-card max-w-sm">
        <div class="text-sm font-semibold mb-2">HR area is PIN-protected</div>
        <p class="text-xs text-ink-500 mb-3">Enter your HR PIN to access salary and contract data.</p>
        <input v-model="pin" type="password" maxlength="8" placeholder="PIN" class="dcgg-input w-full" />
        <div v-if="pinError" class="text-xs text-err mt-2">{{ pinError }}</div>
        <button class="dcgg-btn-primary mt-3" @click="unlockHR">Unlock</button>
      </div>
      <div v-else class="dcgg-card">
        <div class="text-sm font-semibold mb-2">HR unlocked</div>
        <p class="text-xs text-ink-500">Contracts, salaries and benefits data would render here.</p>
      </div>
    </div>

    <div v-else-if="tab === 'holidays'" class="space-y-4">
      <div v-if="!auth.pinVerified" class="dcgg-card max-w-sm">
        <div class="text-sm font-semibold mb-2">Holiday admin is HR-gated</div>
        <p class="text-xs text-ink-500 mb-3">Verify your HR PIN under the HR tab to manage holidays.</p>
      </div>
      <template v-else>
        <div class="dcgg-card">
          <div class="text-sm font-semibold mb-3">New holiday request</div>
          <div class="grid grid-cols-1 md:grid-cols-5 gap-2 items-end">
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Person</span>
              <select v-model="newHoliday.personId" class="dcgg-input w-full mt-1">
                <option value="">Select…</option>
                <option v-for="p in rows" :key="p.id" :value="p.id">{{ (p as any).fullName }}</option>
              </select>
            </label>
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Start</span>
              <input v-model="newHoliday.startDate" type="date" class="dcgg-input w-full mt-1" />
            </label>
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">End</span>
              <input v-model="newHoliday.endDate" type="date" class="dcgg-input w-full mt-1" />
            </label>
            <label class="block md:col-span-1">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Note</span>
              <input v-model="newHoliday.note" class="dcgg-input w-full mt-1" />
            </label>
            <button class="dcgg-btn-primary" @click="submitHoliday">Submit</button>
          </div>
          <div v-if="holidayError" class="text-xs text-err mt-2">{{ holidayError }}</div>
        </div>

        <div class="dcgg-card">
          <div class="flex items-center mb-3">
            <div class="text-sm font-semibold flex-1">Balances (this year)</div>
            <button class="text-xxs text-ink-500 hover:text-brand-600" @click="loadHolidays">Refresh</button>
          </div>
          <table class="w-full text-xs">
            <thead class="text-xxs uppercase text-ink-500">
              <tr><th class="text-left py-1">Person</th><th class="text-right">Quota</th><th class="text-right">Taken</th><th class="text-right">Remaining</th></tr>
            </thead>
            <tbody>
              <tr v-for="b in balances" :key="b.personId" class="border-t border-ink-100">
                <td class="py-1">{{ b.personName }}</td>
                <td class="text-right">{{ b.quota }}</td>
                <td class="text-right">{{ b.taken }}</td>
                <td class="text-right" :class="b.remaining < 0 ? 'text-err font-semibold' : ''">{{ b.remaining }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="dcgg-card">
          <div class="flex items-center mb-3">
            <div class="text-sm font-semibold flex-1">Requests</div>
            <template v-if="selectedHolidays.size > 0">
              <span class="text-xxs text-ink-500 mr-2">{{ selectedHolidays.size }} selected</span>
              <button class="text-xxs text-ok hover:underline mr-3" @click="bulkDecideHolidays('approved')">Approve all</button>
              <button class="text-xxs text-err hover:underline" @click="bulkDecideHolidays('rejected')">Reject all</button>
            </template>
          </div>
          <div v-if="loadingHol" class="text-xs text-ink-400">Loading…</div>
          <div v-else-if="!holidays.length" class="text-xs text-ink-400">No holiday requests yet.</div>
          <table v-else class="w-full text-xs">
            <thead class="text-xxs uppercase text-ink-500">
              <tr>
                <th class="w-6 py-1">
                  <input
                    type="checkbox"
                    :checked="holidays.filter((h) => h.status === 'pending').every((h) => selectedHolidays.has(h.id)) && holidays.some((h) => h.status === 'pending')"
                    @change="toggleAllPendingHolidays"
                  />
                </th>
                <th class="text-left py-1">Person</th><th>Dates</th><th class="text-right">Days</th><th>Status</th><th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="h in holidays" :key="h.id" class="border-t border-ink-100">
                <td class="py-1">
                  <input
                    v-if="h.status === 'pending'"
                    type="checkbox"
                    :checked="selectedHolidays.has(h.id)"
                    @change="toggleHoliday(h.id)"
                  />
                </td>
                <td class="py-1">{{ h.personName }}</td>
                <td>{{ format(new Date(h.startDate), 'PP') }} → {{ format(new Date(h.endDate), 'PP') }}</td>
                <td class="text-right">{{ h.days }}</td>
                <td>
                  <span class="dcgg-tag" :class="{
                    'bg-ok/10 text-ok': h.status==='approved',
                    'bg-err/10 text-err': h.status==='rejected',
                    'bg-warn/10 text-warn': h.status==='pending',
                  }">{{ h.status }}</span>
                </td>
                <td class="text-right">
                  <button v-if="h.status!=='approved'" class="text-xxs text-ok hover:underline mr-2" @click="decideHoliday(h.id,'approved')">Approve</button>
                  <button v-if="h.status!=='rejected'" class="text-xxs text-err hover:underline mr-2" @click="decideHoliday(h.id,'rejected')">Reject</button>
                  <button class="text-xxs text-ink-500 hover:text-err hover:underline" @click="deleteHoliday(h.id)">Delete</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
    <div v-else-if="tab === 'performance'" class="space-y-4">
      <div v-if="!auth.pinVerified" class="dcgg-card max-w-sm text-xs text-ink-500">
        Verify your HR PIN under the HR tab to view reviews.
      </div>
      <template v-else>
        <div class="dcgg-card">
          <div class="text-sm font-semibold mb-3">Log a review</div>
          <div class="grid grid-cols-1 md:grid-cols-5 gap-2 items-end">
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Person</span>
              <select v-model="newReview.personId" class="dcgg-input w-full mt-1">
                <option value="">Select…</option>
                <option v-for="p in rows" :key="p.id" :value="p.id">{{ (p as any).fullName }}</option>
              </select>
            </label>
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Period</span>
              <input v-model="newReview.period" placeholder="e.g. 2026-Q1" class="dcgg-input w-full mt-1" />
            </label>
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Rating</span>
              <select v-model.number="newReview.rating" class="dcgg-input w-full mt-1">
                <option v-for="n in 5" :key="n" :value="n">{{ n }}</option>
              </select>
            </label>
            <label class="block md:col-span-1">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Summary</span>
              <input v-model="newReview.summary" class="dcgg-input w-full mt-1" />
            </label>
            <button class="dcgg-btn-primary" @click="submitReview">Save</button>
          </div>
        </div>
        <div class="dcgg-card">
          <div class="text-sm font-semibold mb-3">Reviews</div>
          <div v-if="!reviews.length" class="text-xs text-ink-400">No reviews logged yet.</div>
          <table v-else class="w-full text-xs">
            <thead class="text-xxs uppercase text-ink-500">
              <tr><th class="text-left py-1">Person</th><th>Period</th><th class="text-right">Rating</th><th>Summary</th></tr>
            </thead>
            <tbody>
              <tr v-for="r in reviews" :key="r.id" class="border-t border-ink-100">
                <td class="py-1">{{ r.personName }}</td>
                <td>{{ r.period }}</td>
                <td class="text-right">{{ r.rating ?? '—' }}</td>
                <td class="text-ink-600">{{ r.summary ?? '' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <div v-else-if="tab === 'expenses'" class="space-y-4">
      <div v-if="!auth.pinVerified" class="dcgg-card max-w-sm text-xs text-ink-500">
        Verify your HR PIN under the HR tab to view expenses.
      </div>
      <template v-else>
        <div class="dcgg-card">
          <div class="text-sm font-semibold mb-3">Submit an expense</div>
          <div class="grid grid-cols-2 md:grid-cols-6 gap-2 items-end">
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Person</span>
              <select v-model="newExpense.personId" class="dcgg-input w-full mt-1">
                <option value="">Select…</option>
                <option v-for="p in rows" :key="p.id" :value="p.id">{{ (p as any).fullName }}</option>
              </select>
            </label>
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Amount</span>
              <input v-model.number="newExpense.amount" type="number" step="0.01" class="dcgg-input w-full mt-1" />
            </label>
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Currency</span>
              <input v-model="newExpense.currency" class="dcgg-input w-full mt-1" />
            </label>
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Category</span>
              <input v-model="newExpense.category" class="dcgg-input w-full mt-1" placeholder="travel / meals…" />
            </label>
            <label class="block">
              <span class="text-xxs font-semibold text-ink-500 uppercase">Date</span>
              <input v-model="newExpense.incurredOn" type="date" class="dcgg-input w-full mt-1" />
            </label>
            <button class="dcgg-btn-primary" @click="submitExpense">Submit</button>
          </div>
        </div>
        <div class="dcgg-card">
          <div class="text-sm font-semibold mb-3">Expenses</div>
          <div v-if="!expenses.length" class="text-xs text-ink-400">No expenses recorded yet.</div>
          <table v-else class="w-full text-xs">
            <thead class="text-xxs uppercase text-ink-500">
              <tr><th class="text-left py-1">Person</th><th>Date</th><th>Category</th><th class="text-right">Amount</th><th>Status</th><th></th></tr>
            </thead>
            <tbody>
              <tr v-for="e in expenses" :key="e.id" class="border-t border-ink-100">
                <td class="py-1">{{ e.personName }}</td>
                <td>{{ format(new Date(e.incurredOn), 'PP') }}</td>
                <td>{{ e.category ?? '—' }}</td>
                <td class="text-right">{{ e.currency }} {{ e.amount.toFixed(2) }}</td>
                <td>
                  <span class="dcgg-tag" :class="{
                    'bg-ok/10 text-ok': e.status==='approved' || e.status==='paid',
                    'bg-err/10 text-err': e.status==='rejected',
                    'bg-warn/10 text-warn': e.status==='submitted',
                  }">{{ e.status }}</span>
                </td>
                <td class="text-right">
                  <button v-if="e.status==='submitted'" class="text-xxs text-ok hover:underline mr-2" @click="decideExpense(e.id,'approved')">Approve</button>
                  <button v-if="e.status==='submitted'" class="text-xxs text-err hover:underline mr-2" @click="decideExpense(e.id,'rejected')">Reject</button>
                  <button v-if="e.status==='approved'" class="text-xxs text-brand-600 hover:underline" @click="decideExpense(e.id,'paid')">Mark paid</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </section>
</template>
