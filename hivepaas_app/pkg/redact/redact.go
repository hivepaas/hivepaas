package redact

import (
	"runtime"
	"sort"
	"strings"
	"sync"
)

const (
	// concurrencyThreshold is the slice length limit at which we switch
	// from sequential execution to parallel execution to optimize performance.
	concurrencyThreshold = 500

	redactionMask = "********"
)

// Redactor handles the masking of sensitive secrets within text data.
type Redactor struct {
	secrets  []string
	replacer *strings.Replacer
}

// New creates a new Redactor initialized with the given secrets.
// Secrets are sorted by length in descending order to prevent partial matching.
func New(secrets []string) *Redactor {
	r := &Redactor{secrets: secrets}
	r.init()
	return r
}

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

// String replaces secrets in a single string sequentially.
func (r *Redactor) String(text string) string {
	return r.replacer.Replace(text)
}

func (r *Redactor) AddSecrets(secrets []string) {
	// TODO: do we need to use mutex
	r.secrets = append(r.secrets, secrets...)
	r.init()
}

// Slice replaces secrets in-place inside the given slice of strings,
// and returns the modified slice. It automatically chooses between
// sequential and parallel execution based on the slice size.
func (r *Redactor) Slice(logs []string) []string {
	numLogs := len(logs)
	// Use sequential processing for small slices to avoid goroutine overhead.
	if numLogs < concurrencyThreshold {
		for idx, log := range logs {
			logs[idx] = r.replacer.Replace(log)
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
			for idx := s; idx < e; idx++ {
				// Each worker processes a disjoint range of indices to prevent write contention.
				logs[idx] = r.replacer.Replace(logs[idx])
			}
		}(start, end)
	}
	wg.Wait()
	return logs
}
