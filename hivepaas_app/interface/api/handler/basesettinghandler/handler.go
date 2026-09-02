package basesettinghandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appfeaturesettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appplacementsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/emailuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/gitcredentialuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/notificationuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/oauthuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslprovideruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/storagesettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/backuprepocleanupuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/sslrenewaluc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systemcleanupuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/usersettings/apikeyuc"
)

type Handler struct {
	*handler.BaseHandler
	AuthHandler            *authhandler.Handler
	AccessTokenUC          *accesstokenuc.UC
	AcmeDnsProviderUC      *acmednsprovideruc.UC
	APIKeyUC               *apikeyuc.UC
	AppFeatureSettingsUC   *appfeaturesettingsuc.UC
	AppPlacementSettingsUC *appplacementsettingsuc.UC
	BackupRepoCleanupUC    *backuprepocleanupuc.UC
	BackupRepoUC           *backuprepouc.UC
	BasicAuthUC            *basicauthuc.UC
	CloudStorageUC         *cloudstorageuc.UC
	ClusterNetworkUC       *networkuc.UC
	ClusterNodeUC          *nodeuc.UC
	ClusterVolumeUC        *volumeuc.UC
	CommandPipeUC          *commandpipeuc.UC
	CommandTemplateUC      *commandtemplateuc.UC
	ConfigFileUC           *configfileuc.UC
	DomainSettingsUC       *domainsettingsuc.UC
	EmailUC                *emailuc.UC
	FileUC                 *fileuc.UC
	GitCredentialUC        *gitcredentialuc.UC
	GithubAppUC            *githubappuc.UC
	ImageBuildUC           *imagebuildsettingsuc.UC
	IMServiceUC            *imserviceuc.UC
	NotificationUC         *notificationuc.UC
	OAuthUC                *oauthuc.UC
	PeriodicJobUC          *periodicjobuc.UC
	RegistryAuthUC         *registryauthuc.UC
	RepoWebhookUC          *repowebhookuc.UC
	SchedJobUC             *schedjobuc.UC
	SecretUC               *secretuc.UC
	SSHKeyUC               *sshkeyuc.UC
	SSLCertUC              *sslcertuc.UC
	SSLProviderUC          *sslprovideruc.UC
	SSLRenewalUC           *sslrenewaluc.UC
	StorageSettingsUC      *storagesettingsuc.UC
	SystemBackupUC         *systembackupuc.UC
	SystemCleanupUC        *systemcleanupuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	accessTokenUC *accesstokenuc.UC,
	acmeDnsProviderUC *acmednsprovideruc.UC,
	apiKeyUC *apikeyuc.UC,
	appFeatureSettingsUC *appfeaturesettingsuc.UC,
	appPlacementSettingsUC *appplacementsettingsuc.UC,
	backupRepoCleanupUC *backuprepocleanupuc.UC,
	backupRepoUC *backuprepouc.UC,
	basicAuthUC *basicauthuc.UC,
	cloudStorageUC *cloudstorageuc.UC,
	clusterNetworkUC *networkuc.UC,
	clusterNodeUC *nodeuc.UC,
	clusterVolumeUC *volumeuc.UC,
	commandPipeUC *commandpipeuc.UC,
	commandTemplateUC *commandtemplateuc.UC,
	configFileUC *configfileuc.UC,
	domainSettingsUC *domainsettingsuc.UC,
	emailUC *emailuc.UC,
	fileUC *fileuc.UC,
	gitCredentialUC *gitcredentialuc.UC,
	githubAppUC *githubappuc.UC,
	imageBuildUC *imagebuildsettingsuc.UC,
	imServiceUC *imserviceuc.UC,
	notificationUC *notificationuc.UC,
	oauthUC *oauthuc.UC,
	periodicJobUC *periodicjobuc.UC,
	registryAuthUC *registryauthuc.UC,
	repoWebhookUC *repowebhookuc.UC,
	schedJobUC *schedjobuc.UC,
	secretUC *secretuc.UC,
	sshKeyUC *sshkeyuc.UC,
	sslCertUC *sslcertuc.UC,
	sslProviderUC *sslprovideruc.UC,
	sslRenewalUC *sslrenewaluc.UC,
	storageSettingsUC *storagesettingsuc.UC,
	systemBackupUC *systembackupuc.UC,
	systemCleanupUC *systemcleanupuc.UC,
) *Handler {
	return &Handler{
		BaseHandler:            baseHandler,
		AuthHandler:            authHandler,
		AccessTokenUC:          accessTokenUC,
		AcmeDnsProviderUC:      acmeDnsProviderUC,
		APIKeyUC:               apiKeyUC,
		AppFeatureSettingsUC:   appFeatureSettingsUC,
		AppPlacementSettingsUC: appPlacementSettingsUC,
		BackupRepoCleanupUC:    backupRepoCleanupUC,
		BackupRepoUC:           backupRepoUC,
		BasicAuthUC:            basicAuthUC,
		CloudStorageUC:         cloudStorageUC,
		ClusterNetworkUC:       clusterNetworkUC,
		ClusterNodeUC:          clusterNodeUC,
		ClusterVolumeUC:        clusterVolumeUC,
		CommandPipeUC:          commandPipeUC,
		CommandTemplateUC:      commandTemplateUC,
		ConfigFileUC:           configFileUC,
		DomainSettingsUC:       domainSettingsUC,
		EmailUC:                emailUC,
		FileUC:                 fileUC,
		GitCredentialUC:        gitCredentialUC,
		GithubAppUC:            githubAppUC,
		ImageBuildUC:           imageBuildUC,
		IMServiceUC:            imServiceUC,
		NotificationUC:         notificationUC,
		OAuthUC:                oauthUC,
		PeriodicJobUC:          periodicJobUC,
		RegistryAuthUC:         registryAuthUC,
		RepoWebhookUC:          repoWebhookUC,
		SchedJobUC:             schedJobUC,
		SecretUC:               secretUC,
		SSHKeyUC:               sshKeyUC,
		SSLCertUC:              sslCertUC,
		SSLProviderUC:          sslProviderUC,
		SSLRenewalUC:           sslRenewalUC,
		StorageSettingsUC:      storageSettingsUC,
		SystemBackupUC:         systemBackupUC,
		SystemCleanupUC:        systemCleanupUC,
	}
}
