import { useMemo } from 'react'
import type { RoadmapDetail, RoadmapNode } from '../../types'

const NODE_W = 228
const NODE_H = 100
const H_GAP = 150
const V_GAP = 24
const PAD = 32

interface Pos {
  x: number
  y: number
}

interface Layout {
  pos: Map<number, Pos>
  width: number
  height: number
}

// Layered DAG layout: a node sits one layer after its deepest prerequisite.
function computeLayout(detail: RoadmapDetail): Layout {
  const depsOf = new Map<number, number[]>()
  for (const e of detail.edges) {
    const arr = depsOf.get(e.to_node_id) ?? []
    arr.push(e.from_node_id)
    depsOf.set(e.to_node_id, arr)
  }

  const layer = new Map<number, number>()
  const layerOf = (id: number): number => {
    const cached = layer.get(id)
    if (cached !== undefined) return cached
    const deps = depsOf.get(id) ?? []
    const l = deps.length === 0 ? 0 : 1 + Math.max(...deps.map(layerOf))
    layer.set(id, l)
    return l
  }

  const ordered = [...detail.nodes].sort((a, b) => a.position - b.position || a.id - b.id)
  for (const n of ordered) layerOf(n.id)

  const byLayer = new Map<number, RoadmapNode[]>()
  for (const n of ordered) {
    const l = layer.get(n.id) ?? 0
    const arr = byLayer.get(l) ?? []
    arr.push(n)
    byLayer.set(l, arr)
  }

  const maxLayer = byLayer.size
  const maxCols = Math.max(1, ...[...byLayer.values()].map((a) => a.length))
  const width = PAD * 2 + maxLayer * NODE_W + Math.max(0, maxLayer - 1) * H_GAP
  const height = PAD * 2 + maxCols * NODE_H + Math.max(0, maxCols - 1) * V_GAP

  const pos = new Map<number, Pos>()
  for (const [l, nodes] of byLayer) {
    const x = PAD + l * (NODE_W + H_GAP)
    const layerH = nodes.length * NODE_H + Math.max(0, nodes.length - 1) * V_GAP
    const y0 = PAD + (height - PAD * 2 - layerH) / 2
    nodes.forEach((n, i) => pos.set(n.id, { x, y: y0 + i * (NODE_H + V_GAP) }))
  }

  return { pos, width, height }
}

function statusColor(node: RoadmapNode): string {
  if (node.status === 'done') return 'border-emerald-500/50 shadow-emerald-500/20'
  if (node.status === 'in_progress') return 'border-amber-500/50 shadow-amber-500/20'
  return 'border-indigo-500/40 shadow-indigo-500/20'
}

function statusIcon(node: RoadmapNode): string {
  if (node.status === 'done') return '✅'
  if (node.status === 'in_progress') return '⏳'
  return '📌'
}

