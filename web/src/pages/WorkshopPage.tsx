import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import type { Roadmap, WorkshopRoadmap } from '../types'
import { Badge, Button, Card, ErrorText, Field, Input, msg, useRefresh } from '../ui'

export default function WorkshopPage({ token, userId }: { token: string; userId: number }) {
  const [roadmaps, setRoadmaps] = useState<Roadmap[]>([])
  const [workshops, setWorkshops] = useState<WorkshopRoadmap[]>([])
  const [error, setError] = useState<string | null>(null)
  const [tick, refresh] = useRefresh()

  const [wTitle, setWTitle] = useState('')
  const [wDesc, setWDesc] = useState('')
  const mine = workshops.filter((w) => w.author_id === userId)

  const load = useCallback(async () => {
    setError(null)
    try {
      const [r, w] = await Promise.all([api.listRoadmaps(token), api.listWorkshops(token)])
      setRoadmaps(r)
      setWorkshops(w)
    } catch (e) {
      setError(msg(e))
    }
  }, [token])

  useEffect(() => {
    void load()
  }, [load, tick])

  async function publish(roadmapId: number) {
    setError(null)
    try {
      await api.createWorkshop(token, {
        roadmap_id: roadmapId,
        title: wTitle || undefined,
        description: wDesc || undefined,
      })
      setWTitle('')
      setWDesc('')
      refresh()
    } catch (e) {
      setError(msg(e))
    }
  }

  async function act(fn: () => Promise<unknown>) {
    setError(null)
    try {
      const r = await fn()
      return r
    } catch (e) {
      setError(msg(e))
      return null
    }
  }

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <div className="space-y-4">
        <Card title="Опубликовать роадмапу в воркшоп">
          <div className="space-y-3">
            <Field label="Заголовок (опционально)">
              <Input value={wTitle} onChange={(e) => setWTitle(e.target.value)} placeholder="Как стать Go-разработчиком" />
            </Field>
            <Field label="Описание (опционально)">
              <Input value={wDesc} onChange={(e) => setWDesc(e.target.value)} placeholder="Курс из 5 шагов" />
            </Field>
            <div className="space-y-1">
              {roadmaps.map((r) => (
                <div key={r.id} className="flex items-center justify-between rounded-lg border border-slate-100 px-3 py-2 text-sm">
                  <span>{r.title}</span>
                  <Button variant="ghost" onClick={() => void publish(r.id)}>
                    Опубликовать
                  </Button>
                </div>
              ))}
              {roadmaps.length === 0 && <div className="text-sm text-slate-400">Сначала создайте роадмапу</div>}
            </div>
          </div>
        </Card>

        <Card title={`Мои публикации (${mine.length})`}>
          <div className="space-y-1">
            {mine.map((w) => (
              <div key={w.id} className="flex items-center justify-between gap-2 rounded-lg border border-slate-100 px-3 py-2 text-sm">
                <span>{w.title}</span>
                <div className="flex gap-1">
                  <Badge color={w.is_published ? 'emerald' : 'slate'}>{w.is_published ? 'опубликована' : 'черновик'}</Badge>
                  <Button
                    variant="ghost"
                    onClick={() => void act(() => api.updateWorkshop(token, w.id, { is_published: !w.is_published }))}
                  >
                    {w.is_published ? 'Снять' : 'Опубликовать'}
                  </Button>
                </div>
              </div>
            ))}
            {mine.length === 0 && <div className="text-sm text-slate-400">Публикаций нет</div>}
          </div>
        </Card>
      </div>

      <Card title={`Воркшоп (${workshops.length})`}>
        <div className="space-y-2">
          {workshops.map((w) => (
            <div key={w.id} className="flex items-start justify-between gap-3 rounded-lg border border-slate-100 p-3">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-medium text-slate-800">{w.title}</span>
                  <Badge color={w.is_published ? 'emerald' : 'slate'}>{w.is_published ? 'опубликована' : 'черновик'}</Badge>
                  <Badge color="indigo">автор #{w.author_id}</Badge>
                </div>
                {w.description && <div className="mt-1 text-sm text-slate-500">{w.description}</div>}
                <div className="mt-1 text-xs text-slate-400">source roadmap #{w.source_roadmap_id}</div>
              </div>
              {w.is_published && (
                <Button
                  onClick={() => void act(() => api.installWorkshop(token, w.id))}
                  title="Установит копию роадмапы в ваш аккаунт"
                >
                  Установить
                </Button>
              )}
            </div>
          ))}
          {workshops.length === 0 && <div className="text-sm text-slate-400">Воркшоп пуст</div>}
        </div>
        <ErrorText error={error} />
      </Card>
    </div>
  )
}
