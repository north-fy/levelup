import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api } from '../api'
import type { QuestType, Roadmap, RoadmapDetail, RoadmapNode } from '../types'
import Button from '../components/ui/Button'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import FormField from '../components/ui/FormField'
import { Input, Select, TextArea } from '../components/ui/Input'
import Spinner from '../components/ui/Spinner'
import RoadmapGraph from '../components/graph/RoadmapGraph'
import { useToast } from '../toast'

const TYPES: QuestType[] = ['simple', 'timed']

export default function RoadmapsPage({ token }: { token: string }) {
  const [roadmaps, setRoadmaps] = useState<Roadmap[] | null>(null)
  const [detail, setDetail] = useState<RoadmapDetail | null>(null)
  const [error, setError] = useState<string | null>(null)
  const { notify } = useToast()

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
      setDetail((prev) => (prev && list.some((r) => r.id === prev.id) ? prev : null))
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }, [token])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    if (!detail) return
    let cancelled = false
    api
      .getRoadmap(token, detail.id)
      .then((d) => {
        if (!cancelled) setDetail(d)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e))
      })
    return () => {
      cancelled = true
    }
  }, [token, detail?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  async function open(r: Roadmap) {
    setError(null)
    try {
      setDetail(await api.getRoadmap(token, r.id))
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  async function createRoadmap(e: FormEvent) {
    e.preventDefault()
    if (!rTitle.trim()) return
    setError(null)
    try {
      const r = await api.createRoadmap(token, { title: rTitle.trim(), description: rDesc })
      setRTitle('')
      setRDesc('')
      notify('Роадмапа создана 🗺️', 'success')
      await load()
      await open(r)
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  async function deleteRoadmap(r: Roadmap) {
    setError(null)
    try {
      await api.deleteRoadmap(token, r.id)
      notify('Роадмапа удалена', 'info')
      setDetail(null)
      await load()
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  async function addNode(e: FormEvent) {
    e.preventDefault()
    if (!detail || !nTitle.trim()) return
    setError(null)
    try {
      await api.addNode(token, detail.id, {
        title: nTitle.trim(),
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
      notify('Узел добавлен в граф ➕', 'success')
      setDetail(await api.getRoadmap(token, detail.id))
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  async function completeNode(node: RoadmapNode) {
    if (!detail) return
    if (node.status === 'done') {
      notify('Узел уже выполнен', 'info')
      return
    }
    const depsDone = !detail.edges.some((e) => e.to_node_id === node.id && detail.nodes.find((n) => n.id === e.from_node_id)?.status !== 'done')
    if (!depsDone) {
      notify('Заблокировано: не выполнены пререквизиты 🔒', 'error')
      return
    }
    setError(null)
    try {
      await api.completeNode(token, detail.id, node.id)
      notify('Узел выполнен 🎉', 'success')
      setDetail(await api.getRoadmap(token, detail.id))
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  const doneCount = detail?.nodes.filter((n) => n.status === 'done').length ?? 0

  return (
    <div className="grid gap-4 lg:grid-cols-[320px_1fr]">
      <div className="space-y-4">
        <Card title="Новая роадмапа">
          <form onSubmit={(e) => void createRoadmap(e)} className="space-y-3" noValidate>
            <FormField label="Название">
              <Input value={rTitle} onChange={(e) => setRTitle(e.target.value)} placeholder="Путь Go-разработчика" />
            </FormField>
            <FormField label="Описание">
              <TextArea value={rDesc} onChange={(e) => setRDesc(e.target.value)} rows={2} />
            </FormField>
            <Button type="submit" disabled={!rTitle.trim()}>
              Создать роадмапу
            </Button>
          </form>
        </Card>

        <Card title={`Роадмапы${roadmaps ? ` (${roadmaps.length})` : ''}`}>
          {!roadmaps ? (
            <Spinner label="Загрузка" />
          ) : roadmaps.length === 0 ? (
            <p className="text-sm text-slate-500 dark:text-slate-400">Роадмап нет — создайте первую</p>
          ) : (
            <ul className="max-h-96 space-y-1 overflow-y-auto">
              {roadmaps.map((r) => (
                <li key={r.id}>
                  <div
                    className={`flex w-full items-center justify-between rounded-xl px-3 py-2 text-left text-sm ${
                      detail?.id === r.id ? 'bg-indigo-600/10 font-semibold text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300' : ''
                    }`}
                  >
                    <button
                      type="button"
                      aria-pressed={detail?.id === r.id}
                      onClick={() => void open(r)}
                      className="min-w-0 flex-1 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500"
                    >
                      {r.title}
                    </button>
                    <button
                      type="button"
                      aria-label={`Удалить роадмапу ${r.title}`}
                      onClick={() => void deleteRoadmap(r)}
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
        {detail ? (
          <>
            <Card
              title={detail.title}
              actions={
                <>
                  {detail.source_type && detail.source_id ? (
                    <Badge color="fuchsia">
                      {detail.source_type} #{detail.source_id}
                    </Badge>
                  ) : (
                    <Badge color="indigo">id {detail.id}</Badge>
                  )}
                  <Badge color={doneCount === detail.nodes.length && detail.nodes.length > 0 ? 'emerald' : 'amber'}>
                    {doneCount}/{detail.nodes.length} узлов
                  </Badge>
                </>
              }
            >
              <p className="text-sm text-slate-600 dark:text-slate-300">{detail.description || 'Без описания'}</p>
            </Card>

            <RoadmapGraph detail={detail} onComplete={(n) => void completeNode(n)} />

            <Card title="Новый узел">
              <form onSubmit={(e) => void addNode(e)} className="grid gap-3 md:grid-cols-2" noValidate>
                <FormField label="Название">
                  <Input value={nTitle} onChange={(e) => setNTitle(e.target.value)} placeholder="Изучить синтаксис" />
                </FormField>
                <FormField label="Тип">
                  <Select value={nType} onChange={(e) => setNType(e.target.value as QuestType)}>
                    {TYPES.map((t) => (
                      <option key={t} value={t}>
                        {t === 'simple' ? 'simple' : 'timed'}
                      </option>
                    ))}
                  </Select>
                </FormField>
                <div className="md:col-span-2">
                  <FormField label="Описание">
                    <TextArea value={nDesc} onChange={(e) => setNDesc(e.target.value)} rows={2} />
                  </FormField>
                </div>
                <FormField label="Reward XP">
                  <Input type="number" min={0} value={nXp} onChange={(e) => setNXp(e.target.value)} />
                </FormField>
                <FormField label="Reward Gold">
                  <Input type="number" min={0} value={nGold} onChange={(e) => setNGold(e.target.value)} />
                </FormField>
                {nType === 'timed' && (
                  <FormField label="Длительность (часы)">
                    <Input type="number" min={1} value={nHours} onChange={(e) => setNHours(e.target.value)} />
                  </FormField>
                )}
                <fieldset className="md:col-span-2">
                  <legend className="mb-1 text-xs font-semibold text-slate-500 dark:text-slate-400">
                    Зависимости (пререквизиты)
                  </legend>
                  {detail.nodes.length === 0 ? (
                    <p className="text-xs text-slate-400">Пока нет узлов — этот будет первым</p>
                  ) : (
                    <div className="grid max-h-40 gap-1 overflow-y-auto rounded-xl border border-slate-200 bg-white p-2 dark:border-slate-700 dark:bg-slate-900">
                      {detail.nodes.map((n) => (
                        <label key={n.id} className="flex items-center gap-2 rounded px-2 py-1 text-sm hover:bg-slate-50 dark:hover:bg-slate-800">
                          <input
                            type="checkbox"
                            checked={nDeps.includes(n.id)}
                            onChange={() =>
                              setNDeps((d) => (d.includes(n.id) ? d.filter((x) => x !== n.id) : [...d, n.id]))
                            }
                            className="h-4 w-4 rounded border-slate-300 text-indigo-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500"
                          />
                          <span className="text-slate-700 dark:text-slate-200">
                            {n.position}. {n.title}
                          </span>
                        </label>
                      ))}
                    </div>
                  )}
                </fieldset>
                <div className="md:col-span-2">
                  <Button type="submit" disabled={!nTitle.trim()}>
                    Добавить узел
                  </Button>
                </div>
              </form>
            </Card>
          </>
        ) : (
          <div className="rounded-2xl border border-dashed border-slate-300 bg-white/60 p-10 text-center text-slate-500 dark:border-slate-700 dark:bg-slate-900/40 dark:text-slate-400">
            Выберите роадмапу слева, чтобы увидеть её граф
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