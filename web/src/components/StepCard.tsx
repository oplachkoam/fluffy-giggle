import { ChevronDownIcon } from 'lucide-react'
import { useState } from 'react'

import { SideBadge, StepStatusText } from '@/components/Badges'
import { CardView } from '@/components/CardView'
import { JsonBlock } from '@/components/JsonBlock'
import { PiiText } from '@/components/PiiText'
import { Button } from '@/components/ui/button'
import { formatDuration } from '@/lib/format'
import { type Card, kindLabels, piiLabels, type Span, type Step } from '@/lib/runs'
import { type Finding, list, num, type Prompt, stepSummary, str } from '@/lib/steps'
import { cn } from '@/lib/utils'

export function StepCard({ step }: { step: Step }) {
  const [open, setOpen] = useState(step.status === 'failed' || step.status === 'blocked')
  const [raw, setRaw] = useState(false)
  const tokens = step.promptTokens + step.completionTokens

  return (
    <li className="bg-card rounded-lg border">
      <button
        type="button"
        className="flex w-full items-start gap-3 p-3 text-left"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
      >
        <span className="text-muted-foreground w-6 shrink-0 pt-0.5 font-mono text-xs tabular-nums">
          {step.seq}
        </span>
        <span className="min-w-0 flex-1">
          <span className="flex flex-wrap items-center gap-2">
            <span className="text-sm font-medium">{kindLabels[step.kind]}</span>
            <SideBadge side={step.side} />
            <StepStatusText status={step.status} />
          </span>
          <span className="text-muted-foreground mt-0.5 block truncate text-xs">
            {stepSummary(step)}
            <span className="sm:hidden">
              {' · '}
              {formatDuration(step.durationMs)}
              {step.model && ` · ${step.model}`}
            </span>
          </span>
        </span>
        <span className="text-muted-foreground hidden shrink-0 text-right font-mono text-[11px] tabular-nums sm:block">
          {formatDuration(step.durationMs)}
          {step.model && (
            <>
              <br />
              {step.model}
              {tokens > 0 && ` · ${tokens} ток.`}
            </>
          )}
        </span>
        <ChevronDownIcon
          aria-hidden
          className={cn(
            'text-muted-foreground mt-1 size-4 shrink-0 transition',
            open && 'rotate-180',
          )}
        />
      </button>
      {open && (
        <div className="space-y-3 border-t px-3 py-3 pl-12">
          {raw ? (
            <>
              <Section title="Вход (JSON)">
                <JsonBlock value={step.input} />
              </Section>
              <Section title="Выход (JSON)">
                <JsonBlock value={step.output} />
              </Section>
            </>
          ) : (
            <StepDetails step={step} />
          )}
          {step.error && <p className="text-negative text-sm">{step.error}</p>}
          <Button variant="ghost" size="sm" onClick={() => setRaw((v) => !v)}>
            {raw ? 'Показать разбор' : 'Показать сырой JSON'}
          </Button>
        </div>
      )}
    </li>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h4 className="text-muted-foreground mb-1 text-xs font-medium uppercase">{title}</h4>
      {children}
    </section>
  )
}

function PromptView({ prompt }: { prompt: Prompt }) {
  return (
    <div className="space-y-2">
      {prompt.system && (
        <details className="text-sm">
          <summary className="text-muted-foreground cursor-pointer">Системный промпт</summary>
          <p className="text-muted-foreground mt-1 text-sm whitespace-pre-wrap">{prompt.system}</p>
        </details>
      )}
      {prompt.image && <p className="text-muted-foreground text-sm">Изображение: {prompt.image}</p>}
      {prompt.user && <PiiText text={prompt.user} />}
    </div>
  )
}

function SpansTable({ spans }: { spans: Span[] }) {
  if (spans.length === 0) return <p className="text-muted-foreground text-sm">Ничего не найдено.</p>
  return (
    <ul className="space-y-0.5 text-sm">
      {spans.map((span, i) => (
        <li key={i} className="flex gap-2">
          <span className="text-muted-foreground w-24 shrink-0">
            {piiLabels[span.type] ?? span.type}
          </span>
          <span className="font-mono">{span.text}</span>
          {span.canonical && span.canonical !== span.text && (
            <span className="text-muted-foreground">→ {span.canonical}</span>
          )}
        </li>
      ))}
    </ul>
  )
}

