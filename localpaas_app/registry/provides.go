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
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/filehandler"
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
	"github.com/localpaas/localpaas/localpaas_app/service/appservice/appserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice/clusterserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/cronjobservice/cronjobserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/dbservice/dbserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice/emailserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice/envvarserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice/fileserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice/lpappserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice/networkserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice/notificationserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice/projectserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice/settingserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice/sslserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice/taskserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice/traefikserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice/userserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/tasks/initializer"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskappdeploy"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskcronjobexec"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskhealthcheck"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/imageuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/nodeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/syserroruc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefikuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/sslrenewaluc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systembackupuc"
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
	filehandler.NewFileHandler,

	// Use case
	syserroruc.New,
	nodeuc.New,
	volumeuc.New,
	imageuc.New,
	networkuc.New,
	sessionuc.New,
	useruc.New,
	projectuc.New,
	appuc.New,
	appdeploymentuc.New,
	settings.New,
	cloudstorageuc.New,
	sshkeyuc.New,
	apikeyuc.New,
	oauthuc.New,
	secretuc.New,
	imserviceuc.New,
	registryauthuc.New,
	basicauthuc.New,
	sslcertuc.New,
	githubappuc.New,
	accesstokenuc.New,
	traefikuc.New,
	lpappuc.New,
	cronjobuc.New,
	healthcheckuc.New,
	taskuc.New,
	emailuc.New,
	webhookuc.New,
	repowebhookuc.New,
	notificationuc.New,
	imagebuilduc.New,
	systemcleanupuc.New,
	gitcredentialuc.New,
	sslrenewaluc.New,
	systembackupuc.New,
	fileuc.New,

	// Service
	clusterserviceimpl.New,
	userserviceimpl.New,
	projectserviceimpl.New,
	networkserviceimpl.New,
	settingserviceimpl.New,
	envvarserviceimpl.New,
	traefikserviceimpl.New,
	lpappserviceimpl.New,
	emailserviceimpl.New,
	notificationserviceimpl.New,
	cronjobserviceimpl.New,
	taskserviceimpl.New,
	dbserviceimpl.New,
	fileserviceimpl.New,
	sslserviceimpl.New,
	appserviceimpl.New,

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
	// Repo: System
	repository.NewSystemStatusRepo,
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
