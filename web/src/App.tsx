import { useEffect, useState } from 'react'
import AuthPage from './pages/AuthPage'
import ProfilePage from './pages/ProfilePage'
import BranchesPage from './pages/BranchesPage'
import ShopPage from './pages/ShopPage'
import RoadmapsPage from './pages/RoadmapsPage'
import WorkshopPage from './pages/WorkshopPage'
import StatsPage from './pages/StatsPage'
import { loadSession, type Session, type SlotId } from './store'

const NAV = [
  { id: 'auth', label: 'Auth' },
  { id: 'profile', label: 'Profile' },
  { id: 'branches', label: 'Branches' },
  { id: 'shop', label: 'Shop' },
  { id: 'roadmaps', label: 'Roadmaps' },
  { id: 'workshop', label: 'Workshop' },
  { id: 'stats', label: 'Stats' },
] as const

type PageId = (typeof NAV)[number]['id']

function NotAuthed() {
  return (
    <div className="rounded-xl border border-dashed border-slate-300 bg-white p-8 text-center text-slate-400">
      Авторизуйтесь во вкладке Auth
    </div>
  )
}

export default function App() {
  const [slot, setSlot] = useState<SlotId>('A')
  const [page, setPage] = useState<PageId>('auth')
  const [session, setSession] = useState<Session | null>(() => loadSession('A'))

  useEffect(() => {
    setSession(loadSession(slot))
  }, [slot])

  const token = session?.access ?? null

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="sticky top-0 z-10 border-b border-slate-200 bg-white/90 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
          <div className="text-lg font-bold text-indigo-600">LevelUp</div>
          <div className="flex items-center gap-2 text-sm">
            {(['A', 'B'] as SlotId[]).map((s) => (
              <button
                key={s}
                onClick={() => setSlot(s)}
                className={`rounded-full px-3 py-1 font-medium ${
                  slot === s ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-500 hover:bg-slate-200'
                }`}
              >
                Slot {s}
              </button>
            ))}
            {session?.user && (
              <span className="text-slate-500">
                · {session.user.nickname} · 💰{session.user.gold} · Lv{session.user.level}
              </span>
            )}
          </div>
        </div>
        <nav className="mx-auto flex max-w-6xl gap-1 overflow-x-auto px-4 pb-2">
          {NAV.map((n) => (
            <button
              key={n.id}
              onClick={() => setPage(n.id)}
              className={`whitespace-nowrap rounded-lg px-3 py-1.5 text-sm font-medium ${
                page === n.id ? 'bg-indigo-50 text-indigo-700' : 'text-slate-500 hover:bg-slate-100'
              }`}
            >
              {n.label}
            </button>
          ))}
        </nav>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-6">
        {page === 'auth' && <AuthPage slot={slot} onSession={setSession} />}
        {page === 'profile' && (token ? <ProfilePage slot={slot} token={token} onSession={setSession} /> : <NotAuthed />)}
        {page === 'branches' && (token ? <BranchesPage token={token} /> : <NotAuthed />)}
        {page === 'shop' && (token ? <ShopPage token={token} /> : <NotAuthed />)}
        {page === 'roadmaps' && (token ? <RoadmapsPage token={token} /> : <NotAuthed />)}
        {page === 'workshop' &&
          (token && session?.user ? <WorkshopPage token={token} userId={session.user.id} /> : <NotAuthed />)}
        {page === 'stats' && (token ? <StatsPage token={token} /> : <NotAuthed />)}
      </main>
    </div>
  )
}
