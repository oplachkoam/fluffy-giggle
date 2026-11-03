package bootstrap

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"fluffy-giggle/core/internal/config"
)

// TestGraph proves the fx graph resolves without opening a database or a port.
func TestGraph(t *testing.T) {
	t.Parallel()

	model := config.Model{BaseURL: "http://localhost", APIKey: "k", Model: "m", Timeout: time.Second}
	cfg := &config.Config{
		App:      config.App{Name: "test", Env: "local"},
		Auth:     config.Auth{Login: "operator", Password: "password-1"},
		LLM:      config.LLM{Local: model, Vision: model, Cloud: model},
		Pipeline: config.Pipeline{Concurrency: 1, Gate2MaxPasses: 3},
	}
	cfg.Logger.Level, cfg.Logger.Format = "error", "text"
	cfg.SQLite.Path = t.TempDir() + "/app.db"

	err := fx.ValidateApp(fx.Supply(cfg), Module, fx.Invoke(func(*http.Server) {}))
	require.NoError(t, err)
}
