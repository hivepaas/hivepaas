import { createContext } from "react";

import {
    AppContainerSettingsApi,
    AppContainerSettingsApiValidator,
    AppDeploymentSettingsApi,
    AppDeploymentSettingsApiValidator,
    AppSecretsApi,
    AppSecretsApiValidator,
    AppServiceSettingsApi,
    AppServiceSettingsApiValidator,
    ProjectAppEnvVarsApi,
    ProjectAppEnvVarsApiValidator,
    ProjectAppsApi,
    ProjectAppsApiValidator,
    ProjectBasicAuthApi,
    ProjectBasicAuthApiValidator,
    ProjectDockerVolumesApi,
    ProjectDockerVolumesApiValidator,
    ProjectDomainSettingsApi,
    ProjectDomainSettingsApiValidator,
    ProjectEnvVarsApi,
    ProjectEnvVarsApiValidator,
    ProjectGitCredentialsApi,
    ProjectGitCredentialsApiValidator,
    ProjectNetworksApi,
    ProjectNetworksApiValidator,
    ProjectNotificationApi,
    ProjectNotificationApiValidator,
    ProjectRegistryAuthApi,
    ProjectRegistryAuthApiValidator,
    ProjectSecretsApi,
    ProjectSecretsApiValidator,
    ProjectSslCertApi,
    ProjectSslCertApiValidator,
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
    const projectSecretsApiValidator = new ProjectSecretsApiValidator();
    const projectEnvVarsApiValidator = new ProjectEnvVarsApiValidator();
    const projectDomainSettingsApiValidator = new ProjectDomainSettingsApiValidator();
    const projectNetworksApiValidator = new ProjectNetworksApiValidator();
    const projectDockerVolumesApiValidator = new ProjectDockerVolumesApiValidator();
    const projectNotificationApiValidator = new ProjectNotificationApiValidator();
    const appContainerSettingsApiValidator = new AppContainerSettingsApiValidator();
    const appDeploymentSettingsApiValidator = new AppDeploymentSettingsApiValidator();
    const appNetworkSettingsApiValidator = new AppNetworkSettingsApiValidator();
    const appResourceSettingsApiValidator = new AppResourceSettingsApiValidator();
    const appStorageSettingsApiValidator = new AppStorageSettingsApiValidator();
    const appHttpSettingsApiValidator = new AppHttpSettingsApiValidator();
    const appServiceSettingsApiValidator = new AppServiceSettingsApiValidator();
    const appSecretsApiValidator = new AppSecretsApiValidator();
    const projectGitCredentialsApiValidator = new ProjectGitCredentialsApiValidator();
    const projectRegistryAuthApiValidator = new ProjectRegistryAuthApiValidator();
    const projectSslCertApiValidator = new ProjectSslCertApiValidator();
    const projectStorageSettingsApiValidator = new ProjectStorageSettingsApiValidator();

    return {
        projects: {
            $: new ProjectsApi(projectsApiValidator),
            apps: {
                $: new ProjectAppsApi(projectAppsApiValidator),
                envVars: {
                    $: new ProjectAppEnvVarsApi(projectAppEnvVarsApiValidator),
                },
                secrets: {
                    $: new AppSecretsApi(appSecretsApiValidator),
                },
                containerSettings: {
                    $: new AppContainerSettingsApi(appContainerSettingsApiValidator),
                },
                deploymentSettings: {
                    $: new AppDeploymentSettingsApi(appDeploymentSettingsApiValidator),
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
            sslCert: {
                $: new ProjectSslCertApi(projectSslCertApiValidator),
            },
            storageSettings: {
                $: new ProjectStorageSettingsApi(projectStorageSettingsApiValidator),
            },
        },
    };
}

export const ProjectsApiContext = createContext({
    api: createApi(),
});
