import { http } from './client';

export interface UserRow {
  id: string;
  email: string;
  name: string;
  active: boolean;
  roles: string[];
  createdAt: string;
}

export const users = {
  async list(): Promise<UserRow[]> {
    const { data } = await http.get<{ data: UserRow[] }>('/users');
    return data.data ?? [];
  },
  async create(payload: { email: string; name: string; password: string; roles: string[] }): Promise<{ id: string }> {
    const { data } = await http.post<{ id: string }>('/users', payload);
    return data;
  },
  async patch(id: string, payload: { name?: string; active?: boolean; roles?: string[] }): Promise<void> {
    await http.patch(`/users/${id}`, payload);
  },
  async resetPassword(id: string, newPassword: string): Promise<void> {
    await http.post(`/users/${id}/reset-password`, { newPassword });
  },
  async changeOwnPassword(currentPassword: string, newPassword: string): Promise<void> {
    await http.post('/auth/change-password', { currentPassword, newPassword });
  },
};
