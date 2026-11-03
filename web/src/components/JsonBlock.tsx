export function JsonBlock({ value }: { value: unknown }) {
  return (
    <pre className="bg-muted max-h-96 overflow-auto rounded-md p-3 font-mono text-xs leading-relaxed whitespace-pre-wrap">
      {typeof value === 'string' ? value : JSON.stringify(value, null, 2)}
    </pre>
  )
}
