package basesettinghandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/awss3uc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/awsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/gitcredentialuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuilduc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/oauthuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/registryauthuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/repowebhookuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/ssluc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systemcleanupuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc"
)

type BaseSettingHandler struct {
	*handler.BaseHandler
	AuthHandler     *authhandler.AuthHandler
	OAuthUC         *oauthuc.OAuthUC
	AWSUC           *awsuc.AWSUC
	AWSS3UC         *awss3uc.AWSS3UC
	SSHKeyUC        *sshkeyuc.SSHKeyUC
	IMServiceUC     *imserviceuc.IMServiceUC
	RegistryAuthUC  *registryauthuc.RegistryAuthUC
	BasicAuthUC     *basicauthuc.BasicAuthUC
	SSLUC           *ssluc.SSLUC
	GithubAppUC     *githubappuc.GithubAppUC
	AccessTokenUC   *accesstokenuc.AccessTokenUC
	CronJobUC       *cronjobuc.CronJobUC
	HealthcheckUC   *healthcheckuc.HealthcheckUC
	SecretUC        *secretuc.SecretUC
	EmailUC         *emailuc.EmailUC
	APIKeyUC        *apikeyuc.APIKeyUC
	RepoWebhookUC   *repowebhookuc.RepoWebhookUC
	NotificationUC  *notificationuc.NotificationUC
	ImageBuildUC    *imagebuilduc.ImageBuildUC
	SystemCleanupUC *systemcleanupuc.SystemCleanupUC
	GitCredentialUC *gitcredentialuc.GitCredentialUC
}

func NewBaseSettingHandler(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.AuthHandler,
	oauthUC *oauthuc.OAuthUC,
	awsUC *awsuc.AWSUC,
	awsS3UC *awss3uc.AWSS3UC,
	sshKeyUC *sshkeyuc.SSHKeyUC,
	imServiceUC *imserviceuc.IMServiceUC,
	registryAuthUC *registryauthuc.RegistryAuthUC,
	basicAuthUC *basicauthuc.BasicAuthUC,
	sslUC *ssluc.SSLUC,
	githubAppUC *githubappuc.GithubAppUC,
	accessTokenUC *accesstokenuc.AccessTokenUC,
	cronJobUC *cronjobuc.CronJobUC,
	healthcheckUC *healthcheckuc.HealthcheckUC,
	secretUC *secretuc.SecretUC,
	emailUC *emailuc.EmailUC,
	apiKeyUC *apikeyuc.APIKeyUC,
	repoWebhookUC *repowebhookuc.RepoWebhookUC,
	notificationUC *notificationuc.NotificationUC,
	imageBuildUC *imagebuilduc.ImageBuildUC,
	systemCleanupUC *systemcleanupuc.SystemCleanupUC,
	gitCredentialUC *gitcredentialuc.GitCredentialUC,
) *BaseSettingHandler {
	return &BaseSettingHandler{
		BaseHandler:     baseHandler,
		AuthHandler:     authHandler,
		OAuthUC:         oauthUC,
		AWSUC:           awsUC,
		AWSS3UC:         awsS3UC,
		SSHKeyUC:        sshKeyUC,
		IMServiceUC:     imServiceUC,
		RegistryAuthUC:  registryAuthUC,
		BasicAuthUC:     basicAuthUC,
		SSLUC:           sslUC,
		GithubAppUC:     githubAppUC,
		AccessTokenUC:   accessTokenUC,
		CronJobUC:       cronJobUC,
		HealthcheckUC:   healthcheckUC,
		SecretUC:        secretUC,
		EmailUC:         emailUC,
		APIKeyUC:        apiKeyUC,
		RepoWebhookUC:   repoWebhookUC,
		NotificationUC:  notificationUC,
		ImageBuildUC:    imageBuildUC,
		SystemCleanupUC: systemCleanupUC,
		GitCredentialUC: gitCredentialUC,
	}
}
