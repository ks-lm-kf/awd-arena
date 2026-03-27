import { get, post, put, del } from './client'

export interface DockerImage {
  id: number
  name: string
  tag: string
  image_id: string
  description: string
  category: 'web' | 'pwn' | 're' | 'crypto' | 'general'
  difficulty: 'easy' | 'medium' | 'hard'
  ports: string
  memory_limit: number
  cpu_limit: number
  flag: string
  initial_score: number
  status: string
  created_at: string
  updated_at: string
}

export interface HostImage {
  repository: string
  tag: string
  image_id: string
  size: string
  created_at: string
}

export interface DockerImageParams {
  name: string
  category: string
  difficulty: string
  tag?: string
  image_id?: string
  description?: string
  ports?: string
  memory_limit?: number
  cpu_limit?: number
  flag?: string
  initial_score?: number
  status?: string
}

export interface PullImageParams {
  name: string
  tag?: string
}

export interface PushImageParams {
  image_ref: string
  auth_config?: {
    username?: string
    password?: string
    email?: string
    serveraddress?: string
  }
}

export interface BuildImageParams {
  context_path?: string
  dockerfile?: string
  tags: string[]
  build_args?: Record<string, string>
  no_cache?: boolean
}

export interface ImageDetails {
  id: string
  repo_tags: string[]
  repo_digests: string[]
  created: string
  size: number
  architecture: string
  os: string
  author: string
}

export const dockerImageApi = {
  // List images from database
  list: (params?: { category?: string; difficulty?: string; search?: string; page?: number; page_size?: number }) =>
    get<{ items: DockerImage[]; total: number }>('/docker-images', params as any),
  
  // Get single image from database
  get: (id: number) => get<DockerImage>(`/docker-images/${id}`),
  
  // Create image record in database
  create: (data: DockerImageParams) => post<DockerImage>('/docker-images', data),
  
  // Update image record in database
  update: (id: number, data: Partial<DockerImageParams>) => put<DockerImage>(`/docker-images/${id}`, data),
  
  // Delete image record from database only
  delete: (id: number) => del<void>(`/docker-images/${id}`),
  
  // Pull image to database record
  pull: (id: number) => post<void>(`/docker-images/${id}/pull`),
  
  // List images from host machine
  hostList: () => get<HostImage[]>('/docker-images/host/list'),

  // ===== New Admin API Routes =====
  
  // Pull image from registry by name
  pullImage: (params: PullImageParams) => post<{ output: string }>('/admin/images/pull', params),
  
  // Push image to registry
  pushImage: (params: PushImageParams) => post<void>('/admin/images/push', params),
  
  // Build image from Dockerfile
  buildImage: (params: BuildImageParams) => post<{ image_id: string }>('/admin/images/build', params),
  
  // Get detailed image info from host
  getImageDetails: (idOrName: string) => get<ImageDetails>(`/admin/images/${idOrName}/details`),
  
  // Remove image from host machine only
  // Note: idOrName should be the image ID (sha256:xxx or short ID)
  removeFromHost: (idOrName: string, force?: boolean) => 
    del<void>(`/admin/images/host/${encodeURIComponent(idOrName)}${force ? '?force=true' : ''}`),
  
  // Remove image from both database and host
  removeFromDBAndHost: (id: number, force?: boolean) => 
    del<void>(`/admin/images/${id}/complete${force ? '?force=true' : ''}`),
}

