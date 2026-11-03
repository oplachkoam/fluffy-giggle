import { useEffect, useState } from 'react'
import { useParams } from 'react-router'

import { ModeBadge, StatusBadge } from '@/components/Badges'
import { CardView } from '@/components/CardView'
import { ErrorCard } from '@/components/ErrorCard'
import { PiiText } from '@/components/PiiText'
import { StepCard } from '@/components/StepCard'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { apiBlob, errorText } from '@/lib/api'
import { formatScore } from '@/lib/format'
import { outcomeLabels, piiLabels, type Run, useReview, useRun } from '@/lib/runs'

export function RunPage() {
  const { id = '' } = useParams()
  const query = useRun(id)

  if (query.isPending) {
    return <Skeleton className="h-64 w-full" />
  }
  if (query.isError) {
    return (
      <ErrorCard
        title="Не удалось загрузить запуск"
        message={errorText(query.error, 'Попробуйте ещё раз.')}
        onRetry={() => void query.refetch()}
      />
    )
  }
  const run = query.data

  return (
    <div className="space-y-8">
      <Summary run={run} />
      {run.status === 'needs_review' && <ReviewPanel run={run} />}
      {run.error && (
        <p role="alert" className="text-negative text-sm">
          {run.error}
        </p>
      )}

      <section aria-label="Трасса">
        <h2 className="mb-2 text-sm font-medium">Шаги</h2>
        {run.steps.length === 0 ? (
          <p className="text-muted-foreground text-sm">Ждём первый шаг…</p>
        ) : (
          <ol className="space-y-2">
            {run.steps.map((step) => (
              <StepCard key={step.id} step={step} />
            ))}
          </ol>
        )}
      </section>

      {run.status === 'done' && <Result run={run} />}

      <section aria-label="Вход" className="space-y-2">
        <h2 className="text-sm font-medium">Вход</h2>
        {run.hasImage && <RunImage id={run.id} />}
        {run.inputText && <PiiText text={run.inputText} className="text-muted-foreground" />}
      </section>
    </div>
  )
}

/** The image needs the Basic header, which a plain <img> cannot send. */
function RunImage({ id }: { id: string }) {
  const [url, setUrl] = useState<string | null>(null)
  useEffect(() => {
    let active = true
    let objectUrl: string | null = null
    apiBlob(`/runs/${id}/image`)
      .then((blob) => {
        if (!active || typeof URL.createObjectURL !== 'function') return
        objectUrl = URL.createObjectURL(blob)
        setUrl(objectUrl)
      })
      .catch(() => {})
    return () => {
      active = false
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [id])
  if (!url) return <Skeleton className="h-40 w-64" />
  return <img src={url} alt="Исходный скан" className="max-h-[32rem] rounded-md border" />
}

function Summary({ run }: { run: Run }) {
  const items = [
    { label: 'Adviser', value: outcomeLabels[run.advisorOutcome] },
    { label: 'Оценка судьи', value: formatScore(run.judgeScore) },
    { label: 'Проходов гейта 2', value: String(run.gate2Passes) },
    {
      label: 'Канарейки в облаке',
      value: run.canariesTotal > 0 ? `${run.canariesLeaked} из ${run.canariesTotal}` : 'не кейс',
    },
    { label: 'В vault', value: String(run.vault.length) },
  ]
  return (
    <section aria-label="Итог" className="space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <StatusBadge status={run.status} />
        <ModeBadge mode={run.advisorMode} />
        {run.escalated && <span className="text-muted-foreground text-xs">была эскалация</span>}
        {run.caseId && (
          <span className="text-muted-foreground font-mono text-xs">{run.caseId}</span>
        )}
      </div>
      <dl className="grid grid-cols-2 gap-3 sm:grid-cols-5">
        {items.map((item) => (
          <div key={item.label}>
            <dt className="text-muted-foreground text-xs">{item.label}</dt>
            <dd className="text-sm font-medium tabular-nums">{item.value}</dd>
          </div>
        ))}
      </dl>
    </section>
  )
}

function ReviewPanel({ run }: { run: Run }) {
  const review = useReview(run.id)
  return (
    <section
      aria-label="Нужна проверка"
      className="border-foreground space-y-3 rounded-lg border p-4"
    >
      <p className="text-sm font-medium">Гейт 2 продолжает находить данные. Решите, что делать.</p>
      <ul className="text-sm">
        {run.reviewFindings.map((f, i) => (
          <li key={i}>
            <span className="text-muted-foreground">{piiLabels[f.type] ?? f.type}: </span>
            <span className="font-mono">{f.text}</span>
          </li>
        ))}
      </ul>
      <PiiText text={run.anonymizedText} className="text-muted-foreground" />
      {review.isError && (
        <p role="alert" className="text-negative text-sm">
          {errorText(review.error, 'Не удалось отправить решение.')}
        </p>
      )}
      <div className="flex flex-wrap gap-2">
        <Button disabled={review.isPending} onClick={() => review.mutate('continue')}>
          Это не ПДн, продолжить
        </Button>
        <Button
          variant="outline"
          disabled={review.isPending}
          onClick={() => review.mutate('finish_without_advisor')}
        >
          Завершить без adviser
        </Button>
      </div>
    </section>
  )
}

function Result({ run }: { run: Run }) {
  const [anonymous, setAnonymous] = useState(false)
  return (
    <section aria-label="Результат" className="space-y-3">
      <div className="flex items-center gap-3">
        <h2 className="text-sm font-medium">Ответ клиенту</h2>
        <Button variant="ghost" size="sm" onClick={() => setAnonymous((v) => !v)}>
          {anonymous ? 'Показать для оператора' : 'Показать обезличенный'}
        </Button>
      </div>
      <PiiText text={anonymous ? run.finalAnonymized : run.finalText} />
      {run.card && (
        <details>
          <summary className="text-muted-foreground cursor-pointer text-sm">
            Карточка обращения
          </summary>
          <div className="mt-2">
            <CardView card={run.card} />
          </div>
        </details>
      )}
      {run.judgeReason && <p className="text-muted-foreground text-sm">Судья: {run.judgeReason}</p>}
    </section>
  )
}
