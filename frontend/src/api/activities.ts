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

export interface ActivityOutput {
  id: string;
  label: string;
  minio_key: string;
  content_type: string;
  size_bytes: number;
  created_at: string;
  url?: string;
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
  async listOutputs(activityId: string): Promise<ActivityOutput[]> {
    const { data } = await http.get<{ data: ActivityOutput[] }>(`/activities/${activityId}/outputs`);
    return data.data ?? [];
  },
  async uploadOutput(activityId: string, file: File): Promise<ActivityOutput> {
    const fd = new FormData();
    fd.append('file', file);
    const { data } = await http.post<ActivityOutput>(`/activities/${activityId}/outputs`, fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return data;
  },
  async deleteOutput(activityId: string, fileId: string): Promise<void> {
    await http.delete(`/activities/${activityId}/outputs/${fileId}`);
  },
};
