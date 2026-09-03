package redact

import (
	"strings"
	"sync"
	"testing"
)

// TestRedactorConcurrentAddSecrets must be run with -race: AddSecrets used to
// swap the replacer while Slice workers were reading it.
func TestRedactorConcurrentAddSecrets(t *testing.T) {
	r := New([]string{"secret-0"})

	logs := make([]string, concurrencyThreshold+10)
	for i := range logs {
		logs[i] = "line with secret-0 inside"
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range 50 {
			r.AddSecrets([]string{"secret-" + string(rune('a'+i%26))})
		}
	})
	wg.Go(func() {
		for range 50 {
			out := make([]string, len(logs))
			copy(out, logs)
			for _, line := range r.Slice(out) {
				if strings.Contains(line, "secret-0") {
					t.Errorf("secret was not redacted: %s", line)
					return
				}
			}
		}
	})
	wg.Wait()
}
