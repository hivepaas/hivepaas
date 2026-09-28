package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A system backup references the repository it goes into.
func TestSystemBackupReferencesItsRepository(t *testing.T) {
	backup := &SystemBackup{IncludeDB: true, TargetRepository: ObjectID{ID: "repo1"}}

	assert.Equal(t, []string{"repo1"}, backup.GetRefObjectIDs().RefSettingIDs)
}

// What a run takes, by name, in the order a person reads it.
func TestSystemBackupIncludes(t *testing.T) {
	assert.Equal(t, []string{"database", "spec"}, (&SystemBackup{IncludeDB: true, IncludeSpec: true}).Includes())
	assert.Equal(t, []string{"spec"}, (&SystemBackup{IncludeSpec: true}).Includes())
	assert.Empty(t, (&SystemBackup{}).Includes())
}
