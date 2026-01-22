package shared

import "time"

type RedisConfig struct {
	URI       string       `yaml:"uri" env:"REDIS_URI" env-default:"localhost:6379"`
	PubStream StreamConfig `yaml:"pub_stream"`
	SubStream StreamConfig `yaml:"sub_stream"`

	Cache CacheConfig `yaml:"cache"`
}

type StreamConfig struct {
	ID string `yaml:"id"`

	MaxBacklog   int64         `yaml:"max_backlog"`
	UseDelApprox bool          `yaml:"use_del_approx"`
	ReadCount    int64         `yaml:"read_count"`
	BlockTime    time.Duration `yaml:"block_time"`

	Group GroupConfig `yaml:"group"`
}

type GroupConfig struct {
	ID                  string `yaml:"id"`
	ConsumerPrimarilyID string `yaml:"consumer_primarily_id"`
}

type CacheConfig struct {
	TTL time.Duration `yaml:"ttl" env:"CACHE_TTL" env-default:"5m"`
}

type BackoffConfig struct {
	Min          time.Duration `yaml:"min" env:"BACKOFF_MIN" env-default:"1s"`
	Max          time.Duration `yaml:"max" env:"BACKOFF_MAX" env-default:"60s"`
	Factor       float64       `yaml:"factor" env:"BACKOFF_FACTOR" env-default:"2"`
	PollInterval time.Duration `yaml:"poll_interval" env:"POLL_INTERVAL" env-default:"1s"`
	MaxRetries   int           `yaml:"max_retries" env:"MAX_RETRIES" env-default:"5"`
}

type OtelConfig struct {
	URI string `yaml:"uri" env:"OTEL_URI"`
}
