import { useState } from 'react'
import { api } from '../api'
import type { BranchStat, OverviewStats, QuestStat, RoadmapStat } from '../types'
import Button from '../components/ui/Button'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import Spinner from '../components/ui/Spinner'
import { useToast } from '../toast'

type Loaded = {
  overview?: OverviewStats
  branches?: BranchStat[]
  roadmaps?: RoadmapStat[]
  quests?: QuestStat[]
}

type LoadKey = keyof Loaded

const LOADERS: Record<LoadKey, { label: string; run: (token: string) => Promise<unknown> }> = {
  overview: { label: 'Обзор', run: (t) => api.statsOverview(t) },
  branches: { label: 'По веткам', run: (t) => api.statsBranches(t) },
  roadmaps: { label: 'По роадмапам', run: (t) => api.statsRoadmaps(t) },
  quests: { label: 'По дням', run: (t) => api.statsQuests(t) },
}

function fmtHours(h: number): string {
  return `${h}ч`
}

export default function StatsPage({ token }: { token: string }) {
  const [data, setData] = useState<Loaded>({})
  const [loading, setLoading] = useState<LoadKey | null>(null)
  const [error, setError] = useState<string | null>(null)
  const { notify } = useToast()

  async function load(k: LoadKey) {
    setLoading(k)
    setError(null)
    try {
      const res = await LOADERS[k].run(token)
      setData((prev) => ({ ...prev, [k]: res }))
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    } finally {
      setLoading(null)
    }
  }

  const { overview } = data

  return (
    <div className="space-y-4">
      <div role="group" aria-label="Какую статистику показать" className="flex flex-wrap gap-2">
        {(Object.keys(LOADERS) as LoadKey[]).map((k) => (
          <Button key={k} variant={data[k] ? 'primary' : 'secondary'} onClick={() => void load(k)} disabled={loading !== null} aria-pressed={!!data[k]}>
            {loading === k ? 'Загружаю…' : LOADERS[k].label}
          </Button>
        ))}
      </div>
      {error && (
        <p role="alert" className="rounded-xl border border-rose-500/30 bg-rose-50 px-3 py-2 text-sm text-rose-700 dark:bg-rose-950/50 dark:text-rose-300">
          {error}
        </p>
      )}

      {overview && (
        <Card title="Обзор (ClickHouse)">
          <div className="flex flex-wrap gap-3">
            <Badge color="indigo">Уровень {overview.level}</Badge>
            <Badge color="amber">XP {overview.xp}</Badge>
            <Badge color="emerald">💰 {overview.gold}</Badge>
            <Badge color="slate">⏱ {fmtHours(overview.hours)}</Badge>
          </div>
        </Card>
      )}

      {data.branches && <StatCard title="По веткам" headers={['branch', 'completed', 'xp', 'gold', 'hours']} rows={data.branches.map((b) => [String(b.branch_id), String(b.completed), String(b.xp), String(b.gold), fmtHours(b.hours)])} />}
      {data.roadmaps && <StatCard title="По роадмапам" headers={['roadmap', 'completed', 'xp', 'gold', 'hours']} rows={data.roadmaps.map((r) => [String(r.roadmap_id), String(r.completed), String(r.xp), String(r.gold), fmtHours(r.hours)])} />}
      {data.quests && <StatCard title="По дням" headers={['date', 'completed', 'xp', 'gold', 'hours']} rows={data.quests.map((q) => [q.date, String(q.completed), String(q.xp), String(q.gold), fmtHours(q.hours)])} />}

      {loading && <Spinner label="Загрузка из ClickHouse" />}
    </div>
  )
}

function StatCard({ title, headers, rows }: { title: string; headers: string[]; rows: string[][] }) {
  return (
    <Card title={title}>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <caption className="sr-only">{title}</caption>
          <thead>
            <tr className="border-b border-slate-200 text-left text-xs text-slate-400 dark:border-slate-700">
              {headers.map((h) => (
                <th key={h} scope="col" className="px-3 py-2 font-medium">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={i} className="border-b border-slate-100 last:border-0 dark:border-slate-800">
                {r.map((c, j) => (
                  <td key={j} className="px-3 py-2 text-slate-700 dark:text-slate-300">
                    {c}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        {rows.length === 0 && <p className="py-3 text-center text-sm text-slate-400">Данных нет</p>}
      </div>
    </Card>
  )
}