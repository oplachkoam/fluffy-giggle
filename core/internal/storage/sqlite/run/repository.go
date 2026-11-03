package run

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
	db "fluffy-giggle/core/internal/storage/sqlite/sqlc"
)

type Repository struct {
	queries *db.Queries
}

var _ port.RunRepository = (*Repository)(nil)

func New(conn *sql.DB) *Repository {
	return &Repository{queries: db.New(conn)}
}

func (r *Repository) Create(ctx context.Context, run *entity.Run) error {
	card := ""
	if c := run.Card(); c != nil {
		card = encode(c)
	}
	err := r.queries.CreateRun(ctx, db.CreateRunParams{
		ID:              run.ID().String(),
		CreatedAt:       formatTime(run.CreatedAt()),
		UpdatedAt:       formatTime(run.UpdatedAt()),
		Status:          string(run.Status()),
		Phase:           string(run.Phase()),
		AdvisorMode:     string(run.AdvisorMode()),
		AdvisorOutcome:  string(run.AdvisorOutcome()),
		ReviewDecision:  string(run.ReviewDecision()),
		Source:          string(run.Source()),
		CaseID:          run.CaseID(),
		BenchmarkID:     benchmarkIDString(run.BenchmarkID()),
		InputText:       run.InputText(),
		Image:           run.Image(),
		ImageMime:       run.ImageMIME(),
		CanariesJson:    encode(run.Canaries()),
		AnonymizedText:  run.Anonymized(),
		VaultJson:       encode(run.Vault().Entries()),
		FindingsJson:    encode(run.Findings()),
		CardJson:        card,
		Draft:           run.Draft(),
		AdviceJson:      encode(run.Advice()),
		FinalAnonymized: run.FinalAnonymous(),
		FinalText:       run.FinalText(),
		JudgeScore:      scorePointer(run.JudgeScore()),
		JudgeReason:     run.JudgeReason(),
		Gate2Passes:     int64(run.Gate2Passes()),
		Escalated:       boolInt(run.Escalated()),
		CanariesLeaked:  int64(run.CanariesLeaked()),
		Error:           run.Error(),
	})
	if err != nil {
		return fmt.Errorf("insert run: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, run *entity.Run) error {
	card := ""
	if c := run.Card(); c != nil {
		card = encode(c)
	}
	err := r.queries.UpdateRun(ctx, db.UpdateRunParams{
		UpdatedAt:       formatTime(run.UpdatedAt()),
		Status:          string(run.Status()),
		Phase:           string(run.Phase()),
		AdvisorOutcome:  string(run.AdvisorOutcome()),
		ReviewDecision:  string(run.ReviewDecision()),
		InputText:       run.InputText(),
		AnonymizedText:  run.Anonymized(),
		VaultJson:       encode(run.Vault().Entries()),
		FindingsJson:    encode(run.Findings()),
		CardJson:        card,
		Draft:           run.Draft(),
		AdviceJson:      encode(run.Advice()),
		FinalAnonymized: run.FinalAnonymous(),
		FinalText:       run.FinalText(),
		JudgeScore:      scorePointer(run.JudgeScore()),
		JudgeReason:     run.JudgeReason(),
		Gate2Passes:     int64(run.Gate2Passes()),
		Escalated:       boolInt(run.Escalated()),
		CanariesLeaked:  int64(run.CanariesLeaked()),
		Error:           run.Error(),
		ID:              run.ID().String(),
	})
	if err != nil {
		return fmt.Errorf("update run: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*entity.Run, error) {
	row, err := r.queries.GetRun(ctx, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select run: %w", err)
	}
	return toEntity(fromRun(row))
}

func (r *Repository) List(ctx context.Context, limit int, before uuid.UUID) ([]*entity.Run, error) {
	// UUIDs are hex and dashes, so "~" sorts after every id: the first page has no cursor.
	cursor := "~"
	if before != uuid.Nil() {
		cursor = before.String()
	}
	rows, err := r.queries.ListRuns(ctx, db.ListRunsParams{ID: cursor, Limit: int64(limit)})
	if err != nil {
		return nil, fmt.Errorf("select runs: %w", err)
	}
	runs := make([]*entity.Run, 0, len(rows))
	for _, row := range rows {
		run, err := toEntity(fromListRow(row))
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

func (r *Repository) ListByBenchmark(ctx context.Context, benchmarkID uuid.UUID) ([]*entity.Run, error) {
	rows, err := r.queries.ListRunsByBenchmark(ctx, benchmarkID.String())
	if err != nil {
		return nil, fmt.Errorf("select benchmark runs: %w", err)
	}
	runs := make([]*entity.Run, 0, len(rows))
	for _, row := range rows {
		run, err := toEntity(fromListRunsByBenchmarkRow(row))
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}
