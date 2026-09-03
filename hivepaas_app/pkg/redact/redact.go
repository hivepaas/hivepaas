package redact

import (
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
)

const (
	// concurrencyThreshold is the slice length limit at which we switch
	// from sequential execution to parallel execution to optimize performance.
	concurrencyThreshold = 500

	redactionMask = "********"
)

// Redactor handles the masking of sensitive secrets within text data.
// It is safe for concurrent use: AddSecrets can swap the replacer while other
// goroutines are redacting.
type Redactor struct {
	mu       sync.RWMutex
	secrets  []string
	replacer *strings.Replacer
}

// New creates a new Redactor initialized with the given secrets.
// Secrets are sorted by length in descending order to prevent partial matching.
func New(secrets []string) *Redactor {
	r := &Redactor{secrets: secrets}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	return r
}

// init rebuilds the replacer from the current secrets. The caller must hold r.mu.
func (r *Redactor) init() {
	sorted := make([]string, len(r.secrets))
	copy(sorted, r.secrets)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i]) > len(sorted[j])
	})
	pairs := make([]string, 0, len(sorted)*2) //nolint:mnd
	for _, s := range sorted {
		pairs = append(pairs, s, redactionMask)
	}
	r.replacer = strings.NewReplacer(pairs...)
}

// currentReplacer returns a snapshot of the replacer. Callers hold on to the
// returned value for the whole operation, so a concurrent AddSecrets swapping
// in a new replacer cannot race with them.
func (r *Redactor) currentReplacer() *strings.Replacer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.replacer
}

// String replaces secrets in a single string sequentially.
func (r *Redactor) String(text string) string {
	return r.currentReplacer().Replace(text)
}

func (r *Redactor) AddSecrets(secrets []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.secrets = append(r.secrets, secrets...)
	r.init()
}

// Slice replaces secrets in-place inside the given slice of strings,
// and returns the modified slice. It automatically chooses between
// sequential and parallel execution based on the slice size.
func (r *Redactor) Slice(logs []string) []string {
	numLogs := len(logs)
	replacer := r.currentReplacer()
	// Use sequential processing for small slices to avoid goroutine overhead.
	if numLogs < concurrencyThreshold {
		for idx, log := range logs {
			logs[idx] = replacer.Replace(log)
		}
		return logs
	}
	// Use a worker pool for larger slices to process in parallel.
	numWorkers := runtime.NumCPU()
	chunkSize := (numLogs + numWorkers - 1) / numWorkers
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if start >= numLogs {
			break
		}
		if end > numLogs {
			end = numLogs
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			defer safego.Recover("redact.sliceWorker")
			for idx := s; idx < e; idx++ {
				// Each worker processes a disjoint range of indices to prevent write contention.
				logs[idx] = replacer.Replace(logs[idx])
			}
		}(start, end)
	}
	wg.Wait()
	return logs
}
