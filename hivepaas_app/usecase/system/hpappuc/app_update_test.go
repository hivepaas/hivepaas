package hpappuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

// fakeHpAppService serves a fixed release.json and records whether an update
// was started. Methods the tests do not reach are left to the nil interface.
type fakeHpAppService struct {
	hpappservice.Service
	info    *hpappservice.AppReleaseInfo
	updated bool
}

func (s *fakeHpAppService) GetAppReleaseInfo(context.Context) (*hpappservice.AppReleaseInfo, error) {
	return s.info, nil
}

func (s *fakeHpAppService) UpdateSystemVersion(context.Context, database.IDB, *base.ReleaseInfo, bool) error {
	s.updated = true
	return nil
}

func releaseOf(appVersion string, canUpdate bool) *hpappservice.ReleaseInfo {
	return &hpappservice.ReleaseInfo{
		ReleaseInfo: base.ReleaseInfo{AppVersion: appVersion},
		CanUpdate:   canUpdate,
	}
}

func TestUpdateHpApp_RefusesTargetNotNewer(t *testing.T) {
	cases := []struct {
		name   string
		info   *hpappservice.AppReleaseInfo
		target string
		err    error
	}{
		{
			name:   "stable lists an older version",
			info:   &hpappservice.AppReleaseInfo{Stable: releaseOf("v0.0.9", false)},
			target: "v0.0.9",
			err:    hperrors.ErrVersionNotNewer,
		},
		{
			name: "beta lists an older version",
			info: &hpappservice.AppReleaseInfo{
				Stable: releaseOf("v0.2.0", true),
				Beta:   releaseOf("v0.0.9-beta1", false),
			},
			target: "v0.0.9-beta1",
			err:    hperrors.ErrVersionNotNewer,
		},
		{
			name:   "target not listed at all",
			info:   &hpappservice.AppReleaseInfo{Stable: releaseOf("v0.2.0", true)},
			target: "v0.3.0",
			err:    hperrors.ErrUpdateVerMismatched,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeHpAppService{info: tc.info}
			// No db: a refusal has to happen before the transaction is opened.
			uc := New(nil, nil, nil, svc)

			resp, err := uc.UpdateHpApp(context.Background(), &basedto.Auth{},
				&hpappdto.UpdateHpAppReq{TargetVersion: tc.target})

			assert.Nil(t, resp)
			assert.ErrorIs(t, err, tc.err)
			assert.False(t, svc.updated, "no update may be started")
		})
	}
}
