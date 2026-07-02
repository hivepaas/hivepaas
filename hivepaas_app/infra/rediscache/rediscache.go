package rediscache

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

type Client interface {
	redis.UniversalClient
	Close() error
}

func NewClient(cfg *config.Config) (Client, error) {
	options, err := redis.ParseURL(cfg.Cache.URL)
	if err != nil {
		return nil, fmt.Errorf("redis parse URL: %w", err)
	}

	options.PoolSize = cfg.Cache.PoolSize

	if cfg.Cache.ReadTimeout > 0 {
		options.ReadTimeout = time.Duration(cfg.Cache.ReadTimeout) * time.Second
	}
	if cfg.Cache.WriteTimeout > 0 {
		options.WriteTimeout = time.Duration(cfg.Cache.WriteTimeout) * time.Second
	}
	if cfg.Cache.MinIdleConns > 0 {
		options.MinIdleConns = cfg.Cache.MinIdleConns
	}
	if cfg.Cache.UseTLS {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
	}

	return redis.NewClient(options), nil //nolint:contextcheck
}
