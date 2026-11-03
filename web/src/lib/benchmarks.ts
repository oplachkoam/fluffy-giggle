import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from '@/lib/api'
import { type AdvisorMode, type Run, type RunDto, toRun } from '@/lib/runs'

export type ModeSummary = {
  mode: AdvisorMode
  runs: number
  done: number
  failed: number
  escalated: number
  blocked: number
  scored: number
  avgScore: number | null
  leaked: number
  gate2Avg: number
}

export type Benchmark = {
  id: string
  createdAt: string
  status: 'running' | 'done'
  total: number
  finished: number
  summary: ModeSummary[]
  runs: Run[]
}

export type BenchmarkDto = {
  id: string
  created_at: string
  status: 'running' | 'done'
  total: number
  finished: number
  summary?: {
    mode: AdvisorMode
    runs: number
    done: number
    failed: number
    escalated: number
    blocked: number
    scored: number
    avg_score: number | null
    leaked: number
    gate2_avg: number
  }[]
  runs?: RunDto[]
}

export function toBenchmark(dto: BenchmarkDto): Benchmark {
  return {
    id: dto.id,
    createdAt: dto.created_at,
    status: dto.status,
    total: dto.total,
    finished: dto.finished,
    summary: (dto.summary ?? []).map((s) => ({
      mode: s.mode,
      runs: s.runs,
      done: s.done,
      failed: s.failed,
      escalated: s.escalated,
      blocked: s.blocked,
      scored: s.scored,
      avgScore: s.avg_score,
      leaked: s.leaked,
      gate2Avg: s.gate2_avg,
    })),
    runs: (dto.runs ?? []).map(toRun),
  }
}

export function useBenchmarks() {
  return useQuery({
    queryKey: ['benchmarks'],
    queryFn: async () => (await api<BenchmarkDto[]>('/benchmarks')).map(toBenchmark),
  })
}

export function useBenchmark(id: string | null) {
  return useQuery({
    queryKey: ['benchmark', id],
    queryFn: async () => toBenchmark(await api<BenchmarkDto>(`/benchmarks/${id}`)),
    enabled: id !== null,
    refetchInterval: (query) => (query.state.data?.status === 'running' ? 5000 : false),
  })
}

export function useStartBenchmark() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => toBenchmark(await api<BenchmarkDto>('/benchmarks', { method: 'POST' })),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['benchmarks'] })
    },
  })
}
