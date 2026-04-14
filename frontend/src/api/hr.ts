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

  async listReviews(params?: Record<string, unknown>): Promise<Review[]> {
    const { data } = await http.get<{ data: Review[] }>('/hr/reviews', { params });
    return data.data ?? [];
  },
  async createReview(payload: { personId: string; period: string; rating?: number; summary?: string; reviewerId?: string }): Promise<{ id: string }> {
    const { data } = await http.post<{ id: string }>('/hr/reviews', payload);
    return data;
  },

  async listExpenses(params?: Record<string, unknown>): Promise<Expense[]> {
    const { data } = await http.get<{ data: Expense[] }>('/hr/expenses', { params });
    return data.data ?? [];
  },
  async createExpense(payload: { personId: string; amount: number; currency?: string; category?: string; incurredOn: string; memo?: string }): Promise<{ id: string }> {
    const { data } = await http.post<{ id: string }>('/hr/expenses', payload);
    return data;
  },
  async setExpenseStatus(id: string, status: 'submitted' | 'approved' | 'rejected' | 'paid'): Promise<void> {
    await http.patch(`/hr/expenses/${id}`, { status });
  },
};

export interface Review {
  id: string;
  personId: string;
  personName: string;
  reviewerId?: string;
  period: string;
  rating?: number;
  summary?: string;
  createdAt: string;
}

export interface Expense {
  id: string;
  personId: string;
  personName: string;
  amount: number;
  currency: string;
  category?: string;
  incurredOn: string;
  memo?: string;
  status: 'submitted' | 'approved' | 'rejected' | 'paid';
  createdAt: string;
}
