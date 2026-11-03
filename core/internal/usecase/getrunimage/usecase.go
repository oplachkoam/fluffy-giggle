package getrunimage

import (
	"context"
	"fmt"
	"uuid"

	"fluffy-giggle/core/internal/domain/errs"
	"fluffy-giggle/core/internal/domain/port"
)

type Output struct {
	Data []byte
	MIME string
}

type Usecase struct {
	runs port.RunRepository
}

func New(runs port.RunRepository) *Usecase {
	return &Usecase{runs: runs}
}

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (*Output, error) {
	run, err := u.runs.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load run: %w", err)
	}
	if run == nil {
		return nil, errs.ErrRunNotFound
	}
	if !run.HasImage() {
		return nil, errs.ErrRunHasNoImage
	}
	return &Output{Data: run.Image(), MIME: run.ImageMIME()}, nil
}
