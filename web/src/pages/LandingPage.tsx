import Button from '../components/ui/Button'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import ThemeToggle from '../components/layout/ThemeToggle'

const FEATURES = [
  {
    icon: '🌿',
    title: 'Ветки и квесты',
    text: 'Делите задачи на ветки-направления и превращайте их в квесты с наградой за выполнение.',
  },
  {
    icon: '⏱️',
    title: 'Таймеры',
    text: 'Timed-квесты засекают время: награда начисляется пропорционально потраченным часам.',
  },
  {
    icon: '🗺️',
    title: 'Роадмапы-графы',
    text: 'Стройте путь из узлов с пререквизитами — полноценный направленный граф прогресса.',
  },
  {
    icon: '🛒',
    title: 'Магазин',
    text: 'Тратьте золото на товары других игроков или продавайте свои за XP и славу.',
  },
  {
    icon: '🧰',
    title: 'Воркшоп',
    text: 'Публикуйте свои роадмапы и устанавливайте чужие — обменивайтесь планами развития.',
  },
  {
    icon: '📊',
    title: 'Статистика',
    text: 'XP, золото, выполненные квесты и часы — сводки по веткам, роадмапам и дням.',
  },
]

const STEPS = [
  { n: '01', title: 'Создайте ветку', text: 'Например «Карьера» или «Спорт».' },
  { n: '02', title: 'Добавляйте квесты', text: 'Назначайте награду в XP и золоте.' },
  { n: '03', title: 'Выполняйте и прокачивайтесь', text: 'Уровень растёт вместе с опытом.' },
  { n: '04', title: 'Тратьте золото в магазине', text: 'И делитесь роадмапами в воркшопе.' },
]

function RoadmapPreview() {
  return (
    <svg
      viewBox="0 0 520 240"
      role="img"
      aria-label="Схема роадмапы: четыре узла, соединённые стрелками"
      className="mx-auto h-auto w-full max-w-xl"
    >
      <defs>
        <marker id="p-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#6366f1" />
        </marker>
      </defs>
      <g stroke="#6366f1" strokeWidth="2" fill="none">
        <path d="M 190 70 C 240 70, 260 70, 310 70" markerEnd="url(#p-arrow)" />
        <path d="M 190 150 C 240 150, 260 150, 310 150" markerEnd="url(#p-arrow)" />
        <path d="M 310 110 C 360 110, 380 110, 430 110" markerEnd="url(#p-arrow)" />
      </g>
      {[
        { x: 60, y: 40, fill: '#818cf8', label: 'Основы' },
        { x: 60, y: 120, fill: '#a78bfa', label: 'Практика' },
        { x: 330, y: 80, fill: '#c084fc', label: 'Проект' },
        { x: 450, y: 80, fill: '#f0abfc', label: 'MVP' },
      ].map((n, i) => (
        <g key={i}>
          <rect x={n.x} y={n.y} width="120" height="60" rx="14" fill={n.fill} opacity="0.9" />
          <text x={n.x + 60} y={n.y + 36} textAnchor="middle" fill="white" fontSize="13" fontWeight="700">
            {n.label}
          </text>
        </g>
      ))}
    </svg>
  )
}

