import { createContext, useCallback, useContext, useRef, useState, type ReactNode } from 'react'

type Kind = 'success' | 'error' | 'info'

interface Toast {
  id: number
  message: string
  kind: Kind
}

interface ToastCtxValue {
  notify: (message: string, kind?: Kind) => void
}

const ToastCtx = createContext<ToastCtxValue>({ notify: () => {} })

const kindStyles: Record<Kind, string> = {
  success: 'border-emerald-500/40 bg-emerald-50 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-200',
  error: 'border-rose-500/40 bg-rose-50 text-rose-800 dark:bg-rose-950/80 dark:text-rose-200',
  info: 'border-indigo-500/40 bg-indigo-50 text-indigo-800 dark:bg-indigo-950/80 dark:text-indigo-200',
}

const kindIcons: Record<Kind, string> = {
  success: '✅',
  error: '⚠️',
  info: 'ℹ️',
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])
  const idRef = useRef(0)

  const notify = useCallback((message: string, kind: Kind = 'info') => {
    const id = ++idRef.current
    setToasts((ts) => [...ts, { id, message, kind }])
    window.setTimeout(() => {
      setToasts((ts) => ts.filter((t) => t.id !== id))
    }, 4500)
  }, [])

  return (
    <ToastCtx.Provider value={{ notify }}>
      {children}
      <div
        aria-live="polite"
        aria-atomic="false"
        className="pointer-events-none fixed bottom-4 right-4 z-50 flex w-full max-w-sm flex-col gap-2 px-4 sm:px-0"
      >
        {toasts.map((t) => (
          <div
            key={t.id}
            role="status"
            className={`pointer-events-auto flex items-start gap-2 rounded-xl border px-3 py-2.5 text-sm shadow-lg ${kindStyles[t.kind]}`}
          >
            <span aria-hidden="true">{kindIcons[t.kind]}</span>
            <span>{t.message}</span>
          </div>
        ))}
      </div>
    </ToastCtx.Provider>
  )
}

export function useToast(): ToastCtxValue {
  return useContext(ToastCtx)
}