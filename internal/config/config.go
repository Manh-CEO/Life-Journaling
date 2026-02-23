package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration parsed from environment variables.
type Config struct {
	App      AppConfig
	DB       DBConfig
	Supabase SupabaseConfig
	Resend   ResendConfig
	Gemini   GeminiConfig
	QStash   QStashConfig
	CF       CFConfig
}

// AppConfig holds HTTP server configuration.
type AppConfig struct {
	Port string `envconfig:"APP_PORT" default:"8080"`
	Env  string `envconfig:"APP_ENV" default:"development"`
}

// DBConfig holds PostgreSQL connection configuration.
type DBConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"postgres"`
	Name     string `envconfig:"DB_NAME" default:"life_journaling"`
	SSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`
}

// DSN returns the PostgreSQL connection string.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// SupabaseConfig holds Supabase authentication configuration.
type SupabaseConfig struct {
	URL       string `envconfig:"SUPABASE_URL" required:"true"`
	JWTSecret string `envconfig:"SUPABASE_JWT_SECRET" required:"true"`
}

// ResendConfig holds email provider configuration.
type ResendConfig struct {
	APIKey    string `envconfig:"RESEND_API_KEY" required:"true"`
	FromEmail string `envconfig:"RESEND_FROM_EMAIL" required:"true"`
}

// GeminiConfig holds LLM provider configuration.
type GeminiConfig struct {
	APIKey string `envconfig:"GEMINI_API_KEY" required:"true"`
}

// QStashConfig holds cron trigger configuration.
type QStashConfig struct {
	SigningKey string `envconfig:"QSTASH_SIGNING_KEY" required:"true"`
}

// CFConfig holds Cloudflare webhook configuration.
type CFConfig struct {
	WebhookSecret string `envconfig:"CF_WEBHOOK_SECRET" required:"true"`
}

// Load parses environment variables into the Config struct.
func Load() (Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg.App); err != nil {
		return Config{}, fmt.Errorf("loading app config: %w", err)
	}
	if err := envconfig.Process("", &cfg.DB); err != nil {
		return Config{}, fmt.Errorf("loading db config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Supabase); err != nil {
		return Config{}, fmt.Errorf("loading supabase config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Resend); err != nil {
		return Config{}, fmt.Errorf("loading resend config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Gemini); err != nil {
		return Config{}, fmt.Errorf("loading gemini config: %w", err)
	}
	if err := envconfig.Process("", &cfg.QStash); err != nil {
		return Config{}, fmt.Errorf("loading qstash config: %w", err)
	}
	if err := envconfig.Process("", &cfg.CF); err != nil {
		return Config{}, fmt.Errorf("loading cloudflare config: %w", err)
	}

	return cfg, nil
}
