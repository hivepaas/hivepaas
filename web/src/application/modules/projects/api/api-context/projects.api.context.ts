import { createContext } from "react";

import {
    ProjectAppEnvVarsApi,
    ProjectAppEnvVarsApiValidator,
    ProjectAppsApi,
    ProjectAppsApiValidator,
    ProjectEnvVarsApi,
    ProjectEnvVarsApiValidator,
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

    return {
        projects: {
            $: new ProjectsApi(projectsApiValidator),
            apps: {
                $: new ProjectAppsApi(projectAppsApiValidator),
                envVars: {
                    $: new ProjectAppEnvVarsApi(projectAppEnvVarsApiValidator),
                },
            },
            secrets: {
                $: new ProjectSecretsApi(projectSecretsApiValidator),
            },
            envVars: {
                $: new ProjectEnvVarsApi(projectEnvVarsApiValidator),
            },
        },
    };
}

export const ProjectsApiContext = createContext({
    api: createApi(),
});
