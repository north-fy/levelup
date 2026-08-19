import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api } from '../api'
import type { Purchase, ShopItem } from '../types'
import Button from '../components/ui/Button'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import FormField from '../components/ui/FormField'
import { Input, TextArea } from '../components/ui/Input'
import Spinner from '../components/ui/Spinner'
import { useToast } from '../toast'

function fmtTime(s: string): string {
  const d = new Date(s)
  return isNaN(d.getTime()) ? s : d.toLocaleString()
}

export default function ShopPage({ token }: { token: string }) {
  const [items, setItems] = useState<ShopItem[] | null>(null)
  const [purchases, setPurchases] = useState<Purchase[]>([])
  const [error, setError] = useState<string | null>(null)
  const { notify } = useToast()

  const [iTitle, setITitle] = useState('')
  const [iDesc, setIDesc] = useState('')
  const [iPrice, setIPrice] = useState('50')

  const load = useCallback(async () => {
    setError(null)
    try {
      const [it, pu] = await Promise.all([api.listItems(token), api.listPurchases(token)])
      setItems(it)
      setPurchases(pu)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }, [token])

  useEffect(() => {
    void load()
  }, [load])

  async function createItem(e: FormEvent) {
    e.preventDefault()
    if (!iTitle.trim()) return
    setError(null)
    try {
      await api.createItem(token, { title: iTitle.trim(), description: iDesc, price_gold: Number(iPrice) || 0 })
      setITitle('')
      setIDesc('')
      notify('Товар создан 🛒', 'success')
      await load()
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  async function act(label: string, fn: () => Promise<unknown>) {
    setError(null)
    try {
      await fn()
      notify(label, 'success')
      await load()
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      setError(m)
      notify(m, 'error')
    }
  }

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <div className="space-y-4">
        <Card title="Новый товар">
          <form onSubmit={(e) => void createItem(e)} className="space-y-3" noValidate>
            <FormField label="Название">
              <Input value={iTitle} onChange={(e) => setITitle(e.target.value)} placeholder="Меч +10" />
            </FormField>
            <FormField label="Описание">
              <TextArea value={iDesc} onChange={(e) => setIDesc(e.target.value)} rows={2} />
            </FormField>
            <FormField label="Цена (gold)">
              <Input type="number" min={1} value={iPrice} onChange={(e) => setIPrice(e.target.value)} />
            </FormField>
            <Button type="submit" disabled={!iTitle.trim()}>
              Создать товар
            </Button>
          </form>
        </Card>

        <Card title={`Мои покупки (${purchases.length})`}>
          {purchases.length === 0 ? (
            <p className="text-sm text-slate-500 dark:text-slate-400">Покупок нет</p>
          ) : (
            <ul className="max-h-72 space-y-1 overflow-y-auto">
              {purchases.map((p) => (
                <li
                  key={p.id}
                  className="flex items-center justify-between gap-2 rounded-xl border border-slate-200 bg-white/60 px-3 py-2 text-sm dark:border-slate-800 dark:bg-slate-900/50"
                >
                  <span className="text-slate-700 dark:text-slate-200">
                    товар <b>#{p.item_id}</b> от продавца #{p.seller_id}
                  </span>
                  <span className="shrink-0 text-xs text-emerald-600 dark:text-emerald-400">
                    −{p.price} 💰 · {fmtTime(p.created_at)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </Card>
      </div>

      <Card title={`Магазин${items ? ` (${items.length})` : ''}`}>
        {!items ? (
          <Spinner label="Загрузка товаров" />
        ) : items.length === 0 ? (
          <p className="text-sm text-slate-500 dark:text-slate-400">Товаров нет — создайте первый</p>
        ) : (
          <ul className="space-y-2">
            {items.map((i) => (
              <li
                key={i.id}
                className="flex flex-wrap items-start justify-between gap-3 rounded-2xl border border-slate-200 bg-white/70 p-3 dark:border-slate-800 dark:bg-slate-900/50"
              >
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-bold text-slate-800 dark:text-slate-100">{i.title}</span>
                    <Badge color={i.is_active ? 'emerald' : 'slate'}>{i.is_active ? 'active' : 'inactive'}</Badge>
                    <Badge color="indigo">продавец #{i.seller_id}</Badge>
                  </div>
                  {i.description && <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{i.description}</p>}
                  <p className="mt-1 text-sm font-bold text-amber-600 dark:text-amber-400">{i.price_gold} 💰</p>
                </div>
                <div className="flex shrink-0 flex-wrap gap-1">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => void act(i.is_active ? 'Товар деактивирован' : 'Товар активирован', () => api.updateItem(token, i.id, { is_active: !i.is_active }))}
                  >
                    {i.is_active ? 'Деактивировать' : 'Активировать'}
                  </Button>
                  <Button size="sm" onClick={() => void act('Покупка совершена 🎉', () => api.buyItem(token, i.id))}>
                    Купить
                  </Button>
                  <Button variant="danger" size="sm" onClick={() => void act('Товар удалён', () => api.deleteItem(token, i.id))}>
                    ✕
                  </Button>
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