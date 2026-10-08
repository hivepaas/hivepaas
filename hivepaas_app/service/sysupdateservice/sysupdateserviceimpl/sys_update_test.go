package sysupdateserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
)

// The task queues are paused right before the app and the worker are scaled
// down, once the images are pulled. A pause from before the pull - minutes on a
// slow link - could run out first, and the app start tasks the scale-down then
// cuts short.
func TestTheTaskQueuesArePausedRightBeforeTheScaleDown(t *testing.T) {
	f := &fakeDocker{}
	hp := &fakeHpApp{app: swarmServiceOn("hivepaas_app", "hivepaas/hivepaas-dev:0.1.0", 2)}
	s := &service{dockerManager: f, hpAppService: hp}
	scaledAtPause := -1
	data := &sysUpdateData{SysUpdateReq: &sysupdateservice.SysUpdateReq{
		PauseTaskQueues: func() error {
			scaledAtPause = len(f.replicaUpdates)
			return nil
		},
	}}

	assert.NoError(t, s.stopServices(context.Background(), data))

	assert.Equal(t, 0, scaledAtPause)
	if assert.NotEmpty(t, f.replicaUpdates) {
		assert.Equal(t, uint64(0), f.replicaUpdates[0].replicas)
	}
}
