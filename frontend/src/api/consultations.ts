import { http, list as listHelper } from './client';
import type { Consultation } from '@/types';

interface BackendConsultation {
  id: string;
  jurisdiction_id: string;
  vertical: string;
  title: string;
  deadline?: string | null;
  status: string;
  assignee_id?: string | null;
  summary?: string;
  metadata?: Record<string, unknown>;
}

function adapt(c: BackendConsultation): Consultation {
  const meta = c.metadata ?? {};
  return {
    ...c,
    deadlineAt: c.deadline ?? null,
    regulator: (meta as { regulator?: string }).regulator ?? '',
  } as unknown as Consultation;
}

export const consultations = {
  async list(params?: Record<string, unknown>): Promise<Consultation[]> {
    return (await listHelper<BackendConsultation>('/consultations', params)).map(adapt);
  },
  async get(id: string): Promise<Consultation> {
    const { data } = await http.get<BackendConsultation>(`/consultations/${id}`);
    return adapt(data);
  },
  async create(payload: Partial<Consultation>): Promise<Consultation> {
    const { data } = await http.post<BackendConsultation>('/consultations', payload);
    return adapt(data);
  },
  async remove(id: string): Promise<void> {
    await http.delete(`/consultations/${id}`);
  },
};
