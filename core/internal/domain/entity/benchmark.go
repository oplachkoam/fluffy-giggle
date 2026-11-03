package entity

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/vo"
)

type BenchmarkStatus string

const (
	BenchmarkRunning BenchmarkStatus = "running"
	BenchmarkDone    BenchmarkStatus = "done"
)

func ParseBenchmarkStatus(s string) (BenchmarkStatus, error) {
	switch BenchmarkStatus(s) {
	case BenchmarkRunning, BenchmarkDone:
		return BenchmarkStatus(s), nil
	}
	return "", fmt.Errorf("unknown benchmark status %q", s)
}

// Benchmark is one pass over the fixture in every advisor mode; it only counts
// finished runs, the numbers come from the runs themselves.
type Benchmark struct {
	id        uuid.UUID
	createdAt time.Time
	status    BenchmarkStatus
	total     int
	finished  int
}

type BenchmarkState struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Status    BenchmarkStatus
	Total     int
	Finished  int
}

var BenchmarkModes = []vo.AdvisorMode{vo.AdvisorOff, vo.AdvisorOn, vo.AdvisorPlacebo}

func NewBenchmark(total int, now time.Time) (*Benchmark, error) {
	return RestoreBenchmark(&BenchmarkState{
		ID:        uuid.NewV7(),
		CreatedAt: now,
		Status:    BenchmarkRunning,
		Total:     total,
	})
}

func RestoreBenchmark(state *BenchmarkState) (*Benchmark, error) {
	if state.ID == uuid.Nil() {
		return nil, errors.New("benchmark id is empty")
	}
	if state.CreatedAt.IsZero() {
		return nil, errors.New("benchmark creation time is empty")
	}
	if _, err := ParseBenchmarkStatus(string(state.Status)); err != nil {
		return nil, err
	}
	if state.Total < 1 {
		return nil, errors.New("benchmark has no runs")
	}
	if state.Finished < 0 || state.Finished > state.Total {
		return nil, errors.New("benchmark finished count is out of range")
	}

	return &Benchmark{
		id:        state.ID,
		createdAt: state.CreatedAt.UTC(),
		status:    state.Status,
		total:     state.Total,
		finished:  state.Finished,
	}, nil
}

func (b *Benchmark) ID() uuid.UUID           { return b.id }
func (b *Benchmark) CreatedAt() time.Time    { return b.createdAt }
func (b *Benchmark) Status() BenchmarkStatus { return b.status }
func (b *Benchmark) Total() int              { return b.total }
func (b *Benchmark) Finished() int           { return b.finished }

// Advance counts one more finished or failed run.
func (b *Benchmark) Advance() {
	if b.finished < b.total {
		b.finished++
	}
	if b.finished == b.total {
		b.status = BenchmarkDone
	}
}
