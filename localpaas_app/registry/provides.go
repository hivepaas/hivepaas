package registry

import (
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/appactionhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/appbasehandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/appdeploymenthandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/apphandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/appsettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/basesettinghandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/clusterhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/devhelperhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/filehandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/imagehandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/localpaashandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/projectbasehandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/projecthandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/projectsettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/sessionhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/settinghandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/supporthandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/systemhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/systemsettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/traefikhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/userhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/usersettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/webhookhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/server"
	"github.com/localpaas/localpaas/localpaas_app/permission/permissionimpl"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice/appdeploymentserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice/appserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice/clusterserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice/containerexecserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/dbservice/dbserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice/emailserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice/envvarserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice/fileserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/healthcheckservice/healthcheckserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice/lpappserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice/networkserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice/notificationserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice/projectserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice/schedjobserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice/settingserviceimpl"
	sslrenewalserviceimpl "github.com/localpaas/localpaas/localpaas_app/service/sslrenewalservice/sslrenewalservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice/sslserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/startupservice/startupserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/sysbackupservice/sysbackupserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice/syscleanupserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/sysupdateservice/sysupdateserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice/taskserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice/traefikserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice/userserviceimpl"
	"github.com/localpaas/localpaas/localpaas_app/tasks/initializer"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue/queueimpl"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskappdeploy"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskdummy"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskhealthcheck"
	"github.com/localpaas/localpaas/localpaas_app/tasks/taskschedjobexec"
	"github.com/localpaas/localpaas/localpaas_app/updater/tasksystemupdate"
	"github.com/localpaas/localpaas/localpaas_app/updater/updaterimpl"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appactionuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appsettingsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/binobjectuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/imageuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/nodeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/devhelperuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectsettingsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accessiblebyprojectsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/configfileuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/domainsettingsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/gitcredentialuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuildsettingsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/oauthuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/registryauthuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/repowebhookuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslcertuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/storagesettingsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/supportuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappsettingsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/syserroruc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefiksettingsuc"
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
	rediscache.NewLock,

	// http server
	server.NewHTTPServer,

	// permission
	permissionimpl.NewManager,

	// Infra
	docker.New,

	// Task queue
	queueimpl.New,
	initializer.NewWorkerInitializer,
	taskdummy.NewExecutor,
	taskappdeploy.NewExecutor,
	taskschedjobexec.NewExecutor,
	taskhealthcheck.NewExecutor,

	// Updater
	updaterimpl.New,
	tasksystemupdate.NewExecutor,

	// Route handler
	server.NewHandlerRegistry, // for all handler list
	handler.New,
	basesettinghandler.New,
	authhandler.New,
	clusterhandler.New,
	sessionhandler.New,
	userhandler.New,
	projectbasehandler.New,
	projecthandler.New,
	projectsettingshandler.New,
	appbasehandler.New,
	apphandler.New,
	appsettingshandler.New,
	appdeploymenthandler.New,
	appactionhandler.New,
	settinghandler.New,
	usersettingshandler.New,
	systemhandler.New,
	systemsettingshandler.New,
	localpaashandler.New,
	traefikhandler.New,
	webhookhandler.New,
	filehandler.New,
	imagehandler.New,
	devhelperhandler.New,
	supporthandler.New,

	// Use case
	syserroruc.New,
	nodeuc.New,
	volumeuc.New,
	imageuc.New,
	networkuc.New,
	sessionuc.New,
	useruc.New,
	projectuc.New,
	projectsettingsuc.New,
	appuc.New,
	appsettingsuc.New,
	appdeploymentuc.New,
	appactionuc.New,
	settings.New,
	cloudstorageuc.New,
	sshkeyuc.New,
	apikeyuc.New,
	oauthuc.New,
	secretuc.New,
	configfileuc.New,
	imserviceuc.New,
	registryauthuc.New,
	basicauthuc.New,
	sslcertuc.New,
	domainsettingsuc.New,
	githubappuc.New,
	accesstokenuc.New,
	traefikuc.New,
	lpappuc.New,
	schedjobuc.New,
	healthcheckuc.New,
	taskuc.New,
	emailuc.New,
	webhookuc.New,
	repowebhookuc.New,
	notificationuc.New,
	imagebuildsettingsuc.New,
	systemcleanupuc.New,
	gitcredentialuc.New,
	sslrenewaluc.New,
	systembackupuc.New,
	fileuc.New,
	storagesettingsuc.New,
	devhelperuc.New,
	lpappsettingsuc.New,
	traefiksettingsuc.New,
	binobjectuc.New,
	accessiblebyprojectsuc.New,
	supportuc.New,

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
	schedjobserviceimpl.New,
	taskserviceimpl.New,
	dbserviceimpl.New,
	fileserviceimpl.New,
	sslserviceimpl.New,
	appserviceimpl.New,
	appdeploymentserviceimpl.New,
	sysbackupserviceimpl.New,
	syscleanupserviceimpl.New,
	sysupdateserviceimpl.New,
	sslrenewalserviceimpl.New,
	containerexecserviceimpl.New,
	healthcheckserviceimpl.New,
	startupserviceimpl.New,

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
	repository.NewResLinkRepo,
	// Repo: File
	repository.NewFileRepo,
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
	repository.NewBinObjectRepo,

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
