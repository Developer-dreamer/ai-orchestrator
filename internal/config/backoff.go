package config

import (
	"time"
)

type Backoff struct {
	Min          time.Duration
	Max          time.Duration
	Factor       float64
	PollInterval time.Duration
}
