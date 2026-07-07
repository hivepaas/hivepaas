import { createContext } from "react";

import {
    AppConfigFilesApi,
    AppConfigFilesApiValidator,
    AppContainerSettingsApi,
    AppContainerSettingsApiValidator,
    AppDeploymentLogsWsApi,
    AppDeploymentSettingsApi,
    AppDeploymentSettingsApiValidator,
    AppDeploymentsApi,
    AppDeploymentsApiValidator,
    AppFeatureSettingsApi,
    AppFeatureSettingsApiValidator,
    AppHealthChecksApi,
    AppHealthChecksApiValidator,
    AppLogsApi,
    AppLogsApiValidator,
    AppLogsWsApi,
    AppPreviewsApi,
    AppPreviewsApiValidator,
    AppScheduledJobTaskLogsWsApi,
    AppScheduledJobsApi,
    AppScheduledJobsApiValidator,
    AppSecretsApi,
    AppSecretsApiValidator,
    AppServiceSettingsApi,
    AppServiceSettingsApiValidator,
    AppServiceTasksApi,
    AppServiceTasksApiValidator,
    AppTerminalApi,
    AppTerminalApiValidator,
    AppTerminalWsApi,
    ProjectAccessTokenApi,
    ProjectAccessTokenApiValidator,
    ProjectAcmeDnsProviderApi,
    ProjectAcmeDnsProviderApiValidator,
    ProjectAppEnvVarsApi,
    ProjectAppEnvVarsApiValidator,
    ProjectAppsApi,
    ProjectAppsApiValidator,
    ProjectBasicAuthApi,
    ProjectBasicAuthApiValidator,
    ProjectCloudStorageApi,
    ProjectCloudStorageApiValidator,
    ProjectClusterVolumesApi,
    ProjectClusterVolumesApiValidator,
    ProjectCommandTemplateApi,
    ProjectCommandTemplateApiValidator,
    ProjectDockerVolumesApi,
    ProjectDockerVolumesApiValidator,
    ProjectDomainSettingsApi,
    ProjectDomainSettingsApiValidator,
    ProjectEmailApi,
    ProjectEmailApiValidator,
    ProjectEnvVarsApi,
    ProjectEnvVarsApiValidator,
    ProjectGitCredentialsApi,
    ProjectGitCredentialsApiValidator,
    ProjectGithubAppApi,
    ProjectGithubAppApiValidator,
    ProjectImServiceApi,
    ProjectImServiceApiValidator,
    ProjectImageBuildSettingsApi,
    ProjectImageBuildSettingsApiValidator,
    ProjectNetworksApi,
    ProjectNetworksApiValidator,
    ProjectNotificationApi,
    ProjectNotificationApiValidator,
    ProjectRegistryAuthApi,
    ProjectRegistryAuthApiValidator,
    ProjectRepoWebhookApi,
    ProjectRepoWebhookApiValidator,
    ProjectSSHKeyApi,
    ProjectSSHKeyApiValidator,
    ProjectSecretsApi,
    ProjectSecretsApiValidator,
    ProjectSettingsImportApi,
    ProjectSettingsImportApiValidator,
    ProjectSslCertApi,
    ProjectSslCertApiValidator,
    ProjectSslProviderApi,
    ProjectSslProviderApiValidator,
    ProjectUserAccessesApi,
    ProjectUserAccessesApiValidator,
    ProjectsApi,
    ProjectsApiValidator,
} from "~/projects/api/services";
import {
    AppHttpSettingsApi,
    AppHttpSettingsApiValidator,
} from "~/projects/api/services/project-apps-services/http-settings";
import {
    AppNetworkSettingsApi,
    AppNetworkSettingsApiValidator,
} from "~/projects/api/services/project-apps-services/network-settings";
import {
    AppResourceSettingsApi,
    AppResourceSettingsApiValidator,
} from "~/projects/api/services/project-apps-services/resource-settings";
import {
    AppStorageSettingsApi,
    AppStorageSettingsApiValidator,
} from "~/projects/api/services/project-apps-services/storage-settings";
import {
    ProjectStorageSettingsApi,
    ProjectStorageSettingsApiValidator,
} from "~/projects/api/services/project-settings-services/storage-settings";

