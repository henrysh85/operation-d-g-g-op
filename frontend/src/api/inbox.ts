import { http } from './client';

export interface InboxItem {
  kind: 'holiday_pending' | 'expense_pending' | 'consultation_deadline' | 'holiday_decided';
  id: string;
  title: string;
  detail?: string;
  when: string;
  link?: string;
}

export const inbox = {
  async tasks(): Promise<{ items: InboxItem[]; counts: Record<string, number> }> {
    const { data } = await http.get<{ data: InboxItem[]; counts: Record<string, number> }>('/me/tasks');
    return { items: data.data ?? [], counts: data.counts ?? {} };
  },
};
