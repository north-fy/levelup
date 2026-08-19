import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import type { Roadmap, WorkshopRoadmap } from '../types'
import Button from '../components/ui/Button'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import FormField from '../components/ui/FormField'
import { Input } from '../components/ui/Input'
import Spinner from '../components/ui/Spinner'
import { useToast } from '../toast'

export default function WorkshopPage({ token, userId }: { token: string; userId: number }) {
  const [roadmaps, setRoadmaps] = useState<Roadmap[]>([])
  const [workshops, setWorkshops] = useState<WorkshopRoadmap[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { notify } = useToast()

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
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }, [token])

  useEffect(() => {
    void load()
  }, [load])

  async function act(label: string, fn: () => Promise<unknown>) {
    setError(null)
    try {
      const result = await fn()
      notify(label, 'success')
      await load()
      return result
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
      return null
    }
  }

  async function publish(roadmapId: number) {
    await act('Опубликовано в воркшоп 🧰', () =>
      api.createWorkshop(token, {
        roadmap_id: roadmapId,
        title: wTitle || undefined,
        description: wDesc || undefined,
      }),
    )
    setWTitle('')
    setWDesc('')
  }

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <div className="space-y-4">
        <Card title="Опубликовать роадмапу">
          <div className="space-y-3">
            <FormField label="Заголовок публикации (опционально)">
              <Input value={wTitle} onChange={(e) => setWTitle(e.target.value)} placeholder="Как стать Go-разработчиком" />
            </FormField>
            <FormField label="Описание (опционально)">
              <Input value={wDesc} onChange={(e) => setWDesc(e.target.value)} placeholder="Курс из 5 шагов" />
            </FormField>
          </div>
          <div className="mt-3 space-y-1">
            {roadmaps.length === 0 ? (
              <p className="text-sm text-slate-500 dark:text-slate-400">Сначала создайте роадмапу</p>
            ) : (
              roadmaps.map((r) => (
                <div
                  key={r.id}
                  className="flex items-center justify-between gap-2 rounded-xl border border-slate-200 bg-white/60 px-3 py-2 text-sm dark:border-slate-800 dark:bg-slate-900/50"
                >
                  <span className="text-slate-700 dark:text-slate-200">{r.title}</span>
                  <Button variant="secondary" size="sm" onClick={() => void publish(r.id)}>
                    Опубликовать
                  </Button>
                </div>
              ))
            )}
          </div>
        </Card>

        <Card title={`Мои публикации (${mine.length})`}>
          {mine.length === 0 ? (
            <p className="text-sm text-slate-500 dark:text-slate-400">Пока ничего не опубликовано</p>
          ) : (
            <ul className="space-y-1">
              {mine.map((w) => (
                <li
                  key={w.id}
                  className="flex items-center justify-between gap-2 rounded-xl border border-slate-200 bg-white/60 px-3 py-2 text-sm dark:border-slate-800 dark:bg-slate-900/50"
                >
                  <span className="flex min-w-0 items-center gap-2">
                    <span className="truncate text-slate-700 dark:text-slate-200">{w.title}</span>
                    <Badge color={w.is_published ? 'emerald' : 'slate'}>{w.is_published ? 'опубликована' : 'черновик'}</Badge>
                  </span>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() =>
                      void act(w.is_published ? 'Снято с публикации' : 'Опубликовано', () =>
                        api.updateWorkshop(token, w.id, { is_published: !w.is_published }),
                      )
                    }
                  >
                    {w.is_published ? 'Снять' : 'Опубликовать'}
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </Card>
      </div>

      <Card title={`Воркшоп (${workshops.length})`}>
        {loading ? (
          <Spinner label="Загрузка воркшопа" />
        ) : workshops.length === 0 ? (
          <p className="text-sm text-slate-500 dark:text-slate-400">Воркшоп пуст</p>
        ) : (
          <ul className="space-y-2">
            {workshops.map((w) => (
              <li
                key={w.id}
                className="rounded-2xl border border-slate-200 bg-white/70 p-3 dark:border-slate-800 dark:bg-slate-900/50"
              >
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-bold text-slate-800 dark:text-slate-100">{w.title}</span>
                      <Badge color={w.is_published ? 'emerald' : 'slate'}>{w.is_published ? 'опубликована' : 'черновик'}</Badge>
                      <Badge color="indigo">автор #{w.author_id}</Badge>
                    </div>
                    {w.description && <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{w.description}</p>}
                    <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">source roadmap #{w.source_roadmap_id}</p>
                  </div>
                  {w.is_published && (
                    <Button
                      size="sm"
                      onClick={() =>
                        void act('Роадмапа установлена в ваш аккаунт 🎉', () => api.installWorkshop(token, w.id))
                      }
                    >
                      Установить
                    </Button>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
        {error && (
          <p role="alert" className="mt-3 rounded-xl border border-rose-500/30 bg-rose-50 px-3 py-2 text-sm text-rose-700 dark:bg-rose-950/50 dark:text-rose-300">
            {error}
          </p>
        )}
      </Card>
    </div>
  )
}