export default function LandingPage({ onEnter }: { onEnter: () => void }) {
  return (
    <div className="bg-aurora min-h-screen">
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-indigo-600 focus:px-3 focus:py-2 focus:text-white"
      >
        Перейти к содержимому
      </a>

      <header className="glass sticky top-0 z-20 border-x-0 border-t-0">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
          <a href="#" className="flex items-center gap-2 text-xl font-black tracking-tight text-slate-800 dark:text-white">
            <span aria-hidden="true" className="text-2xl">🎮</span>
            Level<span className="text-gradient">Up</span>
          </a>
          <nav aria-label="Основная навигация" className="hidden items-center gap-6 text-sm font-medium text-slate-600 md:flex dark:text-slate-300">
            <a href="#features" className="hover:text-indigo-600 dark:hover:text-indigo-300">Возможности</a>
            <a href="#how" className="hover:text-indigo-600 dark:hover:text-indigo-300">Как это работает</a>
          </nav>
          <div className="flex items-center gap-3">
            <ThemeToggle />
            <Button onClick={onEnter}>Открыть приложение</Button>
          </div>
        </div>
      </header>

      <main id="main">
        <section className="mx-auto max-w-6xl px-4 pb-16 pt-16 text-center md:pt-24">
          <Badge color="indigo" className="mb-5">RPG-трекер задач</Badge>
          <h1 className="mx-auto max-w-3xl text-4xl font-black leading-tight tracking-tight md:text-6xl">
            Превратите задачи в <span className="text-gradient">квесты</span>, а прогресс — в <span className="text-gradient">игру</span>
          </h1>
          <p className="mx-auto mt-5 max-w-xl text-lg text-slate-600 dark:text-slate-300">
            Ветки, квесты с наградами, роадмапы-графы, магазин и воркшоп. Прокачивайте уровень за реальные достижения.
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <Button onClick={onEnter} className="px-6 py-3 text-base">Начать бесплатно</Button>
            <Button variant="secondary" onClick={onEnter} className="px-6 py-3 text-base">
              Личный кабинет
            </Button>
          </div>
          <div className="mt-14 animate-float">
            <div className="mx-auto max-w-xl rounded-3xl border border-white/50 bg-white/60 p-6 shadow-glow backdrop-blur dark:border-white/10 dark:bg-slate-900/60">
              <RoadmapPreview />
            </div>
          </div>
        </section>

        <section id="features" aria-labelledby="features-h" className="mx-auto max-w-6xl scroll-mt-20 px-4 py-16">
          <h2 id="features-h" className="text-center text-3xl font-black tracking-tight md:text-4xl">
            Всё для прокачки
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-slate-600 dark:text-slate-300">
            Механики RPG поверх обычного todo: награды, зависимости и экономика.
          </p>
          <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {FEATURES.map((f) => (
              <Card key={f.title} className="transition hover:-translate-y-1">
                <span aria-hidden="true" className="text-3xl">{f.icon}</span>
                <h3 className="mt-3 text-lg font-bold text-slate-800 dark:text-slate-100">{f.title}</h3>
                <p className="mt-1.5 text-sm leading-relaxed text-slate-600 dark:text-slate-300">{f.text}</p>
              </Card>
            ))}
          </div>
        </section>

        <section id="how" aria-labelledby="how-h" className="mx-auto max-w-6xl scroll-mt-20 px-4 py-16">
          <h2 id="how-h" className="text-center text-3xl font-black tracking-tight md:text-4xl">
            Как это работает
          </h2>
          <ol className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {STEPS.map((s) => (
              <li key={s.n} className="glass rounded-2xl p-5">
                <span className="text-gradient text-3xl font-black">{s.n}</span>
                <h3 className="mt-2 font-bold text-slate-800 dark:text-slate-100">{s.title}</h3>
                <p className="mt-1 text-sm text-slate-600 dark:text-slate-300">{s.text}</p>
              </li>
            ))}
          </ol>
        </section>
      </main>

      <footer className="border-t border-slate-200 py-8 dark:border-slate-800">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-4 text-sm text-slate-500 dark:text-slate-400">
          <span>🎮 LevelUp — RPG todo на Go + React</span>
          <div className="flex gap-4">
            <a href="/api/v1" className="hover:text-indigo-600 dark:hover:text-indigo-300">API</a>
            <a href="/swagger/index.html" className="hover:text-indigo-600 dark:hover:text-indigo-300">Swagger</a>
            <a href="/metrics" className="hover:text-indigo-600 dark:hover:text-indigo-300">Metrics</a>
          </div>
        </div>
      </footer>
    </div>
  )
}