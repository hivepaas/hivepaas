package taskdatafileload

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

// A load is done when a command ran, ended well, and read the whole file:
// anything less leaves the app with a part of it, or nothing, and has to say so.
func TestLoadOutcome(t *testing.T) {
	ran := func(code int) *schedjobexecservice.RunCommandResp {
		return &schedjobexecservice.RunCommandResp{ExitCode: &code}
	}
	whole := readResult{n: 10, eof: true}

	assert.NoError(t, loadOutcome(ran(0), nil, whole), "read whole, ended well")
	assert.Error(t, loadOutcome(&schedjobexecservice.RunCommandResp{}, nil, readResult{}),
		"no container ran the command")
	assert.Error(t, loadOutcome(ran(0), nil, readResult{n: 4}), "ended before it read the whole file")
	assert.Error(t, loadOutcome(ran(0), nil, readResult{n: 4, err: errors.New("truncated")}),
		"the file could not be read")
	assert.Error(t, loadOutcome(ran(1), errors.New("exit 1"), whole), "the command failed")
}
