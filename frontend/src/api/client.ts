import axios, { AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from 'axios';

const baseURL = import.meta.env.VITE_API_BASE ?? '/api';

export const http: AxiosInstance = axios.create({
  baseURL,
  timeout: 20_000,
  headers: { 'Content-Type': 'application/json' },
});

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('dcgg.token');
  if (token) {
    config.headers = config.headers ?? {};
    (config.headers as Record<string, string>).Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (res) => res,
  (err: AxiosError<{ message?: string }>) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('dcgg.token');
      if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
        const next = encodeURIComponent(window.location.pathname + window.location.search);
        window.location.assign(`/login?redirect=${next}`);
      }
    }
    const msg = err.response?.data?.message ?? err.message ?? 'Request failed';
    return Promise.reject(new Error(msg));
  },
);

export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export async function list<T>(path: string, params?: Record<string, unknown>): Promise<T[]> {
  const { data } = await http.get<T[] | Page<T>>(path, { params });
  return Array.isArray(data) ? data : data.items;
}
