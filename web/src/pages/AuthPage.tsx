import { useState } from 'react'
import { api } from '../api'
import { loadSession, saveSession, sessionFromResponse, type Session, type SlotId } from '../store'
import { Button, Card, ErrorText, Field, Input } from '../ui'

export default function AuthPage({ slot, onSession }: { slot: SlotId; onSession: (s: Session | null) => void }) {
  const [mode, setMode] = useState<'login' | 'register'>('register')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [nickname, setNickname] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const session = loadSession(slot)

  async function submit() {
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
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Card
        title={`Аккаунт слота ${slot}`}
        actions={
          session && (
            <Button variant="ghost" onClick={() => { saveSession(slot, null); onSession(null) }}>
              Сбросить
            </Button>
          )
        }
      >
        {session ? (
          <div className="space-y-2 text-sm">
            <div>
              Пользователь: <b>{session.user?.nickname ?? '—'}</b> ({session.user?.email ?? '—'})
            </div>
            <div>
              Gold <b>{session.user?.gold}</b> · XP <b>{session.user?.xp}</b> · Level <b>{session.user?.level}</b>
            </div>
            <div className="truncate text-xs text-slate-400">access: {session.access.slice(0, 28)}…</div>
            <div className="truncate text-xs text-slate-400">refresh: {session.refresh.slice(0, 28)}…</div>
            <div className="pt-2">
              <a href="/api/v1/auth/github/redirect" className="text-sm text-indigo-600 underline">
                Войти через GitHub
              </a>
            </div>
          </div>
        ) : (
          <div className="text-sm text-slate-400">Нет сессии. Зарегистрируйте или войдите.</div>
        )}
      </Card>

      <Card title={mode === 'register' ? 'Регистрация' : 'Вход'}>
        <div className="space-y-3">
          <Field label="Email">
            <Input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="user@example.com" />
          </Field>
          <Field label="Password">
            <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="••••••••" />
          </Field>
          {mode === 'register' && (
            <Field label="Nickname">
              <Input value={nickname} onChange={(e) => setNickname(e.target.value)} placeholder="hero" />
            </Field>
          )}
          <ErrorText error={error} />
          <div className="flex items-center gap-2">
            <Button onClick={submit} disabled={busy || !email || !password}>
              {busy ? '…' : mode === 'register' ? 'Зарегистрироваться' : 'Войти'}
            </Button>
            <Button variant="ghost" onClick={() => setMode(mode === 'register' ? 'login' : 'register')}>
              {mode === 'register' ? 'Есть аккаунт — войти' : 'Нет аккаунта — регистрация'}
            </Button>
          </div>
        </div>
      </Card>
    </div>
  )
}
