import axios, { AxiosError } from 'axios';

import type { Account } from '@/types/account';
import type { GeneratedReply, WorkerStatusResponse } from '@/types/api';
import type { Negotiation, NegotiationMessage } from '@/types/negotiation';
import type { StatsResponse } from '@/types/stats';

const baseURL = import.meta.env.VITE_API_URL?.trim() || '';

export const api = axios.create({
  baseURL,
  timeout: 30_000,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => config);

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: string; message?: string }>) => {
    const message =
      error.response?.data?.error ??
      error.response?.data?.message ??
      error.message ??
      'Request failed';

    return Promise.reject(new Error(message));
  },
);

export const apiClient = {
  async getStats(): Promise<StatsResponse> {
    const { data } = await api.get<StatsResponse>('/api/stats');
    return data;
  },
  async getAccounts(): Promise<Account[]> {
    const { data } = await api.get<Account[]>('/api/accounts');
    return data;
  },
  async getNegotiations(): Promise<Negotiation[]> {
    const { data } = await api.get<Negotiation[]>('/api/negotiations');
    return data;
  },
  async getNegotiationMessages(id: string): Promise<NegotiationMessage[]> {
    const { data } = await api.get<NegotiationMessage[]>(`/api/negotiations/${id}/messages`);
    return data;
  },
  async startApply(): Promise<WorkerStatusResponse> {
    const { data } = await api.post<WorkerStatusResponse>('/api/apply/start');
    return data;
  },
  async stopApply(): Promise<WorkerStatusResponse> {
    const { data } = await api.post<WorkerStatusResponse>('/api/apply/stop');
    return data;
  },
  async publishResume(): Promise<WorkerStatusResponse> {
    const { data } = await api.post<WorkerStatusResponse>('/api/resume/publish');
    return data;
  },
  async sendReply(id: string, payload: { text?: string; quickReplyOptionId?: string }): Promise<WorkerStatusResponse> {
    const body = {
      text: payload.text,
      quick_reply_option_id: payload.quickReplyOptionId,
    };
    const { data } = await api.post<WorkerStatusResponse>(`/api/negotiations/${id}/reply`, body);
    return data;
  },
  async generateReply(id: string): Promise<GeneratedReply> {
    const { data } = await api.post<GeneratedReply>(`/api/negotiations/${id}/generate-reply`);
    return data;
  },
};
