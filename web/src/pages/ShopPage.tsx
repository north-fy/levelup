import { useCallback, useEffect, useState } from 'react'
import { api } from '../api'
import type { Purchase, ShopItem } from '../types'
import { Badge, Button, Card, ErrorText, Field, Input, TextArea, fmtTime, msg, useRefresh } from '../ui'

export default function ShopPage({ token }: { token: string }) {
  const [items, setItems] = useState<ShopItem[]>([])
  const [purchases, setPurchases] = useState<Purchase[]>([])
  const [error, setError] = useState<string | null>(null)
  const [tick, refresh] = useRefresh()

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
      setError(msg(e))
    }
  }, [token])

  useEffect(() => {
    void load()
  }, [load, tick])

  async function createItem() {
    if (!iTitle.trim()) return
    setError(null)
    try {
      await api.createItem(token, { title: iTitle, description: iDesc, price_gold: Number(iPrice) || 0 })
      setITitle('')
      setIDesc('')
      refresh()
    } catch (e) {
      setError(msg(e))
    }
  }

  async function act(fn: () => Promise<unknown>) {
    setError(null)
    try {
      await fn()
      refresh()
    } catch (e) {
      setError(msg(e))
    }
  }

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <div className="space-y-4">
        <Card title="Новый товар">
          <div className="space-y-3">
            <Field label="Название">
              <Input value={iTitle} onChange={(e) => setITitle(e.target.value)} placeholder="Меч +10" />
            </Field>
            <Field label="Описание">
              <TextArea value={iDesc} onChange={(e) => setIDesc(e.target.value)} rows={2} />
            </Field>
            <Field label="Цена (gold)">
              <Input type="number" value={iPrice} onChange={(e) => setIPrice(e.target.value)} />
            </Field>
            <Button onClick={() => void createItem()} disabled={!iTitle.trim()}>
              Создать товар
            </Button>
          </div>
        </Card>

        <Card title={`Мои покупки (${purchases.length})`}>
          <div className="max-h-72 space-y-1 overflow-y-auto">
            {purchases.map((p) => (
              <div key={p.id} className="flex justify-between rounded-lg border border-slate-100 px-3 py-2 text-sm">
                <span>
                  товар <b>#{p.item_id}</b> от #{p.seller_id}
                </span>
                <span className="text-emerald-600">−{p.price} 💰 · {fmtTime(p.created_at)}</span>
              </div>
            ))}
            {purchases.length === 0 && <div className="text-sm text-slate-400">Покупок нет</div>}
          </div>
        </Card>
      </div>

      <Card title={`Магазин (${items.length})`}>
        <div className="space-y-2">
          {items.map((i) => (
            <div key={i.id} className="flex items-start justify-between gap-3 rounded-lg border border-slate-100 p-3">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-medium text-slate-800">{i.title}</span>
                  <Badge color={i.is_active ? 'emerald' : 'slate'}>{i.is_active ? 'active' : 'inactive'}</Badge>
                  <Badge color="indigo">от #{i.seller_id}</Badge>
                </div>
                {i.description && <div className="mt-1 text-sm text-slate-500">{i.description}</div>}
                <div className="mt-1 text-sm font-medium text-amber-600">{i.price_gold} 💰</div>
              </div>
              <div className="flex shrink-0 flex-wrap gap-1">
                <Button
                  variant="ghost"
                  onClick={() => void act(() => api.updateItem(token, i.id, { is_active: !i.is_active }))}
                >
                  {i.is_active ? 'Деактивировать' : 'Активировать'}
                </Button>
                <Button onClick={() => void act(() => api.buyItem(token, i.id))}>Купить</Button>
                <Button variant="danger" onClick={() => void act(() => api.deleteItem(token, i.id))}>
                  ✕
                </Button>
              </div>
            </div>
          ))}
          {items.length === 0 && <div className="text-sm text-slate-400">Товаров нет</div>}
        </div>
        <ErrorText error={error} />
      </Card>
    </div>
  )
}
