import { createContext } from "react";

import {
    AppsPublicApi,
    AppsPublicApiValidator,
    ProjectsPublicApi,
    ProjectsPublicApiValidator,
    UsersPublicApi,
    UsersPublicApiValidator,
} from "@application/shared/api-public/services";

function createApplicationPublicApi() {
    return {
        projects: new ProjectsPublicApi(new ProjectsPublicApiValidator()),
        users: new UsersPublicApi(new UsersPublicApiValidator()),
        apps: new AppsPublicApi(new AppsPublicApiValidator()),
    };
}

export const ApplicationPublicApiContext = createContext({
    api: createApplicationPublicApi(),
});
