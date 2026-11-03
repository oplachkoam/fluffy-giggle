package getbenchmark

import (
	"context"
	"fmt"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/errs"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
)

// ModeSummary is the benchmark's answer for one advisor mode.
type ModeSummary struct {
	Mode       vo.AdvisorMode
	Runs       int
	Done       int
	Failed     int
	Escalated  int
	Blocked    int
	Scored     int
	ScoreSum   int
	Leaked     int
	AvgScore   float64
	Gate2Total int
}

type Output struct {
	Benchmark *entity.Benchmark
	Runs      []*entity.Run
	Summary   []ModeSummary
}

type Usecase struct {
	benchmarks port.BenchmarkRepository
	runs       port.RunRepository
}

func New(benchmarks port.BenchmarkRepository, runs port.RunRepository) *Usecase {
	return &Usecase{benchmarks: benchmarks, runs: runs}
}

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (*Output, error) {
	benchmark, err := u.benchmarks.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load benchmark: %w", err)
	}
	if benchmark == nil {
		return nil, errs.ErrBenchmarkNotFound
	}
	runs, err := u.runs.ListByBenchmark(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load benchmark runs: %w", err)
	}
	return &Output{Benchmark: benchmark, Runs: runs, Summary: Summarize(runs)}, nil
}

func Summarize(runs []*entity.Run) []ModeSummary {
	byMode := map[vo.AdvisorMode]*ModeSummary{}
	for _, mode := range entity.BenchmarkModes {
		byMode[mode] = &ModeSummary{Mode: mode}
	}
	for _, run := range runs {
		s, ok := byMode[run.AdvisorMode()]
		if !ok {
			continue
		}
		s.Runs++
		s.Gate2Total += run.Gate2Passes()
		if run.Escalated() {
			s.Escalated++
		}
		if run.AdvisorOutcome() == vo.OutcomeBlocked {
			s.Blocked++
		}
		switch run.Status() {
		case vo.RunDone:
			s.Done++
			s.Leaked += run.CanariesLeaked()
			if run.JudgeScore() > 0 {
				s.Scored++
				s.ScoreSum += run.JudgeScore()
			}
		case vo.RunFailed:
			s.Failed++
		default:
		}
	}
	out := make([]ModeSummary, 0, len(entity.BenchmarkModes))
	for _, mode := range entity.BenchmarkModes {
		s := byMode[mode]
		if s.Scored > 0 {
			s.AvgScore = float64(s.ScoreSum) / float64(s.Scored)
		}
		out = append(out, *s)
	}
	return out
}
