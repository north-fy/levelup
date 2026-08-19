import { useEffect, useState, type FormEvent } from 'react'
import { api } from '../api'
import { patchSessionUser, type Session, type SlotId } from '../store'
import type { User } from '../types'
import Button from '../components/ui/Button'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import FormField from '../components/ui/FormField'
import { Input, TextArea } from '../components/ui/Input'
import Spinner from '../components/ui/Spinner'
import { useToast } from '../toast'

function levelThreshold(level: number): number {
  return 50 * (level - 1) * (level - 1)
}

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
  const { notify } = useToast()

  useEffect(() => {
    let cancelled = false
    setError(null)
    api
      .me(token)
      .then((u) => {
        if (!cancelled) setUser(u)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e))
      })
    return () => {
      cancelled = true
    }
  }, [token])

  useEffect(() => {
    if (user) {
      setNickname(user.nickname)
      setStatus(user.status)
      setAvatar(user.avatar_url)
    }
  }, [user])

  async function save(e: FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      const u = await api.updateMe(token, { nickname, status, avatar_url: avatar })
      setUser(u)
      const next = patchSessionUser(slot, u)
      if (next) onSession(next)
      notify('Профиль обновлён ✅', 'success')
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    } finally {
      setBusy(false)
    }
  }

  if (!user) return error ? <p role="alert" className="rounded-xl border border-rose-500/30 bg-rose-50 px-3 py-2 text-sm text-rose-700 dark:bg-rose-950/50 dark:text-rose-300">{error}</p> : <Spinner />

  const cur = levelThreshold(user.level)
  const next = levelThreshold(user.level + 1)
  const pct = Math.min(100, Math.max(0, Math.round(((user.xp - cur) / Math.max(1, next - cur)) * 100)))

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card title="Персонаж">
        <div className="flex items-center gap-4">
          {user.avatar_url ? (
            <img src={user.avatar_url} alt="Аватар" className="h-16 w-16 rounded-2xl object-cover ring-2 ring-indigo-500/40" />
          ) : (
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-indigo-500 to-fuchsia-500 text-2xl font-black text-white">
              {user.nickname.slice(0, 1).toUpperCase()}
            </div>
          )}
          <div>
            <div className="text-lg font-black text-slate-800 dark:text-slate-100">{user.nickname}</div>
            <div className="text-sm text-slate-500 dark:text-slate-400">{user.email}</div>
          </div>
        </div>

        <div className="mt-5 flex flex-wrap gap-2">
          <Badge color="indigo">Уровень {user.level}</Badge>
          <Badge color="amber">XP {user.xp}</Badge>
          <Badge color="emerald">💰 {user.gold}</Badge>
        </div>

        <div className="mt-5">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500 dark:text-slate-400">
            <span>Прогресс до уровня {user.level + 1}</span>
            <span>{pct}%</span>
          </div>
          <div
            role="progressbar"
            aria-valuenow={pct}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label={`Прогресс до уровня ${user.level + 1}: ${pct} процентов`}
            className="mt-1.5 h-3 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-800"
          >
            <div
              className="h-full rounded-full bg-gradient-to-r from-indigo-500 via-violet-500 to-fuchsia-500 transition-all"
              style={{ width: `${pct}%` }}
            />
          </div>
          <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">
            {user.xp} XP сейчас · {next} XP нужно для уровня {user.level + 1}
          </p>
        </div>

        {user.status && (
          <p className="mt-4 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm italic text-slate-600 dark:border-slate-800 dark:bg-slate-800/50 dark:text-slate-300">
            «{user.status}»
          </p>
        )}
      </Card>

      <Card title="Редактирование">
        <form onSubmit={(e) => void save(e)} className="space-y-4" noValidate>
          <FormField label="Никнейм">
            <Input value={nickname} onChange={(e) => setNickname(e.target.value)} />
          </FormField>
          <FormField label="Статус" hint="Короткая строка-«цитата» персонажа">
            <TextArea value={status} onChange={(e) => setStatus(e.target.value)} rows={2} />
          </FormField>
          <FormField label="URL аватара">
            <Input value={avatar} onChange={(e) => setAvatar(e.target.value)} placeholder="https://…" />
          </FormField>
          {error && (
            <p role="alert" className="rounded-xl border border-rose-500/30 bg-rose-50 px-3 py-2 text-sm text-rose-700 dark:bg-rose-950/50 dark:text-rose-300">
              {error}
            </p>
          )}
          <Button type="submit" disabled={busy || !nickname.trim()}>
            {busy ? 'Сохраняем…' : 'Сохранить'}
          </Button>
        </form>
      </Card>
    </div>
  )
}