function createApi() {
    /**
     * Projects
     */
    const projectsApiValidator = new ProjectsApiValidator();
    const projectAppsApiValidator = new ProjectAppsApiValidator();
    const appLogsApiValidator = new AppLogsApiValidator();
    const appLogsApi = new AppLogsApi(appLogsApiValidator);
    const appTerminalApiValidator = new AppTerminalApiValidator();
    const appTerminalApi = new AppTerminalApi(appTerminalApiValidator);
    const projectAppEnvVarsApiValidator = new ProjectAppEnvVarsApiValidator();
    const projectAcmeDnsProviderApiValidator = new ProjectAcmeDnsProviderApiValidator();
    const projectBasicAuthApiValidator = new ProjectBasicAuthApiValidator();
    const projectSettingsImportApiValidator = new ProjectSettingsImportApiValidator();
    const projectGithubAppApiValidator = new ProjectGithubAppApiValidator();
    const projectRepoWebhookApiValidator = new ProjectRepoWebhookApiValidator();
    const projectSecretsApiValidator = new ProjectSecretsApiValidator();
    const projectEnvVarsApiValidator = new ProjectEnvVarsApiValidator();
    const projectDomainSettingsApiValidator = new ProjectDomainSettingsApiValidator();
    const projectNetworksApiValidator = new ProjectNetworksApiValidator();
    const projectClusterVolumesApiValidator = new ProjectClusterVolumesApiValidator();
    const projectDockerVolumesApiValidator = new ProjectDockerVolumesApiValidator();
    const projectNotificationApiValidator = new ProjectNotificationApiValidator();
    const appContainerSettingsApiValidator = new AppContainerSettingsApiValidator();
    const appConfigFilesApiValidator = new AppConfigFilesApiValidator();
    const appDeploymentsApiValidator = new AppDeploymentsApiValidator();
    const appDeploymentsApi = new AppDeploymentsApi(appDeploymentsApiValidator);
    const appPreviewsApiValidator = new AppPreviewsApiValidator();
    const appDeploymentSettingsApiValidator = new AppDeploymentSettingsApiValidator();
    const appFeatureSettingsApiValidator = new AppFeatureSettingsApiValidator();
    const appHealthChecksApiValidator = new AppHealthChecksApiValidator();
    const appScheduledJobsApiValidator = new AppScheduledJobsApiValidator();
    const appScheduledJobsApi = new AppScheduledJobsApi(appScheduledJobsApiValidator);
    const appNetworkSettingsApiValidator = new AppNetworkSettingsApiValidator();
    const appResourceSettingsApiValidator = new AppResourceSettingsApiValidator();
    const appStorageSettingsApiValidator = new AppStorageSettingsApiValidator();
    const appHttpSettingsApiValidator = new AppHttpSettingsApiValidator();
    const appServiceSettingsApiValidator = new AppServiceSettingsApiValidator();
    const appServiceTasksApiValidator = new AppServiceTasksApiValidator();
    const appSecretsApiValidator = new AppSecretsApiValidator();
    const projectGitCredentialsApiValidator = new ProjectGitCredentialsApiValidator();
    const projectRegistryAuthApiValidator = new ProjectRegistryAuthApiValidator();
    const projectSslCertApiValidator = new ProjectSslCertApiValidator();
    const projectSslProviderApiValidator = new ProjectSslProviderApiValidator();
    const projectImServiceApiValidator = new ProjectImServiceApiValidator();
    const projectEmailApiValidator = new ProjectEmailApiValidator();
    const projectStorageSettingsApiValidator = new ProjectStorageSettingsApiValidator();
    const projectSSHKeyApiValidator = new ProjectSSHKeyApiValidator();
    const projectAccessTokenApiValidator = new ProjectAccessTokenApiValidator();
    const projectCloudStorageApiValidator = new ProjectCloudStorageApiValidator();
    const projectCommandTemplateApiValidator = new ProjectCommandTemplateApiValidator();
    const projectUserAccessesApiValidator = new ProjectUserAccessesApiValidator();
    const projectImageBuildSettingsApiValidator = new ProjectImageBuildSettingsApiValidator();

    return {
        projects: {
            $: new ProjectsApi(projectsApiValidator),
            userAccesses: {
                $: new ProjectUserAccessesApi(projectUserAccessesApiValidator),
            },
            apps: {
                $: new ProjectAppsApi(projectAppsApiValidator),
                logs: {
                    $: appLogsApi,
                    stream: {
                        $: new AppLogsWsApi(),
                    },
                },
                terminal: {
                    $: appTerminalApi,
                    stream: {
                        $: new AppTerminalWsApi(),
                    },
                },
                envVars: {
                    $: new ProjectAppEnvVarsApi(projectAppEnvVarsApiValidator),
                },
                secrets: {
                    $: new AppSecretsApi(appSecretsApiValidator),
                },
                configFiles: {
                    $: new AppConfigFilesApi(appConfigFilesApiValidator),
                },
                deployments: {
                    $: appDeploymentsApi,
                    logs: {
                        $: new AppDeploymentLogsWsApi(),
                    },
                },
                previews: {
                    $: new AppPreviewsApi(appPreviewsApiValidator),
                },
                containerSettings: {
                    $: new AppContainerSettingsApi(appContainerSettingsApiValidator),
                },
                deploymentSettings: {
                    $: new AppDeploymentSettingsApi(appDeploymentSettingsApiValidator),
                },
                featureSettings: {
                    $: new AppFeatureSettingsApi(appFeatureSettingsApiValidator),
                },
                healthChecks: {
                    $: new AppHealthChecksApi(appHealthChecksApiValidator),
                },
                scheduledJobs: {
                    $: appScheduledJobsApi,
                    taskLogs: {
                        $: new AppScheduledJobTaskLogsWsApi(),
                    },
                },
                networkSettings: {
                    $: new AppNetworkSettingsApi(appNetworkSettingsApiValidator),
                },
                resourceSettings: {
                    $: new AppResourceSettingsApi(appResourceSettingsApiValidator),
                },
                storageSettings: {
                    $: new AppStorageSettingsApi(appStorageSettingsApiValidator),
                },
                httpSettings: {
                    $: new AppHttpSettingsApi(appHttpSettingsApiValidator),
                },
                serviceSettings: {
                    $: new AppServiceSettingsApi(appServiceSettingsApiValidator),
                },
                serviceTasks: {
                    $: new AppServiceTasksApi(appServiceTasksApiValidator),
                },
            },
            secrets: {
                $: new ProjectSecretsApi(projectSecretsApiValidator),
            },
            acmeDnsProvider: {
                $: new ProjectAcmeDnsProviderApi(projectAcmeDnsProviderApiValidator),
            },
            basicAuth: {
                $: new ProjectBasicAuthApi(projectBasicAuthApiValidator),
            },
            settingsImport: {
                $: new ProjectSettingsImportApi(projectSettingsImportApiValidator),
            },
            githubApp: {
                $: new ProjectGithubAppApi(projectGithubAppApiValidator),
            },
            repoWebhook: {
                $: new ProjectRepoWebhookApi(projectRepoWebhookApiValidator),
            },
            envVars: {
                $: new ProjectEnvVarsApi(projectEnvVarsApiValidator),
            },
            domainSettings: {
                $: new ProjectDomainSettingsApi(projectDomainSettingsApiValidator),
            },
            networks: {
                $: new ProjectNetworksApi(projectNetworksApiValidator),
            },
            clusterVolumes: {
                $: new ProjectClusterVolumesApi(projectClusterVolumesApiValidator),
            },
            dockerVolumes: {
                $: new ProjectDockerVolumesApi(projectDockerVolumesApiValidator),
            },
            notifications: {
                $: new ProjectNotificationApi(projectNotificationApiValidator),
            },
            gitCredentials: {
                $: new ProjectGitCredentialsApi(projectGitCredentialsApiValidator),
            },
            registryAuth: {
                $: new ProjectRegistryAuthApi(projectRegistryAuthApiValidator),
            },
            imService: {
                $: new ProjectImServiceApi(projectImServiceApiValidator),
            },
            email: {
                $: new ProjectEmailApi(projectEmailApiValidator),
            },
            sslCert: {
                $: new ProjectSslCertApi(projectSslCertApiValidator),
            },
            sslProvider: {
                $: new ProjectSslProviderApi(projectSslProviderApiValidator),
            },
            storageSettings: {
                $: new ProjectStorageSettingsApi(projectStorageSettingsApiValidator),
            },
            imageBuildSettings: {
                $: new ProjectImageBuildSettingsApi(projectImageBuildSettingsApiValidator),
            },
            sshKey: {
                $: new ProjectSSHKeyApi(projectSSHKeyApiValidator),
            },
            accessToken: {
                $: new ProjectAccessTokenApi(projectAccessTokenApiValidator),
            },
            cloudStorage: {
                $: new ProjectCloudStorageApi(projectCloudStorageApiValidator),
            },
            commandTemplates: {
                $: new ProjectCommandTemplateApi(projectCommandTemplateApiValidator),
            },
        },
    };
}

export const ProjectsApiContext = createContext({
    api: createApi(),
});
