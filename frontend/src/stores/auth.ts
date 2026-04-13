import { defineStore } from 'pinia';
import { auth as authApi } from '@/api/auth';
import type { User } from '@/types';

interface State {
  token: string | null;
  user: User | null;
  pinVerified: boolean;
  loading: boolean;
  error: string | null;
}

export const useAuthStore = defineStore('auth', {
  state: (): State => ({
    token: localStorage.getItem('dcgg.token'),
    user: null,
    pinVerified: false,
    loading: false,
    error: null,
  }),
  getters: {
    isAuthed: (s) => !!s.token,
    roles:    (s) => s.user?.roles ?? [],
    hasRole:  (s) => (role: string) => (s.user?.roles ?? []).includes(role),
  },
  actions: {
    async login(email: string, password: string) {
      this.loading = true;
      this.error = null;
      try {
        const { token, user } = await authApi.login({ email, password });
        this.token = token;
        this.user  = user;
        localStorage.setItem('dcgg.token', token);
      } catch (e) {
        this.error = (e as Error).message;
        throw e;
      } finally {
        this.loading = false;
      }
    },
    async hydrate() {
      if (!this.token || this.user) return;
      try { this.user = await authApi.me(); }
      catch { this.logout(); }
    },
    async verifyPin(pin: string) {
      const { ok } = await authApi.verifyPin(pin);
      this.pinVerified = ok;
      return ok;
    },
    logout() {
      this.token = null;
      this.user = null;
      this.pinVerified = false;
      localStorage.removeItem('dcgg.token');
      void authApi.logout();
    },
  },
});
