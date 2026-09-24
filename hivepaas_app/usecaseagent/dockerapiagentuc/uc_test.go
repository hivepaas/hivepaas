package dockerapiagentuc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
)

// fakePolicies answers Policies with what a test sets, or fails.
type fakePolicies struct {
	dockerapiservice.Service
	policies []*dockerproxy.Policy
	err      error
}

func (f *fakePolicies) Policies(context.Context, database.IDB) ([]*dockerproxy.Policy, error) {
	return f.policies, f.err
}

func newTestUC(t *testing.T, policies *fakePolicies, node *fakeNode) *UC {
	t.Helper()
	host, _, _ := newTestHost(t)
	return &UC{logger: logging.GlobalLogger(), dockerManager: node, dockerAPIService: policies, host: host}
}

func TestSyncServesTheAppsThatHaveAccess(t *testing.T) {
	uc := newTestUC(t, &fakePolicies{policies: []*dockerproxy.Policy{testPolicy("app1")}}, aNode())
	served, err := uc.Sync(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, served)
}

var errDatabaseDown = errors.New("database down")

func TestAFailedReadChangesNothing(t *testing.T) {
	policies := &fakePolicies{policies: []*dockerproxy.Policy{testPolicy("app1")}}
	node := aNode()
	uc := newTestUC(t, policies, node)
	_, err := uc.Sync(context.Background())
	assert.NoError(t, err)

	policies.err = errDatabaseDown
	_, err = uc.Sync(context.Background())
	assert.ErrorIs(t, err, errDatabaseDown)
	assert.Equal(t, []string{"app1"}, uc.host.served(), "the socket stays")

	_, err = uc.Collect(context.Background())
	assert.ErrorIs(t, err, errDatabaseDown)
	assert.Empty(t, node.removed, "without knowing who has access, nothing is anybody's to remove")
}

func TestRemoveAppStopsServingItAndRemovesItsObjects(t *testing.T) {
	node := aNode()
	uc := newTestUC(t, &fakePolicies{policies: []*dockerproxy.Policy{testPolicy("app1")}}, node)
	_, err := uc.Sync(context.Background())
	assert.NoError(t, err)

	removed, err := uc.RemoveApp(context.Background(), "app1")
	assert.NoError(t, err)
	assert.Empty(t, uc.host.served())
	assert.Equal(t, 3, removed.Containers)
}
