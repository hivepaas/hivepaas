package getstarteduc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
)

func TestRefuseWhileObtainingRefusesOnlyAnAttemptUnderWay(t *testing.T) {
	err := refuseWhileObtaining(&getstartedservice.Item{Status: getstartedservice.ItemStatusObtaining})
	assert.True(t, errors.Is(err, hperrors.ErrConflict), "got %v", err)

	for _, status := range []getstartedservice.ItemStatus{
		getstartedservice.ItemStatusTodo, getstartedservice.ItemStatusFailed, getstartedservice.ItemStatusDone,
	} {
		assert.NoError(t, refuseWhileObtaining(&getstartedservice.Item{Status: status}), string(status))
	}
}

func TestRefuseNotAskedSaysWhy(t *testing.T) {
	assert.NoError(t, refuseNotAsked(&getstartedservice.CertRequest{Tasks: []*entity.Task{{ID: "task-1"}}}))

	err := refuseNotAsked(&getstartedservice.CertRequest{NotAsked: "dash.example.com has a certificate attached"})
	assert.True(t, errors.Is(err, hperrors.ErrConflict), "got %v", err)
	var hpErr hperrors.HPError
	if assert.True(t, errors.As(err, &hpErr)) {
		assert.Contains(t, hpErr.Build(translation.LangEn).Detail, "dash.example.com has a certificate attached")
	}
}

func TestRequireAdmin(t *testing.T) {
	asRole := func(role base.UserRole) *basedto.Auth {
		return &basedto.Auth{User: &basedto.User{User: &entity.User{Role: role}}}
	}

	assert.NoError(t, requireAdmin(asRole(base.UserRoleAdmin)))
	assert.True(t, errors.Is(requireAdmin(asRole(base.UserRoleMember)), hperrors.ErrForbidden))
	assert.True(t, errors.Is(requireAdmin(nil), hperrors.ErrForbidden))
	assert.True(t, errors.Is(requireAdmin(&basedto.Auth{}), hperrors.ErrForbidden))
}
