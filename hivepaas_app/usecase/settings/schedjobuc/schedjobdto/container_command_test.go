package schedjobdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

// A container command runs in a running container of an app: the app is
// named, or there is nothing to run it in.
func TestAContainerCommandNamesItsApp(t *testing.T) {
	req := NewCreateSchedJobReq()
	req.SchedJobBaseReq = &SchedJobBaseReq{Name: "migrate", JobType: base.SchedJobTypeContainerCommand,
		Command: &commandtemplatedto.CommandTemplateBaseReq{Command: "./migrate up"}}
	assert.Equal(t, "app", invalidFields(t, req))

	req.App = basedto.ObjectIDReq{ID: backupApp}
	assert.Equal(t, "", invalidFields(t, req))
}
