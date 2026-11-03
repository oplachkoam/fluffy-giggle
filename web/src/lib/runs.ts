import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from '@/lib/api'

export type RunStatus = 'queued' | 'running' | 'needs_review' | 'done' | 'failed'
export type AdvisorMode = 'off' | 'on' | 'placebo'
export type AdvisorOutcome = 'pending' | 'applied' | 'placebo' | 'skipped' | 'blocked'
export type Side = 'local' | 'cloud' | 'system' | 'placebo'
export type StepStatus = 'ok' | 'failed' | 'blocked' | 'skipped'
export type StepKind =
  | 'ingest'
  | 'vision'
  | 'gate1'
  | 'gate2'
  | 'review'
  | 'draft'
  | 'egress'
  | 'advise'
  | 'finalize'
  | 'judge'
  | 'reidentify'
export type PiiType =
  | 'PERSON'
  | 'PHONE'
  | 'EMAIL'
  | 'PASSPORT'
  | 'SNILS'
  | 'INN'
  | 'CARD'
  | 'ACCOUNT'
  | 'ADDRESS'
  | 'DATE'
  | 'ORG'
  | 'OTHER'

export type Span = { text: string; type: PiiType; canonical: string }
export type VaultEntry = { placeholder: string; type: PiiType; canonical: string; forms: string[] }
export type Card = {
  category: string
  subject: string
  amountRange: string
  eventPeriod: string
  requestedAction: string
  summary: string
}

export type Step = {
  id: string
  runId: string
  seq: number
  kind: StepKind
  side: Side
  status: StepStatus
  startedAt: string
  finishedAt: string
  durationMs: number
  model: string
  promptTokens: number
  completionTokens: number
  input: Record<string, unknown>
  output: Record<string, unknown>
  error: string
}

export type Run = {
  id: string
  createdAt: string
  updatedAt: string
  status: RunStatus
  advisorMode: AdvisorMode
  advisorOutcome: AdvisorOutcome
  reviewDecision: string
  source: 'text' | 'image'
  caseId: string
  benchmarkId: string
  hasImage: boolean
  inputText: string
  anonymizedText: string
  vault: VaultEntry[]
  reviewFindings: Span[]
  card: Card | null
  draft: string
  advice: string[]
  finalAnonymized: string
  finalText: string
  judgeScore: number | null
  judgeReason: string
  gate2Passes: number
  escalated: boolean
  canariesTotal: number
  canariesLeaked: number
  error: string
  steps: Step[]
}

export type RunDto = {
  id: string
  created_at: string
  updated_at: string
  status: RunStatus
  advisor_mode: AdvisorMode
  advisor_outcome: AdvisorOutcome
  review_decision: string
  source: 'text' | 'image'
  case_id: string
  benchmark_id: string
  has_image: boolean
  input_text: string
  anonymized_text: string
  vault: VaultEntry[]
  review_findings: Span[]
  card: {
    category: string
    subject: string
    amount_range: string
    event_period: string
    requested_action: string
    summary: string
  } | null
  draft: string
  advice: string[]
  final_anonymized: string
  final_text: string
  judge_score: number | null
  judge_reason: string
  gate2_passes: number
  escalated: boolean
  canaries_total: number
  canaries_leaked: number
  error: string
  steps?: StepDto[]
}

export type StepDto = {
  id: string
  run_id: string
  seq: number
  kind: StepKind
  side: Side
  status: StepStatus
  started_at: string
  finished_at: string
  duration_ms: number
  model: string
  prompt_tokens: number
  completion_tokens: number
  input: Record<string, unknown>
  output: Record<string, unknown>
  error: string
}

export function toStep(dto: StepDto): Step {
  return {
    id: dto.id,
    runId: dto.run_id,
    seq: dto.seq,
    kind: dto.kind,
    side: dto.side,
    status: dto.status,
    startedAt: dto.started_at,
    finishedAt: dto.finished_at,
    durationMs: dto.duration_ms,
    model: dto.model,
    promptTokens: dto.prompt_tokens,
    completionTokens: dto.completion_tokens,
    input: dto.input ?? {},
    output: dto.output ?? {},
    error: dto.error,
  }
}

