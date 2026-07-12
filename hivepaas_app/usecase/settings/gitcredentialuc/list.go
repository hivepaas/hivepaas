package gitcredentialuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/gitcredentialuc/gitcredentialdto"
)

func (uc *UC) ListGitCredential(
	ctx context.Context,
	auth *basedto.Auth,
	req *gitcredentialdto.ListGitCredentialReq,
) (*gitcredentialdto.ListGitCredentialResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhereGroup(
			// Github app
			bunex.SelectWhereIn("setting.type = ?", base.SettingTypeGithubApp),
			// All access tokens of kind `git`
			bunex.SelectWhereOrGroup(
				bunex.SelectWhere("setting.type = ?", base.SettingTypeAccessToken),
				bunex.SelectWhereIn("setting.kind IN (?)", base.AllGitAccessTokenKinds...),
			),
			// All ssh keys of kind `git`
			bunex.SelectWhereOrGroup(
				bunex.SelectWhere("setting.type = ?", base.SettingTypeSSHKey),
				bunex.SelectWhereIn("setting.kind IN (?)", base.AllGitSSHKeyKinds...),
			),
		),
	}
	if len(req.Statuses) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("setting.status IN (?)", req.Statuses...))
	}
	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("setting.name ILIKE ?", keyword),
			),
		)
	}
	if len(auth.AllowObjectIDs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhereIn("setting.id IN (?)", auth.AllowObjectIDs...),
		)
	}

	settings, pagingMeta, err := uc.SettingRepo.List(ctx, uc.DB, req.Scope, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, setting := range settings {
		setting.CurrentObjectID = req.Scope.MainObjectID()
	}

	refObjects, err := uc.SettingService.LoadReferenceObjects(ctx, uc.DB, req.Scope, true,
		false, settings...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := gitcredentialdto.TransformGitCredentials(settings, refObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &gitcredentialdto.ListGitCredentialResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: respData,
	}, nil
}
