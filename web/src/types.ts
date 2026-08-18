export interface User {
  id: number
  email: string
  nickname: string
  status: string
  avatar_url: string
  level: number
  xp: number
  gold: number
  created_at: string
  updated_at: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  access_expires_at: number
  refresh_expires_at: number
  user: User
}

export interface Branch {
  id: number
  user_id: number
  name: string
  description: string
  color: string
  icon: string
  created_at: string
  updated_at: string
}

export type QuestType = 'simple' | 'timed'
export type QuestStatus = 'todo' | 'in_progress' | 'done' | 'cancelled'

export interface Quest {
  id: number
  branch_id: number
  user_id: number
  title: string
  description: string
  type: QuestType
  reward_xp: number
  reward_gold: number
  duration_hours: number
  status: QuestStatus
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface ShopItem {
  id: number
  seller_id: number
  title: string
  description: string
  price_gold: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Purchase {
  id: number
  item_id: number
  buyer_id: number
  seller_id: number
  price: number
  created_at: string
}

export interface Roadmap {
  id: number
  user_id: number
  title: string
  description: string
  source_type: string
  source_id: number
  created_at: string
  updated_at: string
}

export interface RoadmapNode {
  id: number
  roadmap_id: number
  title: string
  description: string
  position: number
  type: QuestType
  reward_xp: number
  reward_gold: number
  duration_hours: number
  status: QuestStatus
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface RoadmapEdge {
  id: number
  roadmap_id: number
  from_node_id: number
  to_node_id: number
}

export interface RoadmapDetail extends Roadmap {
  nodes: RoadmapNode[]
  edges: RoadmapEdge[]
}

export interface WorkshopRoadmap {
  id: number
  author_id: number
  source_roadmap_id: number
  title: string
  description: string
  is_published: boolean
  created_at: string
  updated_at: string
}

export interface OverviewStats {
  xp: number
  gold: number
  hours: number
  level: number
}

export interface BranchStat {
  branch_id: number
  completed: number
  xp: number
  gold: number
  hours: number
}

export interface RoadmapStat {
  roadmap_id: number
  completed: number
  xp: number
  gold: number
  hours: number
}

export interface QuestStat {
  date: string
  completed: number
  xp: number
  gold: number
  hours: number
}
