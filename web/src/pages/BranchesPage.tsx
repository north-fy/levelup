import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api } from '../api'
import type { Branch, Quest, QuestType } from '../types'
import Button from '../components/ui/Button'
import Card from '../components/ui/Card'
import Badge, { type BadgeColor } from '../components/ui/Badge'
import FormField from '../components/ui/FormField'
import { Input, Select, TextArea } from '../components/ui/Input'
import Spinner from '../components/ui/Spinner'
import { useToast } from '../toast'

const TYPES: QuestType[] = ['simple', 'timed']

const STATUS_COLOR: Record<Quest['status'], BadgeColor> = {
  todo: 'slate',
  in_progress: 'amber',
  done: 'emerald',
  cancelled: 'rose',
}

const STATUS_LABEL: Record<Quest['status'], string> = {
  todo: 'todo',
  in_progress: 'in_progress',
  done: 'done',
  cancelled: 'cancelled',
}

function fmtTime(s: string | null): string {
  if (!s) return '—'
  const d = new Date(s)
  return isNaN(d.getTime()) ? s : d.toLocaleString()
}

export default function BranchesPage({ token }: { token: string }) {
  const [branches, setBranches] = useState<Branch[] | null>(null)
  const [selected, setSelected] = useState<Branch | null>(null)
  const [quests, setQuests] = useState<Quest[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const { notify } = useToast()

  const [bName, setBName] = useState('')
  const [bDesc, setBDesc] = useState('')
  const [bColor, setBColor] = useState('#6366f1')

  const [qTitle, setQTitle] = useState('')
  const [qDesc, setQDesc] = useState('')
  const [qType, setQType] = useState<QuestType>('simple')
  const [qXp, setQXp] = useState('20')
  const [qGold, setQGold] = useState('10')
  const [qHours, setQHours] = useState('1')

  const loadBranches = useCallback(async () => {
    setError(null)
    try {
      const list = await api.listBranches(token)
      setBranches(list)
      setSelected((prev) => (prev ? list.find((b) => b.id === prev.id) ?? list[0] ?? null : list[0] ?? null))
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }, [token])

  useEffect(() => {
    void loadBranches()
  }, [loadBranches])

  useEffect(() => {
    if (!selected) {
      setQuests(null)
      return
    }
    let cancelled = false
    setError(null)
    api
      .listQuests(token, selected.id)
      .then((q) => {
        if (!cancelled) setQuests(q)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e))
      })
    return () => {
      cancelled = true
    }
  }, [token, selected])

  async function createBranch(e: FormEvent) {
    e.preventDefault()
    if (!bName.trim()) return
    setError(null)
    try {
      const b = await api.createBranch(token, { name: bName.trim(), description: bDesc, color: bColor })
      setBName('')
      setBDesc('')
      notify('Ветка создана 🌿', 'success')
      setSelected(b)
      await loadBranches()
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  async function deleteBranch(b: Branch) {
    setError(null)
    try {
      await api.deleteBranch(token, b.id)
      notify('Ветка удалена', 'info')
      setSelected(null)
      await loadBranches()
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  async function createQuest(e: FormEvent) {
    e.preventDefault()
    if (!selected || !qTitle.trim()) return
    setError(null)
    try {
      await api.createQuest(token, selected.id, {
        title: qTitle.trim(),
        description: qDesc,
        type: qType,
        reward_xp: Number(qXp) || 0,
        reward_gold: Number(qGold) || 0,
        duration_hours: qType === 'timed' ? Number(qHours) || 1 : 0,
      })
      setQTitle('')
      setQDesc('')
      notify('Квест создан 📜', 'success')
      const list = await api.listQuests(token, selected.id)
      setQuests(list)
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  async function act(label: string, fn: () => Promise<unknown>) {
    setError(null)
    try {
      await fn()
      notify(label, 'success')
      if (selected) setQuests(await api.listQuests(token, selected.id))
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  return (
    <div className="grid gap-4 lg:grid-cols-[340px_1fr]">
      <div className="space-y-4">
        <Card title="Новая ветка">
          <form onSubmit={(e) => void createBranch(e)} className="space-y-3" noValidate>
            <FormField label="Название">
              <Input value={bName} onChange={(e) => setBName(e.target.value)} placeholder="Карьера" />
            </FormField>
            <FormField label="Описание">
              <TextArea value={bDesc} onChange={(e) => setBDesc(e.target.value)} rows={2} />
            </FormField>
            <FormField label="Цвет">
              <Input type="color" value={bColor} onChange={(e) => setBColor(e.target.value)} className="h-10 p-1" />
            </FormField>
            <Button type="submit" disabled={!bName.trim()}>
              Создать ветку
            </Button>
          </form>
        </Card>

        <Card title={`Ветки${branches ? ` (${branches.length})` : ''}`}>
          {!branches ? (
            <Spinner label="Загрузка веток" />
          ) : branches.length === 0 ? (
            <p className="text-sm text-slate-500 dark:text-slate-400">Веток нет — создайте первую</p>
          ) : (
            <ul className="max-h-96 space-y-1 overflow-y-auto">
              {branches.map((b) => (
                <li key={b.id}>
                  <div
                    className={`flex w-full items-center justify-between gap-1 rounded-xl px-3 py-2 text-left text-sm transition ${
                      selected?.id === b.id
                        ? 'bg-indigo-600/10 font-semibold text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300'
                        : 'text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800'
                    }`}
                  >
                    <button
                      type="button"
                      aria-pressed={selected?.id === b.id}
                      onClick={() => setSelected(b)}
                      className="flex min-w-0 flex-1 items-center gap-2 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500"
                    >
                      <span aria-hidden="true" className="h-3 w-3 shrink-0 rounded-full" style={{ background: b.color || '#94a3b8' }} />
                      <span className="truncate">{b.name}</span>
                    </button>
                    <button
                      type="button"
                      aria-label={`Удалить ветку ${b.name}`}
                      onClick={() => void deleteBranch(b)}
                      className="rounded p-1 text-slate-400 hover:text-rose-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-rose-500"
                    >
                      ✕
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </Card>
      </div>

      <div className="space-y-4">
        {selected && branches ? (
          <>
            <Card title={selected.name} actions={<Badge color="indigo">id {selected.id}</Badge>}>
              <p className="text-sm text-slate-600 dark:text-slate-300">{selected.description || 'Без описания'}</p>
            </Card>

            <Card title="Новый квест">
              <form onSubmit={(e) => void createQuest(e)} className="grid gap-3 md:grid-cols-2" noValidate>
                <FormField label="Название">
                  <Input value={qTitle} onChange={(e) => setQTitle(e.target.value)} placeholder="Выучить Go" />
                </FormField>
                <FormField label="Тип">
                  <Select value={qType} onChange={(e) => setQType(e.target.value as QuestType)}>
                    {TYPES.map((t) => (
                      <option key={t} value={t}>
                        {t === 'simple' ? 'simple (мгновенная награда)' : 'timed (по часам)'}
                      </option>
                    ))}
                  </Select>
                </FormField>
                <div className="md:col-span-2">
                  <FormField label="Описание">
                    <TextArea value={qDesc} onChange={(e) => setQDesc(e.target.value)} rows={2} />
                  </FormField>
                </div>
                <FormField label="Reward XP">
                  <Input type="number" min={0} value={qXp} onChange={(e) => setQXp(e.target.value)} />
                </FormField>
                <FormField label="Reward Gold">
                  <Input type="number" min={0} value={qGold} onChange={(e) => setQGold(e.target.value)} />
                </FormField>
                {qType === 'timed' && (
                  <FormField label="Длительность (часы)">
                    <Input type="number" min={1} value={qHours} onChange={(e) => setQHours(e.target.value)} />
                  </FormField>
                )}
                <div className="md:col-span-2">
                  <Button type="submit" disabled={!qTitle.trim()}>
                    Создать квест
                  </Button>
                </div>
              </form>
            </Card>

            <Card title={`Квесты${quests ? ` (${quests.length})` : ''}`}>
              {!quests ? (
                <Spinner label="Загрузка квестов" />
              ) : quests.length === 0 ? (
                <p className="text-sm text-slate-500 dark:text-slate-400">Квестов нет — добавьте первый</p>
              ) : (
                <ul className="space-y-2">
                  {quests.map((q) => (
                    <li key={q.id} className="rounded-2xl border border-slate-200 bg-white/70 p-3 dark:border-slate-800 dark:bg-slate-900/50">
                      <div className="flex flex-wrap items-start justify-between gap-2">
                        <div className="min-w-0">
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="font-bold text-slate-800 dark:text-slate-100">{q.title}</span>
                            <Badge color={q.type === 'timed' ? 'amber' : 'slate'}>{q.type}</Badge>
                            <Badge color={STATUS_COLOR[q.status]}>{STATUS_LABEL[q.status]}</Badge>
                          </div>
                          {q.description && <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{q.description}</p>}
                          <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">
                            +{q.reward_xp} XP · +{q.reward_gold} 💰
                            {q.duration_hours > 0 && ` · до ${q.duration_hours}ч`}
                            {q.type === 'timed' && ` · старт ${fmtTime(q.started_at)}`}
                          </p>
                        </div>
                        <div className="flex shrink-0 flex-wrap gap-1">
                          {q.status === 'todo' && (
                            <Button variant="secondary" size="sm" onClick={() => void act('Таймер запущен ⏳', () => api.startQuest(token, q.id))}>
                              Старт
                            </Button>
                          )}
                          {q.status === 'in_progress' && (
                            <Button variant="secondary" size="sm" onClick={() => void act('Таймер остановлен', () => api.stopQuest(token, q.id))}>
                              Стоп
                            </Button>
                          )}
                          {q.status !== 'done' && (
                            <Button variant="success" size="sm" onClick={() => void act('Квест выполнен 🎉', () => api.completeQuest(token, q.id))}>
                              Выполнить
                            </Button>
                          )}
                          <Button variant="danger" size="sm" onClick={() => void act('Квест удалён', () => api.deleteQuest(token, q.id))}>
                            ✕
                          </Button>
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </Card>
          </>
        ) : (
          <div className="rounded-2xl border border-dashed border-slate-300 bg-white/60 p-10 text-center text-slate-500 dark:border-slate-700 dark:bg-slate-900/40 dark:text-slate-400">
            Выберите ветку слева
          </div>
        )}
        {error && (
          <p role="alert" className="rounded-xl border border-rose-500/30 bg-rose-50 px-3 py-2 text-sm text-rose-700 dark:bg-rose-950/50 dark:text-rose-300">
            {error}
          </p>
        )}
      </div>
    </div>
  )
}