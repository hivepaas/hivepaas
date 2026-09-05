package config

import (
	"sync"
	"time"
)

const (
	// appDomainTTL is how long a loaded domain is trusted. A change does not have to
	// wait it out: SetAppDomainToNeedReload forces the next call to reload.
	appDomainTTL = 10 * time.Minute

	// appDomainErrorTTL applies when the load failed. It is short, because until it
	// succeeds every caller building a URL works from a stale domain - but it is not
	// zero, because retrying on every call turns a database blip into a query storm.
	appDomainErrorTTL = 30 * time.Second
)

var (
	appDomainMu       sync.Mutex
	appDomainLoadFunc func() (string, error)
	appDomainSeeded   bool
	// appDomain is the cached domain. It lives here rather than on Config because
	// Config.AppDomain is read from other goroutines without this lock, so writing
	// the refreshed value there would be a data race.
	appDomain string
	// appDomainNextLoad is when the cached value stops being trusted. Storing the
	// deadline rather than the last load is what lets a failure be retried on a
	// different schedule from a success.
	appDomainNextLoad time.Time
)

// loadAppDomain returns the app domain, refreshing it from the database when the
// cached one has expired.
//
// Every outcome moves the deadline, including a load that legitimately comes back
// empty. Only advancing it on a non-empty result meant that an installation with no
// domain configured - the state every new installation starts in - ran the loader's
// two-join query on every single call, serialized behind this lock.
func (cfg *Config) loadAppDomain() string {
	appDomainMu.Lock()
	defer appDomainMu.Unlock()

	if !appDomainSeeded {
		appDomain = cfg.AppDomain
		appDomainSeeded = true
	}
	if time.Now().Before(appDomainNextLoad) {
		return appDomain
	}
	if appDomainLoadFunc == nil {
		// Registered during startup. Leaving the deadline alone means the first call
		// after registration loads, instead of waiting out a window we set before
		// there was anything to load from.
		return appDomain
	}

	loadedDomain, err := appDomainLoadFunc()
	if err != nil {
		appDomainNextLoad = time.Now().Add(appDomainErrorTTL)
		return appDomain
	}
	// An empty result keeps the previous domain: it is what the config file or the
	// environment supplied, and serving "https://" to callers would be worse than
	// serving something out of date.
	if loadedDomain != "" {
		appDomain = loadedDomain
	}
	appDomainNextLoad = time.Now().Add(appDomainTTL)

	return appDomain
}

func SetAppDomainReloadFunc(fn func() (string, error)) {
	appDomainMu.Lock()
	defer appDomainMu.Unlock()
	appDomainLoadFunc = fn
}

// SetAppDomainToNeedReload expires the cached domain so the next read reloads it.
// It is what makes the TTL above an upper bound rather than a delay: a domain change
// takes effect immediately.
func SetAppDomainToNeedReload() {
	appDomainMu.Lock()
	defer appDomainMu.Unlock()
	appDomainNextLoad = time.Time{}
}
