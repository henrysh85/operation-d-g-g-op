import { http } from './client';

export const files = {
  async upload(file: File, meta?: Record<string, string>): Promise<{ id: string; url: string }> {
    const fd = new FormData();
    fd.append('file', file);
    if (meta) for (const [k, v] of Object.entries(meta)) fd.append(k, v);
    const { data } = await http.post<{ id: string; url: string }>('/files', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return data;
  },
  async download(id: string): Promise<Blob> {
    const { data } = await http.get(`/files/${id}`, { responseType: 'blob' });
    return data as Blob;
  },
  async remove(id: string): Promise<void> {
    await http.delete(`/files/${id}`);
  },
};
