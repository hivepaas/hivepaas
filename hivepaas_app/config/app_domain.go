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
	appDomain := cfg.AppDomain
	if appDomain == "" || time.Since(appDomainLastLoad) > 10*time.Minute {
		mu.Lock()
		defer mu.Unlock()
		if appDomainLoadFunc != nil {
			appDomain, _ = appDomainLoadFunc()
		}
		if appDomain != "" {
			cfg.AppDomain = appDomain
			appDomainLastLoad = time.Now()
		}
	}
	return appDomain
}

func SetAppDomainReloadFunc(fn func() (string, error)) {
	mu.Lock()
	defer mu.Unlock()
	appDomainLoadFunc = fn
}

func SetAppDomainToNeedReload() {
	mu.Lock()
	defer mu.Unlock()
	Current.AppDomain = ""
}
