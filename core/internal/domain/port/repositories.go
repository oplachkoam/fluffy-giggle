package port

import (
	"context"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
)

type RunRepository interface {
	Create(ctx context.Context, run *entity.Run) error
	Update(ctx context.Context, run *entity.Run) error
	// Get returns (nil, nil) when there is no such run.
	Get(ctx context.Context, id uuid.UUID) (*entity.Run, error)
	// List returns runs newest first; before is the id to continue after, empty for
	// the first page.
	List(ctx context.Context, limit int, before uuid.UUID) ([]*entity.Run, error)
	ListByBenchmark(ctx context.Context, benchmarkID uuid.UUID) ([]*entity.Run, error)
}

type StepRepository interface {
	Create(ctx context.Context, step *entity.Step) error
	// ListByRun returns steps in order.
	ListByRun(ctx context.Context, runID uuid.UUID) ([]*entity.Step, error)
}

type BenchmarkRepository interface {
	Create(ctx context.Context, benchmark *entity.Benchmark) error
	Update(ctx context.Context, benchmark *entity.Benchmark) error
	// Get returns (nil, nil) when there is no such benchmark.
	Get(ctx context.Context, id uuid.UUID) (*entity.Benchmark, error)
	// List returns benchmarks newest first.
	List(ctx context.Context) ([]*entity.Benchmark, error)
}

type CaseCatalog interface {
	List() []*entity.Case
	// Get returns nil when there is no such case.
	Get(id string) *entity.Case
}
