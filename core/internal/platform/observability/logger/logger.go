package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level  string `env:"LEVEL"  env-required:"true" validate:"required,oneof=debug info warn error"`
	Format string `env:"FORMAT" env-required:"true" validate:"required,oneof=text json"`
}

// Setup installs the default slog logger; code calls slog.InfoContext and friends.
func Setup(cfg Config) error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(cfg.Level))); err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))

	return nil
}
