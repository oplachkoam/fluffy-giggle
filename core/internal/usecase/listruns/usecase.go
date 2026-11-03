package listruns

import (
	"context"
	"fmt"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
)

const PageSize = 50

type Input struct {
	// Before is the id of the last run on the previous page; Nil for the first.
	Before uuid.UUID
}

type Output struct {
	Runs []*entity.Run
	// Next is the cursor for the next page, Nil when this was the last.
	Next uuid.UUID
}

type Usecase struct {
	runs port.RunRepository
}

func New(runs port.RunRepository) *Usecase {
	return &Usecase{runs: runs}
}

func (u *Usecase) Execute(ctx context.Context, in Input) (*Output, error) {
	runs, err := u.runs.List(ctx, PageSize+1, in.Before)
	if err != nil {
		return nil, fmt.Errorf("failed to list runs: %w", err)
	}
	out := &Output{Runs: runs, Next: uuid.Nil()}
	if len(runs) > PageSize {
		out.Runs = runs[:PageSize]
		out.Next = runs[PageSize-1].ID()
	}
	return out, nil
}
