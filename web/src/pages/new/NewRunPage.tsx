import { type FormEvent, useState } from 'react'
import { useNavigate } from 'react-router'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'
import { errorText } from '@/lib/api'
import { useCases } from '@/lib/cases'
import { type AdvisorMode, modeLabels, useCreateRun } from '@/lib/runs'
import { cn } from '@/lib/utils'

type SourceKind = 'text' | 'image' | 'case'

const sources: { kind: SourceKind; label: string }[] = [
  { kind: 'text', label: 'Текст' },
  { kind: 'image', label: 'Скан' },
  { kind: 'case', label: 'Кейс из набора' },
]

export function NewRunPage() {
  const navigate = useNavigate()
  const cases = useCases()
  const create = useCreateRun()
  const [source, setSource] = useState<SourceKind>('text')
  const [text, setText] = useState('')
  const [image, setImage] = useState<File | null>(null)
  const [caseId, setCaseId] = useState('')
  const [mode, setMode] = useState<AdvisorMode>('on')

  const ready =
    (source === 'text' && text.trim() !== '') ||
    (source === 'image' && image !== null) ||
    (source === 'case' && caseId !== '')

  const submit = (event: FormEvent) => {
    event.preventDefault()
    create.mutate(
      {
        advisorMode: mode,
        text: source === 'text' ? text : undefined,
        image: source === 'image' ? (image ?? undefined) : undefined,
        caseId: source === 'case' ? caseId : undefined,
      },
      { onSuccess: (run) => void navigate(`/runs/${run.id}`) },
    )
  }

  return (
    <form onSubmit={submit} noValidate className="max-w-3xl space-y-6">
      <p className="text-muted-foreground text-sm">
        Вход обезличивается двумя гейтами, сильная модель видит только карточку и черновик. Каждый
        шаг и всё, что ушло в облако, появится на странице запуска.
      </p>

      <div className="bg-muted flex gap-1 rounded-md p-1" role="tablist" aria-label="Источник">
        {sources.map((item) => (
          <button
            key={item.kind}
            type="button"
            role="tab"
            aria-selected={source === item.kind}
            onClick={() => setSource(item.kind)}
            className={cn(
              'flex-1 rounded-sm px-3 py-1.5 text-sm transition-colors',
              source === item.kind ? 'bg-background shadow-sm' : 'text-muted-foreground',
            )}
          >
            {item.label}
          </button>
        ))}
      </div>

      {source === 'text' && (
        <div className="space-y-1.5">
          <Label htmlFor="text">Текст обращения</Label>
          <Textarea
            id="text"
            rows={12}
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Вставьте письмо клиента с персональными данными"
          />
        </div>
      )}
      {source === 'image' && (
        <div className="space-y-1.5">
          <Label htmlFor="image">Скан или фото, PNG или JPEG до 10 МБ</Label>
          <Input
            id="image"
            type="file"
            accept="image/png,image/jpeg"
            onChange={(e) => setImage(e.target.files?.[0] ?? null)}
          />
          <p className="text-muted-foreground text-xs">
            Картинка уходит только модели распознавания на локальной стороне.
          </p>
        </div>
      )}
      {source === 'case' && (
        <div className="space-y-1.5">
          <Label htmlFor="case">Кейс</Label>
          <NativeSelect id="case" value={caseId} onChange={(e) => setCaseId(e.target.value)}>
            <NativeSelectOption value="">Выберите кейс</NativeSelectOption>
            {(cases.data ?? []).map((c) => (
              <NativeSelectOption key={c.id} value={c.id}>
                {c.title}
                {c.hasImage ? ' · скан' : ''} · канареек: {c.canaries}
              </NativeSelectOption>
            ))}
          </NativeSelect>
          <p className="text-muted-foreground text-xs">
            У кейса известны все подставленные значения, после запуска видно, сколько их дошло до
            облака.
          </p>
        </div>
      )}

      <div className="space-y-1.5">
        <Label htmlFor="mode">Adviser</Label>
        <NativeSelect
          id="mode"
          value={mode}
          onChange={(e) => setMode(e.target.value as AdvisorMode)}
        >
          {(Object.keys(modeLabels) as AdvisorMode[]).map((m) => (
            <NativeSelectOption key={m} value={m}>
              {modeLabels[m]}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </div>

      {create.isError && (
        <p role="alert" className="text-negative text-sm">
          {errorText(create.error, 'Не удалось создать запуск.')}
        </p>
      )}
      <Button type="submit" disabled={!ready || create.isPending}>
        Запустить
      </Button>
    </form>
  )
}
