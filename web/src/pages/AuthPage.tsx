import { useState, type FormEvent } from 'react'
import { api } from '../api'
import { loadSession, saveSession, sessionFromResponse, type Session, type SlotId } from '../store'
import Button from '../components/ui/Button'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import FormField from '../components/ui/FormField'
import { Input } from '../components/ui/Input'
import { useToast } from '../toast'

export default function AuthPage({ slot, onSession }: { slot: SlotId; onSession: (s: Session | null) => void }) {
  const [mode, setMode] = useState<'login' | 'register'>('register')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [nickname, setNickname] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const { notify } = useToast()
  const session = loadSession(slot)

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!email || !password) {
      setError('Заполните email и пароль')
      return
    }
    setBusy(true)
    setError(null)
    try {
      const r =
        mode === 'register'
          ? await api.register(email, password, nickname || email.split('@')[0])
          : await api.login(email, password)
      const s = sessionFromResponse(r)
      saveSession(slot, s)
      onSession(s)
      notify(mode === 'register' ? 'Аккаунт создан 🎉' : 'Вход выполнен ✅', 'success')
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e)
      setError(message)
      notify(message, 'error')
    } finally {
      setBusy(false)
    }
  }

  function reset() {
    saveSession(slot, null)
    onSession(null)
    notify('Сессия слота сброшена', 'info')
  }

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card title={`Аккаунт слота ${slot}`} actions={session ? <Button variant="secondary" size="sm" onClick={reset}>Сбросить</Button> : undefined}>
        {session?.user ? (
          <div className="space-y-3 text-sm">
            <div className="flex items-center gap-3">
              <div className="flex h-11 w-11 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 to-fuchsia-500 text-lg font-black text-white">
                {session.user.nickname.slice(0, 1).toUpperCase()}
              </div>
              <div>
                <div className="font-bold text-slate-800 dark:text-slate-100">{session.user.nickname}</div>
                <div className="text-xs text-slate-500 dark:text-slate-400">{session.user.email}</div>
              </div>
            </div>
            <div className="flex flex-wrap gap-2">
              <Badge color="indigo">Lv {session.user.level}</Badge>
              <Badge color="amber">XP {session.user.xp}</Badge>
              <Badge color="emerald">💰 {session.user.gold}</Badge>
            </div>
            <dl className="space-y-1 text-xs">
              <div className="flex justify-between gap-3">
                <dt className="text-slate-400">access_token</dt>
                <dd className="truncate font-mono text-slate-500 dark:text-slate-300">{session.access.slice(0, 32)}…</dd>
              </div>
              <div className="flex justify-between gap-3">
                <dt className="text-slate-400">refresh_token</dt>
                <dd className="truncate font-mono text-slate-500 dark:text-slate-300">{session.refresh.slice(0, 32)}…</dd>
              </div>
            </dl>
            <a href="/api/v1/auth/github/redirect" className="inline-block text-sm font-medium text-indigo-600 underline-offset-2 hover:underline dark:text-indigo-300">
              Войти через GitHub →
            </a>
          </div>
        ) : (
          <p className="text-sm text-slate-500 dark:text-slate-400">
            Нет сессии. Зарегистрируйте новый аккаунт или войдите — форма справа.
          </p>
        )}
      </Card>

      <Card title={mode === 'register' ? 'Регистрация' : 'Вход'}>
        <form onSubmit={(e) => void submit(e)} className="space-y-4" noValidate>
          <FormField label="Email">
            <Input
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="user@example.com"
            />
          </FormField>
          <FormField label="Пароль">
            <Input
              type="password"
              autoComplete={mode === 'register' ? 'new-password' : 'current-password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
            />
          </FormField>
          {mode === 'register' && (
            <FormField label="Никнейм" hint="Необязательно. По умолчанию — часть email до @">
              <Input value={nickname} onChange={(e) => setNickname(e.target.value)} placeholder="hero" />
            </FormField>
          )}
          {error && (
            <p role="alert" className="rounded-xl border border-rose-500/30 bg-rose-50 px-3 py-2 text-sm text-rose-700 dark:bg-rose-950/50 dark:text-rose-300">
              {error}
            </p>
          )}
          <div className="flex flex-wrap items-center gap-3">
            <Button type="submit" disabled={busy || !email || !password}>
              {busy ? '…' : mode === 'register' ? 'Зарегистрироваться' : 'Войти'}
            </Button>
            <Button type="button" variant="ghost" onClick={() => setMode(mode === 'register' ? 'login' : 'register')}>
              {mode === 'register' ? 'Есть аккаунт — войти' : 'Нет аккаунта — регистрация'}
            </Button>
          </div>
        </form>
      </Card>
    </div>
  )
}