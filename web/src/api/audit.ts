import client from './client';

export interface AuditLog {
  id: number;
  user_id: number;
  username: string;
  action: string;
  target_type: string;
  target_id: number;
  details: string;
  ip_address: string;
  created_at: string;
}

export interface AuditLogListResponse {
  logs: AuditLog[];
  total: number;
  page: number;
  page_size: number;
}

export const auditApi = {
  getLogs: async (page: number = 1, pageSize: number = 20): Promise<AuditLogListResponse> => {
    const response = await client.get('/api/v1/judge/logs', {
      params: {
        page,
        page_size: pageSize,
      },
    });
    return response.data;
  },
};

export default auditApi;
