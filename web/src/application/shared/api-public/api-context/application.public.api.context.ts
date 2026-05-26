import { createContext } from "react";

import {
    ProjectsPublicApi,
    ProjectsPublicApiValidator,
    UsersPublicApi,
    UsersPublicApiValidator,
} from "@application/shared/api-public/services";

function createApplicationPublicApi() {
    return {
        projects: new ProjectsPublicApi(new ProjectsPublicApiValidator()),
        users: new UsersPublicApi(new UsersPublicApiValidator()),
    };
}

export const ApplicationPublicApiContext = createContext({
    api: createApplicationPublicApi(),
});
