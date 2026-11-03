import type { PiiType, Span, Step } from '@/lib/runs'

export type Finding = { type: PiiType; value: string }
export type Prompt = { system?: string; user?: string; image?: string }

export function str(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

export function num(value: unknown): number | null {
  return typeof value === 'number' ? value : null
}

export function list<T>(value: unknown): T[] {
  return Array.isArray(value) ? (value as T[]) : []
}

/** One line that says what the step did, visible without expanding. */
export function stepSummary(step: Step): string {
  const { input, output } = step
  if (step.error) return step.error
  switch (step.kind) {
    case 'vision':
      return `распознано ${str(output.text).length} символов`
    case 'gate1':
    case 'gate2': {
      const spans = list<Span>(output.spans)
      const pass = num(input.pass)
      const prefix = pass ? `проход ${pass}: ` : ''
      return prefix + (spans.length ? `найдено ${spans.length}` : 'ничего не найдено')
    }
    case 'review':
      return str(output.decision) === 'continue'
        ? 'это не ПДн, продолжаем'
        : 'завершить без adviser'
    case 'draft':
      return `черновик ${str(output.draft).length} символов`
    case 'egress': {
      const findings = list<Finding>(output.findings)
      return findings.length
        ? `найдено в payload: ${findings.map((f) => f.value).join(', ')}`
        : `payload чист, цель: ${str(input.target)}`
    }
    case 'advise': {
      const remarks = list<string>(output.remarks)
      if (step.status === 'skipped') return str(input.reason)
      if (step.status === 'blocked') return 'вызов не сделан'
      return remarks.length ? `правок: ${remarks.length}` : 'правок нет'
    }
    case 'finalize':
      return step.status === 'skipped'
        ? 'финал равен черновику'
        : `${str(output.text).length} символов`
    case 'judge':
      return step.status === 'ok' ? `оценка ${num(output.score)} / 5` : 'оценки нет'
    case 'reidentify':
      return 'плейсхолдеры заменены значениями'
    default:
      return ''
  }
}
