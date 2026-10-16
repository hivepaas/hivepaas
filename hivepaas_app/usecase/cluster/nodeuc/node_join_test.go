package nodeuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

// A join past the plan's node limit is refused before anything else: before the
// join token is revealed or read, and before the SSH key is decrypted. The UC
// here has nothing else to reach, so going further would panic.
func TestNodeJoinRefusedFirst(t *testing.T) {
	ctx := context.Background()
	auth := &basedto.Auth{}

	license := &fakeLicenseService{err: hperrors.ErrPlanLimitReached}
	uc := &UC{licenseService: license}

	_, err := uc.GetNodeJoinCommand(ctx, auth, &nodedto.GetNodeJoinCommandReq{JoinAsManager: true})
	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)

	_, err = uc.JoinNode(ctx, auth, &nodedto.JoinNodeReq{Host: "10.0.0.9"})
	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)

	assert.Equal(t, []*licenseservice.CheckReq{
		licenseservice.Adding(base.LicenseLimitNodes, 1),
		licenseservice.Adding(base.LicenseLimitNodes, 1),
	}, license.checks)
}

type fakeLicenseService struct {
	licenseservice.Service
	checks []*licenseservice.CheckReq
	err    error
}

func (f *fakeLicenseService) Check(_ context.Context, _ database.IDB, req *licenseservice.CheckReq) error {
	f.checks = append(f.checks, req)
	return f.err
}
