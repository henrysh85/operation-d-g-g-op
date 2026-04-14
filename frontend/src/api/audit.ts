import { http } from './client';

export interface AuditEntry {
  id: string;
  actorEmail?: string | null;
  action: string;
  entity: string;
  entityId?: string | null;
  metadata: Record<string, unknown>;
  createdAt: string;
}

export const audit = {
  async list(params?: { entity?: string; actor?: string }): Promise<AuditEntry[]> {
    const { data } = await http.get<{ data: AuditEntry[] }>('/audit-log', { params });
    return data.data ?? [];
  },
};
