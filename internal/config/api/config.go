package api

import (
	"ai-orchestrator/internal/config/shared"
	"fmt"
)

type Config struct {
	App      AppConfig          `yaml:"app"`
	Postgres PostgresConfig     `yaml:"postgres"`
	Redis    shared.RedisConfig `yaml:"redis"`
	OTEL     shared.OtelConfig  `yaml:"otel"`
}

type AppConfig struct {
	ID            string `yaml:"id" env:"APP_ID" env-default:"api"`
	Port          string `yaml:"port" env:"APP_PORT" env-default:"8080"`
	Environment   string `yaml:"env" env:"APP_ENV" env-default:"development"`
	MigrationsDir string `yaml:"migrations_dir" env:"MIGRATIONS_DIR"`

	Backoff shared.BackoffConfig `yaml:"backoff"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
	Password string `yaml:"-" env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"db_name" env:"POSTGRES_DB"`
}

func (pc PostgresConfig) GetConnectionString() string {
	asdlfkj
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", pc.User, pc.Password, pc.Host, pc.Port, pc.DBName)
}
