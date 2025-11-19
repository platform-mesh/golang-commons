package ratelimiter

import (
	"time"
)

type Config struct {
	StaticRequeueDelay        time.Duration
	StaticWindow              time.Duration
	ExponentialInitialBackoff time.Duration
	ExponentialMaxBackoff     time.Duration
}

var defaultConfig = Config{
	StaticRequeueDelay:        5 * time.Second,
	StaticWindow:              1 * time.Minute,
	ExponentialInitialBackoff: 5 * time.Second,
	ExponentialMaxBackoff:     2 * time.Minute,
}

type Option func(*Config)

func WithStaticWindow(d time.Duration) Option {
	return func(c *Config) {
		c.StaticWindow = d
	}
}

func WithRequeueDelay(d time.Duration) Option {
	return func(c *Config) {
		c.StaticRequeueDelay = d
	}
}

func WithExponentialInitialBackoff(d time.Duration) Option {
	return func(c *Config) {
		c.ExponentialInitialBackoff = d
	}
}

func WithExponentialMaxBackoff(d time.Duration) Option {
	return func(c *Config) {
		c.ExponentialMaxBackoff = d
	}
}

func NewConfig(options ...Option) Config {
	cfg := defaultConfig

	for _, option := range options {
		option(&cfg)
	}

	return cfg
}