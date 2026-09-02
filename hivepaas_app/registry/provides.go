package registry

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	agentserver "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appactionhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appcontainerhandler"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectenvhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectenvsettingshandler"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice/agentserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice/appcloneserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice/apppreviewserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice/approutingserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice/appserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backuprepocleanupservice/backuprepocleanupserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice/backupreposerviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice/clustercleanupserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice/clustersecretserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice/clusterserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandpipeexecservice/commandpipeexecserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice/commandserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice/containerexecserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice/containerfileserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dbservice/dbserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice/domainserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice/emailserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envvarserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice/fileserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice/healthcheckserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice/hpappserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice/imagebuildserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice/networkserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice/nodeexecserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice/notificationserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice/placementserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice/projectserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice/repocheckoutserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/reslinkservice/reslinkserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice/schedjobserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice/scopeserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice/settingeventserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settinginitservice/settinginitserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice/settingserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslrenewalservice/sslrenewalserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice/sslserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/startupservice/startupserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysbackupservice/sysbackupserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice/syscleanupserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice/systemeventbusserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice/sysupdateserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice/taskserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice/traefikserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice/userserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice/volumeserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/initializer"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue/queueimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskappclone"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskappdeploy"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskapppreview"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskdummy"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskperiodicjobexec"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskschedjobexec"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/taskworkflow"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater/tasksystemupdate"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater/updaterimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/supportuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/sysstatusuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/backuprepocleanupuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/sslrenewaluc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systemcleanupuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/usersettings/apikeyuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/webhookuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc"
	"github.com/hivepaas/hivepaas/services/docker"
)

