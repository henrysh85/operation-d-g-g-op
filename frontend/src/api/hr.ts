import { http } from './client';

export interface Holiday {
  id: string;
  personId: string;
  personName: string;
  startDate: string;
  endDate: string;
  days: number;
  status: 'pending' | 'approved' | 'rejected';
  note?: string | null;
  createdAt: string;
}

export interface HolidayBalance {
  personId: string;
  personName: string;
  quota: number;
  taken: number;
  remaining: number;
}

export const hr = {
  async listHolidays(params?: Record<string, unknown>): Promise<Holiday[]> {
    const { data } = await http.get<{ data: Holiday[] }>('/hr/holidays', { params });
    return data.data ?? [];
  },
  async createHoliday(payload: { personId: string; startDate: string; endDate: string; days?: number; note?: string }): Promise<{ id: string }> {
    const { data } = await http.post<{ id: string }>('/hr/holidays', payload);
    return data;
  },
  async setHolidayStatus(id: string, status: 'approved' | 'rejected' | 'pending'): Promise<void> {
    await http.patch(`/hr/holidays/${id}`, { status });
  },
  async deleteHoliday(id: string): Promise<void> {
    await http.delete(`/hr/holidays/${id}`);
  },
  async balances(year?: number): Promise<{ year: number; data: HolidayBalance[] }> {
    const { data } = await http.get<{ year: number; data: HolidayBalance[] }>('/hr/holidays/balances', {
      params: year ? { year } : undefined,
    });
    return data;
  },
};
