import { http, list as listHelper } from './client';
import type { Template } from '@/types';

export interface TemplateVariable {
  name: string;
  type?: string;
  required?: boolean;
}

export interface TemplateFull extends Template {
  slug: string;
  body: string;
  variables: TemplateVariable[];
  tags: string[];
}

interface BackendTemplate {
  id: string;
  slug: string;
  name: string;
  kind: string;
  body: string;
  variables: TemplateVariable[] | null;
  tags: string[] | null;
  created_at: string;
  updated_at: string;
}

function adapt(t: BackendTemplate): TemplateFull {
  return {
    id: t.id,
    slug: t.slug,
    name: t.name,
    category: t.kind,
    body: t.body,
    variables: t.variables ?? [],
    tags: t.tags ?? [],
    updatedAt: t.updated_at,
  };
}

export const templates = {
  async list(params?: Record<string, unknown>): Promise<TemplateFull[]> {
    return (await listHelper<BackendTemplate>('/templates', params)).map(adapt);
  },
  async get(id: string): Promise<TemplateFull> {
    const { data } = await http.get<BackendTemplate>(`/templates/${id}`);
    return adapt(data);
  },
  async render(id: string, params: Record<string, unknown>): Promise<string> {
    const { data } = await http.post<{ rendered: string }>(`/templates/${id}/render`, { params });
    return data.rendered;
  },
};