export function toRun(dto: RunDto): Run {
  return {
    id: dto.id,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
    status: dto.status,
    advisorMode: dto.advisor_mode,
    advisorOutcome: dto.advisor_outcome,
    reviewDecision: dto.review_decision,
    source: dto.source,
    caseId: dto.case_id,
    benchmarkId: dto.benchmark_id,
    hasImage: dto.has_image,
    inputText: dto.input_text,
    anonymizedText: dto.anonymized_text,
    vault: dto.vault ?? [],
    reviewFindings: dto.review_findings ?? [],
    card: dto.card
      ? {
          category: dto.card.category,
          subject: dto.card.subject,
          amountRange: dto.card.amount_range,
          eventPeriod: dto.card.event_period,
          requestedAction: dto.card.requested_action,
          summary: dto.card.summary,
        }
      : null,
    draft: dto.draft,
    advice: dto.advice ?? [],
    finalAnonymized: dto.final_anonymized,
    finalText: dto.final_text,
    judgeScore: dto.judge_score,
    judgeReason: dto.judge_reason,
    gate2Passes: dto.gate2_passes,
    escalated: dto.escalated,
    canariesTotal: dto.canaries_total,
    canariesLeaked: dto.canaries_leaked,
    error: dto.error,
    steps: (dto.steps ?? []).map(toStep),
  }
}

export const statusLabels: Record<RunStatus, string> = {
  queued: 'в очереди',
  running: 'выполняется',
  needs_review: 'нужна проверка',
  done: 'готово',
  failed: 'ошибка',
}

export const modeLabels: Record<AdvisorMode, string> = {
  off: 'без adviser',
  on: 'с adviser',
  placebo: 'плацебо',
}

export const outcomeLabels: Record<AdvisorOutcome, string> = {
  pending: 'ещё не дошли',
  applied: 'совет применён',
  placebo: 'плацебо применено',
  skipped: 'пропущен',
  blocked: 'заблокирован egress',
}

export const sideLabels: Record<Side, string> = {
  local: 'локально',
  cloud: 'облако',
  system: 'код',
  placebo: 'плацебо',
}

export const kindLabels: Record<StepKind, string> = {
  ingest: 'Вход',
  vision: 'Распознавание скана',
  gate1: 'Гейт 1: детекторы',
  gate2: 'Гейт 2: слабая модель',
  review: 'Решение оператора',
  draft: 'Карточка и черновик',
  egress: 'Проверка перед отправкой',
  advise: 'Совет сильной модели',
  finalize: 'Финальный текст',
  judge: 'Оценка судьи',
  reidentify: 'Обратная подстановка',
}

export const stepStatusLabels: Record<StepStatus, string> = {
  ok: 'ок',
  failed: 'ошибка',
  blocked: 'заблокировано',
  skipped: 'пропущено',
}

export const piiLabels: Record<PiiType, string> = {
  PERSON: 'имя',
  PHONE: 'телефон',
  EMAIL: 'почта',
  PASSPORT: 'паспорт',
  SNILS: 'СНИЛС',
  INN: 'ИНН',
  CARD: 'карта',
  ACCOUNT: 'счёт',
  ADDRESS: 'адрес',
  DATE: 'дата',
  ORG: 'организация',
  OTHER: 'прочее',
}

export function useRun(id: string) {
  return useQuery({
    queryKey: ['run', id],
    queryFn: async () => toRun(await api<RunDto>(`/runs/${id}`)),
    refetchInterval: (query) => {
      const status = query.state.data?.status
      return status === 'queued' || status === 'running' ? 5000 : false
    },
  })
}

export function useRuns(before: string | null) {
  return useQuery({
    queryKey: ['runs', before],
    queryFn: async () => {
      const data = await api<{ runs: RunDto[]; next: string }>(
        before ? `/runs?before=${encodeURIComponent(before)}` : '/runs',
      )
      return { runs: data.runs.map(toRun), next: data.next }
    },
  })
}

export type CreateRunInput = {
  advisorMode: AdvisorMode
  text?: string
  image?: File
  caseId?: string
}

export function useCreateRun() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateRunInput) => {
      if (input.image) {
        const form = new FormData()
        form.set('advisor_mode', input.advisorMode)
        form.set('image', input.image)
        return toRun(await api<RunDto>('/runs', { method: 'POST', form }))
      }
      return toRun(
        await api<RunDto>('/runs', {
          method: 'POST',
          body: { advisor_mode: input.advisorMode, text: input.text, case_id: input.caseId },
        }),
      )
    },
    onSuccess: (run) => {
      queryClient.setQueryData(['run', run.id], run)
      void queryClient.invalidateQueries({ queryKey: ['runs'] })
    },
  })
}

export type ReviewDecision = 'continue' | 'finish_without_advisor'

export function useReview(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (decision: ReviewDecision) =>
      toRun(await api<RunDto>(`/runs/${id}/review`, { method: 'POST', body: { decision } })),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['run', id] })
    },
  })
}
