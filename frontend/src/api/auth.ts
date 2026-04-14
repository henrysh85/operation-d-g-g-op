import { http } from './client';
import type { User } from '@/types';

export interface LoginPayload { email: string; password: string; }
export interface LoginResponse { token: string; user: User; }

export const auth = {
  async login(payload: LoginPayload): Promise<LoginResponse> {
    const { data } = await http.post<LoginResponse>('/auth/login', payload);
    return data;
  },
  async me(): Promise<User> {
    const { data } = await http.get<User>('/auth/me');
    return data;
  },
  async logout(): Promise<void> {
    await http.post('/auth/logout').catch(() => void 0);
  },
  async verifyPin(pin: string): Promise<{ token: string; expires_in: number }> {
    const { data } = await http.post<{ token: string; expires_in: number }>('/auth/verify-pin', { pin });
    return data;
  },
};
