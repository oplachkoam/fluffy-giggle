package listbenchmarks

import (
	"context"
	"fmt"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
)

type Usecase struct {
	benchmarks port.BenchmarkRepository
}

func New(benchmarks port.BenchmarkRepository) *Usecase {
	return &Usecase{benchmarks: benchmarks}
}

func (u *Usecase) Execute(ctx context.Context) ([]*entity.Benchmark, error) {
	out, err := u.benchmarks.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmarks: %w", err)
	}
	return out, nil
}
