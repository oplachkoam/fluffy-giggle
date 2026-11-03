import { Button } from '@/components/ui/button'

type ErrorCardProps = {
  title: string
  message: string
  onRetry?: () => void
}

export function ErrorCard({ title, message, onRetry }: ErrorCardProps) {
  return (
    <div role="alert" className="rounded-lg border p-4">
      <p className="text-sm font-medium">{title}</p>
      <p className="text-muted-foreground mt-1 text-sm">{message}</p>
      {onRetry && (
        <Button variant="outline" size="sm" className="mt-3" onClick={onRetry}>
          Повторить
        </Button>
      )}
    </div>
  )
}
