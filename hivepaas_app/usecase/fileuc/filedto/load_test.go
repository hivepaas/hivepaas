package filedto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

// A load is a command to feed the file to: without one there is nothing to
// load into; with a TTY asked for, the command gets none, as the file would not
// pass through one as is.
func TestLoadDataFileReq(t *testing.T) {
	req := NewLoadDataFileReq()
	req.ID = "01JAB9XED0GTXBSQDFVYAJ8WJ1"
	assert.NoError(t, req.ModifyRequest())
	assert.NotEmpty(t, req.Validate(), "no command")

	req.Command = &commandtemplatedto.CommandTemplateBaseReq{Command: "psql -U postgres", TTY: true}
	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, req.Validate())
	assert.False(t, req.Command.TTY)
}
