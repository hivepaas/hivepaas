package basesettinghandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/fileuc"
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
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslcertuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/sslrenewaluc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systembackupuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systemcleanupuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc"
)

type BaseSettingHandler struct {
	*handler.BaseHandler
	AuthHandler     *authhandler.AuthHandler
	OAuthUC         *oauthuc.UC
	CloudStorageUC  *cloudstorageuc.UC
	SSHKeyUC        *sshkeyuc.UC
	IMServiceUC     *imserviceuc.UC
	RegistryAuthUC  *registryauthuc.UC
	BasicAuthUC     *basicauthuc.UC
	SSLCertUC       *sslcertuc.UC
	GithubAppUC     *githubappuc.UC
	AccessTokenUC   *accesstokenuc.UC
	CronJobUC       *cronjobuc.UC
	HealthcheckUC   *healthcheckuc.UC
	SecretUC        *secretuc.UC
	EmailUC         *emailuc.UC
	APIKeyUC        *apikeyuc.UC
	RepoWebhookUC   *repowebhookuc.UC
	NotificationUC  *notificationuc.UC
	ImageBuildUC    *imagebuilduc.UC
	SystemCleanupUC *systemcleanupuc.UC
	SystemBackupUC  *systembackupuc.UC
	GitCredentialUC *gitcredentialuc.UC
	SSLRenewalUC    *sslrenewaluc.UC
	FileUC          *fileuc.UC
}

func NewBaseSettingHandler(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.AuthHandler,
	oauthUC *oauthuc.UC,
	cloudStorageUC *cloudstorageuc.UC,
	sshKeyUC *sshkeyuc.UC,
	imServiceUC *imserviceuc.UC,
	registryAuthUC *registryauthuc.UC,
	basicAuthUC *basicauthuc.UC,
	sslCertUC *sslcertuc.UC,
	githubAppUC *githubappuc.UC,
	accessTokenUC *accesstokenuc.UC,
	cronJobUC *cronjobuc.UC,
	healthcheckUC *healthcheckuc.UC,
	secretUC *secretuc.UC,
	emailUC *emailuc.UC,
	apiKeyUC *apikeyuc.UC,
	repoWebhookUC *repowebhookuc.UC,
	notificationUC *notificationuc.UC,
	imageBuildUC *imagebuilduc.UC,
	systemCleanupUC *systemcleanupuc.UC,
	systemBackupUC *systembackupuc.UC,
	gitCredentialUC *gitcredentialuc.UC,
	sslRenewalUC *sslrenewaluc.UC,
	fileUC *fileuc.UC,
) *BaseSettingHandler {
	return &BaseSettingHandler{
		BaseHandler:     baseHandler,
		AuthHandler:     authHandler,
		OAuthUC:         oauthUC,
		CloudStorageUC:  cloudStorageUC,
		SSHKeyUC:        sshKeyUC,
		IMServiceUC:     imServiceUC,
		RegistryAuthUC:  registryAuthUC,
		BasicAuthUC:     basicAuthUC,
		SSLCertUC:       sslCertUC,
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
		SystemBackupUC:  systemBackupUC,
		GitCredentialUC: gitCredentialUC,
		SSLRenewalUC:    sslRenewalUC,
		FileUC:          fileUC,
	}
}
