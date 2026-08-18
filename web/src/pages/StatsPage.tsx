import { useState } from 'react'
import { api } from '../api'
import type { BranchStat, OverviewStats, QuestStat, RoadmapStat } from '../types'
import { Badge, Button, Card, ErrorText, Spinner, msg } from '../ui'

type Loaded = {
  overview?: OverviewStats
  branches?: BranchStat[]
  roadmaps?: RoadmapStat[]
  quests?: QuestStat[]
}

export default function StatsPage({ token }: { token: string }) {
  const [data, setData] = useState<Loaded>({})
  const [loading, setLoading] = useState<keyof Loaded | null>(null)
  const [error, setError] = useState<string | null>(null)

  async function load(k: keyof Loaded) {
    setLoading(k)
    setError(null)
    try {
      const fn = {
        overview: () => api.statsOverview(token),
        branches: () => api.statsBranches(token),
        roadmaps: () => api.statsRoadmaps(token),
        quests: () => api.statsQuests(token),
      }[k]
      const res = await fn()
      setData((prev) => ({ ...prev, [k]: res }))
    } catch (e) {
      setError(msg(e))
    } finally {
      setLoading(null)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {(['overview', 'branches', 'roadmaps', 'quests'] as (keyof Loaded)[]).map((k) => (
          <Button key={k} onClick={() => void load(k)} disabled={loading !== null}>
            {loading === k ? '…' : k}
          </Button>
        ))}
      </div>
      <ErrorText error={error} />

      {loading === 'overview' && !data.overview && <Spinner />}
      {data.overview && (
        <Card title="Overview (ClickHouse)">
          <div className="flex flex-wrap gap-3">
            <Badge color="indigo">Lv {data.overview.level}</Badge>
            <Badge color="amber">XP {data.overview.xp}</Badge>
            <Badge color="emerald">💰 {data.overview.gold}</Badge>
            <Badge color="slate">⏱ {data.overview.hours}ч</Badge>
          </div>
        </Card>
      )}

      {data.branches && (
        <Card title="По веткам">
          <StatTable
            headers={['branch', 'completed', 'xp', 'gold', 'hours']}
            rows={data.branches.map((b) => [String(b.branch_id), String(b.completed), String(b.xp), String(b.gold), `${b.hours}ч`])}
          />
        </Card>
      )}

      {data.roadmaps && (
        <Card title="По роадмапам">
          <StatTable
            headers={['roadmap', 'completed', 'xp', 'gold', 'hours']}
            rows={data.roadmaps.map((r) => [String(r.roadmap_id), String(r.completed), String(r.xp), String(r.gold), `${r.hours}ч`])}
          />
        </Card>
      )}

      {data.quests && (
        <Card title="По дням">
          <StatTable
            headers={['date', 'completed', 'xp', 'gold', 'hours']}
            rows={data.quests.map((q) => [q.date, String(q.completed), String(q.xp), String(q.gold), `${q.hours}ч`])}
          />
        </Card>
      )}
    </div>
  )
}

function StatTable({ headers, rows }: { headers: string[]; rows: string[][] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-200 text-left text-xs text-slate-400">
            {headers.map((h) => (
              <th key={h} className="px-3 py-2 font-medium">
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => (
            <tr key={i} className="border-b border-slate-100 last:border-0">
              {r.map((c, j) => (
                <td key={j} className="px-3 py-2 text-slate-700">
                  {c}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      {rows.length === 0 && <div className="py-3 text-center text-sm text-slate-400">Данных нет</div>}
    </div>
  )
}
