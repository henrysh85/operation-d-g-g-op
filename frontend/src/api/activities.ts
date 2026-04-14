import { http, list as listHelper } from './client';
import type { Activity } from '@/types';

interface BackendActivity {
  id: string;
  title: string;
  description?: string;
  type: string;
  vertical?: string;
  occurred_on: string;
  owner_id?: string | null;
  status?: string;
  impact?: number;
  highlight?: boolean;
}

function adapt(a: BackendActivity): Activity {
  return {
    ...a,
    occurredAt: a.occurred_on,
  } as unknown as Activity;
}

export const activities = {
  async list(params?: Record<string, unknown>): Promise<Activity[]> {
    return (await listHelper<BackendActivity>('/activities', params)).map(adapt);
  },
  async get(id: string): Promise<Activity> {
    const { data } = await http.get<BackendActivity>(`/activities/${id}`);
    return adapt(data);
  },
  async create(payload: Partial<Activity>): Promise<Activity> {
    const { data } = await http.post<BackendActivity>('/activities', payload);
    return adapt(data);
  },
  async remove(id: string): Promise<void> {
    await http.delete(`/activities/${id}`);
  },
};
