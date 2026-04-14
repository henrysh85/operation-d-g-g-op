import { http } from './client';
import type { KPI, Activity, Consultation } from '@/types';

export interface DashboardSummary {
  kpis: KPI[];
  upcomingDeadlines: Consultation[];
  recentActivity: Activity[];
}

interface BackendSummary {
  counts: Record<string, number>;
  deltas?: Record<string, number>;
  highlights: Array<{ id: string; title: string; occurred_on: string; vertical?: string; type?: string }>;
  deadlines?: Array<{ id: string; title: string; deadline: string; regulator?: string; impact?: string }>;
}

const COUNT_LABELS: Record<string, string> = {
  activities: 'Activities',
  clients: 'Clients',
  contacts: 'Contacts',
  jurisdictions: 'Jurisdictions',
  open_consultations: 'Open consultations',
  people: 'People',
};

export const dashboard = {
  async summary(params?: Record<string, unknown>): Promise<DashboardSummary> {
    const { data } = await http.get<BackendSummary>('/dashboard/summary', { params });
    const counts = data.counts ?? {};
    const deltas = data.deltas ?? {};
    const order = ['people', 'activities', 'open_consultations', 'clients', 'contacts', 'jurisdictions'];
    const kpis: KPI[] = order
      .filter((k) => k in counts)
      .map((key) => ({
        key,
        label: COUNT_LABELS[key] ?? key,
        value: counts[key],
        delta: deltas[key],
      } as KPI));
    const recentActivity: Activity[] = (data.highlights ?? []).map((h) => ({
      id: h.id,
      title: h.title,
      type: h.type ?? 'activity',
      occurredAt: h.occurred_on,
      vertical: h.vertical,
    } as unknown as Activity));
    const upcomingDeadlines: Consultation[] = (data.deadlines ?? []).map((d) => ({
      id: d.id,
      title: d.title,
      deadlineAt: d.deadline,
      regulator: d.regulator,
      impact: d.impact,
    } as unknown as Consultation));
    return { kpis, recentActivity, upcomingDeadlines };
  },
};
