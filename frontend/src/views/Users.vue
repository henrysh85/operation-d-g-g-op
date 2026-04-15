<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { users } from '@/api';
import type { UserRow } from '@/api/users';
import Modal from '@/components/Modal.vue';
import { useToastStore } from '@/stores/toast';
import { format } from 'date-fns';

const ROLES = ['admin', 'lead', 'staff', 'readonly', 'hr'];

const rows = ref<UserRow[]>([]);
const loading = ref(true);
const toasts = useToastStore();

const creating = ref(false);
const form = ref({ email: '', name: '', password: '', roles: ['staff'] as string[] });

const pwReset = ref<UserRow | null>(null);
const newPw = ref('');

async function load() {
  loading.value = true;
  try { rows.value = await users.list(); }
  finally { loading.value = false; }
}
onMounted(load);

async function submitCreate() {
  if (!form.value.email || !form.value.name || form.value.password.length < 10) {
    toasts.error('Email, name, and a 10+ char password are required.');
    return;
  }
  try {
    await users.create(form.value);
    toasts.success(`Created ${form.value.email}`);
    form.value = { email: '', name: '', password: '', roles: ['staff'] };
    creating.value = false;
    await load();
  } catch { /* toast already shown by interceptor */ }
}

function toggleRole(u: UserRow, role: string) {
  const next = u.roles.includes(role) ? u.roles.filter((r) => r !== role) : [...u.roles, role];
  users.patch(u.id, { roles: next }).then(() => {
    u.roles = next;
    toasts.success(`${u.email}: ${next.join(', ') || '(no roles)'}`);
  }).catch(() => void 0);
}

function toggleActive(u: UserRow) {
  const next = !u.active;
  users.patch(u.id, { active: next }).then(() => {
    u.active = next;
    toasts.success(`${u.email} ${next ? 'enabled' : 'disabled'}`);
  }).catch(() => void 0);
}

async function submitReset() {
  if (!pwReset.value || newPw.value.length < 10) return;
  try {
    await users.resetPassword(pwReset.value.id, newPw.value);
    toasts.success(`Password reset for ${pwReset.value.email}`);
    pwReset.value = null;
    newPw.value = '';
  } catch { /* toast shown */ }
}
</script>

<template>
  <header class="px-6 py-4 bg-white border-b border-ink-200 flex gap-3 items-start">
    <div class="flex-1">
      <h1 class="text-base font-semibold text-ink-900">Users</h1>
      <p class="text-xs text-ink-500 mt-0.5">Manage sign-in accounts, roles and password resets.</p>
    </div>
    <button class="dcgg-btn-primary" @click="creating = true">+ Add user</button>
  </header>

  <section class="flex-1 overflow-y-auto p-6">
    <div v-if="loading" class="text-xs text-ink-400">Loading…</div>
    <table v-else class="dcgg-card !p-0 w-full text-xs">
      <thead class="text-xxs uppercase text-ink-500 bg-ink-50">
        <tr>
          <th class="text-left px-3 py-2">Email</th>
          <th class="text-left px-3 py-2">Name</th>
          <th class="text-left px-3 py-2">Roles</th>
          <th class="text-left px-3 py-2">Active</th>
          <th class="text-left px-3 py-2">Created</th>
          <th class="text-left px-3 py-2"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="u in rows" :key="u.id" class="border-t border-ink-100">
          <td class="px-3 py-2 font-mono text-xxs">{{ u.email }}</td>
          <td class="px-3 py-2">{{ u.name }}</td>
          <td class="px-3 py-2">
            <div class="flex flex-wrap gap-1">
              <button
                v-for="r in ROLES"
                :key="r"
                class="text-xxs px-1.5 py-0.5 rounded border"
                :class="u.roles.includes(r)
                  ? 'bg-brand-50 text-brand-600 border-brand-200'
                  : 'bg-white text-ink-400 border-ink-200 hover:border-ink-400'"
                @click="toggleRole(u, r)"
              >{{ r }}</button>
            </div>
          </td>
          <td class="px-3 py-2">
            <button class="text-xxs" :class="u.active ? 'text-ok' : 'text-ink-400'" @click="toggleActive(u)">
              {{ u.active ? '● active' : '○ disabled' }}
            </button>
          </td>
          <td class="px-3 py-2 text-ink-500">{{ format(new Date(u.createdAt), 'PP') }}</td>
          <td class="px-3 py-2 text-right">
            <button class="text-xxs text-brand-600 hover:underline" @click="pwReset = u">Reset password</button>
          </td>
        </tr>
      </tbody>
    </table>

    <Modal :open="creating" title="Add user" width="420px" @close="creating = false">
      <div class="space-y-3">
        <label class="block">
          <span class="text-xxs font-semibold text-ink-500 uppercase">Email</span>
          <input v-model="form.email" type="email" class="dcgg-input w-full mt-1" />
        </label>
        <label class="block">
          <span class="text-xxs font-semibold text-ink-500 uppercase">Name</span>
          <input v-model="form.name" class="dcgg-input w-full mt-1" />
        </label>
        <label class="block">
          <span class="text-xxs font-semibold text-ink-500 uppercase">Password (10+ chars)</span>
          <input v-model="form.password" type="password" class="dcgg-input w-full mt-1" />
        </label>
        <div>
          <div class="text-xxs font-semibold text-ink-500 uppercase mb-1">Roles</div>
          <div class="flex flex-wrap gap-1">
            <button
              v-for="r in ROLES"
              :key="r"
              class="text-xxs px-1.5 py-0.5 rounded border"
              :class="form.roles.includes(r)
                ? 'bg-brand-50 text-brand-600 border-brand-200'
                : 'bg-white text-ink-400 border-ink-200'"
              @click="form.roles = form.roles.includes(r) ? form.roles.filter((x) => x !== r) : [...form.roles, r]"
            >{{ r }}</button>
          </div>
        </div>
      </div>
      <template #footer>
        <button class="dcgg-btn" @click="creating = false">Cancel</button>
        <button class="dcgg-btn-primary" @click="submitCreate">Create</button>
      </template>
    </Modal>

    <Modal :open="!!pwReset" :title="pwReset ? `Reset password — ${pwReset.email}` : ''" width="360px" @close="pwReset = null">
      <label class="block">
        <span class="text-xxs font-semibold text-ink-500 uppercase">New password (10+ chars)</span>
        <input v-model="newPw" type="password" class="dcgg-input w-full mt-1" />
      </label>
      <template #footer>
        <button class="dcgg-btn" @click="pwReset = null">Cancel</button>
        <button class="dcgg-btn-primary" :disabled="newPw.length < 10" @click="submitReset">Reset</button>
      </template>
    </Modal>
  </section>
</template>
