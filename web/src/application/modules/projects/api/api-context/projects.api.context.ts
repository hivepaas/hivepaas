import { createContext } from "react";

import {
    AppDeploymentSettingsApi,
    AppDeploymentSettingsApiValidator,
    AppSecretsApi,
    AppSecretsApiValidator,
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

function createApi() {
    /**
     * Projects
     */
    const projectsApiValidator = new ProjectsApiValidator();
    const projectAppsApiValidator = new ProjectAppsApiValidator();
    const projectAppEnvVarsApiValidator = new ProjectAppEnvVarsApiValidator();
    const projectSecretsApiValidator = new ProjectSecretsApiValidator();
    const projectEnvVarsApiValidator = new ProjectEnvVarsApiValidator();
    const appDeploymentSettingsApiValidator = new AppDeploymentSettingsApiValidator();
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
                deploymentSettings: {
                    $: new AppDeploymentSettingsApi(appDeploymentSettingsApiValidator),
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
