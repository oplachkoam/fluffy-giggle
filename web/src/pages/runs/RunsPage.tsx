import { useState } from 'react'
import { Link } from 'react-router'

import { ModeBadge, StatusBadge } from '@/components/Badges'
import { ErrorCard } from '@/components/ErrorCard'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { errorText } from '@/lib/api'
import { formatDateTime, formatScore } from '@/lib/format'
import { type Run, useRuns } from '@/lib/runs'

export function RunsPage() {
  const [pages, setPages] = useState<(string | null)[]>([null])
  return (
    <div className="space-y-4">
      {pages.map((cursor, i) => (
        <RunsChunk
          key={cursor ?? 'first'}
          before={cursor}
          last={i === pages.length - 1}
          onMore={(next) => setPages((p) => [...p, next])}
        />
      ))}
    </div>
  )
}

function RunsChunk({
  before,
  last,
  onMore,
}: {
  before: string | null
  last: boolean
  onMore: (next: string) => void
}) {
  const query = useRuns(before)
  if (query.isPending) {
    return <Skeleton className="h-40 w-full" />
  }
  if (query.isError) {
    return (
      <ErrorCard
        title="Не удалось загрузить запуски"
        message={errorText(query.error, 'Попробуйте ещё раз.')}
        onRetry={() => void query.refetch()}
      />
    )
  }
  if (before === null && query.data.runs.length === 0) {
    return (
      <p className="text-muted-foreground text-sm">
        Запусков пока нет.{' '}
        <Link to="/" className="underline">
          Создайте первый.
        </Link>
      </p>
    )
  }
  return (
    <>
      <Table>
        {before === null && (
          <TableHeader>
            <TableRow>
              <TableHead>Создан</TableHead>
              <TableHead>Вход</TableHead>
              <TableHead>Режим</TableHead>
              <TableHead>Статус</TableHead>
              <TableHead className="text-right">Оценка</TableHead>
              <TableHead className="text-right">Утечки</TableHead>
            </TableRow>
          </TableHeader>
        )}
        <TableBody>
          {query.data.runs.map((run) => (
            <RunRow key={run.id} run={run} />
          ))}
        </TableBody>
      </Table>
      {last && query.data.next && (
        <Button variant="outline" size="sm" onClick={() => onMore(query.data.next)}>
          Показать ещё
        </Button>
      )}
    </>
  )
}

function RunRow({ run }: { run: Run }) {
  return (
    <TableRow>
      <TableCell className="tabular-nums">
        <Link to={`/runs/${run.id}`} className="underline-offset-4 hover:underline">
          {formatDateTime(run.createdAt)}
        </Link>
      </TableCell>
      <TableCell className="text-muted-foreground">
        {run.source === 'image' ? 'скан' : 'текст'}
        {run.caseId && ` · ${run.caseId}`}
      </TableCell>
      <TableCell>
        <ModeBadge mode={run.advisorMode} />
      </TableCell>
      <TableCell>
        <StatusBadge status={run.status} />
      </TableCell>
      <TableCell className="text-right tabular-nums">{formatScore(run.judgeScore)}</TableCell>
      <TableCell className="text-right tabular-nums">
        {run.canariesTotal > 0 ? `${run.canariesLeaked} / ${run.canariesTotal}` : '—'}
      </TableCell>
    </TableRow>
  )
}
