import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import type { Branch, Quest, QuestType } from '../types'
import { Badge, Button, Card, ErrorText, Field, Input, Select, STATUS_COLOR, TextArea, fmtTime, msg, useRefresh } from '../ui'

const TYPES: QuestType[] = ['simple', 'timed']

export default function BranchesPage({ token }: { token: string }) {
  const [branches, setBranches] = useState<Branch[]>([])
  const [selected, setSelected] = useState<Branch | null>(null)
  const [quests, setQuests] = useState<Quest[]>([])
  const [error, setError] = useState<string | null>(null)
  const [tick, refresh] = useRefresh()

  const [bName, setBName] = useState('')
  const [bDesc, setBDesc] = useState('')
  const [bColor, setBColor] = useState('')

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
      setError(msg(e))
    }
  }, [token])

  useEffect(() => {
    void loadBranches()
  }, [loadBranches, tick])

  useEffect(() => {
    if (!selected) {
      setQuests([])
      return
    }
    api
      .listQuests(token, selected.id)
      .then(setQuests)
      .catch((e) => setError(msg(e)))
  }, [token, selected, tick])

  async function createBranch() {
    if (!bName.trim()) return
    setError(null)
    try {
      const b = await api.createBranch(token, { name: bName, description: bDesc, color: bColor })
      setBName('')
      setBDesc('')
      await loadBranches()
      setSelected(b)
    } catch (e) {
      setError(msg(e))
    }
  }

  async function deleteBranch(b: Branch) {
    setError(null)
    try {
      await api.deleteBranch(token, b.id)
      setSelected(null)
      await loadBranches()
    } catch (e) {
      setError(msg(e))
    }
  }

  async function createQuest() {
    if (!selected || !qTitle.trim()) return
    setError(null)
    try {
      await api.createQuest(token, selected.id, {
        title: qTitle,
        description: qDesc,
        type: qType,
        reward_xp: Number(qXp) || 0,
        reward_gold: Number(qGold) || 0,
        duration_hours: qType === 'timed' ? Number(qHours) || 1 : 0,
      })
      setQTitle('')
      setQDesc('')
      refresh()
    } catch (e) {
      setError(msg(e))
    }
  }

  async function act(fn: () => Promise<unknown>) {
    setError(null)
    try {
      await fn()
      refresh()
    } catch (e) {
      setError(msg(e))
    }
  }

  return (
    <div className="grid gap-4 lg:grid-cols-[320px_1fr]">
      <div className="space-y-4">
        <Card title="Новая ветка">
          <div className="space-y-3">
            <Field label="Название">
              <Input value={bName} onChange={(e) => setBName(e.target.value)} placeholder="Карьера" />
            </Field>
            <Field label="Описание">
              <TextArea value={bDesc} onChange={(e) => setBDesc(e.target.value)} rows={2} />
            </Field>
            <div className="grid grid-cols-2 gap-2">
              <Field label="Цвет">
                <Input value={bColor} onChange={(e) => setBColor(e.target.value)} placeholder="#6366f1" />
              </Field>
              <Field label="Иконка">
                <Input value="" disabled placeholder="—" />
              </Field>
            </div>
            <Button onClick={() => void createBranch()} disabled={!bName.trim()}>
              Создать ветку
            </Button>
          </div>
        </Card>

        <Card title={`Ветки (${branches.length})`}>
          <div className="max-h-96 space-y-1 overflow-y-auto">
            {branches.map((b) => (
              <div
                key={b.id}
                className={`flex cursor-pointer items-center justify-between rounded-lg px-3 py-2 text-sm ${
                  selected?.id === b.id ? 'bg-indigo-50 text-indigo-700' : 'hover:bg-slate-50'
                }`}
                onClick={() => setSelected(b)}
              >
                <span className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full" style={{ background: b.color || '#94a3b8' }} />
                  {b.name}
                </span>
                <button
                  className="text-xs text-slate-400 hover:text-rose-500"
                  onClick={(e) => {
                    e.stopPropagation()
                    void deleteBranch(b)
                  }}
                >
                  ✕
                </button>
              </div>
            ))}
            {branches.length === 0 && <div className="text-sm text-slate-400">Веток нет</div>}
          </div>
        </Card>
      </div>

      <div className="space-y-4">
        {selected && (
          <>
            <Card
              title={selected.name}
              actions={<Badge color="indigo">id {selected.id}</Badge>}
            >
              <p className="text-sm text-slate-600">{selected.description || 'Без описания'}</p>
            </Card>

            <Card title="Новый квест">
              <div className="grid gap-3 md:grid-cols-2">
                <Field label="Название">
                  <Input value={qTitle} onChange={(e) => setQTitle(e.target.value)} placeholder="Выучить Go" />
                </Field>
                <Field label="Тип">
                  <Select value={qType} onChange={(e) => setQType(e.target.value as QuestType)}>
                    {TYPES.map((t) => (
                      <option key={t} value={t}>
                        {t}
                      </option>
                    ))}
                  </Select>
                </Field>
                <div className="md:col-span-2">
                  <Field label="Описание">
                    <TextArea value={qDesc} onChange={(e) => setQDesc(e.target.value)} rows={2} />
                  </Field>
                </div>
                <Field label="Reward XP">
                  <Input type="number" value={qXp} onChange={(e) => setQXp(e.target.value)} />
                </Field>
                <Field label="Reward Gold">
                  <Input type="number" value={qGold} onChange={(e) => setQGold(e.target.value)} />
                </Field>
                {qType === 'timed' && (
                  <Field label="Длительность (часы)">
                    <Input type="number" value={qHours} onChange={(e) => setQHours(e.target.value)} />
                  </Field>
                )}
              </div>
              <div className="mt-3">
                <Button onClick={() => void createQuest()} disabled={!qTitle.trim()}>
                  Создать квест
                </Button>
              </div>
            </Card>

            <Card title={`Квесты (${quests.length})`}>
              <div className="space-y-2">
                {quests.map((q) => (
                  <div key={q.id} className="flex items-start justify-between gap-3 rounded-lg border border-slate-100 p-3">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-medium text-slate-800">{q.title}</span>
                        <Badge color={q.type === 'timed' ? 'amber' : 'slate'}>{q.type}</Badge>
                        <Badge color={STATUS_COLOR[q.status]}>{q.status}</Badge>
                      </div>
                      {q.description && <div className="mt-1 text-sm text-slate-500">{q.description}</div>}
                      <div className="mt-1 text-xs text-slate-400">
                        +{q.reward_xp} XP · +{q.reward_gold} 💰
                        {q.duration_hours > 0 && ` · ${q.duration_hours}ч`} · старт {fmtTime(q.started_at)}
                      </div>
                    </div>
                    <div className="flex shrink-0 flex-wrap gap-1">
                      {q.status === 'todo' && (
                        <Button variant="ghost" onClick={() => void act(() => api.startQuest(token, q.id))}>
                          Старт
                        </Button>
                      )}
                      {q.status === 'in_progress' && (
                        <Button variant="ghost" onClick={() => void act(() => api.stopQuest(token, q.id))}>
                          Стоп
                        </Button>
                      )}
                      {q.status !== 'done' && (
                        <Button
                          variant="success"
                          onClick={() => void act(() => api.completeQuest(token, q.id))}
                        >
                          Выполнить
                        </Button>
                      )}
                      <Button
                        variant="danger"
                        onClick={() => void act(() => api.deleteQuest(token, q.id))}
                      >
                        ✕
                      </Button>
                    </div>
                  </div>
                ))}
                {quests.length === 0 && <div className="text-sm text-slate-400">Квестов нет</div>}
              </div>
            </Card>
          </>
        )}
        <ErrorText error={error} />
      </div>
    </div>
  )
}
