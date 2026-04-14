<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import Sidebar from '@/components/Sidebar.vue';

const route = useRoute();
const isAuthRoute = computed(() => route.name === 'login');
</script>

<template>
  <template v-if="isAuthRoute">
    <router-view />
  </template>
  <template v-else>
    <div class="flex h-screen w-screen overflow-hidden bg-ink-50">
      <Sidebar />
      <main class="flex-1 flex flex-col overflow-hidden">
        <router-view />
      </main>
    </div>
  </template>
</template>

<style>
.fade-enter-active, .fade-leave-active { transition: opacity .12s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
