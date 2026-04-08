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
    ProjectEnvVarsApi,
    ProjectEnvVarsApiValidator,
    ProjectGitCredentialsApi,
    ProjectGitCredentialsApiValidator,
    ProjectNetworksApi,
    ProjectNetworksApiValidator,
    ProjectRegistryAuthApi,
    ProjectRegistryAuthApiValidator,
    ProjectSecretsApi,
    ProjectSecretsApiValidator,
    ProjectsApi,
    ProjectsApiValidator,
} from "~/projects/api/services";
import {
    AppNetworkSettingsApi,
    AppNetworkSettingsApiValidator,
} from "~/projects/api/services/project-apps-services/network-settings";
import {
    AppResourceSettingsApi,
    AppResourceSettingsApiValidator,
} from "~/projects/api/services/project-apps-services/resource-settings";

function createApi() {
    /**
     * Projects
     */
    const projectsApiValidator = new ProjectsApiValidator();
    const projectAppsApiValidator = new ProjectAppsApiValidator();
    const projectAppEnvVarsApiValidator = new ProjectAppEnvVarsApiValidator();
    const projectSecretsApiValidator = new ProjectSecretsApiValidator();
    const projectEnvVarsApiValidator = new ProjectEnvVarsApiValidator();
    const projectNetworksApiValidator = new ProjectNetworksApiValidator();
    const appContainerSettingsApiValidator = new AppContainerSettingsApiValidator();
    const appDeploymentSettingsApiValidator = new AppDeploymentSettingsApiValidator();
    const appNetworkSettingsApiValidator = new AppNetworkSettingsApiValidator();
    const appResourceSettingsApiValidator = new AppResourceSettingsApiValidator();
    const appServiceSettingsApiValidator = new AppServiceSettingsApiValidator();
    const appSecretsApiValidator = new AppSecretsApiValidator();
    const projectGitCredentialsApiValidator = new ProjectGitCredentialsApiValidator();
    const projectRegistryAuthApiValidator = new ProjectRegistryAuthApiValidator();

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
                serviceSettings: {
                    $: new AppServiceSettingsApi(appServiceSettingsApiValidator),
                },
            },
            secrets: {
                $: new ProjectSecretsApi(projectSecretsApiValidator),
            },
            envVars: {
                $: new ProjectEnvVarsApi(projectEnvVarsApiValidator),
            },
            networks: {
                $: new ProjectNetworksApi(projectNetworksApiValidator),
            },
            gitCredentials: {
                $: new ProjectGitCredentialsApi(projectGitCredentialsApiValidator),
            },
            registryAuth: {
                $: new ProjectRegistryAuthApi(projectRegistryAuthApiValidator),
            },
        },
    };
}

export const ProjectsApiContext = createContext({
    api: createApi(),
});
