package entity_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// ArgsAsX and OutputAsX read like getters and are not: the first caller parses
// and caches into the task. Steps of a system update run side by side, so two of
// them reaching a cold cache at the same time has to be safe.
//
// Run under -race, which scripts/test.sh does; without the lock this fails there
// rather than here.
func TestTaskArgsAndOutputSurviveConcurrentReaders(t *testing.T) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeSystemUpdate}
	task.MustSetArgs(&entity.TaskSystemUpdateArgs{
		TargetVersion: &base.ReleaseInfo{AppVersion: "v0.2.0"},
	})
	task.MustSetOutput(&entity.TaskSystemUpdateOutput{})

	// A task loaded from the database has the strings and no parsed cache, which
	// is the state where the first reader writes.
	cold := &entity.Task{ID: task.ID, Type: task.Type, Args: task.Args, Output: task.Output}

	const readers = 8
	var wg sync.WaitGroup
	results := make([]string, readers)

	for i := range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			args, err := cold.ArgsAsSystemUpdate()
			if err == nil && args != nil && args.TargetVersion != nil {
				results[i] = args.TargetVersion.AppVersion
			}
			_, _ = cold.OutputAsSystemUpdate()
		}()
	}
	wg.Wait()

	for i, got := range results {
		assert.Equal(t, "v0.2.0", got, "reader %d", i)
	}
}
