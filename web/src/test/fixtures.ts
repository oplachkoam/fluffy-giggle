import type { RunDto, StepDto } from '@/lib/runs'

export function runDto(overrides: Partial<RunDto> = {}): RunDto {
  return {
    id: '019a3f8e-0000-7000-8000-000000000001',
    created_at: '2026-11-03T12:00:00Z',
    updated_at: '2026-11-03T12:00:05Z',
    status: 'done',
    advisor_mode: 'on',
    advisor_outcome: 'applied',
    review_decision: '',
    source: 'text',
    case_id: 'case-01',
    benchmark_id: '',
    has_image: false,
    input_text: 'Я, Иванов Пётр Сергеевич, тел. +7 916 123-45-67, требую выплату.',
    anonymized_text: 'Я, [PERSON_1], тел. [PHONE_1], требую выплату.',
    vault: [
      {
        placeholder: '[PHONE_1]',
        type: 'PHONE',
        canonical: '79161234567',
        forms: ['+7 916 123-45-67'],
      },
      {
        placeholder: '[PERSON_1]',
        type: 'PERSON',
        canonical: 'Иванов Пётр Сергеевич',
        forms: ['Иванов Пётр Сергеевич'],
      },
    ],
    review_findings: [],
    card: {
      category: 'выплата',
      subject: 'задержка',
      amount_range: 'до 50 тыс. ₽',
      event_period: 'не указан',
      requested_action: 'выплатить',
      summary: 'Клиент ждёт выплату.',
    },
    draft: 'Уважаемый [PERSON_1], мы получили обращение.',
    advice: ['Укажите срок.'],
    final_anonymized: 'Уважаемый [PERSON_1], ответим в течение 10 дней.',
    final_text: 'Уважаемый Иванов Пётр Сергеевич, ответим в течение 10 дней.',
    judge_score: 4,
    judge_reason: 'Полный ответ.',
    gate2_passes: 2,
    escalated: false,
    canaries_total: 16,
    canaries_leaked: 0,
    error: '',
    steps: [],
    ...overrides,
  }
}

export function stepDto(overrides: Partial<StepDto> = {}): StepDto {
  return {
    id: '019a3f8e-0000-7000-8000-00000000000a',
    run_id: '019a3f8e-0000-7000-8000-000000000001',
    seq: 1,
    kind: 'gate1',
    side: 'system',
    status: 'ok',
    started_at: '2026-11-03T12:00:00Z',
    finished_at: '2026-11-03T12:00:00.120Z',
    duration_ms: 120,
    model: '',
    prompt_tokens: 0,
    completion_tokens: 0,
    input: { text: 'Я, Иванов Пётр Сергеевич' },
    output: {
      spans: [{ text: '+7 916 123-45-67', type: 'PHONE', canonical: '79161234567' }],
      text: 'Я, Иванов Пётр Сергеевич, тел. [PHONE_1]',
    },
    error: '',
    ...overrides,
  }
}
