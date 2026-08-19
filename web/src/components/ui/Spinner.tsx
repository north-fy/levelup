export default function Spinner({ label = 'Загрузка' }: { label?: string }) {
  return (
    <div role="status" aria-live="polite" className="flex flex-col items-center gap-2 py-10">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-indigo-300 border-t-indigo-600 dark:border-indigo-700 dark:border-t-indigo-400" />
      <span className="text-xs text-slate-400">{label}</span>
    </div>
  )
}