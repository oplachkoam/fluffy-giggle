// Package resolvereview applies the operator's decision to a parked run.
package resolvereview

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/errs"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
)

type Input struct {
	RunID    uuid.UUID
	Decision string
}

type Output struct {
	Run *entity.Run
}

type Usecase struct {
	runs   port.RunRepository
	steps  port.StepRepository
	queue  port.RunQueue
	events port.EventBus
	now    func() time.Time
}

func New(
	runs port.RunRepository,
	steps port.StepRepository,
	queue port.RunQueue,
	events port.EventBus,
) *Usecase {
	return &Usecase{runs: runs, steps: steps, queue: queue, events: events, now: time.Now}
}

func (u *Usecase) Execute(ctx context.Context, in Input) (*Output, error) {
	decision, err := vo.ParseReviewDecision(in.Decision)
	if err != nil {
		return nil, errs.NewValidationError(fmt.Errorf("invalid field: decision: %w", err))
	}
	run, err := u.runs.Get(ctx, in.RunID)
	if err != nil {
		return nil, fmt.Errorf("failed to load run: %w", err)
	}
	if run == nil {
		return nil, errs.ErrRunNotFound
	}
	findings := run.Findings()
	if err := run.Resolve(decision, u.now()); err != nil {
		slog.DebugContext(ctx, "review rejected", slog.String("run_id", run.ID().String()),
			slog.String("error", err.Error()))
		return nil, errs.ErrRunNotAwaitingReview
	}

	existing, err := u.steps.ListByRun(ctx, run.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to load steps: %w", err)
	}
	step, err := entity.NewStep(&entity.StepState{
		RunID:     run.ID(),
		Seq:       len(existing) + 1,
		Kind:      vo.StepReview,
		Side:      vo.SideSystem,
		Status:    vo.StepOK,
		StartedAt: u.now(),
		Input:     mustJSON(map[string]any{"findings": findings}),
		Output:    mustJSON(map[string]any{"decision": decision}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build review step: %w", err)
	}
	if err := u.steps.Create(ctx, step); err != nil {
		return nil, fmt.Errorf("failed to store review step: %w", err)
	}
	if err := u.runs.Update(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to store run: %w", err)
	}
	u.events.Publish(port.Event{Name: "step", Data: step})
	u.events.Publish(port.Event{Name: "run", Data: run})
	u.queue.Enqueue(run.ID())

	slog.InfoContext(ctx, "review resolved",
		slog.String("run_id", run.ID().String()), slog.String("decision", string(decision)))

	return &Output{Run: run}, nil
}
