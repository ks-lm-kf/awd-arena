import { get } from './client';

export interface AuditLog {
  id: number;
  user_id: number;
  username: string;
  action: string;
  resource_type: string;
  resource_id: number;
  details: string;
  ip_address: string;
  created_at: string;
}

export const auditApi = {
  getLogs: (page: number = 1, pageSize: number = 20, action?: string, resourceType?: string) =>
    get<{ items: AuditLog[]; total: number; page: number; page_size: number }>('/judge/logs', {
      page, page_size: pageSize, action, resource_type: resourceType,
    } as Record<string, unknown>),
};

export default auditApi;
