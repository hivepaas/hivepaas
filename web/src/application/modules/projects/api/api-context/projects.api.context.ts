import { createContext } from "react";

import { ProjectsApi, ProjectsApiValidator } from "~/projects/api/services";

function createApi() {
    /**
     * Projects
     */
    const projectsApiValidator = new ProjectsApiValidator();

    return {
        projects: {
            $: new ProjectsApi(projectsApiValidator),
        },
    };
}

export const ProjectsApiContext = createContext({
    api: createApi(),
});
