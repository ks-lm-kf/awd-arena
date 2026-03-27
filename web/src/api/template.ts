import client from './client';

export interface ChallengeTemplate {
  id: number;
  name: string;
  description: string;
  category: string;
  difficulty: string;
  points: number;
  image_url?: string;
  docker_image: string;
  internal_port: number;
  flag: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface TemplateListResponse {
  templates: ChallengeTemplate[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateTemplateRequest {
  name: string;
  description: string;
  category: string;
  difficulty: string;
  points: number;
  image_url?: string;
  docker_image: string;
  internal_port: number;
  flag: string;
  is_active?: boolean;
}

export interface UpdateTemplateRequest extends Partial<CreateTemplateRequest> {
  id: number;
}

export const templateApi = {
  list: async (page: number = 1, pageSize: number = 20): Promise<TemplateListResponse> => {
    const response = await client.get(`/api/v1/templates`, {
      params: {
        page,
        page_size: pageSize,
      },
    });
    return response.data;
  },

  get: async (id: number): Promise<ChallengeTemplate> => {
    const response = await client.get(`/api/v1/templates/${id}`);
    return response.data;
  },

  create: async (data: CreateTemplateRequest): Promise<ChallengeTemplate> => {
    const response = await client.post(`/api/v1/templates`, data);
    return response.data;
  },

  update: async (id: number, data: Partial<CreateTemplateRequest>): Promise<ChallengeTemplate> => {
    const response = await client.put(`/api/v1/templates/${id}`, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await client.delete(`/api/v1/templates/${id}`);
  },
};

export default templateApi;
