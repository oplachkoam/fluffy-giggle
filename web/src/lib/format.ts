const dateTime = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

const timeOfDay = new Intl.DateTimeFormat('ru-RU', {
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
})

export function formatDateTime(value: string | Date): string {
  return dateTime.format(new Date(value))
}

export function formatTimeOfDay(value: string | Date): string {
  return timeOfDay.format(new Date(value))
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms} мс`
  return `${(ms / 1000).toFixed(1)} с`
}

export function formatScore(score: number | null): string {
  return score === null ? '—' : `${score} / 5`
}

export function formatAverage(value: number | null): string {
  return value === null ? '—' : value.toFixed(2)
}
