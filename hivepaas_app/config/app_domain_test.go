package config

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// The cache is package state, so each test has to start from a known point.
func resetAppDomainState() {
	appDomainMu.Lock()
	defer appDomainMu.Unlock()
	appDomainLoadFunc = nil
	appDomainSeeded = false
	appDomain = ""
	appDomainNextLoad = time.Time{}
}

// The bug this replaces: a load that legitimately returns no domain never moved the
// deadline, so every call ran the loader's query. That is the state a new
// installation is in until someone configures a domain.
func TestLoadAppDomain_EmptyResultIsCached(t *testing.T) {
	resetAppDomainState()
	calls := 0
	SetAppDomainReloadFunc(func() (string, error) {
		calls++
		return "", nil
	})

	cfg := &Config{}
	for range 100 {
		assert.Equal(t, "", cfg.loadAppDomain())
	}
	assert.Equal(t, 1, calls, "an empty result must be cached, not re-queried on every call")
}

func TestLoadAppDomain_SuccessIsCached(t *testing.T) {
	resetAppDomainState()
	calls := 0
	SetAppDomainReloadFunc(func() (string, error) {
		calls++
		return "app.example.com", nil
	})

	cfg := &Config{}
	for range 100 {
		assert.Equal(t, "app.example.com", cfg.loadAppDomain())
	}
	assert.Equal(t, 1, calls)
}

// The TTL has to be an upper bound, not a delay: changing the domain must take
// effect at once.
func TestLoadAppDomain_NeedReloadForcesAFreshLoad(t *testing.T) {
	resetAppDomainState()
	domain := "first.example.com"
	calls := 0
	SetAppDomainReloadFunc(func() (string, error) {
		calls++
		return domain, nil
	})

	cfg := &Config{}
	assert.Equal(t, "first.example.com", cfg.loadAppDomain())

	domain = "second.example.com"
	assert.Equal(t, "first.example.com", cfg.loadAppDomain(), "still inside the window")

	SetAppDomainToNeedReload()
	assert.Equal(t, "second.example.com", cfg.loadAppDomain())
	assert.Equal(t, 2, calls)
}

// A failure must keep serving what is known and retry on its own schedule, rather
// than either forgetting the domain or querying on every call.
func TestLoadAppDomain_ErrorKeepsValueAndBacksOff(t *testing.T) {
	resetAppDomainState()
	calls := 0
	SetAppDomainReloadFunc(func() (string, error) {
		calls++
		if calls == 1 {
			return "app.example.com", nil
		}
		return "", errors.New("database is down")
	})

	cfg := &Config{}
	assert.Equal(t, "app.example.com", cfg.loadAppDomain())

	SetAppDomainToNeedReload()
	assert.Equal(t, "app.example.com", cfg.loadAppDomain(), "the known domain must survive a failure")
	assert.Equal(t, 2, calls)

	for range 50 {
		cfg.loadAppDomain()
	}
	assert.Equal(t, 2, calls, "a failure must not be retried on every call")

	// The retry window is the short one, so the domain is not stuck for the full TTL.
	appDomainMu.Lock()
	retryIn := time.Until(appDomainNextLoad)
	appDomainMu.Unlock()
	assert.LessOrEqual(t, retryIn, appDomainErrorTTL)
	assert.Less(t, retryIn, appDomainTTL)
}

// Before the loader is registered there is nothing to load from, so that call must
// not consume the window - otherwise the first real load would wait it out.
func TestLoadAppDomain_NoLoaderDoesNotStartTheWindow(t *testing.T) {
	resetAppDomainState()
	cfg := &Config{AppDomain: "seed.example.com"}

	assert.Equal(t, "seed.example.com", cfg.loadAppDomain())
	appDomainMu.Lock()
	assert.True(t, appDomainNextLoad.IsZero())
	appDomainMu.Unlock()

	calls := 0
	SetAppDomainReloadFunc(func() (string, error) {
		calls++
		return "loaded.example.com", nil
	})
	assert.Equal(t, "loaded.example.com", cfg.loadAppDomain())
	assert.Equal(t, 1, calls, "the first call after registration must load")
}

// Config.AppDomain is read from other goroutines without this lock, so the refreshed
// value must not be written back into it.
func TestLoadAppDomain_DoesNotWriteToConfig(t *testing.T) {
	resetAppDomainState()
	SetAppDomainReloadFunc(func() (string, error) { return "loaded.example.com", nil })

	cfg := &Config{AppDomain: "seed.example.com"}
	assert.Equal(t, "loaded.example.com", cfg.loadAppDomain())
	assert.Equal(t, "seed.example.com", cfg.AppDomain)
}

func TestLoadAppDomain_ConcurrentUseIsSafe(t *testing.T) {
	resetAppDomainState()
	SetAppDomainReloadFunc(func() (string, error) { return "app.example.com", nil })

	cfg := &Config{}
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				_ = cfg.loadAppDomain()
				SetAppDomainToNeedReload()
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, "app.example.com", cfg.loadAppDomain())
}
