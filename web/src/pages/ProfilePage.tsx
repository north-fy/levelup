import { useEffect, useState } from 'react'
import { api } from '../api'
import { patchSessionUser, type Session, type SlotId } from '../store'
import type { User } from '../types'
import { Badge, Button, Card, ErrorText, Field, Input, TextArea, msg } from '../ui'

export default function ProfilePage({
  slot,
  token,
  onSession,
}: {
  slot: SlotId
  token: string
  onSession: (s: Session | null) => void
}) {
  const [user, setUser] = useState<User | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [nickname, setNickname] = useState('')
  const [status, setStatus] = useState('')
  const [avatar, setAvatar] = useState('')

  async function load() {
    setError(null)
    try {
      setUser(await api.me(token))
    } catch (e) {
      setError(msg(e))
    }
  }

  useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token])

  useEffect(() => {
    if (user) {
      setNickname(user.nickname)
      setStatus(user.status)
      setAvatar(user.avatar_url)
    }
  }, [user])

  async function save() {
    setBusy(true)
    setError(null)
    try {
      const u = await api.updateMe(token, { nickname, status, avatar_url: avatar })
      setUser(u)
      const next = patchSessionUser(slot, u)
      if (next) onSession(next)
    } catch (e) {
      setError(msg(e))
    } finally {
      setBusy(false)
    }
  }

  if (!user) return error ? <ErrorText error={error} /> : <div className="flex justify-center py-10">Loading…</div>

  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Card title="Профиль">
        <div className="space-y-2 text-sm">
          <div className="flex items-center gap-2">
            {user.avatar_url && <img src={user.avatar_url} alt="" className="h-10 w-10 rounded-full" />}
            <div>
              <div className="font-semibold">{user.nickname}</div>
              <div className="text-slate-400">{user.email}</div>
            </div>
          </div>
          <div className="flex gap-2">
            <Badge color="indigo">Lv {user.level}</Badge>
            <Badge color="amber">XP {user.xp}</Badge>
            <Badge color="emerald">💰 {user.gold}</Badge>
          </div>
          <div className="pt-1 text-xs text-slate-400">
            id={user.id} · создан {new Date(user.created_at).toLocaleDateString()}
          </div>
          {user.status && <div className="text-slate-600">{user.status}</div>}
        </div>
      </Card>

      <Card title="Изменить">
        <div className="space-y-3">
          <Field label="Nickname">
            <Input value={nickname} onChange={(e) => setNickname(e.target.value)} />
          </Field>
          <Field label="Status">
            <TextArea value={status} onChange={(e) => setStatus(e.target.value)} rows={2} />
          </Field>
          <Field label="Avatar URL">
            <Input value={avatar} onChange={(e) => setAvatar(e.target.value)} placeholder="https://…" />
          </Field>
          <ErrorText error={error} />
          <Button onClick={() => void save()} disabled={busy}>
            {busy ? '…' : 'Сохранить'}
          </Button>
        </div>
      </Card>
    </div>
  )
}