function StepDetails({ step }: { step: Step }) {
  const { input, output } = step
  switch (step.kind) {
    case 'vision':
      return (
        <>
          <Section title="Что ушло модели">
            <PromptView prompt={(input.prompt as Prompt) ?? {}} />
          </Section>
          <Section title="Распознанный текст">
            <PiiText text={str(output.text)} />
          </Section>
        </>
      )
    case 'gate1':
      return (
        <>
          <Section title="Найдено">
            <SpansTable spans={list<Span>(output.spans)} />
          </Section>
          <Section title="Текст после гейта">
            <PiiText text={str(output.text)} />
          </Section>
        </>
      )
    case 'gate2':
      return (
        <>
          <Section title="Что ушло модели">
            <PromptView prompt={(input.prompt as Prompt) ?? {}} />
          </Section>
          <Section title="Найдено">
            <SpansTable spans={list<Span>(output.spans)} />
          </Section>
          {str(output.text) && (
            <Section title="Текст после прохода">
              <PiiText text={str(output.text)} />
            </Section>
          )}
          {str(output.raw) && (
            <Section title="Сырой ответ модели">
              <JsonBlock value={str(output.raw)} />
            </Section>
          )}
        </>
      )
    case 'review':
      return (
        <Section title="Что оставалось после гейта 2">
          <SpansTable spans={list<Span>(input.findings)} />
        </Section>
      )
    case 'draft':
      return (
        <>
          <Section title="Что ушло модели">
            <PromptView prompt={(input.prompt as Prompt) ?? {}} />
          </Section>
          {output.card ? (
            <Section title="Карточка">
              <CardView card={toCard(output.card as Record<string, string>)} />
            </Section>
          ) : null}
          <Section title="Черновик">
            <PiiText text={str(output.draft)} />
          </Section>
          {str(output.raw) && (
            <Section title="Сырой ответ модели">
              <JsonBlock value={str(output.raw)} />
            </Section>
          )}
        </>
      )
    case 'egress': {
      const findings = list<Finding>(output.findings)
      return (
        <>
          <Section
            title={`Payload для шага «${kindLabels[str(input.target) as Step['kind']] ?? str(input.target)}»`}
          >
            <PiiText text={str(input.payload)} marks={findings.map((f) => f.value)} />
          </Section>
          {findings.length > 0 && (
            <Section title="Находки">
              <ul className="text-sm">
                {findings.map((f, i) => (
                  <li key={i}>
                    <span className="text-muted-foreground">{piiLabels[f.type] ?? f.type}: </span>
                    <span className="text-negative font-mono">{f.value}</span>
                  </li>
                ))}
              </ul>
            </Section>
          )}
        </>
      )
    }
    case 'advise':
      return (
        <>
          {input.prompt ? (
            <Section title="Что ушло в облако">
              <PromptView prompt={input.prompt as Prompt} />
            </Section>
          ) : (
            <p className="text-muted-foreground text-sm">{str(input.reason)}</p>
          )}
          {list<string>(output.remarks).length > 0 && (
            <Section title="Правки">
              <ol className="list-decimal space-y-0.5 pl-5 text-sm">
                {list<string>(output.remarks).map((r, i) => (
                  <li key={i}>{r}</li>
                ))}
              </ol>
            </Section>
          )}
        </>
      )
    case 'finalize':
      return (
        <>
          {input.prompt ? (
            <Section title="Что ушло модели">
              <PromptView prompt={input.prompt as Prompt} />
            </Section>
          ) : (
            <p className="text-muted-foreground text-sm">{str(input.reason)}</p>
          )}
          <Section title="Текст">
            <PiiText text={str(output.text)} />
          </Section>
        </>
      )
    case 'judge':
      return (
        <>
          {input.prompt ? (
            <Section title="Что ушло в облако">
              <PromptView prompt={input.prompt as Prompt} />
            </Section>
          ) : (
            <p className="text-muted-foreground text-sm">{str(input.reason)}</p>
          )}
          {num(output.score) !== null && (
            <Section title="Оценка">
              <p className="text-sm">
                <span className="text-2xl font-semibold tabular-nums">{num(output.score)}</span>
                <span className="text-muted-foreground"> / 5 · {str(output.reason)}</span>
              </p>
            </Section>
          )}
        </>
      )
    case 'reidentify':
      return (
        <>
          <Section title="Обезличенный">
            <PiiText text={str(input.text)} />
          </Section>
          <Section title="Для оператора">
            <PiiText text={str(output.text)} />
          </Section>
        </>
      )
    default:
      return <JsonBlock value={output} />
  }
}

function toCard(raw: Record<string, string>): Card {
  return {
    category: raw.category ?? '',
    subject: raw.subject ?? '',
    amountRange: raw.amount_range ?? '',
    eventPeriod: raw.event_period ?? '',
    requestedAction: raw.requested_action ?? '',
    summary: raw.summary ?? '',
  }
}
