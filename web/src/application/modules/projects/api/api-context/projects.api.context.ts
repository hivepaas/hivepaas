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

function createApi() {
    /**
     * Projects
     */
    const projectsApiValidator = new ProjectsApiValidator();
    const projectAppsApiValidator = new ProjectAppsApiValidator();
    const projectAppEnvVarsApiValidator = new ProjectAppEnvVarsApiValidator();
    const projectSecretsApiValidator = new ProjectSecretsApiValidator();
    const projectEnvVarsApiValidator = new ProjectEnvVarsApiValidator();
    const appContainerSettingsApiValidator = new AppContainerSettingsApiValidator();
    const appDeploymentSettingsApiValidator = new AppDeploymentSettingsApiValidator();
    const appNetworkSettingsApiValidator = new AppNetworkSettingsApiValidator();
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
