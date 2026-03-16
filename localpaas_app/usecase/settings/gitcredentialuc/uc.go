package gitcredentialuc

import (
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GitCredentialUC struct {
	*settings.BaseSettingUC
}

func NewGitCredentialUC(
	baseSettingUC *settings.BaseSettingUC,
) *GitCredentialUC {
	return &GitCredentialUC{
		BaseSettingUC: baseSettingUC,
	}
}
