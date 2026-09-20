package volumeserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDirWritableCmd(t *testing.T) {
	cmd := makeDirWritableCmd("/mnt/data/project_data/p1")

	// The directory itself is the part that has to work.
	assert.Contains(t, cmd, "mkdir -p '/mnt/data/project_data/p1' && chmod 777 '/mnt/data/project_data/p1'")

	// What is already inside it is attempted and forgiven: a unix socket left by a
	// deleted app cannot be chmod'd on a Docker Desktop bind mount, and used to
	// take the whole command - and with it the project - down with it.
	assert.Contains(t, cmd, "|| true")
	assert.Contains(t, cmd, "-type d -o -type f")
	assert.NotContains(t, cmd, "chmod -R")

	// Nothing runs after a failed mkdir.
	before, _, found := strings.Cut(cmd, "&&")
	assert.True(t, found)
	assert.Contains(t, before, "mkdir -p")
}
