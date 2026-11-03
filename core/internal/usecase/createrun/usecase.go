// Package createrun accepts an input and hands the run to the workers.
package createrun

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/errs"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
)

type Usecase struct {
	runs      port.RunRepository
	cases     port.CaseCatalog
	queue     port.RunQueue
	events    port.EventBus
	validator *validator.Validate
	now       func() time.Time
}

func New(
	runs port.RunRepository, cases port.CaseCatalog, queue port.RunQueue, events port.EventBus,
	validator *validator.Validate,
) *Usecase {
	return &Usecase{
		runs:      runs,
		cases:     cases,
		queue:     queue,
		events:    events,
		validator: validator,
		now:       time.Now,
	}
}

func (u *Usecase) Execute(ctx context.Context, in Input) (*Output, error) {
	if err := u.validator.StructCtx(ctx, in); err != nil {
		return nil, errs.NewValidationError(err)
	}
	mode, _ := vo.ParseAdvisorMode(in.AdvisorMode)

	input := entity.NewRunInput{AdvisorMode: mode, Source: vo.SourceText}
	switch {
	case in.CaseID != "":
		c := u.cases.Get(in.CaseID)
		if c == nil {
			return nil, errs.ErrCaseNotFound
		}
		input.CaseID, input.Canaries = c.ID(), c.Canaries()
		if c.HasImage() {
			input.Source, input.Image, input.ImageMIME = vo.SourceImage, c.Image(), c.ImageMIME()
		} else {
			input.Text = c.Text()
		}
	case len(in.Image) > 0:
		if len(in.Image) > entity.MaxImageBytes {
			return nil, errs.NewValidationError(errors.New("invalid field: image: larger than 10 MB"))
		}
		if !vo.IsImageMIME(in.ImageMIME) {
			return nil, errs.NewValidationError(errors.New("invalid field: image: only PNG and JPEG"))
		}
		input.Source, input.Image, input.ImageMIME = vo.SourceImage, in.Image, in.ImageMIME
	case strings.TrimSpace(in.Text) != "":
		input.Text = strings.TrimSpace(in.Text)
	default:
		return nil, errs.NewValidationError(
			errors.New("invalid field: text: text, image or case_id is required"),
		)
	}

	run, err := entity.NewRun(input, u.now())
	if err != nil {
		return nil, errs.NewValidationError(err)
	}
	if err := u.runs.Create(ctx, run); err != nil {
		msg := "failed to store run"
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	u.events.Publish(port.Event{Name: "run", Data: run})
	u.queue.Enqueue(run.ID())

	slog.InfoContext(ctx, "run created",
		slog.String("run_id", run.ID().String()),
		slog.String("source", string(run.Source())),
		slog.String("advisor_mode", string(run.AdvisorMode())),
	)

	return &Output{Run: run}, nil
}
