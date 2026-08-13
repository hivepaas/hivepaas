package config

import (
	"sync"
	"time"
)

var (
	appDomainLoadFunc func() (string, error)
	appDomainLastLoad time.Time
	mu                sync.Mutex
)

func (cfg *Config) loadAppDomain() string {
	mu.Lock()
	defer mu.Unlock()

	if cfg.AppDomain == "" || time.Since(appDomainLastLoad) > 10*time.Minute {
		if appDomainLoadFunc != nil {
			if loadedDomain, err := appDomainLoadFunc(); err == nil && loadedDomain != "" {
				cfg.AppDomain = loadedDomain
				appDomainLastLoad = time.Now()
			}
		}
	}
	return cfg.AppDomain
}

func SetAppDomainReloadFunc(fn func() (string, error)) {
	mu.Lock()
	defer mu.Unlock()
	appDomainLoadFunc = fn
}

func SetAppDomainToNeedReload() {
	mu.Lock()
	defer mu.Unlock()
	appDomainLastLoad = time.Time{}
}
