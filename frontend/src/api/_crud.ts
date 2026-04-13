import { http, list as listHelper } from './client';

export function crud<T extends { id: string | number }, C = Partial<T>, U = Partial<T>>(base: string) {
  return {
    async list(params?: Record<string, unknown>): Promise<T[]> {
      return listHelper<T>(base, params);
    },
    async get(id: T['id']): Promise<T> {
      const { data } = await http.get<T>(`${base}/${id}`);
      return data;
    },
    async create(payload: C): Promise<T> {
      const { data } = await http.post<T>(base, payload);
      return data;
    },
    async update(id: T['id'], payload: U): Promise<T> {
      const { data } = await http.put<T>(`${base}/${id}`, payload);
      return data;
    },
    async remove(id: T['id']): Promise<void> {
      await http.delete(`${base}/${id}`);
    },
  };
}
