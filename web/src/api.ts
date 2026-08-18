import type {
  Branch,
  BranchStat,
  OverviewStats,
  Purchase,
  Quest,
  QuestStat,
  QuestType,
  Roadmap,
  RoadmapDetail,
  RoadmapNode,
  RoadmapStat,
  ShopItem,
  TokenResponse,
  User,
  WorkshopRoadmap,
} from './types'

const BASE = ''

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message)
  }
}

async function request<T>(method: string, path: string, token?: string | null, body?: unknown): Promise<T> {
  const res = await fetch(BASE + path, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  const text = await res.text()
  let data: unknown = null
  try {
    data = text ? JSON.parse(text) : null
  } catch {
    data = text
  }
  if (!res.ok) {
    const msg =
      data && typeof data === 'object' && 'error' in data
        ? String((data as { error: string }).error)
        : `HTTP ${res.status}`
    throw new ApiError(msg, res.status)
  }
  return data as T
}

export const api = {
  register: (email: string, password: string, nickname: string) =>
    request<TokenResponse>('POST', '/api/v1/auth/register', null, { email, password, nickname }),
  login: (email: string, password: string) =>
    request<TokenResponse>('POST', '/api/v1/auth/login', null, { email, password }),
  refresh: (refreshToken: string) =>
    request<TokenResponse>('POST', '/api/v1/auth/refresh', null, { refresh_token: refreshToken }),
  logout: (access: string, refresh: string, token: string) =>
    request<unknown>('POST', '/api/v1/auth/logout', token, { access_token: access, refresh_token: refresh }),

  me: (token: string) => request<User>('GET', '/api/v1/users/me', token),
  updateMe: (token: string, patch: { nickname?: string; status?: string; avatar_url?: string }) =>
    request<User>('PATCH', '/api/v1/users/me', token, patch),

  listBranches: (token: string) => request<Branch[]>('GET', '/api/v1/branches', token),
  createBranch: (token: string, body: { name: string; description?: string; color?: string; icon?: string }) =>
    request<Branch>('POST', '/api/v1/branches', token, body),
  updateBranch: (token: string, id: number, body: { name?: string; description?: string; color?: string; icon?: string }) =>
    request<Branch>('PATCH', `/api/v1/branches/${id}`, token, body),
  deleteBranch: (token: string, id: number) => request<unknown>('DELETE', `/api/v1/branches/${id}`, token),

  listQuests: (token: string, branchId: number) =>
    request<Quest[]>('GET', `/api/v1/branches/${branchId}/quests`, token),
  createQuest: (
    token: string,
    branchId: number,
    body: { title: string; description?: string; type: QuestType; reward_xp?: number; reward_gold?: number; duration_hours?: number },
  ) => request<Quest>('POST', `/api/v1/branches/${branchId}/quests`, token, body),
  completeQuest: (token: string, id: number) => request<Quest>('POST', `/api/v1/quests/${id}/complete`, token),
  startQuest: (token: string, id: number) => request<Quest>('POST', `/api/v1/quests/${id}/start`, token),
  stopQuest: (token: string, id: number) => request<Quest>('POST', `/api/v1/quests/${id}/stop`, token),
  deleteQuest: (token: string, id: number) => request<unknown>('DELETE', `/api/v1/quests/${id}`, token),

  listItems: (token: string) => request<ShopItem[]>('GET', '/api/v1/shop/items', token),
  createItem: (token: string, body: { title: string; description?: string; price_gold: number }) =>
    request<ShopItem>('POST', '/api/v1/shop/items', token, body),
  updateItem: (token: string, id: number, body: { title?: string; description?: string; price_gold?: number; is_active?: boolean }) =>
    request<ShopItem>('PATCH', `/api/v1/shop/items/${id}`, token, body),
  deleteItem: (token: string, id: number) => request<unknown>('DELETE', `/api/v1/shop/items/${id}`, token),
  buyItem: (token: string, id: number) => request<ShopItem>('POST', `/api/v1/shop/items/${id}/buy`, token),
  listPurchases: (token: string) => request<Purchase[]>('GET', '/api/v1/shop/purchases', token),

  listRoadmaps: (token: string) => request<Roadmap[]>('GET', '/api/v1/roadmaps', token),
  createRoadmap: (token: string, body: { title: string; description?: string }) =>
    request<Roadmap>('POST', '/api/v1/roadmaps', token, body),
  getRoadmap: (token: string, id: number) => request<RoadmapDetail>('GET', `/api/v1/roadmaps/${id}`, token),
  deleteRoadmap: (token: string, id: number) => request<unknown>('DELETE', `/api/v1/roadmaps/${id}`, token),
  addNode: (
    token: string,
    id: number,
    body: { title: string; description?: string; type: QuestType; reward_xp?: number; reward_gold?: number; duration_hours?: number; dependencies?: number[] },
  ) => request<RoadmapNode>('POST', `/api/v1/roadmaps/${id}/nodes`, token, body),
  completeNode: (token: string, id: number, nodeId: number) =>
    request<RoadmapNode>('POST', `/api/v1/roadmaps/${id}/nodes/${nodeId}/complete`, token),

  listWorkshops: (token: string) => request<WorkshopRoadmap[]>('GET', '/api/v1/workshop/roadmaps', token),
  createWorkshop: (token: string, body: { roadmap_id: number; title?: string; description?: string }) =>
    request<WorkshopRoadmap>('POST', '/api/v1/workshop/roadmaps', token, body),
  updateWorkshop: (token: string, id: number, body: { title?: string; description?: string; is_published?: boolean }) =>
    request<WorkshopRoadmap>('PATCH', `/api/v1/workshop/roadmaps/${id}`, token, body),
  installWorkshop: (token: string, id: number) =>
    request<RoadmapDetail>('POST', `/api/v1/workshop/roadmaps/${id}/install`, token),

  statsOverview: (token: string) => request<OverviewStats>('GET', '/api/v1/stats/overview', token),
  statsBranches: (token: string) => request<BranchStat[]>('GET', '/api/v1/stats/branches', token),
  statsRoadmaps: (token: string) => request<RoadmapStat[]>('GET', '/api/v1/stats/roadmaps', token),
  statsQuests: (token: string) => request<QuestStat[]>('GET', '/api/v1/stats/quests', token),
}