export default function RoadmapGraph({
  detail,
  onComplete,
}: {
  detail: RoadmapDetail
  onComplete: (node: RoadmapNode) => void
}) {
  const { pos, width, height } = useMemo(() => computeLayout(detail), [detail])
  const byId = useMemo(() => new Map(detail.nodes.map((n) => [n.id, n])), [detail])

  const isLocked = (n: RoadmapNode) =>
    detail.edges.some((e) => e.to_node_id === n.id && byId.get(e.from_node_id)?.status !== 'done')

  // Reading order: top-to-bottom, then left-to-right (source code order = tab order).
  const ordered = [...detail.nodes].sort((a, b) => {
    const pa = pos.get(a.id)!
    const pb = pos.get(b.id)!
    return pa.y - pb.y || pa.x - pb.x
  })

  return (
    <div>
      <div
        role="region"
        aria-label="Граф роадмапы: узлы и зависимости"
        className="overflow-x-auto rounded-2xl border border-slate-200 bg-white/50 p-2 dark:border-slate-800 dark:bg-slate-900/40"
      >
        <div className="relative" style={{ width, height }}>
          <svg className="absolute inset-0" width={width} height={height} aria-hidden="true">
            <defs>
              <marker
                id="arrowhead"
                viewBox="0 0 10 10"
                refX="9"
                refY="5"
                markerWidth="7"
                markerHeight="7"
                orient="auto-start-reverse"
              >
                <path d="M 0 0 L 10 5 L 0 10 z" className="fill-slate-300 dark:fill-slate-600" />
              </marker>
            </defs>
            {detail.edges.map((e) => {
              const from = pos.get(e.from_node_id)
              const to = pos.get(e.to_node_id)
              if (!from || !to) return null
              const src = byId.get(e.from_node_id)
              const done = src?.status === 'done'
              const sx = from.x + NODE_W
              const sy = from.y + NODE_H / 2
              const tx = to.x
              const ty = to.y + NODE_H / 2
              const d = `M ${sx} ${sy} C ${sx + H_GAP / 2} ${sy}, ${tx - H_GAP / 2} ${ty}, ${tx} ${ty}`
              return (
                <path
                  key={`${e.from_node_id}-${e.to_node_id}`}
                  d={d}
                  fill="none"
                  strokeWidth={2}
                  markerEnd="url(#arrowhead)"
                  className={done ? 'stroke-emerald-400' : 'stroke-slate-300 dark:stroke-slate-600'}
                />
              )
            })}
          </svg>

          {ordered.map((n) => {
            const p = pos.get(n.id)!
            const locked = isLocked(n)
            const done = n.status === 'done'
            return (
              <div key={n.id} className="absolute" style={{ left: p.x, top: p.y, width: NODE_W }}>
                <button
                  type="button"
                  onClick={() => onComplete(n)}
                  aria-pressed={done}
                  aria-label={`Узел ${n.position} «${n.title}», статус: ${n.status}${locked ? ', заблокирован пререквизитами' : ''}`}
                  className={`flex w-full flex-col rounded-2xl border bg-white p-3 text-left shadow-lg transition hover:scale-[1.02] hover:border-indigo-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 ${statusColor(n)} ${
                    locked ? 'opacity-60 saturate-50 dark:opacity-50' : ''
                  } dark:bg-slate-900`}
                  style={{ height: NODE_H }}
                >
                  <span className="flex items-start justify-between gap-1">
                    <span className="text-sm font-bold leading-tight text-slate-800 dark:text-slate-100">
                      {n.position}. {n.title}
                    </span>
                    <span aria-hidden="true">{locked ? '🔒' : statusIcon(n)}</span>
                  </span>
                  {n.description && (
                    <span className="mt-0.5 line-clamp-2 text-[11px] leading-snug text-slate-500 dark:text-slate-400">
                      {n.description}
                    </span>
                  )}
                  <span className="mt-auto flex gap-2 text-[11px] font-semibold">
                    <span className="text-amber-600 dark:text-amber-400">+{n.reward_xp} XP</span>
                    <span className="text-emerald-600 dark:text-emerald-400">+{n.reward_gold} 💰</span>
                  </span>
                </button>
              </div>
            )
          })}
        </div>
      </div>

      <ul className="mt-3 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500 dark:text-slate-400" aria-label="Легенда графа">
        <LegendItem color="bg-indigo-400">todo</LegendItem>
        <LegendItem color="bg-amber-400">in_progress</LegendItem>
        <LegendItem color="bg-emerald-500">done</LegendItem>
        <LegendItem color="bg-slate-400/70 dark:bg-slate-600">заблокирован</LegendItem>
        <LegendItem color="bg-transparent border border-dashed border-slate-300 dark:border-slate-600">
          стрелка = пререквизит
        </LegendItem>
      </ul>
    </div>
  )
}

function LegendItem({ color, children }: { color: string; children: string }) {
  return (
    <li className="flex items-center gap-1.5">
      <span aria-hidden="true" className={`inline-block h-2.5 w-2.5 rounded-full ${color}`} />
      {children}
    </li>
  )
}