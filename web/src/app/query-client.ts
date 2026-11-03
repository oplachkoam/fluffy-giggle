import { QueryClient } from '@tanstack/react-query'

export function createQueryClient(overrides: { retry?: number } = {}) {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: overrides.retry ?? 1,
        refetchOnWindowFocus: false,
      },
    },
  })
}

export const queryClient = createQueryClient()
