package getstarteduc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
)

// certService answers DashboardCert with what the test sets and counts the
// calls to Finish. The embedded interface leaves the rest nil: reading the
// certificate reaches nothing else.
type certService struct {
	getstartedservice.Service
	item     getstartedservice.Item
	finished int
}

func (s *certService) DashboardCert(context.Context, database.IDB) (*getstartedservice.Item, error) {
	item := s.item
	return &item, nil
}

func (s *certService) Finish(context.Context, database.IDB) error {
	s.finished++
	return nil
}

func admin() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{Role: base.UserRoleAdmin}}}
}

func TestGetDashboardCertFinishesTheStepOnceTheCertificateIsDone(t *testing.T) {
	service := &certService{item: getstartedservice.Item{
		Status: getstartedservice.ItemStatusDone, Domain: "dash.example.com"}}
	uc := &UC{getStartedService: service}

	resp, err := uc.GetDashboardCert(context.Background(), admin(), getstarteddto.NewGetDashboardCertReq())

	assert.NoError(t, err)
	assert.Equal(t, 1, service.finished)
	if assert.NotNil(t, resp) && assert.NotNil(t, resp.Data) {
		assert.Equal(t, "done", resp.Data.Status, "the card still says so, until the page is loaded again")
		assert.Equal(t, "dash.example.com", resp.Data.Domain)
	}
}

func TestGetDashboardCertLeavesTheStepWhileTheCertificateIsNotDone(t *testing.T) {
	for _, status := range []getstartedservice.ItemStatus{
		getstartedservice.ItemStatusTodo, getstartedservice.ItemStatusObtaining, getstartedservice.ItemStatusFailed,
	} {
		service := &certService{item: getstartedservice.Item{Status: status, Error: "acme: NXDOMAIN"}}
		uc := &UC{getStartedService: service}

		resp, err := uc.GetDashboardCert(context.Background(), admin(), getstarteddto.NewGetDashboardCertReq())

		assert.NoError(t, err)
		assert.Zero(t, service.finished, string(status))
		if assert.NotNil(t, resp) && assert.NotNil(t, resp.Data) {
			assert.Equal(t, string(status), resp.Data.Status)
		}
	}
}

func TestGetDashboardCertIsForAdmins(t *testing.T) {
	service := &certService{item: getstartedservice.Item{Status: getstartedservice.ItemStatusDone}}
	uc := &UC{getStartedService: service}
	member := &basedto.Auth{User: &basedto.User{User: &entity.User{Role: base.UserRoleMember}}}

	_, err := uc.GetDashboardCert(context.Background(), member, getstarteddto.NewGetDashboardCertReq())

	assert.True(t, errors.Is(err, hperrors.ErrForbidden), "got %v", err)
	assert.Zero(t, service.finished)
}
