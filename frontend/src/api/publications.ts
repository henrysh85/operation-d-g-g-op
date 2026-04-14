import { http, list as listHelper } from './client';
import type { Publication } from '@/types';

interface BackendPublication {
  id: string;
  title: string;
  authors: string[];
  venue?: string;
  url?: string;
  abstract?: string;
  published_on?: string | null;
}

function adapt(p: BackendPublication): Publication {
  return {
    ...p,
    publishedAt: p.published_on ?? null,
  } as unknown as Publication;
}

export const publications = {
  async list(params?: Record<string, unknown>): Promise<Publication[]> {
    return (await listHelper<BackendPublication>('/publications', params)).map(adapt);
  },
  async get(id: string): Promise<Publication> {
    const { data } = await http.get<BackendPublication>(`/publications/${id}`);
    return adapt(data);
  },
  async create(payload: Partial<Publication>): Promise<Publication> {
    const { data } = await http.post<BackendPublication>('/publications', payload);
    return adapt(data);
  },
  async remove(id: string): Promise<void> {
    await http.delete(`/publications/${id}`);
  },
};
