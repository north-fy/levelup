import { useEffect, useState } from 'react'
import { loadSession, type Session, type SlotId } from './store'
import ThemeToggle from './components/layout/ThemeToggle'
import Badge from './components/ui/Badge'
import LandingPage from './pages/LandingPage'
import AuthPage from './pages/AuthPage'
import ProfilePage from './pages/ProfilePage'
import BranchesPage from './pages/BranchesPage'
import ShopPage from './pages/ShopPage'
import RoadmapsPage from './pages/RoadmapsPage'
import WorkshopPage from './pages/WorkshopPage'
import StatsPage from './pages/StatsPage'

const TABS = [
  { id: 'auth', label: 'Аккаунт' },
  { id: 'profile', label: 'Профиль' },
  { id: 'branches', label: 'Ветки' },
  { id: 'shop', label: 'Магазин' },
  { id: 'roadmaps', label: 'Роадмапы' },
  { id: 'workshop', label: 'Воркшоп' },
  { id: 'stats', label: 'Статистика' },
] as const

type TabId = (typeof TABS)[number]['id']

function NotAuthed({ onGoAuth }: { onGoAuth: () => void }) {
  return (
    <div className="rounded-2xl border border-dashed border-slate-300 bg-white/60 p-10 text-center dark:border-slate-700 dark:bg-slate-900/40">
      <p className="text-slate-500 dark:text-slate-400">
        Нет активной сессии для этого слота. Зарегистрируйтесь или войдите.
      </p>
      <button
        type="button"
        onClick={onGoAuth}
        className="mt-4 rounded-xl bg-gradient-to-r from-indigo-600 to-violet-600 px-4 py-2 text-sm font-semibold text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 focus-visible:ring-offset-2"
      >
        Перейти к аккаунту
      </button>
    </div>
  )
}

export default function App() {
  const [view, setView] = useState<'landing' | 'app'>('landing')
  const [slot, setSlot] = useState<SlotId>('A')
  const [session, setSession] = useState<Session | null>(null)
  const [tab, setTab] = useState<TabId>('profile')

  useEffect(() => {
    const s = loadSession(slot)
    setSession(s)
    setTab(s ? 'profile' : 'auth')
  }, [slot])

  if (view === 'landing') {
    return <LandingPage onEnter={() => setView('app')} />
  }

  const token = session?.access ?? null
  const userId = session?.user?.id ?? 0

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950">
      <a
        href="#app-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-indigo-600 focus:px-3 focus:py-2 focus:text-white"
      >
        Перейти к содержимому
      </a>

      <header className="glass sticky top-0 z-20 border-x-0 border-t-0">
        <div className="mx-auto max-w-6xl px-4">
          <div className="flex items-center justify-between py-3">
            <button
              type="button"
              onClick={() => setView('landing')}
              className="flex items-center gap-2 text-xl font-black tracking-tight text-slate-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 focus-visible:ring-offset-2 dark:text-white"
            >
              <span aria-hidden="true" className="text-2xl">
                🎮
              </span>
              Level<span className="text-gradient">Up</span>
            </button>

            <div className="flex items-center gap-3">
              <div role="group" aria-label="Активный слот аккаунта" className="flex rounded-full bg-slate-100 p-1 dark:bg-slate-800">
                {(['A', 'B'] as SlotId[]).map((s) => (
                  <button
                    key={s}
                    type="button"
                    aria-pressed={slot === s}
                    onClick={() => setSlot(s)}
                    className={`rounded-full px-3 py-1 text-xs font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 ${
                      slot === s
                        ? 'bg-gradient-to-r from-indigo-600 to-violet-600 text-white'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200'
                    }`}
                  >
                    Слот {s}
                  </button>
                ))}
              </div>
              <ThemeToggle />
            </div>
          </div>

          <nav aria-label="Разделы приложения" className="flex gap-1 overflow-x-auto pb-2">
            {TABS.map((t) => (
              <button
                key={t.id}
                type="button"
                aria-current={tab === t.id ? 'page' : undefined}
                onClick={() => setTab(t.id)}
                className={`whitespace-nowrap rounded-lg px-3 py-1.5 text-sm font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 ${
                  tab === t.id
                    ? 'bg-indigo-600/10 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300'
                    : 'text-slate-500 hover:bg-slate-100 hover:text-slate-700 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-200'
                }`}
              >
                {t.label}
              </button>
            ))}
          </nav>

          {session?.user && (
            <div className="flex flex-wrap items-center gap-2 pb-3 text-xs">
              <span className="font-semibold text-slate-700 dark:text-slate-200">{session.user.nickname}</span>
              <Badge color="indigo">Lv {session.user.level}</Badge>
              <Badge color="amber">XP {session.user.xp}</Badge>
              <Badge color="emerald">💰 {session.user.gold}</Badge>
            </div>
          )}
        </div>
      </header>

      <main id="app-main" className="mx-auto max-w-6xl px-4 py-6">
        {tab === 'auth' && <AuthPage slot={slot} onSession={setSession} />}
        {tab === 'profile' && (token ? <ProfilePage slot={slot} token={token} onSession={setSession} /> : <NotAuthed onGoAuth={() => setTab('auth')} />)}
        {tab === 'branches' && (token ? <BranchesPage token={token} /> : <NotAuthed onGoAuth={() => setTab('auth')} />)}
        {tab === 'shop' && (token ? <ShopPage token={token} /> : <NotAuthed onGoAuth={() => setTab('auth')} />)}
        {tab === 'roadmaps' && (token ? <RoadmapsPage token={token} /> : <NotAuthed onGoAuth={() => setTab('auth')} />)}
        {tab === 'workshop' && (token && userId ? <WorkshopPage token={token} userId={userId} /> : <NotAuthed onGoAuth={() => setTab('auth')} />)}
        {tab === 'stats' && (token ? <StatsPage token={token} /> : <NotAuthed onGoAuth={() => setTab('auth')} />)}
      </main>

      <footer className="border-t border-slate-200 py-6 text-center text-xs text-slate-400 dark:border-slate-800 dark:text-slate-500">
        🎮 LevelUp — RPG todo на Go + React
      </footer>
    </div>
  )
}