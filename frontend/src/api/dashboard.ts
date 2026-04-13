import { http } from './client';
import type { KPI, Activity, Consultation } from '@/types';

export interface DashboardSummary {
  kpis: KPI[];
  upcomingDeadlines: Consultation[];
  recentActivity: Activity[];
}

export const dashboard = {
  async summary(params?: Record<string, unknown>): Promise<DashboardSummary> {
    const { data } = await http.get<DashboardSummary>('/dashboard/summary', { params });
    return data;
  },
};
