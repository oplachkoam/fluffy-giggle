import { MonitorIcon, MoonIcon, SunIcon } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { nextTheme, setTheme, themeLabels, useTheme } from '@/lib/theme'

const icons = { system: MonitorIcon, light: SunIcon, dark: MoonIcon } as const

export function ThemeButton() {
  const theme = useTheme()
  const Icon = icons[theme]
  return (
    <Button
      variant="ghost"
      size="icon"
      aria-label={`Тема: ${themeLabels[theme]}`}
      onClick={() => setTheme(nextTheme(theme))}
    >
      <Icon aria-hidden />
    </Button>
  )
}
