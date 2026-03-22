package registry

import (
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/apphandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/basesettinghandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/clusterhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/projecthandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/sessionhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/settinghandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/systemhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/systemsettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/userhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/usersettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/webhookhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/server"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/localpaas_app/service/cronjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/dbservice"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/nginxservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/initializer"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskappdeploy"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskcronjobexec"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskhealthcheck"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/imageuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/nodeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/nginxuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/syserroruc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/sslrenewaluc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systemcleanupuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/webhookuc"
	"github.com/localpaas/localpaas/services/docker"
)

var Provides = []any{
	// configuration
	config.LoadConfig,

	// logger
	logging.NewZapLogger,

	// db
	database.NewDB,

	// cache
	rediscache.NewClient,

	// http server
	server.NewHTTPServer,

	// permission
	permission.NewManager,

	// Infra
	docker.New,

	// Task queue
	queue.NewTaskQueue,
	initializer.NewWorkerInitializer,
	taskappdeploy.NewExecutor,
	taskcronjobexec.NewExecutor,
	taskhealthcheck.NewExecutor,

	// Route handler
	server.NewHandlerRegistry, // for all handler list
	handler.NewBaseHandler,
	basesettinghandler.NewBaseSettingHandler,
	authhandler.NewAuthHandler,
	clusterhandler.NewClusterHandler,
	sessionhandler.NewSessionHandler,
	userhandler.NewUserHandler,
	projecthandler.NewProjectHandler,
	apphandler.NewAppHandler,
	settinghandler.NewSettingHandler,
	usersettingshandler.NewUserSettingsHandler,
	systemhandler.NewSystemHandler,
	systemsettingshandler.NewSystemSettingsHandler,
	webhookhandler.NewWebhookHandler,

	// Use case
	syserroruc.NewSysErrorUC,
	nodeuc.NewNodeUC,
	volumeuc.NewVolumeUC,
	imageuc.NewImageUC,
	sessionuc.NewSessionUC,
	useruc.NewUserUC,
	projectuc.NewProjectUC,
	appuc.NewAppUC,
	appdeploymentuc.NewAppDeploymentUC,
	settings.NewBaseSettingUC,
	awsuc.NewAWSUC,
	awss3uc.NewAWSS3UC,
	sshkeyuc.NewSSHKeyUC,
	apikeyuc.NewAPIKeyUC,
	oauthuc.NewOAuthUC,
	secretuc.NewSecretUC,
	imserviceuc.NewIMServiceUC,
	registryauthuc.NewRegistryAuthUC,
	basicauthuc.NewBasicAuthUC,
	ssluc.NewSSLUC,
	githubappuc.NewGithubAppUC,
	accesstokenuc.NewAccessTokenUC,
	nginxuc.NewNginxUC,
	lpappuc.NewLpAppUC,
	cronjobuc.NewCronJobUC,
	healthcheckuc.NewHealthcheckUC,
	taskuc.NewTaskUC,
	emailuc.NewEmailUC,
	webhookuc.NewWebhookUC,
	repowebhookuc.NewRepoWebhookUC,
	notificationuc.NewNotificationUC,
	imagebuilduc.NewImageBuildUC,
	systemcleanupuc.NewSystemCleanupUC,
	gitcredentialuc.NewGitCredentialUC,
	sslrenewaluc.NewSSLRenewalUC,

	// Service
	clusterservice.NewClusterService,
	userservice.NewUserService,
	projectservice.NewProjectService,
	appservice.NewAppService,
	networkservice.NewNetworkService,
	settingservice.NewSettingService,
	envvarservice.NewEnvVarService,
	nginxservice.NewNginxService,
	lpappservice.NewLpAppService,
	emailservice.NewEmailService,
	notificationservice.NewNotificationService,
	cronjobservice.NewCronJobService,
	taskservice.NewTaskService,
	dbservice.NewDBService,

	// Repo: User
	repository.NewUserRepo,
	// Repo: Permission
	repository.NewACLPermissionRepo,
	// Repo: Project
	repository.NewProjectRepo,
	repository.NewProjectTagRepo,
	repository.NewProjectSharedSettingRepo,
	// Repo: App
	repository.NewAppRepo,
	repository.NewAppTagRepo,
	// Repo: App deployment
	repository.NewDeploymentRepo,
	repository.NewTaskLogRepo,
	// Repo: Setting
	repository.NewSettingRepo,
	// Repo: Task
	repository.NewTaskRepo,
	// Repo: Sys error
	repository.NewSysErrorRepo,
	// Migration
	repository.NewDataMigrationRepo,
	// Others
	repository.NewLoginTrustedDeviceRepo,
	repository.NewLockRepo,

	// Cache Repo
	cacherepository.NewUserTokenRepo,
	cacherepository.NewMFAPasscodeRepo,
	cacherepository.NewTaskInfoRepo,
	cacherepository.NewTaskControlRepo,
	cacherepository.NewDeploymentInfoRepo,
	cacherepository.NewLoginAttemptRepo,
	cacherepository.NewHealthcheckNotifEventRepo,
	cacherepository.NewHealthcheckSettingsRepo,
	cacherepository.NewGithubAppManifestRepo,
}
