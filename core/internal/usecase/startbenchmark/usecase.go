// Package startbenchmark queues every fixture case in every advisor mode.
package startbenchmark

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/errs"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
)

type Output struct {
	Benchmark *entity.Benchmark
}

type Usecase struct {
	benchmarks port.BenchmarkRepository
	runs       port.RunRepository
	cases      port.CaseCatalog
	queue      port.RunQueue
	events     port.EventBus
	now        func() time.Time
}

func New(
	benchmarks port.BenchmarkRepository, runs port.RunRepository, cases port.CaseCatalog,
	queue port.RunQueue, events port.EventBus,
) *Usecase {
	return &Usecase{
		benchmarks: benchmarks,
		runs:       runs,
		cases:      cases,
		queue:      queue,
		events:     events,
		now:        time.Now,
	}
}

func (u *Usecase) Execute(ctx context.Context) (*Output, error) {
	existing, err := u.benchmarks.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmarks: %w", err)
	}
	for _, b := range existing {
		if b.Status() == entity.BenchmarkRunning {
			return nil, errs.ErrBenchmarkRunning
		}
	}

	cases := u.cases.List()
	benchmark, err := entity.NewBenchmark(len(cases)*len(entity.BenchmarkModes), u.now())
	if err != nil {
		return nil, fmt.Errorf("failed to build benchmark: %w", err)
	}
	if err := u.benchmarks.Create(ctx, benchmark); err != nil {
		return nil, fmt.Errorf("failed to store benchmark: %w", err)
	}

	for _, mode := range entity.BenchmarkModes {
		for _, c := range cases {
			input := entity.NewRunInput{
				Source: vo.SourceText, AdvisorMode: mode, Text: c.Text(),
				CaseID: c.ID(), Canaries: c.Canaries(), BenchmarkID: benchmark.ID(),
			}
			if c.HasImage() {
				input.Source, input.Image, input.ImageMIME, input.Text = vo.SourceImage, c.Image(), c.ImageMIME(), ""
			}
			run, err := entity.NewRun(input, u.now())
			if err != nil {
				return nil, fmt.Errorf("failed to build run: %w", err)
			}
			if err := u.runs.Create(ctx, run); err != nil {
				return nil, fmt.Errorf("failed to store run: %w", err)
			}
			u.queue.Enqueue(run.ID())
		}
	}
	u.events.Publish(port.Event{Name: "benchmark", Data: benchmark})

	slog.InfoContext(ctx, "benchmark started",
		slog.String("benchmark_id", benchmark.ID().String()), slog.Int("runs", benchmark.Total()))

	return &Output{Benchmark: benchmark}, nil
}
