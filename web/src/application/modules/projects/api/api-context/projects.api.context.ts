import { createContext } from "react";

import {
    AppConfigFilesApi,
    AppConfigFilesApiValidator,
    AppContainerSettingsApi,
    AppContainerSettingsApiValidator,
    AppDeploymentSettingsApi,
    AppDeploymentSettingsApiValidator,
    AppHealthChecksApi,
    AppHealthChecksApiValidator,
    AppScheduledJobsApi,
    AppScheduledJobsApiValidator,
    AppSecretsApi,
    AppSecretsApiValidator,
    AppServiceSettingsApi,
    AppServiceSettingsApiValidator,
    ProjectAccessTokenApi,
    ProjectAccessTokenApiValidator,
    ProjectAppEnvVarsApi,
    ProjectAppEnvVarsApiValidator,
    ProjectAppsApi,
    ProjectAppsApiValidator,
    ProjectBasicAuthApi,
    ProjectBasicAuthApiValidator,
    ProjectCloudStorageApi,
    ProjectCloudStorageApiValidator,
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
    ProjectNetworksApi,
    ProjectNetworksApiValidator,
    ProjectNotificationApi,
    ProjectNotificationApiValidator,
    ProjectRegistryAuthApi,
    ProjectRegistryAuthApiValidator,
    ProjectSSHKeyApi,
    ProjectSSHKeyApiValidator,
    ProjectSecretsApi,
    ProjectSecretsApiValidator,
    ProjectSettingsImportApi,
    ProjectSettingsImportApiValidator,
    ProjectSslCertApi,
    ProjectSslCertApiValidator,
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
    const projectAppEnvVarsApiValidator = new ProjectAppEnvVarsApiValidator();
    const projectBasicAuthApiValidator = new ProjectBasicAuthApiValidator();
    const projectSettingsImportApiValidator = new ProjectSettingsImportApiValidator();
    const projectGithubAppApiValidator = new ProjectGithubAppApiValidator();
    const projectSecretsApiValidator = new ProjectSecretsApiValidator();
    const projectEnvVarsApiValidator = new ProjectEnvVarsApiValidator();
    const projectDomainSettingsApiValidator = new ProjectDomainSettingsApiValidator();
    const projectNetworksApiValidator = new ProjectNetworksApiValidator();
    const projectDockerVolumesApiValidator = new ProjectDockerVolumesApiValidator();
    const projectNotificationApiValidator = new ProjectNotificationApiValidator();
    const appContainerSettingsApiValidator = new AppContainerSettingsApiValidator();
    const appConfigFilesApiValidator = new AppConfigFilesApiValidator();
    const appDeploymentSettingsApiValidator = new AppDeploymentSettingsApiValidator();
    const appHealthChecksApiValidator = new AppHealthChecksApiValidator();
    const appScheduledJobsApiValidator = new AppScheduledJobsApiValidator();
    const appNetworkSettingsApiValidator = new AppNetworkSettingsApiValidator();
    const appResourceSettingsApiValidator = new AppResourceSettingsApiValidator();
    const appStorageSettingsApiValidator = new AppStorageSettingsApiValidator();
    const appHttpSettingsApiValidator = new AppHttpSettingsApiValidator();
    const appServiceSettingsApiValidator = new AppServiceSettingsApiValidator();
    const appSecretsApiValidator = new AppSecretsApiValidator();
    const projectGitCredentialsApiValidator = new ProjectGitCredentialsApiValidator();
    const projectRegistryAuthApiValidator = new ProjectRegistryAuthApiValidator();
    const projectSslCertApiValidator = new ProjectSslCertApiValidator();
    const projectImServiceApiValidator = new ProjectImServiceApiValidator();
    const projectEmailApiValidator = new ProjectEmailApiValidator();
    const projectStorageSettingsApiValidator = new ProjectStorageSettingsApiValidator();
    const projectSSHKeyApiValidator = new ProjectSSHKeyApiValidator();
    const projectAccessTokenApiValidator = new ProjectAccessTokenApiValidator();
    const projectCloudStorageApiValidator = new ProjectCloudStorageApiValidator();
    const projectUserAccessesApiValidator = new ProjectUserAccessesApiValidator();

    return {
        projects: {
            $: new ProjectsApi(projectsApiValidator),
            userAccesses: {
                $: new ProjectUserAccessesApi(projectUserAccessesApiValidator),
            },
            apps: {
                $: new ProjectAppsApi(projectAppsApiValidator),
                envVars: {
                    $: new ProjectAppEnvVarsApi(projectAppEnvVarsApiValidator),
                },
                secrets: {
                    $: new AppSecretsApi(appSecretsApiValidator),
                },
                configFiles: {
                    $: new AppConfigFilesApi(appConfigFilesApiValidator),
                },
                containerSettings: {
                    $: new AppContainerSettingsApi(appContainerSettingsApiValidator),
                },
                deploymentSettings: {
                    $: new AppDeploymentSettingsApi(appDeploymentSettingsApiValidator),
                },
                healthChecks: {
                    $: new AppHealthChecksApi(appHealthChecksApiValidator),
                },
                scheduledJobs: {
                    $: new AppScheduledJobsApi(appScheduledJobsApiValidator),
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
            },
            secrets: {
                $: new ProjectSecretsApi(projectSecretsApiValidator),
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
            envVars: {
                $: new ProjectEnvVarsApi(projectEnvVarsApiValidator),
            },
            domainSettings: {
                $: new ProjectDomainSettingsApi(projectDomainSettingsApiValidator),
            },
            networks: {
                $: new ProjectNetworksApi(projectNetworksApiValidator),
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
            storageSettings: {
                $: new ProjectStorageSettingsApi(projectStorageSettingsApiValidator),
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
        },
    };
}

export const ProjectsApiContext = createContext({
    api: createApi(),
});
