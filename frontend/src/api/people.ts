import { http, list as listHelper } from './client';
import type { Person } from '@/types';

interface BackendPerson {
  id: string;
  name: string;
  dept?: string;
  title?: string;
  location?: string;
  email?: string;
  reports_to?: string | null;
  photo_key?: string | null;
  start_date?: string | null;
  status?: string;
}

function adapt(p: BackendPerson): Person {
  return {
    id: p.id,
    fullName: p.name,
    role: p.title ?? '',
    team: p.dept ?? '',
    region: p.location ?? '',
    email: p.email ?? '',
    reportsTo: p.reports_to ?? null,
    photoKey: p.photo_key ?? null,
    joinedAt: p.start_date ?? null,
    status: p.status,
  } as unknown as Person;
}

export const people = {
  async list(params?: Record<string, unknown>): Promise<Person[]> {
    const rows = await listHelper<BackendPerson>('/people', params);
    return rows.map(adapt);
  },
  async get(id: string): Promise<Person> {
    const { data } = await http.get<BackendPerson>(`/people/${id}`);
    return adapt(data);
  },
  async create(payload: Partial<Person>): Promise<Person> {
    const { data } = await http.post<BackendPerson>('/people', payload);
    return adapt(data);
  },
  async update(id: string, payload: Partial<Person>): Promise<Person> {
    const { data } = await http.patch<BackendPerson>(`/people/${id}`, payload);
    return adapt(data);
  },
  async remove(id: string): Promise<void> {
    await http.delete(`/people/${id}`);
  },
};
