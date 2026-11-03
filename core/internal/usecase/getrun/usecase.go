package getrun

import (
	"context"
	"fmt"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/errs"
	"fluffy-giggle/core/internal/domain/port"
)

type Output struct {
	Run   *entity.Run
	Steps []*entity.Step
}

type Usecase struct {
	runs  port.RunRepository
	steps port.StepRepository
}

func New(runs port.RunRepository, steps port.StepRepository) *Usecase {
	return &Usecase{runs: runs, steps: steps}
}

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (*Output, error) {
	run, err := u.runs.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load run: %w", err)
	}
	if run == nil {
		return nil, errs.ErrRunNotFound
	}
	steps, err := u.steps.ListByRun(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load steps: %w", err)
	}
	return &Output{Run: run, Steps: steps}, nil
}
