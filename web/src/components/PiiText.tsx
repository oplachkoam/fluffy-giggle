import { Fragment } from 'react'

import { type PiiType, piiLabels } from '@/lib/runs'
import { cn } from '@/lib/utils'

const placeholderRe = /\[([A-Z]+)_(\d+)\]/g

type PiiTextProps = {
  text: string
  /** Raw values to flag inside the text, for example egress findings. */
  marks?: string[]
  className?: string
}

/** Text with placeholders underlined and known raw values flagged. */
export function PiiText({ text, marks = [], className }: PiiTextProps) {
  const parts: React.ReactNode[] = []
  const pattern = new RegExp(
    [placeholderRe.source, ...marks.filter(Boolean).map(escape)].join('|'),
    'g',
  )
  let last = 0
  for (const match of text.matchAll(pattern)) {
    const index = match.index ?? 0
    if (index > last) parts.push(text.slice(last, index))
    const value = match[0]
    if (match[1]) {
      const type = match[1] as PiiType
      parts.push(
        <mark
          key={index}
          title={piiLabels[type] ?? type}
          className="bg-accent text-foreground rounded-sm px-0.5 font-mono text-[0.9em] underline decoration-dotted underline-offset-2"
        >
          {value}
        </mark>,
      )
    } else {
      parts.push(
        <mark
          key={index}
          title="значение из vault в payload"
          className="bg-negative/15 text-negative rounded-sm px-0.5"
        >
          {value}
        </mark>,
      )
    }
    last = index + value.length
  }
  if (last < text.length) parts.push(text.slice(last))

  return (
    <p className={cn('text-sm leading-relaxed whitespace-pre-wrap', className)}>
      {parts.map((part, i) => (
        <Fragment key={i}>{part}</Fragment>
      ))}
    </p>
  )
}

function escape(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}
