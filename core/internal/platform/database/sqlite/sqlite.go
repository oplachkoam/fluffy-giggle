package sqlitedb

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Config struct {
	Path string `env:"PATH" env-required:"true" validate:"required"`
}

// Open creates the file and its directory when missing. WAL lets the SSE readers
// and the pipeline writers coexist; busy_timeout turns lock contention into a wait.
func Open(cfg Config) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o750); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}
	dsn := "file:" + cfg.Path +
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(4)
	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}

// Migrate applies every *.sql file from files in name order once, recording each in
// schema_migrations. Applied files are never run again.
func Migrate(ctx context.Context, db *sql.DB, files fs.FS) error {
	_, err := db.ExecContext(ctx, `create table if not exists schema_migrations (
		name text primary key, applied_at text not null)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		var applied int
		err := db.QueryRowContext(ctx,
			`select count(*) from schema_migrations where name = ?`, name).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied > 0 {
			continue
		}
		body, err := fs.ReadFile(files, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := apply(ctx, db, name, string(body)); err != nil {
			return err
		}
		slog.Info("migration applied", slog.String("name", name))
	}

	return nil
}

func apply(ctx context.Context, db *sql.DB, name, body string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, body); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	_, err = tx.ExecContext(ctx,
		`insert into schema_migrations (name, applied_at) values (?, ?)`,
		name, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}

	return nil
}

// Clean strips a leading "file:" so error messages show a path, not a DSN.
func Clean(path string) string {
	return strings.TrimPrefix(path, "file:")
}
