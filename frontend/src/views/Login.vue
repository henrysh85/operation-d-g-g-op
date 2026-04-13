<script setup lang="ts">
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const email = ref('');
const password = ref('');
const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

async function submit() {
  try {
    await auth.login(email.value, password.value);
    const redirect = (route.query.redirect as string) || '/dashboard';
    router.replace(redirect);
  } catch { /* error shown via store */ }
}
</script>

<template>
  <div class="min-h-screen grid place-items-center bg-ink-50 p-4">
    <form class="dcgg-card w-full max-w-sm space-y-4" @submit.prevent="submit">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 rounded-md bg-brand-600 grid place-items-center text-white text-xs font-bold">DC</div>
        <div>
          <div class="text-sm font-semibold text-ink-900">DCGG Intelligence</div>
          <div class="text-xxs text-ink-500">Sign in to continue</div>
        </div>
      </div>

      <label class="block">
        <span class="text-xxs font-semibold text-ink-500 uppercase tracking-wider">Email</span>
        <input v-model="email" type="email" required autocomplete="username" class="dcgg-input w-full mt-1" />
      </label>

      <label class="block">
        <span class="text-xxs font-semibold text-ink-500 uppercase tracking-wider">Password</span>
        <input v-model="password" type="password" required autocomplete="current-password" class="dcgg-input w-full mt-1" />
      </label>

      <div v-if="auth.error" class="text-xs text-err">{{ auth.error }}</div>

      <button type="submit" class="dcgg-btn-primary w-full justify-center" :disabled="auth.loading">
        {{ auth.loading ? 'Signing in…' : 'Sign in' }}
      </button>
    </form>
  </div>
</template>
