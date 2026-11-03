import { useState } from 'react'
import { Link } from 'react-router'

import { StatusBadge } from '@/components/Badges'
import { ErrorCard } from '@/components/ErrorCard'
import { Button } from '@/components/ui/button'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
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
import { useBenchmark, useBenchmarks, useStartBenchmark } from '@/lib/benchmarks'
import { formatAverage, formatDateTime, formatScore } from '@/lib/format'
import { type AdvisorMode, modeLabels, type Run } from '@/lib/runs'

const modeOrder: Record<AdvisorMode, number> = { off: 0, on: 1, placebo: 2 }

/** Side by side per case: the three modes of one letter next to each other. */
function byCase(runs: Run[]): Run[] {
  return [...runs].sort(
    (a, b) =>
      a.caseId.localeCompare(b.caseId) || modeOrder[a.advisorMode] - modeOrder[b.advisorMode],
  )
}

export function BenchmarkPage() {
  const benchmarks = useBenchmarks()
  const start = useStartBenchmark()
  const [picked, setPicked] = useState<string | null>(null)
  const selected = picked ?? benchmarks.data?.[0]?.id ?? null
  const benchmark = useBenchmark(selected)
  const running = benchmarks.data?.some((b) => b.status === 'running') ?? false

  return (
    <div className="space-y-6">
      <p className="text-muted-foreground text-sm">
        Все кейсы набора в трёх режимах. Adviser не фикция, если «с adviser» выше «плацебо», а
        «плацебо» примерно равно «без adviser», при нуле утечек.
      </p>
      <div className="flex flex-wrap items-center gap-3">
        <Button disabled={running || start.isPending} onClick={() => start.mutate()}>
          Прогнать набор
        </Button>
        {benchmarks.data && benchmarks.data.length > 0 && (
          <NativeSelect
            aria-label="Бенчмарк"
            value={selected ?? ''}
            onChange={(e) => setPicked(e.target.value)}
          >
            {benchmarks.data.map((b) => (
              <NativeSelectOption key={b.id} value={b.id}>
                {formatDateTime(b.createdAt)} · {b.finished} / {b.total}
              </NativeSelectOption>
            ))}
          </NativeSelect>
        )}
        {start.isError && (
          <p role="alert" className="text-negative text-sm">
            {errorText(start.error, 'Не удалось запустить.')}
          </p>
        )}
      </div>

      {selected === null && !benchmarks.isPending && (
        <p className="text-muted-foreground text-sm">Бенчмарков ещё не было.</p>
      )}
      {benchmark.isPending && selected !== null && <Skeleton className="h-40 w-full" />}
      {benchmark.isError && (
        <ErrorCard
          title="Не удалось загрузить бенчмарк"
          message={errorText(benchmark.error, 'Попробуйте ещё раз.')}
          onRetry={() => void benchmark.refetch()}
        />
      )}
      {benchmark.data && (
        <>
          <p className="text-sm">
            Выполнено {benchmark.data.finished} из {benchmark.data.total}
            {benchmark.data.status === 'running' && (
              <span className="text-muted-foreground"> · идёт</span>
            )}
          </p>
          <Table aria-label="Сводка по режимам">
            <TableHeader>
              <TableRow>
                <TableHead>Режим</TableHead>
                <TableHead className="text-right">Готово</TableHead>
                <TableHead className="text-right">Средний балл</TableHead>
                <TableHead className="text-right">Утечки</TableHead>
                <TableHead className="text-right">Эскалации</TableHead>
                <TableHead className="text-right">Блокировки egress</TableHead>
                <TableHead className="text-right">Ошибки</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {benchmark.data.summary.map((s) => (
                <TableRow key={s.mode}>
                  <TableCell>{modeLabels[s.mode]}</TableCell>
                  <TableCell className="text-right tabular-nums">
                    {s.done} / {s.runs}
                  </TableCell>
                  <TableCell className="text-right font-medium tabular-nums">
                    {formatAverage(s.avgScore)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">{s.leaked}</TableCell>
                  <TableCell className="text-right tabular-nums">{s.escalated}</TableCell>
                  <TableCell className="text-right tabular-nums">{s.blocked}</TableCell>
                  <TableCell className="text-right tabular-nums">{s.failed}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Table aria-label="Запуски бенчмарка">
            <TableHeader>
              <TableRow>
                <TableHead>Кейс</TableHead>
                <TableHead>Режим</TableHead>
                <TableHead>Статус</TableHead>
                <TableHead className="text-right">Оценка</TableHead>
                <TableHead className="text-right">Утечки</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {byCase(benchmark.data.runs).map((run) => (
                <TableRow key={run.id}>
                  <TableCell>
                    <Link
                      to={`/runs/${run.id}`}
                      className="font-mono text-xs underline-offset-4 hover:underline"
                    >
                      {run.caseId}
                    </Link>
                  </TableCell>
                  <TableCell>{modeLabels[run.advisorMode]}</TableCell>
                  <TableCell>
                    <StatusBadge status={run.status} />
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {formatScore(run.judgeScore)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {run.canariesLeaked} / {run.canariesTotal}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </>
      )}
    </div>
  )
}
