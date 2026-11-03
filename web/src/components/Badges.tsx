import { Badge } from '@/components/ui/badge'
import {
  type AdvisorMode,
  modeLabels,
  type RunStatus,
  type Side,
  sideLabels,
  statusLabels,
  type StepStatus,
  stepStatusLabels,
} from '@/lib/runs'
import { cn } from '@/lib/utils'

export function StatusBadge({ status }: { status: RunStatus }) {
  return (
    <Badge
      variant={status === 'needs_review' ? 'default' : 'outline'}
      className={cn(
        status === 'failed' && 'border-negative text-negative',
        status === 'running' && 'animate-pulse',
      )}
    >
      {statusLabels[status]}
    </Badge>
  )
}

/** Cloud is filled: the one side that is outside the perimeter must stand out. */
export function SideBadge({ side }: { side: Side }) {
  return (
    <Badge
      variant={side === 'cloud' ? 'default' : side === 'placebo' ? 'secondary' : 'outline'}
      className={cn('font-mono text-[11px] uppercase', side === 'system' && 'border-dashed')}
    >
      {sideLabels[side]}
    </Badge>
  )
}

export function ModeBadge({ mode }: { mode: AdvisorMode }) {
  return <Badge variant="outline">{modeLabels[mode]}</Badge>
}

export function StepStatusText({ status }: { status: StepStatus }) {
  return (
    <span
      className={cn(
        'text-xs',
        status === 'ok' && 'text-muted-foreground',
        status === 'failed' && 'text-negative',
        status === 'blocked' && 'text-negative font-medium',
        status === 'skipped' && 'text-muted-foreground italic',
      )}
    >
      {stepStatusLabels[status]}
    </span>
  )
}
