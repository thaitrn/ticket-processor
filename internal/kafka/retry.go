package kafka

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

type RetryConfig struct {
	MaxRetries      uint64
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

func NewRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      5,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     2 * time.Second,
		Multiplier:      2,
	}
}

func (rc *RetryConfig) CreateBackOff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = rc.InitialInterval
	b.MaxInterval = rc.MaxInterval
	b.Multiplier = rc.Multiplier
	b.MaxElapsedTime = time.Duration(rc.MaxRetries) * rc.MaxInterval
	return b
}
