import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import type { QuestType, Roadmap, RoadmapDetail, RoadmapNode } from '../types'
import { Badge, Button, Card, ErrorText, Field, Input, Select, STATUS_COLOR, TextArea, fmtTime, msg, useRefresh } from '../ui'

const TYPES: QuestType[] = ['simple', 'timed']

export default function RoadmapsPage({ token }: { token: string }) {
  const [roadmaps, setRoadmaps] = useState<Roadmap[]>([])
  const [detail, setDetail] = useState<RoadmapDetail | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [tick, refresh] = useRefresh()

  const [rTitle, setRTitle] = useState('')
  const [rDesc, setRDesc] = useState('')

  const [nTitle, setNTitle] = useState('')
  const [nDesc, setNDesc] = useState('')
  const [nType, setNType] = useState<QuestType>('simple')
  const [nXp, setNXp] = useState('30')
  const [nGold, setNGold] = useState('15')
  const [nHours, setNHours] = useState('1')
  const [nDeps, setNDeps] = useState<number[]>([])

  const load = useCallback(async () => {
    setError(null)
    try {
      const list = await api.listRoadmaps(token)
      setRoadmaps(list)
      setDetail((prev) => {
        if (!prev) return null
        return list.find((r) => r.id === prev.id) ? prev : null
      })
    } catch (e) {
      setError(msg(e))
    }
  }, [token])

  useEffect(() => {
    void load()
  }, [load, tick])

  useEffect(() => {
    if (!detail) return
    api
      .getRoadmap(token, detail.id)
      .then(setDetail)
      .catch((e) => setError(msg(e)))
  }, [token, tick]) // eslint-disable-line react-hooks/exhaustive-deps

  async function open(r: Roadmap) {
    try {
      setDetail(await api.getRoadmap(token, r.id))
    } catch (e) {
      setError(msg(e))
    }
  }

  async function createRoadmap() {
    if (!rTitle.trim()) return
    setError(null)
    try {
      const r = await api.createRoadmap(token, { title: rTitle, description: rDesc })
      setRTitle('')
      setRDesc('')
      await load()
      await open(r)
    } catch (e) {
      setError(msg(e))
    }
  }

  async function addNode() {
    if (!detail || !nTitle.trim()) return
    setError(null)
    try {
      await api.addNode(token, detail.id, {
        title: nTitle,
        description: nDesc,
        type: nType,
        reward_xp: Number(nXp) || 0,
        reward_gold: Number(nGold) || 0,
        duration_hours: nType === 'timed' ? Number(nHours) || 1 : 0,
        dependencies: nDeps,
      })
      setNTitle('')
      setNDesc('')
      setNDeps([])
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

  const depNames = (node: RoadmapNode) =>
    detail
      ? detail.edges
          .filter((e) => e.to_node_id === node.id)
          .map((e) => detail?.nodes.find((n) => n.id === e.from_node_id)?.title ?? `#${e.from_node_id}`)
      : []

  return (
    <div className="grid gap-4 lg:grid-cols-[300px_1fr]">
      <div className="space-y-4">
        <Card title="Новая роадмапа">
          <div className="space-y-3">
            <Field label="Название">
              <Input value={rTitle} onChange={(e) => setRTitle(e.target.value)} placeholder="Путь Go-разработчика" />
            </Field>
            <Field label="Описание">
              <TextArea value={rDesc} onChange={(e) => setRDesc(e.target.value)} rows={2} />
            </Field>
            <Button onClick={() => void createRoadmap()} disabled={!rTitle.trim()}>
              Создать
            </Button>
          </div>
        </Card>

        <Card title={`Роадмапы (${roadmaps.length})`}>
          <div className="max-h-96 space-y-1 overflow-y-auto">
            {roadmaps.map((r) => (
              <div
                key={r.id}
                className={`flex cursor-pointer items-center justify-between rounded-lg px-3 py-2 text-sm ${
                  detail?.id === r.id ? 'bg-indigo-50 text-indigo-700' : 'hover:bg-slate-50'
                }`}
                onClick={() => void open(r)}
              >
                {r.title}
                <button
                  className="text-xs text-slate-400 hover:text-rose-500"
                  onClick={(e) => {
                    e.stopPropagation()
                    void act(() => api.deleteRoadmap(token, r.id)).then(() => setDetail(null))
                  }}
                >
                  ✕
                </button>
              </div>
            ))}
            {roadmaps.length === 0 && <div className="text-sm text-slate-400">Роадмап нет</div>}
          </div>
        </Card>
      </div>

      <div className="space-y-4">
        {detail ? (
          <>
            <Card
              title={detail.title}
              actions={
                detail.source_type && detail.source_id ? (
                  <Badge color="indigo">
                    {detail.source_type} #{detail.source_id}
                  </Badge>
                ) : (
                  <Badge color="slate">id {detail.id}</Badge>
                )
              }
            >
              <p className="text-sm text-slate-600">{detail.description || 'Без описания'}</p>
            </Card>

            <Card title="Новый узел">
              <div className="grid gap-3 md:grid-cols-2">
                <Field label="Название">
                  <Input value={nTitle} onChange={(e) => setNTitle(e.target.value)} placeholder="Изучить синтаксис" />
                </Field>
                <Field label="Тип">
                  <Select value={nType} onChange={(e) => setNType(e.target.value as QuestType)}>
                    {TYPES.map((t) => (
                      <option key={t} value={t}>
                        {t}
                      </option>
                    ))}
                  </Select>
                </Field>
                <div className="md:col-span-2">
                  <Field label="Описание">
                    <TextArea value={nDesc} onChange={(e) => setNDesc(e.target.value)} rows={2} />
                  </Field>
                </div>
                <Field label="Reward XP">
                  <Input type="number" value={nXp} onChange={(e) => setNXp(e.target.value)} />
                </Field>
                <Field label="Reward Gold">
                  <Input type="number" value={nGold} onChange={(e) => setNGold(e.target.value)} />
                </Field>
                {nType === 'timed' && (
                  <Field label="Длительность (часы)">
                    <Input type="number" value={nHours} onChange={(e) => setNHours(e.target.value)} />
                  </Field>
                )}
                <Field label="Зависимости (пререквизиты)">
                  <select
                    multiple
                    className="h-28 w-full rounded-lg border border-slate-300 px-2 py-1 text-sm outline-none focus:border-indigo-400"
                    value={nDeps.map(String)}
                    onChange={(e) =>
                      setNDeps(Array.from(e.target.selectedOptions, (o) => Number(o.value)))
                    }
                  >
                    {detail.nodes.map((n) => (
                      <option key={n.id} value={n.id}>
                        {n.title}
                      </option>
                    ))}
                  </select>
                </Field>
              </div>
              <div className="mt-3">
                <Button onClick={() => void addNode()} disabled={!nTitle.trim()}>
                  Добавить узел
                </Button>
              </div>
            </Card>

            <Card title={`Узлы (${detail.nodes.length})`}>
              <div className="space-y-2">
                {detail.nodes.map((n) => (
                  <div key={n.id} className="rounded-lg border border-slate-100 p-3">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-medium text-slate-800">
                          {n.position}. {n.title}
                        </span>
                        <Badge color={n.type === 'timed' ? 'amber' : 'slate'}>{n.type}</Badge>
                        <Badge color={STATUS_COLOR[n.status]}>{n.status}</Badge>
                      </div>
                      <div className="flex gap-1">
                        {n.status !== 'done' && (
                          <Button variant="success" onClick={() => void act(() => api.completeNode(token, detail.id, n.id))}>
                            Выполнить
                          </Button>
                        )}
                      </div>
                    </div>
                    {n.description && <div className="mt-1 text-sm text-slate-500">{n.description}</div>}
                    <div className="mt-1 text-xs text-slate-400">
                      +{n.reward_xp} XP · +{n.reward_gold} 💰
                      {n.duration_hours > 0 && ` · ${n.duration_hours}ч`}
                      {depNames(n).length > 0 && (
                        <span className="text-indigo-500"> · пререквизиты: {depNames(n).join(', ')}</span>
                      )}
                      {n.completed_at && <span> · выполнено {fmtTime(n.completed_at)}</span>}
                    </div>
                  </div>
                ))}
                {detail.nodes.length === 0 && <div className="text-sm text-slate-400">Узлов нет</div>}
              </div>
            </Card>
          </>
        ) : (
          <div className="rounded-xl border border-dashed border-slate-300 bg-white p-8 text-center text-slate-400">
            Выберите роадмапу
          </div>
        )}
        <ErrorText error={error} />
      </div>
    </div>
  )
}
