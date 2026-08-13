package appdeploymentserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
)

func (s *service) repoDeployStepSourceCheckout(
	ctx context.Context,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepRepoCheckout
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	checkoutReq := &repocheckoutservice.RepoCheckoutReq{
		App:         data.App,
		RepoSource:  repoSource,
		CredSetting: data.RefObjects.RefSettings[repoSource.Credentials.ID],
		RefObjects:  data.RefObjects,
		LogStore:    data.LogStore,
		TempDir:     data.TempDir,
		CheckoutDir: data.CheckoutDir,
	}
	if deployment.Settings.NoCache || (data.ImageBuildSettings != nil && data.ImageBuildSettings.NoCache) {
		checkoutReq.NoCache = true
	}

	checkoutResp, err := s.repoCheckoutService.Checkout(ctx, checkoutReq)
	if err != nil {
		return apperrors.Wrap(err)
	}

	repoSource.CommitHash = checkoutResp.CommitHash
	data.Deployment.Output.CommitHash = checkoutResp.CommitHash
	data.Deployment.Output.CommitMessage = checkoutResp.CommitMessage
	data.Deployment.Output.CommitTitle = checkoutResp.CommitTitle
	data.Deployment.Output.CommitAuthor = checkoutResp.CommitAuthor

	return nil
}
