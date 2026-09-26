package sessionuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// checklistService answers Checklist with what the test sets and counts the
// calls to Finish. The embedded interface leaves the rest nil: GetMe reaches
// nothing else.
type checklistService struct {
	getstartedservice.Service
	checklist    getstartedservice.Checklist
	hasTwoFactor bool
	finished     int
}

func (s *checklistService) Checklist(
	_ context.Context, _ database.IDB, hasTwoFactor bool,
) (*getstartedservice.Checklist, error) {
	s.hasTwoFactor = hasTwoFactor
	checklist := s.checklist
	return &checklist, nil
}

func (s *checklistService) Finish(context.Context, database.IDB) error {
	s.finished++
	return nil
}

func adminWithTotp(secret string) *basedto.User {
	return &basedto.User{User: &entity.User{Role: base.UserRoleAdmin, TotpSecret: secret}}
}

func TestSetupChecklistIsGivenWhileTheStepIsGetStarted(t *testing.T) {
	service := &checklistService{checklist: getstartedservice.Checklist{
		DashboardCert: getstartedservice.Item{Status: getstartedservice.ItemStatusFailed,
			Domain: "dash.example.com", Error: "acme: NXDOMAIN"},
		TwoFactor: getstartedservice.Item{Status: getstartedservice.ItemStatusDone},
		GithubApp: getstartedservice.Item{Status: getstartedservice.ItemStatusTodo},
	}}
	uc := &UC{getStartedService: service}
	respData := &sessiondto.GetMeDataResp{NextStep: base.InstallationStepGetStarted}

	err := uc.addSetupChecklist(context.Background(), adminWithTotp("secret"), respData)

	assert.NoError(t, err)
	assert.True(t, service.hasTwoFactor)
	assert.Zero(t, service.finished)
	assert.Equal(t, base.InstallationStepGetStarted, respData.NextStep)
	if assert.NotNil(t, respData.SetupChecklist) {
		assert.Equal(t, "failed", respData.SetupChecklist.DashboardCert.Status)
		assert.Equal(t, "dash.example.com", respData.SetupChecklist.DashboardCert.Domain)
		assert.Equal(t, "acme: NXDOMAIN", respData.SetupChecklist.DashboardCert.Error)
		assert.Equal(t, "done", respData.SetupChecklist.TwoFactor.Status)
		assert.Equal(t, "todo", respData.SetupChecklist.GithubApp.Status)
	}
}

func TestSetupChecklistAllDoneClearsTheStep(t *testing.T) {
	done := getstartedservice.Item{Status: getstartedservice.ItemStatusDone}
	service := &checklistService{checklist: getstartedservice.Checklist{
		DashboardCert: done, TwoFactor: done, GithubApp: done,
	}}
	uc := &UC{getStartedService: service}
	respData := &sessiondto.GetMeDataResp{NextStep: base.InstallationStepGetStarted}

	err := uc.addSetupChecklist(context.Background(), adminWithTotp("secret"), respData)

	assert.NoError(t, err)
	assert.Equal(t, 1, service.finished)
	assert.Empty(t, respData.NextStep)
	assert.Nil(t, respData.SetupChecklist)
}

func TestSetupChecklistIsAbsentForAnotherStep(t *testing.T) {
	service := &checklistService{}
	uc := &UC{getStartedService: service}
	respData := &sessiondto.GetMeDataResp{NextStep: base.InstallationStepInitData}

	err := uc.addSetupChecklist(context.Background(), adminWithTotp(""), respData)

	assert.NoError(t, err)
	assert.Nil(t, respData.SetupChecklist)
	assert.Zero(t, service.finished)
}
