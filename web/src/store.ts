import type { TokenResponse, User } from './types'

export type SlotId = 'A' | 'B'

export interface Session {
  access: string
  refresh: string
  user: User | null
}

const KEY = (s: SlotId) => `levelup.session.${s}`

export function loadSession(slot: SlotId): Session | null {
  const raw = localStorage.getItem(KEY(slot))
  if (!raw) return null
  try {
    return JSON.parse(raw) as Session
  } catch {
    return null
  }
}

export function saveSession(slot: SlotId, s: Session | null): void {
  if (s) localStorage.setItem(KEY(slot), JSON.stringify(s))
  else localStorage.removeItem(KEY(slot))
}

export function sessionFromResponse(r: TokenResponse): Session {
  return { access: r.access_token, refresh: r.refresh_token, user: r.user }
}

export function patchSessionUser(slot: SlotId, user: User): Session | null {
  const s = loadSession(slot)
  if (!s) return null
  const next = { ...s, user }
  saveSession(slot, next)
  return next
}
