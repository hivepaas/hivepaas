import { createContext } from "react";

import { ProjectsPublicApi, ProjectsPublicApiValidator } from "@application/shared/api-public/services";

function createApplicationPublicApi() {
    return {
        projects: new ProjectsPublicApi(new ProjectsPublicApiValidator()),
    };
}

export const ApplicationPublicApiContext = createContext({
    api: createApplicationPublicApi(),
});
