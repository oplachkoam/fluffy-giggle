package benchmark

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
	db "fluffy-giggle/core/internal/storage/sqlite/sqlc"
)

type Repository struct {
	queries *db.Queries
}

var _ port.BenchmarkRepository = (*Repository)(nil)

func New(conn *sql.DB) *Repository {
	return &Repository{queries: db.New(conn)}
}

func (r *Repository) Create(ctx context.Context, b *entity.Benchmark) error {
	err := r.queries.CreateBenchmark(ctx, db.CreateBenchmarkParams{
		ID:        b.ID().String(),
		CreatedAt: b.CreatedAt().UTC().Format(time.RFC3339Nano),
		Status:    string(b.Status()),
		Total:     int64(b.Total()),
		Finished:  int64(b.Finished()),
	})
	if err != nil {
		return fmt.Errorf("insert benchmark: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, b *entity.Benchmark) error {
	err := r.queries.UpdateBenchmark(ctx, db.UpdateBenchmarkParams{
		Status:   string(b.Status()),
		Finished: int64(b.Finished()),
		ID:       b.ID().String(),
	})
	if err != nil {
		return fmt.Errorf("update benchmark: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*entity.Benchmark, error) {
	row, err := r.queries.GetBenchmark(ctx, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select benchmark: %w", err)
	}
	return toEntity(row)
}

func (r *Repository) List(ctx context.Context) ([]*entity.Benchmark, error) {
	rows, err := r.queries.ListBenchmarks(ctx)
	if err != nil {
		return nil, fmt.Errorf("select benchmarks: %w", err)
	}
	out := make([]*entity.Benchmark, 0, len(rows))
	for _, row := range rows {
		b, err := toEntity(row)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func toEntity(row db.Benchmark) (*entity.Benchmark, error) {
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("parse benchmark id: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse benchmark time: %w", err)
	}
	return entity.RestoreBenchmark(&entity.BenchmarkState{
		ID:        id,
		CreatedAt: createdAt,
		Status:    entity.BenchmarkStatus(row.Status),
		Total:     int(row.Total),
		Finished:  int(row.Finished),
	})
}
