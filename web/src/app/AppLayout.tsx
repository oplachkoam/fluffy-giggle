import { useQueryClient } from '@tanstack/react-query'
import { BarChart3Icon, ListIcon, LogOutIcon, PlusIcon } from 'lucide-react'
import { useCallback } from 'react'
import { matchPath, Navigate, NavLink, Outlet, useLocation } from 'react-router'

import { ThemeButton } from '@/components/ThemeButton'
import { Button } from '@/components/ui/button'
import { clearCredentials, useCredentials } from '@/lib/auth'
import { applyEvent, type AppEvent, useEvents } from '@/lib/events'
import { cn } from '@/lib/utils'

const sections = [
  { to: '/', label: 'Новый запуск', icon: PlusIcon, end: true },
  { to: '/runs', label: 'Запуски', icon: ListIcon },
  { to: '/benchmark', label: 'Бенчмарк', icon: BarChart3Icon },
] as const

const titles = [
  { pattern: '/', title: 'Новый запуск' },
  { pattern: '/runs', title: 'Запуски' },
  { pattern: '/runs/:id', title: 'Запуск' },
  { pattern: '/benchmark', title: 'Бенчмарк' },
]

export function AppLayout() {
  const credentials = useCredentials()
  const queryClient = useQueryClient()
  const onEvent = useCallback((event: AppEvent) => applyEvent(queryClient, event), [queryClient])
  useEvents(credentials !== null, onEvent)
  const location = useLocation()

  if (credentials === null) {
    return <Navigate to="/login" replace />
  }
  const title = titles.find((item) => matchPath(item.pattern, location.pathname))?.title ?? ''

  return (
    <div className="flex min-h-svh">
      <aside className="bg-card sticky top-0 hidden h-svh w-56 shrink-0 flex-col border-r md:flex">
        <div className="flex h-14 items-center px-4 font-mono text-sm font-semibold tracking-tight">
          fluffy-giggle
        </div>
        <nav aria-label="Разделы" className="flex flex-1 flex-col gap-0.5 px-2">
          {sections.map((section) => (
            <NavLink
              key={section.to}
              to={section.to}
              end={'end' in section}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors',
                  isActive
                    ? 'bg-accent text-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                )
              }
            >
              <section.icon className="size-4" aria-hidden />
              {section.label}
            </NavLink>
          ))}
        </nav>
        <div className="border-t px-2 py-2">
          <Button
            variant="ghost"
            size="sm"
            className="text-muted-foreground w-full justify-start"
            onClick={() => clearCredentials()}
          >
            <LogOutIcon className="size-4" aria-hidden />
            Выйти
          </Button>
        </div>
      </aside>
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="bg-background sticky top-0 z-10 flex h-14 items-center gap-3 border-b px-4 sm:px-6">
          <nav aria-label="Разделы" className="flex gap-1 md:hidden">
            {sections.map((section) => (
              <NavLink
                key={section.to}
                to={section.to}
                end={'end' in section}
                aria-label={section.label}
                className={({ isActive }) =>
                  cn('rounded-md p-2', isActive ? 'bg-accent' : 'text-muted-foreground')
                }
              >
                <section.icon className="size-4" aria-hidden />
              </NavLink>
            ))}
          </nav>
          <h1 className="min-w-0 flex-1 truncate text-base font-semibold tracking-tight">
            {title}
          </h1>
          <ThemeButton />
        </header>
        <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-6 sm:px-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
