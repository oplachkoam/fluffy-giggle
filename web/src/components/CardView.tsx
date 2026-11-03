import type { Card } from '@/lib/runs'

const fields: { key: keyof Card; label: string }[] = [
  { key: 'category', label: 'Категория' },
  { key: 'subject', label: 'Суть' },
  { key: 'amountRange', label: 'Сумма' },
  { key: 'eventPeriod', label: 'Период' },
  { key: 'requestedAction', label: 'Требование' },
  { key: 'summary', label: 'Кратко' },
]

export function CardView({ card }: { card: Card }) {
  return (
    <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
      {fields.map(({ key, label }) => (
        <div key={key} className="contents">
          <dt className="text-muted-foreground">{label}</dt>
          <dd>{card[key] || '—'}</dd>
        </div>
      ))}
    </dl>
  )
}
