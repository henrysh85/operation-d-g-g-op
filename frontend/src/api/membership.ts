import { http } from './client';
import { crud } from './_crud';
import type { Member } from '@/types';

const base = crud<Member>('/members');

export interface QuoteInput {
  entityType: string;
  jurisdictionId: string;
  tier: 'standard' | 'premium' | 'enterprise';
}
export interface QuoteResult {
  currency: string;
  setupFee: number;
  annualFee: number;
  lineItems: Array<{ label: string; amount: number }>;
}

export interface MemberIntel {
  risks: Array<{ key: string; label: string; value: number; note?: string }>;
  recentRegChanges: Array<{ id: string; title: string; when?: string; tag?: string }>;
  openConsults: Array<{ id: string; title: string; when?: string; tag?: string }>;
  activities: Array<{ id: string; title: string; when?: string; tag?: string }>;
}

export const membership = {
  ...base,
  async quote(input: QuoteInput): Promise<QuoteResult> {
    const { data } = await http.post<QuoteResult>('/membership/quote', input);
    return data;
  },
  async submitApplication(payload: Record<string, unknown>): Promise<{ id: string; pdfUrl: string }> {
    const { data } = await http.post<{ id: string; pdfUrl: string }>('/membership/applications', payload);
    return data;
  },
  async intel(id: string): Promise<MemberIntel> {
    const { data } = await http.get<MemberIntel>(`/members/${id}/intel`);
    return data;
  },
};