var Provides = []any{
	context.Background,

	// Configuration
	config.LoadConfig,

	// Logger
	logging.NewZapLogger,

	// DB
	database.NewDB,

	// Cache
	rediscache.NewClient,
	rediscache.NewLock,

	// HTTP server
	server.NewHTTPServer,

	// Permission
	permissionimpl.NewManager,

	// Infra
	docker.New,

	// Task queue
	initializer.NewWorkerInitializer,
	queueimpl.New,
	taskappclone.NewExecutor,
	taskappdeploy.NewExecutor,
	taskapppreview.NewExecutor,
	taskdummy.NewExecutor,
	taskperiodicjobexec.NewExecutor,
	taskschedjobexec.NewExecutor,
	taskworkflow.NewExecutor,

	// Updater
	tasksystemupdate.NewExecutor,
	updaterimpl.New,

	// Route handler
	appactionhandler.New,
	appbasehandler.New,
	appcontainerhandler.New,
	appdeploymenthandler.New,
	apphandler.New,
	apppreviewhandler.New,
	appsettingshandler.New,
	authhandler.New,
	basesettinghandler.New,
	clusterhandler.New,
	devhelperhandler.New,
	filehandler.New,
	handler.New,
	hivepaashandler.New,
	imagehandler.New,
	projectbasehandler.New,
	projectenvhandler.New,
	projectenvsettingshandler.New,
	projecthandler.New,
	projectsettingshandler.New,
	server.NewHandlerRegistry,
	sessionhandler.New,
	settinghandler.New,
	supporthandler.New,
	systemhandler.New,
	systemsettingshandler.New,
	traefikhandler.New,
	userhandler.New,
	usersettingshandler.New,
	webhookhandler.New,

	// Use case
	accesstokenuc.New,
	acmednsprovideruc.New,
	apikeyuc.New,
	appactionuc.New,
	appcontaineruc.New,
	appdeploymentuc.New,
	appfeaturesettingsuc.New,
	appplacementsettingsuc.New,
	apppreviewuc.New,
	appsettingsuc.New,
	appuc.New,
	backuprepocleanupuc.New,
	backuprepouc.New,
	basicauthuc.New,
	binobjectuc.New,
	builduc.New,
	cloudstorageuc.New,
	commandpipeuc.New,
	commandtemplateuc.New,
	configfileuc.New,
	devhelperuc.New,
	domainsettingsuc.New,
	emailuc.New,
	fileuc.New,
	gitcredentialuc.New,
	githubappuc.New,
	hpappsettingsuc.New,
	hpappuc.New,
	imagebuildsettingsuc.New,
	imageuc.New,
	imserviceuc.New,
	networkuc.New,
	nodeuc.New,
	notificationuc.New,
	oauthuc.New,
	periodicjobuc.New,
	projectenvsettingsuc.New,
	projectenvuc.New,
	projectsettingsuc.New,
	projectuc.New,
	registryauthuc.New,
	repowebhookuc.New,
	schedjobuc.New,
	secretuc.New,
	sessionuc.New,
	settings.New,
	sshkeyuc.New,
	sslcertuc.New,
	sslprovideruc.New,
	sslrenewaluc.New,
	storagesettingsuc.New,
	supportuc.New,
	syserroruc.New,
	sysstatusuc.New,
	systembackupuc.New,
	systemcleanupuc.New,
	taskuc.New,
	traefiksettingsuc.New,
	traefikuc.New,
	useruc.New,
	volumeuc.New,
	webhookuc.New,

	// Service
	agentserviceimpl.New,
	appcloneserviceimpl.New,
	appdeploymentserviceimpl.New,
	apppreviewserviceimpl.New,
	approutingserviceimpl.New,
	appserviceimpl.New,
	backuprepocleanupserviceimpl.New,
	backupreposerviceimpl.New,
	clustercleanupserviceimpl.New,
	clustersecretserviceimpl.New,
	clusterserviceimpl.New,
	commandpipeexecserviceimpl.New,
	commandserviceimpl.New,
	containerexecserviceimpl.New,
	containerfileserviceimpl.New,
	dbserviceimpl.New,
	domainserviceimpl.New,
	emailserviceimpl.New,
	envvarserviceimpl.New,
	fileserviceimpl.New,
	healthcheckserviceimpl.New,
	hpappserviceimpl.New,
	imagebuildserviceimpl.New,
	networkserviceimpl.New,
	nodeexecserviceimpl.New,
	notificationserviceimpl.New,
	placementserviceimpl.New,
	projectserviceimpl.New,
	repocheckoutserviceimpl.New,
	reslinkserviceimpl.New,
	schedjobexecserviceimpl.New,
	schedjobserviceimpl.New,
	scopeserviceimpl.New,
	settingeventserviceimpl.New,
	settinginitserviceimpl.New,
	settingserviceimpl.New,
	sslrenewalserviceimpl.New,
	sslserviceimpl.New,
	startupserviceimpl.New,
	sysbackupserviceimpl.New,
	syscleanupserviceimpl.New,
	systemeventbusserviceimpl.New,
	sysupdateserviceimpl.New,
	taskserviceimpl.New,
	traefikserviceimpl.New,
	userserviceimpl.New,
	volumeserviceimpl.New,

	// Repository
	repository.NewACLPermissionRepo,
	repository.NewAppRepo,
	repository.NewBinObjectRepo,
	repository.NewDataMigrationRepo,
	repository.NewDeploymentRepo,
	repository.NewFileRepo,
	repository.NewLockRepo,
	repository.NewLoginTrustedDeviceRepo,
	repository.NewProjectEnvRepo,
	repository.NewProjectRepo,
	repository.NewResLinkRepo,
	repository.NewSettingRepo,
	repository.NewSharedSettingRepo,
	repository.NewSysErrorRepo,
	repository.NewSystemStatusRepo,
	repository.NewTagRepo,
	repository.NewTaskLogRepo,
	repository.NewTaskRepo,
	repository.NewUserRepo,

	// Cache repository
	cacherepository.NewDeploymentInfoRepo,
	cacherepository.NewGithubAppManifestRepo,
	cacherepository.NewHealthcheckStateRepo,
	cacherepository.NewLoginAttemptRepo,
	cacherepository.NewMFAPasscodeRepo,
	cacherepository.NewPeriodicSettingsRepo,
	cacherepository.NewTaskControlRepo,
	cacherepository.NewTaskInfoRepo,
	cacherepository.NewUserTokenRepo,

	// Agent server
	agentserver.NewAgentServer,

	// Use case: Agent
	containeragentuc.New,
	imagebuildagentuc.New,
	nodeagentuc.New,
	nodecleanupagentuc.New,
}
