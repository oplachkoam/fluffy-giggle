package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"

	sqlitedb "fluffy-giggle/core/internal/platform/database/sqlite"
	httpserver "fluffy-giggle/core/internal/platform/http/server"
	"fluffy-giggle/core/internal/platform/observability/logger"
)

type Config struct {
	App        App               `env-prefix:"APP_"`
	Auth       Auth              `env-prefix:"AUTH_"`
	Logger     logger.Config     `env-prefix:"LOGGER_"`
	HTTPServer httpserver.Config `env-prefix:"HTTP_SERVER_"`
	SQLite     sqlitedb.Config   `env-prefix:"SQLITE_"`
	LLM        LLM               `env-prefix:"LLM_"`
	Pipeline   Pipeline          `env-prefix:"PIPELINE_"`
}

type App struct {
	Name string `env:"NAME" env-required:"true" validate:"required,max=64"`
	Env  string `env:"ENV"  env-required:"true" validate:"required,oneof=local dev stage prod"`
}

// Auth is the single operator account. The password is compared in constant time
// and never logged.
type Auth struct {
	Login    string `env:"LOGIN"    env-required:"true" validate:"required,max=64"`
	Password string `env:"PASSWORD" env-required:"true" validate:"required,min=8"`
}

// LLM names the three OpenAI-compatible endpoints. Local and Vision stand in for
// models inside the perimeter; Cloud is the adviser and the judge.
type LLM struct {
	Local  Model `env-prefix:"LOCAL_"`
	Vision Model `env-prefix:"VISION_"`
	Cloud  Model `env-prefix:"CLOUD_"`
}

type Model struct {
	BaseURL    string        `env:"BASE_URL"    env-required:"true" validate:"required,url"`
	APIKey     string        `env:"API_KEY"     env-required:"true" validate:"required"`
	Model      string        `env:"MODEL"       env-required:"true" validate:"required"`
	Timeout    time.Duration `env:"TIMEOUT"     env-required:"true"`
	MaxRetries int           `env:"MAX_RETRIES" env-required:"true" validate:"min=0,max=10"`
	// DisableThinking turns off the reasoning phase of models that have one (Qwen 3
	// and alike): a gate that returns a short JSON must not think for a minute.
	DisableThinking bool `env:"DISABLE_THINKING" env-default:"false"`
	// Temperature below zero means "do not send": some models accept only their default.
	Temperature float64 `env:"TEMPERATURE" env-default:"-1" validate:"max=2"`
}

type Pipeline struct {
	// Concurrency is how many runs execute at once; the benchmark queues them all.
	Concurrency int `env:"CONCURRENCY" env-required:"true" validate:"min=1,max=32"`
	// Gate2MaxPasses bounds the weak model's anonymization loop before escalation.
	Gate2MaxPasses int `env:"GATE2_MAX_PASSES" env-required:"true" validate:"min=1,max=10"`
}

func Load() (*Config, error) {
	cfg := new(Config)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}
