import { defineStore } from 'pinia';
import { auth as authApi } from '@/api/auth';
import type { User } from '@/types';

interface State {
  token: string | null;
  hrToken: string | null;
  hrTokenExpiresAt: number | null;
  user: User | null;
  loading: boolean;
  error: string | null;
}

export const useAuthStore = defineStore('auth', {
  state: (): State => ({
    token:    localStorage.getItem('dcgg.token'),
    hrToken:  null, // intentionally not persisted — PIN must be re-entered each session
    hrTokenExpiresAt: null,
    user: null,
    loading: false,
    error: null,
  }),
  getters: {
    isAuthed:    (s) => !!s.token,
    pinVerified: (s) => !!s.hrToken && (s.hrTokenExpiresAt ?? 0) > Date.now(),
    roles:       (s) => s.user?.roles ?? [],
    hasRole:     (s) => (role: string) => (s.user?.roles ?? []).includes(role),
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
      try {
        const { token, expires_in } = await authApi.verifyPin(pin);
        this.hrToken = token;
        this.hrTokenExpiresAt = Date.now() + expires_in * 1000;
        (window as unknown as { __DCGG_HR_TOKEN__?: string }).__DCGG_HR_TOKEN__ = token;
        return true;
      } catch { return false; }
    },
    logout() {
      this.token = null;
      this.hrToken = null;
      this.hrTokenExpiresAt = null;
      this.user = null;
      localStorage.removeItem('dcgg.token');
      delete (window as unknown as { __DCGG_HR_TOKEN__?: string }).__DCGG_HR_TOKEN__;
      void authApi.logout();
    },
  },
});
