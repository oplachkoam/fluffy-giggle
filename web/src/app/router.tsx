import { createBrowserRouter, Navigate, type RouteObject } from 'react-router'

import { AppLayout } from '@/app/AppLayout'
import { LoginPage } from '@/pages/auth/LoginPage'
import { BenchmarkPage } from '@/pages/benchmark/BenchmarkPage'
import { NewRunPage } from '@/pages/new/NewRunPage'
import { RunPage } from '@/pages/runs/RunPage'
import { RunsPage } from '@/pages/runs/RunsPage'

export const routes: RouteObject[] = [
  { path: '/login', Component: LoginPage },
  {
    path: '/',
    Component: AppLayout,
    children: [
      { index: true, Component: NewRunPage },
      { path: 'runs', Component: RunsPage },
      { path: 'runs/:id', Component: RunPage },
      { path: 'benchmark', Component: BenchmarkPage },
    ],
  },
  { path: '*', element: <Navigate to="/" replace /> },
]

export const router = createBrowserRouter(routes)
