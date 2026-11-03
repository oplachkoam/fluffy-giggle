import { useQuery } from '@tanstack/react-query'

import { api } from '@/lib/api'

export type Case = { id: string; title: string; hasImage: boolean; canaries: number }

type CaseDto = { id: string; title: string; has_image: boolean; canaries: number }

export function useCases() {
  return useQuery({
    queryKey: ['cases'],
    queryFn: async () =>
      (await api<CaseDto[]>('/cases')).map((c) => ({
        id: c.id,
        title: c.title,
        hasImage: c.has_image,
        canaries: c.canaries,
      })),
    staleTime: Infinity,
  })
}
