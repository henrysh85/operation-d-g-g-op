<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { users as usersApi } from '@/api';
import { useToastStore } from '@/stores/toast';
import Modal from '@/components/Modal.vue';

interface Nav { to: string; label: string; icon: string; role?: string; }

const allNavs: Nav[] = [
  { to: '/dashboard',     label: 'Dashboard',     icon: 'D' },
  { to: '/regulatory',    label: 'Regulatory',    icon: 'R' },
  { to: '/stakeholders',  label: 'Stakeholders',  icon: 'S' },
  { to: '/membership',    label: 'Membership',    icon: 'M' },
  { to: '/consultations', label: 'Consultations', icon: 'C' },
  { to: '/calendar',      label: 'Calendar',      icon: 'K' },
  { to: '/activities',    label: 'Activities',    icon: 'A' },
  { to: '/templates',     label: 'Templates',     icon: 'T' },
  { to: '/people',        label: 'People',        icon: 'P' },
  { to: '/engagement',    label: 'Engagement',    icon: 'E' },
  { to: '/members',       label: 'Members',       icon: 'U' },
  { to: '/users',         label: 'Users',         icon: 'V', role: 'admin' },
  { to: '/audit',         label: 'Audit log',     icon: 'L', role: 'admin' },
];

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const navs = computed(() => allNavs.filter((n) => !n.role || auth.hasRole(n.role)));

function signOut() {
  auth.logout();
  router.push('/login');
}

const pwOpen = ref(false);
const pwCurrent = ref('');
const pwNew = ref('');
const toasts = useToastStore();
async function changePassword() {
  if (pwNew.value.length < 10) { toasts.error('New password must be at least 10 characters.'); return; }
  try {
    await usersApi.changeOwnPassword(pwCurrent.value, pwNew.value);
    toasts.success('Password updated.');
    pwOpen.value = false;
    pwCurrent.value = ''; pwNew.value = '';
  } catch { /* interceptor shows the error */ }
}
const openOnMobile = ref(false);
const collapsed = ref(false);

const initials = computed(() =>
  (auth.user?.name ?? 'DC')
    .split(' ').map((s) => s[0]).slice(0, 2).join('').toUpperCase(),
);
</script>

<template>
  <!-- Mobile menu toggle -->
  <button
    class="md:hidden fixed top-3 left-3 z-50 w-9 h-9 rounded-md bg-white border border-ink-200 grid place-items-center"
    aria-label="Toggle navigation"
    @click="openOnMobile = !openOnMobile"
  >
    <span class="block w-4 h-[2px] bg-ink-700 mb-[3px]" />
    <span class="block w-4 h-[2px] bg-ink-700 mb-[3px]" />
    <span class="block w-4 h-[2px] bg-ink-700" />
  </button>

  <aside
    class="bg-white border-r border-ink-200 flex flex-col transition-all duration-150
           md:static md:translate-x-0 fixed inset-y-0 left-0 z-40"
    :class="[
      collapsed ? 'w-16' : 'w-56',
      openOnMobile ? 'translate-x-0' : '-translate-x-full md:translate-x-0',
    ]"
  >
    <!-- Brand -->
    <div class="h-14 flex items-center gap-2 px-3 border-b border-ink-200">
      <div class="w-7 h-7 rounded-md bg-brand-600 grid place-items-center text-white text-xs font-bold">
        DC
      </div>
      <div v-if="!collapsed" class="text-sm font-semibold text-ink-900 truncate">
        DCGG Intelligence
      </div>
      <button
        class="ml-auto text-ink-400 hover:text-ink-700 text-xs hidden md:block"
        :title="collapsed ? 'Expand' : 'Collapse'"
        @click="collapsed = !collapsed"
      >
        {{ collapsed ? '»' : '«' }}
      </button>
    </div>

    <!-- Nav -->
    <nav class="flex-1 overflow-y-auto py-2">
      <router-link
        v-for="n in navs"
        :key="n.to"
        :to="n.to"
        class="group flex items-center gap-2 mx-2 my-0.5 px-2 py-2 rounded-md text-sm text-ink-700
               hover:bg-ink-100 hover:text-ink-900"
        :class="(route.path === n.to || route.path.startsWith(n.to + '/')) && 'bg-brand-50 text-brand-600 font-semibold'"
        @click="openOnMobile = false"
      >
        <span class="w-6 h-6 rounded grid place-items-center text-xxs font-bold
                     bg-ink-100 text-ink-500 group-hover:bg-ink-200"
              :class="(route.path === n.to || route.path.startsWith(n.to + '/')) && '!bg-brand-600 !text-white'">
          {{ n.icon }}
        </span>
        <span v-if="!collapsed" class="truncate">{{ n.label }}</span>
      </router-link>
    </nav>

    <!-- Footer / user -->
    <div class="border-t border-ink-200 p-3 flex items-center gap-2">
      <div class="w-7 h-7 rounded-full bg-ink-200 grid place-items-center text-xxs font-bold text-ink-700">
        {{ initials }}
      </div>
      <div v-if="!collapsed" class="flex-1 min-w-0">
        <div class="text-xs font-semibold text-ink-900 truncate">{{ auth.user?.name ?? 'Signed out' }}</div>
        <div class="flex gap-2">
          <button class="text-xxs text-ink-500 hover:text-brand-600" @click="pwOpen = true">
            Password
          </button>
          <button class="text-xxs text-ink-500 hover:text-brand-600" @click="signOut">
            Sign out
          </button>
        </div>
      </div>
    </div>
  </aside>

  <!-- Mobile backdrop -->
  <div
    v-if="openOnMobile"
    class="md:hidden fixed inset-0 bg-ink-900/30 z-30"
    @click="openOnMobile = false"
  />

  <Modal :open="pwOpen" title="Change password" width="360px" @close="pwOpen = false">
    <div class="space-y-3">
      <label class="block">
        <span class="text-xxs font-semibold text-ink-500 uppercase">Current password</span>
        <input v-model="pwCurrent" type="password" class="dcgg-input w-full mt-1" />
      </label>
      <label class="block">
        <span class="text-xxs font-semibold text-ink-500 uppercase">New password (10+ chars)</span>
        <input v-model="pwNew" type="password" class="dcgg-input w-full mt-1" />
      </label>
    </div>
    <template #footer>
      <button class="dcgg-btn" @click="pwOpen = false">Cancel</button>
      <button class="dcgg-btn-primary" :disabled="pwNew.length < 10 || !pwCurrent" @click="changePassword">
        Save
      </button>
    </template>
  </Modal>
</template>
