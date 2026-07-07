package registry

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	agentserver "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appactionhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appdeploymenthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/apphandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/apppreviewhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/basesettinghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/clusterhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/devhelperhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/filehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/hivepaashandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/imagehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projecthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/sessionhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/settinghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/supporthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/systemhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/systemsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/traefikhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/userhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/usersettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/webhookhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/server"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission/permissionimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice/agentserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcopyservice/appcopyserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice/apppreviewserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice/appserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice/clusterserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice/containerexecserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dbservice/dbserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice/domainserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice/emailserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envvarserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice/fileserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice/healthcheckserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice/hpappserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice/imagebuildserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice/networkserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice/notificationserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice/projectserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice/repocheckoutserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/reslinkservice/reslinkserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice/schedjobserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice/settingserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslrenewalservice/sslrenewalserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice/sslserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/startupservice/startupserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysbackupservice/sysbackupserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice/syscleanupserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice/sysupdateserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice/taskserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice/traefikserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice/userserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/initializer"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue/queueimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskappdeploy"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskdummy"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskhealthcheck"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskschedjobexec"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater/tasksystemupdate"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater/updaterimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apppreviewuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/binobjectuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/builduc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/imageuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accessiblebyprojectsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appfeaturesettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/emailuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/gitcredentialuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/healthcheckuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/notificationuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/oauthuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslprovideruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/storagesettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/supportuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/sslrenewaluc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systemcleanupuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/usersettings/apikeyuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/webhookuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
	"github.com/hivepaas/hivepaas/services/docker"
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
	apppreviewhandler.New,
	settinghandler.New,
	usersettingshandler.New,
	systemhandler.New,
	systemsettingshandler.New,
	hivepaashandler.New,
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
	apppreviewuc.New,
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
	commandtemplateuc.New,
	sslprovideruc.New,
	sslcertuc.New,
	domainsettingsuc.New,
	githubappuc.New,
	accesstokenuc.New,
	acmednsprovideruc.New,
	traefikuc.New,
	hpappuc.New,
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
	appfeaturesettingsuc.New,
	devhelperuc.New,
	hpappsettingsuc.New,
	traefiksettingsuc.New,
	binobjectuc.New,
	accessiblebyprojectsuc.New,
	supportuc.New,
	builduc.New,

	// Service
	clusterserviceimpl.New,
	userserviceimpl.New,
	projectserviceimpl.New,
	networkserviceimpl.New,
	settingserviceimpl.New,
	envvarserviceimpl.New,
	traefikserviceimpl.New,
	hpappserviceimpl.New,
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
	domainserviceimpl.New,
	reslinkserviceimpl.New,
	repocheckoutserviceimpl.New,
	imagebuildserviceimpl.New,
	agentserviceimpl.New,
	appcopyserviceimpl.New,
	apppreviewserviceimpl.New,
	schedjobexecserviceimpl.New,

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

	// Agent
	agentserver.NewAgentServer,

	// Use case of Agent
	containeragentuc.New,
}
