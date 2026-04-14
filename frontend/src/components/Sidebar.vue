<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

interface Nav { to: string; label: string; icon: string; }

const navs: Nav[] = [
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
  { to: '/audit',         label: 'Audit log',     icon: 'L' },
];

const route = useRoute();
const auth = useAuthStore();
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
        :class="route.path.startsWith(n.to) && 'bg-brand-50 text-brand-600 font-semibold'"
        @click="openOnMobile = false"
      >
        <span class="w-6 h-6 rounded grid place-items-center text-xxs font-bold
                     bg-ink-100 text-ink-500 group-hover:bg-ink-200"
              :class="route.path.startsWith(n.to) && '!bg-brand-600 !text-white'">
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
        <button class="text-xxs text-ink-500 hover:text-brand-600" @click="auth.logout()">
          Sign out
        </button>
      </div>
    </div>
  </aside>

  <!-- Mobile backdrop -->
  <div
    v-if="openOnMobile"
    class="md:hidden fixed inset-0 bg-ink-900/30 z-30"
    @click="openOnMobile = false"
  />
</template>
