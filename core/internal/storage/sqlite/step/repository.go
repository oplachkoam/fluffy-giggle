package step

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
	db "fluffy-giggle/core/internal/storage/sqlite/sqlc"
)

type Repository struct {
	queries *db.Queries
}

var _ port.StepRepository = (*Repository)(nil)

func New(conn *sql.DB) *Repository {
	return &Repository{queries: db.New(conn)}
}

func (r *Repository) Create(ctx context.Context, step *entity.Step) error {
	err := r.queries.CreateStep(ctx, db.CreateStepParams{
		ID:               step.ID().String(),
		RunID:            step.RunID().String(),
		Seq:              int64(step.Seq()),
		Kind:             string(step.Kind()),
		Side:             string(step.Side()),
		Status:           string(step.Status()),
		StartedAt:        step.StartedAt().UTC().Format(time.RFC3339Nano),
		FinishedAt:       step.FinishedAt().UTC().Format(time.RFC3339Nano),
		Model:            step.Model(),
		PromptTokens:     int64(step.PromptTokens()),
		CompletionTokens: int64(step.CompletionTokens()),
		InputJson:        string(step.Input()),
		OutputJson:       string(step.Output()),
		Error:            step.Error(),
	})
	if err != nil {
		return fmt.Errorf("insert step: %w", err)
	}
	return nil
}

func (r *Repository) ListByRun(ctx context.Context, runID uuid.UUID) ([]*entity.Step, error) {
	rows, err := r.queries.ListStepsByRun(ctx, runID.String())
	if err != nil {
		return nil, fmt.Errorf("select steps: %w", err)
	}
	steps := make([]*entity.Step, 0, len(rows))
	for _, row := range rows {
		step, err := toEntity(row)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func toEntity(row db.Step) (*entity.Step, error) {
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("parse step id: %w", err)
	}
	runID, err := uuid.Parse(row.RunID)
	if err != nil {
		return nil, fmt.Errorf("parse step run id: %w", err)
	}
	startedAt, err := time.Parse(time.RFC3339Nano, row.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("parse step start: %w", err)
	}
	finishedAt, err := time.Parse(time.RFC3339Nano, row.FinishedAt)
	if err != nil {
		return nil, fmt.Errorf("parse step finish: %w", err)
	}
	return entity.RestoreStep(&entity.StepState{
		ID:               id,
		RunID:            runID,
		Seq:              int(row.Seq),
		Kind:             vo.StepKind(row.Kind),
		Side:             vo.Side(row.Side),
		Status:           vo.StepStatus(row.Status),
		StartedAt:        startedAt,
		FinishedAt:       finishedAt,
		Model:            row.Model,
		PromptTokens:     int(row.PromptTokens),
		CompletionTokens: int(row.CompletionTokens),
		Input:            json.RawMessage(row.InputJson),
		Output:           json.RawMessage(row.OutputJson),
		Error:            row.Error,
	})
}
